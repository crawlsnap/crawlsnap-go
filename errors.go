// Package crawlsnap — error types.
//
// Hierarchy (mirrors the other CrawlSnap SDKs):
//
//	CrawlSnapError                      marker interface for everything the SDK returns
//	├── ConfigError                     misconfiguration (e.g. missing API key)
//	├── DecodeError                     a 2xx body could not be decoded
//	├── APIConnectionError              transport failed (no HTTP response)
//	│   └── APITimeoutError             the request timed out
//	└── APIStatusError                  the API returned a non-success status
//	    ├── BadRequestError             400  invalid input
//	    ├── AuthenticationError         401  missing / invalid API key
//	    ├── QuotaExceededError          402  out of credits / monthly quota
//	    ├── SubscriptionInactiveError   403  subscription not active
//	    ├── NotFoundError               404  no data for the indicator
//	    ├── RateLimitError              429  request limit exceeded
//	    └── ServerError                 5xx  server / upstream failure
//
// Discover the concrete type with errors.As:
//
//	if _, err := client.VectorSnap.IP(ctx, "8.8.8.8"); err != nil {
//	    var nf *crawlsnap.NotFoundError
//	    if errors.As(err, &nf) { /* no data for this indicator */ }
//
//	    var status *crawlsnap.APIStatusError   // catches any HTTP status error
//	    if errors.As(err, &status) { log.Println(status.StatusCode, status.RequestID) }
//
//	    var any crawlsnap.CrawlSnapError       // catches anything the SDK returns
//	    if errors.As(err, &any) { /* ... */ }
//	}
package crawlsnap

import "fmt"

// CrawlSnapError is implemented by every error the SDK returns. Use it with
// errors.As to catch anything originating from this package.
type CrawlSnapError interface {
	error
	crawlsnapError()
}

// ConfigError is returned for SDK misconfiguration detected before any request
// is made (e.g. a missing API key). It is a config-time error, distinct from the
// transport-level APIConnectionError.
type ConfigError struct{ Message string }

func (e *ConfigError) Error() string   { return e.Message }
func (e *ConfigError) crawlsnapError() {}

// DecodeError is returned when a successful (2xx) response body could not be
// decoded into the expected typed payload — an application-layer failure, not a
// transport one.
type DecodeError struct {
	Message string
	Err     error
}

func (e *DecodeError) Error() string   { return e.Message }
func (e *DecodeError) Unwrap() error   { return e.Err }
func (e *DecodeError) crawlsnapError() {}

// APIStatusError is the base error for any non-success HTTP status. The seven
// status-specific types embed it; errors.As(err, &(*APIStatusError)(nil)) on any
// of them yields this base value.
type APIStatusError struct {
	// StatusCode is the effective HTTP status (the body response_code when it
	// disagrees with and supersedes the transport status).
	StatusCode int
	// Message is the human-readable error from the response envelope.
	Message string
	// RequestID is the x-request-id response header, useful for support.
	RequestID string
	// Body is the raw response body (the BaseResponse envelope), when available.
	Body []byte
}

func (e *APIStatusError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("[%d] %s (request_id: %s)", e.StatusCode, e.Message, e.RequestID)
	}
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
}

func (e *APIStatusError) crawlsnapError() {}

// As lets errors.As extract the embedded *APIStatusError from any of the
// status-specific subtypes (they embed APIStatusError and promote this method).
func (e *APIStatusError) As(target any) bool {
	if t, ok := target.(**APIStatusError); ok {
		*t = e
		return true
	}
	return false
}

// BadRequestError is returned for HTTP 400 (invalid indicator).
type BadRequestError struct{ APIStatusError }

// AuthenticationError is returned for HTTP 401 (missing / invalid API key).
type AuthenticationError struct{ APIStatusError }

// QuotaExceededError is returned for HTTP 402 (out of credits or monthly quota).
type QuotaExceededError struct{ APIStatusError }

// SubscriptionInactiveError is returned for HTTP 403 (subscription not active).
type SubscriptionInactiveError struct{ APIStatusError }

// NotFoundError is returned for HTTP 404 (no data for the supplied indicator).
type NotFoundError struct{ APIStatusError }

// RateLimitError is returned for HTTP 429 (request limit exceeded).
type RateLimitError struct {
	APIStatusError
	// RetryAfter is the Retry-After header value in seconds, if the server sent one.
	RetryAfter float64
}

// ServerError is returned for HTTP 5xx (server or upstream failure).
type ServerError struct{ APIStatusError }

// APIConnectionError is returned when the request never reached the API
// (DNS, connection refused, TLS, ...). Unwrap to inspect the underlying cause.
type APIConnectionError struct {
	Message string
	Err     error
}

func (e *APIConnectionError) Error() string   { return e.Message }
func (e *APIConnectionError) Unwrap() error   { return e.Err }
func (e *APIConnectionError) crawlsnapError() {}

// As lets errors.As extract the embedded *APIConnectionError from *APITimeoutError.
func (e *APIConnectionError) As(target any) bool {
	if t, ok := target.(**APIConnectionError); ok {
		*t = e
		return true
	}
	return false
}

// APITimeoutError is returned when the request timed out. It embeds
// *APIConnectionError, so it is also matched by errors.As as a connection error.
type APITimeoutError struct {
	*APIConnectionError
}
