// Package crawlsnap is the official Go client for the CrawlSnap data
// intelligence platform.
//
//	client, err := crawlsnap.NewClient("sk-cs-...")   // or set CRAWLSNAP_API_KEY
//	if err != nil { log.Fatal(err) }
//
//	ip, err := client.VectorSnap.IP(ctx, "8.8.8.8")
//	if err != nil { log.Fatal(err) }
//	fmt.Println(ip.GetReputation(), ip.GetAsOwner())
//
// Every call returns the unwrapped, typed payload on success and a typed error
// otherwise (see errors.go) — the caller never inspects the response envelope.
package crawlsnap

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ErrNoAPIKey is returned by NewClient when no API key is supplied and
// CRAWLSNAP_API_KEY is unset.
var ErrNoAPIKey = &ConfigError{Message: "no API key provided: pass it to NewClient or set CRAWLSNAP_API_KEY"}

// Client is the CrawlSnap API client. It is safe for concurrent use.
type Client struct {
	apiKey     string
	baseURL    string
	maxRetries int
	httpClient *http.Client

	// Resource groups.
	VectorSnap *VectorSnap
	PulseSnap  *PulseSnap
	SubdoSnap  *SubdoSnap
	SportSnap  *SportSnap
}

type clientConfig struct {
	baseURL    string
	timeout    time.Duration
	maxRetries int
	httpClient *http.Client
	transport  http.RoundTripper
}

// Option configures a Client in NewClient.
type Option func(*clientConfig)

// WithBaseURL overrides the API host (default https://api.crawlsnap.com, or
// $CRAWLSNAP_BASE_URL).
func WithBaseURL(baseURL string) Option {
	return func(c *clientConfig) { c.baseURL = baseURL }
}

// WithTimeout sets the per-request timeout (default 30s). Ignored when
// WithHTTPClient is supplied.
func WithTimeout(d time.Duration) Option {
	return func(c *clientConfig) { c.timeout = d }
}

// WithMaxRetries sets retries for 429 / 5xx / connection errors (default 2).
func WithMaxRetries(n int) Option {
	return func(c *clientConfig) { c.maxRetries = n }
}

// WithHTTPClient supplies your own *http.Client (advanced). Takes precedence
// over WithTimeout and WithTransport.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *clientConfig) { c.httpClient = hc }
}

// WithTransport supplies a custom http.RoundTripper, e.g. a mock for testing.
func WithTransport(rt http.RoundTripper) Option {
	return func(c *clientConfig) { c.transport = rt }
}

// NewClient creates a client. The API key falls back to $CRAWLSNAP_API_KEY when
// empty; if neither is set, ErrNoAPIKey is returned.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	cfg := clientConfig{
		baseURL:    "",
		timeout:    DefaultTimeout,
		maxRetries: DefaultMaxRetries,
	}
	for _, o := range opts {
		o(&cfg)
	}

	if apiKey == "" {
		apiKey = os.Getenv("CRAWLSNAP_API_KEY")
	}
	if apiKey == "" {
		return nil, ErrNoAPIKey
	}

	baseURL := cfg.baseURL
	if baseURL == "" {
		baseURL = os.Getenv("CRAWLSNAP_BASE_URL")
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	hc := cfg.httpClient
	if hc == nil {
		hc = &http.Client{Timeout: cfg.timeout, Transport: cfg.transport}
	}

	c := &Client{
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		maxRetries: max(0, cfg.maxRetries),
		httpClient: hc,
	}
	c.VectorSnap = newVectorSnap(c)
	c.PulseSnap = newPulseSnap(c)
	c.SubdoSnap = newSubdoSnap(c)
	c.SportSnap = newSportSnap(c)
	return c, nil
}

// requestConfig holds per-call options.
type requestConfig struct {
	raw *RawResponse
}

// RequestOption customizes a single request.
type RequestOption func(*requestConfig)

// WithRawResponse populates dst with the full response envelope (status,
// headers, request id, is_success, message) alongside the normal typed return.
func WithRawResponse(dst *RawResponse) RequestOption {
	return func(rc *requestConfig) { rc.raw = dst }
}

// do issues a single GET request with auth + standard headers applied.
func (c *Client) do(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "crawlsnap-go/"+Version)
	return c.httpClient.Do(req)
}

// doGet runs the request pipeline (retries + envelope unwrap) and returns the
// typed payload T or a typed error.
func doGet[T any](ctx context.Context, c *Client, path string, params url.Values, opts ...RequestOption) (T, error) {
	var zero T
	var rc requestConfig
	for _, o := range opts {
		o(&rc)
	}

	attempt := 0
	for {
		resp, err := c.do(ctx, path, params)
		if err != nil {
			if attempt < c.maxRetries {
				if serr := sleep(ctx, backoffDelay(attempt, 0, false)); serr != nil {
					return zero, connError(serr)
				}
				attempt++
				continue
			}
			return zero, connError(err)
		}

		if isRetryableStatus(resp.StatusCode) && attempt < c.maxRetries {
			ra, hasRA := parseRetryAfter(resp.Header.Get("Retry-After"))
			resp.Body.Close()
			if serr := sleep(ctx, backoffDelay(attempt, secondsToDuration(ra), hasRA)); serr != nil {
				return zero, connError(serr)
			}
			attempt++
			continue
		}

		body, status, header, rerr := readResponse(resp)
		if rerr != nil {
			return zero, connError(rerr)
		}
		return processResponse[T](body, status, header, &rc)
	}
}

// readResponse consumes and closes the response body.
func readResponse(resp *http.Response) ([]byte, int, http.Header, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, nil, err
	}
	return body, resp.StatusCode, resp.Header, nil
}

// secondsToDuration converts a fractional-second Retry-After value to a Duration.
func secondsToDuration(secs float64) time.Duration {
	return time.Duration(secs * float64(time.Second))
}

// sleep waits for d, returning early with the context error if the context is
// cancelled first.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
