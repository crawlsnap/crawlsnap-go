package crawlsnap

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// roundTripFunc adapts a function to an http.RoundTripper for mocked transport.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResp(status int, body string, headers map[string]string) *http.Response {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func ok(data string) *http.Response {
	return jsonResp(200,
		`{"data":`+data+`,"is_success":true,"message":"Success","response_code":200}`,
		map[string]string{"x-request-id": "req_ok"})
}

func errResp(status int, message string) *http.Response {
	return jsonResp(status,
		`{"data":null,"is_success":false,"message":"`+message+`","response_code":`+itoa(status)+`}`,
		map[string]string{"x-request-id": "req_err"})
}

func itoa(n int) string {
	switch n {
	case 402:
		return "402"
	case 404:
		return "404"
	default:
		return "500"
	}
}

// mockHandler routes mocked responses and tracks the URL-endpoint call count so
// the retry test can observe a 429-then-success sequence.
type mockHandler struct{ urlCalls int }

func (m *mockHandler) handle(r *http.Request) (*http.Response, error) {
	path := r.URL.Path
	query := r.URL.Query().Get("query")
	switch path {
	case "/v1/ioc/search/ip":
		return ok(`{"hash_id":"h","search_type":"ip","ip":"` + query + `","reputation":3,"as_owner":"GOOGLE","tags":["dns"]}`), nil
	case "/v1/ioc/search/domain":
		return errResp(404, "No IoC data found"), nil
	case "/v1/ioc/search/hash":
		return errResp(402, "Monthly quota exceeded"), nil
	case "/v1/ioc/search/url":
		m.urlCalls++
		if m.urlCalls == 1 {
			return jsonResp(429,
				`{"data":null,"is_success":false,"message":"Rate limited","response_code":429}`,
				map[string]string{"Retry-After": "0", "x-request-id": "req_429"}), nil
		}
		return ok(`{"hash_id":"h","search_type":"url","url":"` + query + `"}`), nil
	case "/v1/subdo-snap/scan":
		if r.URL.Query().Get("cursor") == "" {
			return ok(`{"hash_id":"h","search_type":"domain","subdomains":[{"subdomain":"a.example.com"}],"cursor":"c1","count":2}`), nil
		}
		return ok(`{"hash_id":"h","search_type":"domain","subdomains":[{"subdomain":"b.example.com"}],"cursor":"","count":2}`), nil
	case "/v1/sport-snap/channels/bein-connect-turkey":
		return ok(`{"slug":"bein-connect-turkey","name":"beIN CONNECT Turkey","country":"Turkey","broadcast_rights":[{"competition":"England - Premier League","year_start":2024,"year_end":2027}],"updated_at":"2026-07-05T10:00:00Z"}`), nil
	case "/v1/sport-snap/channels/bein-connect-turkey/schedule":
		return ok(`{"slug":"bein-connect-turkey","name":"beIN CONNECT Turkey","entries":[{"date":"2026-07-06","kickoff_utc":"2026-07-06T19:00:00Z","match_id":5542814,"match_title":"Brazil vs Norway","competition":"Friendly"}],"updated_at":"2026-07-05T10:00:00Z"}`), nil
	case "/v1/sport-snap/matches/5542814":
		return ok(`{"id":5542814,"status":"finished","competition":{"name":"Friendly"},"home_team":{"name":"Brazil"},"away_team":{"name":"Norway"},"score":{"home":2,"away":1},"events":[],"stats":[],"broadcasts":[{"country":"Turkey","country_slug":"turkey","channels":[{"name":"beIN CONNECT Turkey","slug":"bein-connect-turkey"}]}],"updated_at":"2026-07-05T10:00:00Z"}`), nil
	case "/v1/sport-snap/matches/404":
		return errResp(404, "Unknown match id"), nil
	case "/v1/sport-snap/countries/turkey/channels":
		return ok(`{"country":"Turkey","country_slug":"turkey","channels":[{"name":"beIN CONNECT Turkey","slug":"bein-connect-turkey","last_seen":"2026-07-05T10:00:00Z"}],"updated_at":"2026-07-05T10:00:00Z"}`), nil
	case "/v1/sport-snap/schedules/2026-07-05":
		return ok(`{"date":"2026-07-05","competitions":[{"competition":"Friendly","matches":[{"id":5542814,"title":"Brazil vs Norway","status":"scheduled","kickoff_utc":"2026-07-06T19:00:00Z","channels":[{"name":"beIN CONNECT Turkey","slug":"bein-connect-turkey"}]}]}],"updated_at":"2026-07-05T10:00:00Z"}`), nil
	}
	return jsonResp(500, `{"data":null,"is_success":false,"message":"unexpected","response_code":500}`, nil), nil
}

