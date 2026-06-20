package crawlsnap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Defaults for client configuration.
const (
	DefaultBaseURL    = "https://api.crawlsnap.com"
	DefaultTimeout    = 30 * time.Second
	DefaultMaxRetries = 2
)

const (
	retryBaseDelay = 500 * time.Millisecond
	retryMaxDelay  = 8 * time.Second
)

// retryableStatuses are retried with backoff: rate limit + transient upstream failures.
var retryableStatuses = map[int]bool{429: true, 500: true, 502: true, 503: true, 504: true}

func isRetryableStatus(status int) bool { return retryableStatuses[status] }

// envelope is the BaseResponse wrapper every endpoint returns.
type envelope struct {
	Data         json.RawMessage `json:"data"`
	IsSuccess    bool            `json:"is_success"`
	Message      string          `json:"message"`
	ResponseCode *int            `json:"response_code"`
}

// RawResponse is the full envelope, populated when a call is made with
// WithRawResponse. Data holds the parsed typed payload.
type RawResponse struct {
	StatusCode   int
	IsSuccess    bool
	Data         any
	Message      string
	ResponseCode *int
	RequestID    string
	Headers      http.Header
}

// parseRetryAfter reads a Retry-After header value in seconds. The API emits the
// numeric form only; the HTTP-date form is ignored.
func parseRetryAfter(value string) (float64, bool) {
	if value == "" {
		return 0, false
	}
	secs, err := strconv.ParseFloat(value, 64)
	if err != nil || secs < 0 {
		return 0, false
	}
	return secs, true
}

// backoffDelay is the time to wait before the next attempt. When a 429 carries
// Retry-After, that value is honored (clamped to [0, retryMaxDelay] so a huge or
// absurd value can't hang the goroutine). Otherwise it's exponential backoff
// with jitter. The exponential term is capped in the float domain before the
// Duration conversion, so a large attempt count cannot overflow int64.
func backoffDelay(attempt int, retryAfter time.Duration, hasRetryAfter bool) time.Duration {
	if hasRetryAfter {
		if retryAfter < 0 {
			retryAfter = 0
		}
		if retryAfter > retryMaxDelay {
			retryAfter = retryMaxDelay
		}
		return retryAfter
	}
	maxSec := float64(retryMaxDelay) / float64(time.Second)
	delaySec := (float64(retryBaseDelay) / float64(time.Second)) * math.Pow(2, float64(attempt))
	if delaySec > maxSec {
		delaySec = maxSec
	}
	delay := time.Duration(delaySec * float64(time.Second))
	jitter := time.Duration(rand.Int63n(int64(retryBaseDelay / 2)))
	return delay + jitter
}

// connError wraps a transport (or wait-aborting) error as the appropriate typed
// SDK error. The underlying cause is preserved via Unwrap, so callers can still
// match it with errors.Is (e.g. against context.Canceled).
func connError(err error) error {
	conn := &APIConnectionError{Err: err}
	switch {
	case isTimeout(err):
		conn.Message = "request timed out: " + err.Error()
		return &APITimeoutError{APIConnectionError: conn}
	case errors.Is(err, context.Canceled):
		conn.Message = "request canceled: " + err.Error()
	default:
		conn.Message = "connection error: " + err.Error()
	}
	return conn
}

func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	return false
}

// statusToError maps an effective HTTP status to the matching typed error.
func statusToError(status int, message, requestID string, body []byte, retryAfter float64, hasRetryAfter bool) error {
	base := APIStatusError{StatusCode: status, Message: message, RequestID: requestID, Body: body}
	switch status {
	case http.StatusBadRequest:
		return &BadRequestError{base}
	case http.StatusUnauthorized:
		return &AuthenticationError{base}
	case http.StatusPaymentRequired:
		return &QuotaExceededError{base}
	case http.StatusForbidden:
		return &SubscriptionInactiveError{base}
	case http.StatusNotFound:
		return &NotFoundError{base}
	case http.StatusTooManyRequests:
		e := &RateLimitError{APIStatusError: base}
		if hasRetryAfter {
			e.RetryAfter = retryAfter
		}
		return e
	}
	if status >= 500 {
		return &ServerError{base}
	}
	return &base
}

// processResponse unwraps the envelope into typed data T, or returns a typed
// error. It contains no IO beyond reading the (already received) response body.
func processResponse[T any](body []byte, status int, header http.Header, rc *requestConfig) (T, error) {
	var zero T
	requestID := header.Get("x-request-id")

	var env envelope
	_ = json.Unmarshal(body, &env) // tolerate an unparseable body; handled below

	is2xx := status >= 200 && status < 300
	effectiveStatus := status
	ok := is2xx && env.IsSuccess
	if !is2xx || !env.IsSuccess {
		ok = false
		if env.ResponseCode != nil && *env.ResponseCode >= 400 {
			effectiveStatus = *env.ResponseCode
		}
	}

	if ok {
		var data T
		if len(env.Data) > 0 && string(env.Data) != "null" {
			if err := json.Unmarshal(env.Data, &data); err != nil {
				return zero, &DecodeError{Message: "failed to decode response data: " + err.Error(), Err: err}
			}
		}
		if rc.raw != nil {
			*rc.raw = RawResponse{
				StatusCode:   status,
				IsSuccess:    true,
				Data:         data,
				Message:      env.Message,
				ResponseCode: env.ResponseCode,
				RequestID:    requestID,
				Headers:      header,
			}
		}
		return data, nil
	}

	message := env.Message
	if message == "" {
		message = fmt.Sprintf("HTTP %d", effectiveStatus)
	}
	ra, hasRA := parseRetryAfter(header.Get("Retry-After"))
	return zero, statusToError(effectiveStatus, message, requestID, body, ra, hasRA)
}