func newTestClient(t *testing.T, maxRetries int) (*Client, *mockHandler) {
	t.Helper()
	h := &mockHandler{}
	c, err := NewClient("sk-cs-test",
		WithMaxRetries(maxRetries),
		WithTransport(roundTripFunc(h.handle)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, h
}

func TestSuccessReturnsTypedData(t *testing.T) {
	c, _ := newTestClient(t, 2)
	ip, err := c.VectorSnap.IP(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip.GetAsOwner() != "GOOGLE" {
		t.Errorf("as_owner = %q, want GOOGLE", ip.GetAsOwner())
	}
	if ip.GetIp() != "8.8.8.8" {
		t.Errorf("ip = %q, want 8.8.8.8", ip.GetIp())
	}
	if got := ip.GetTags(); len(got) != 1 || got[0] != "dns" {
		t.Errorf("tags = %v, want [dns]", got)
	}
}

func TestPinnedVersionHitsSameEndpointAndIsCached(t *testing.T) {
	c, _ := newTestClient(t, 2)
	ip, err := c.VectorSnap.V1().IP(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip.GetAsOwner() != "GOOGLE" {
		t.Errorf("as_owner = %q, want GOOGLE", ip.GetAsOwner())
	}
	if c.VectorSnap.V1() != c.VectorSnap.V1() {
		t.Error("V1() should return the same cached instance")
	}
}

func TestNotFoundRaises(t *testing.T) {
	c, _ := newTestClient(t, 2)
	_, err := c.VectorSnap.Domain(context.Background(), "nope.example")
	var nf *NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("want *NotFoundError, got %T (%v)", err, err)
	}
	if nf.StatusCode != 404 {
		t.Errorf("status = %d, want 404", nf.StatusCode)
	}
}

func TestQuotaExceededRaises(t *testing.T) {
	c, _ := newTestClient(t, 2)
	_, err := c.VectorSnap.Hash(context.Background(), "deadbeef")
	var qe *QuotaExceededError
	if !errors.As(err, &qe) {
		t.Fatalf("want *QuotaExceededError, got %T (%v)", err, err)
	}
	if qe.StatusCode != 402 {
		t.Errorf("status = %d, want 402", qe.StatusCode)
	}
	if qe.RequestID != "req_err" {
		t.Errorf("request_id = %q, want req_err", qe.RequestID)
	}
	// Catchable as the base APIStatusError too.
	var base *APIStatusError
	if !errors.As(err, &base) {
		t.Error("QuotaExceededError should be matchable as *APIStatusError")
	}
	// And as the top-level marker interface.
	var any CrawlSnapError
	if !errors.As(err, &any) {
		t.Error("QuotaExceededError should be matchable as CrawlSnapError")
	}
}

func TestRetryThenSuccess(t *testing.T) {
	c, h := newTestClient(t, 2)
	res, err := c.VectorSnap.URL(context.Background(), "https://x.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GetUrl() != "https://x.com" {
		t.Errorf("url = %q, want https://x.com", res.GetUrl())
	}
	if h.urlCalls != 2 {
		t.Errorf("urlCalls = %d, want 2 (one 429, then a retry)", h.urlCalls)
	}
}

func TestRateLimitExhaustedRaises(t *testing.T) {
	c, _ := newTestClient(t, 0) // no retries -> the 429 surfaces
	_, err := c.VectorSnap.URL(context.Background(), "https://x.com")
	var rl *RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("want *RateLimitError, got %T (%v)", err, err)
	}
	if rl.RetryAfter != 0 {
		t.Errorf("retry_after = %v, want 0", rl.RetryAfter)
	}
}

func TestPaginationIteratesAllPages(t *testing.T) {
	c, _ := newTestClient(t, 2)
	var got []string
	for sub, err := range c.SubdoSnap.ScanIter(context.Background(), "example.com") {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got = append(got, sub["subdomain"].(string))
	}
	want := []string{"a.example.com", "b.example.com"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("subdomains = %v, want %v", got, want)
	}
}

func TestRawResponse(t *testing.T) {
	c, _ := newTestClient(t, 2)
	var raw RawResponse
	ip, err := c.VectorSnap.IP(context.Background(), "8.8.8.8", WithRawResponse(&raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw.StatusCode != 200 {
		t.Errorf("status = %d, want 200", raw.StatusCode)
	}
	if raw.RequestID != "req_ok" {
		t.Errorf("request_id = %q, want req_ok", raw.RequestID)
	}
	if !raw.IsSuccess {
		t.Error("raw.IsSuccess should be true")
	}
	if ip.GetAsOwner() != "GOOGLE" {
		t.Errorf("typed return still expected, got %q", ip.GetAsOwner())
	}
}

func TestMissingAPIKeyRaises(t *testing.T) {
	t.Setenv("CRAWLSNAP_API_KEY", "")
	_, err := NewClient("")
	if !errors.Is(err, ErrNoAPIKey) {
		t.Fatalf("want ErrNoAPIKey, got %v", err)
	}
	// It is a config error, NOT a transport (connection) error, so it must not
	// be caught by callers handling connection failures.
	var cfg *ConfigError
	if !errors.As(err, &cfg) {
		t.Errorf("want *ConfigError, got %T", err)
	}
	var conn *APIConnectionError
	if errors.As(err, &conn) {
		t.Error("missing-key error must not match *APIConnectionError")
	}
}

func TestHashSigmaAcceptsArrayShape(t *testing.T) {
	// Regression: opaque upstream pass-through fields (sigma, votes_result, ...)
	// are free-form, so a response where the API returns sigma as an ARRAY (the
	// real-world shape that previously caused a DecodeError) decodes cleanly.
	c, err := NewClient("sk-cs-test", WithMaxRetries(0), WithTransport(
		roundTripFunc(func(*http.Request) (*http.Response, error) {
			return ok(`{"hash_id":"h","search_type":"hash","sigma":[{"rule_title":"x"}],"sigma_stats":{"high":1}}`), nil
		}),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	res, err := c.VectorSnap.Hash(context.Background(), "deadbeef")
	if err != nil {
		t.Fatalf("sigma-as-array must decode, got error: %v", err)
	}
	arr, ok := res.GetSigma().([]interface{})
	if !ok || len(arr) != 1 {
		t.Errorf("sigma = %#v, want a 1-element array", res.GetSigma())
	}
}

func TestDecodeErrorOnMalformedData(t *testing.T) {
	// A 2xx, is_success:true response whose data field is the wrong shape for the
	// typed payload surfaces a DecodeError, not a connection error.
	c, err := NewClient("sk-cs-test", WithMaxRetries(0), WithTransport(
		roundTripFunc(func(*http.Request) (*http.Response, error) {
			return jsonResp(200, `{"data":"not-an-object","is_success":true,"message":"ok","response_code":200}`, nil), nil
		}),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.VectorSnap.IP(context.Background(), "8.8.8.8")
	var de *DecodeError
	if !errors.As(err, &de) {
		t.Fatalf("want *DecodeError, got %T (%v)", err, err)
	}
	var conn *APIConnectionError
	if errors.As(err, &conn) {
		t.Error("decode error must not match *APIConnectionError")
	}
}

func TestBackoffHonorsRetryAfterButCaps(t *testing.T) {
	// Retry-After is honored but clamped to retryMaxDelay so an absurd value
	// cannot hang the caller.
	if got := backoffDelay(0, 2*time.Second, true); got != 2*time.Second {
		t.Errorf("retry-after 2s = %v, want 2s", got)
	}
	if got := backoffDelay(0, 100*time.Hour, true); got != retryMaxDelay {
		t.Errorf("absurd retry-after = %v, want clamp to %v", got, retryMaxDelay)
	}
	if got := backoffDelay(0, -5*time.Second, true); got != 0 {
		t.Errorf("negative retry-after = %v, want 0", got)
	}
	// Exponential path stays within the cap (+ jitter under retryBaseDelay/2)
	// even for a large attempt count — no int64 overflow.
	if got := backoffDelay(1000, 0, false); got < retryMaxDelay || got > retryMaxDelay+retryBaseDelay {
		t.Errorf("exp backoff(1000) = %v, want ~%v", got, retryMaxDelay)
	}
}

func TestDefaultVersionIsStableNotLatest(t *testing.T) {
	c, _ := newTestClient(t, 2)
	// A direct (unpinned) call targets the product's pinned default version,
	// which must equal the explicit .V1() endpoint.
	if c.VectorSnap.version != "v1" {
		t.Errorf("default version = %q, want v1", c.VectorSnap.version)
	}
	if c.VectorSnap.V1().version != "v1" {
		t.Errorf("V1 version = %q, want v1", c.VectorSnap.V1().version)
	}
	// Pinning is per-product: it leaves the others untouched.
	if c.PulseSnap.version != "v1" {
		t.Errorf("pulse version = %q, want v1", c.PulseSnap.version)
	}
}

func TestConnectionErrorIsTyped(t *testing.T) {
	wire := errors.New("dial tcp: connection refused")
	c, err := NewClient("sk-cs-test", WithMaxRetries(0), WithTransport(
		roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, wire }),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.VectorSnap.IP(context.Background(), "8.8.8.8")
	var ce *APIConnectionError
	if !errors.As(err, &ce) {
		t.Fatalf("want *APIConnectionError, got %T (%v)", err, err)
	}
}
