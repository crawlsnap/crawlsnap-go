# CrawlSnap Go SDK

Official Go client for the [CrawlSnap](https://crawlsnap.com) data intelligence
platform. Typed responses, typed errors, automatic retries, and per-product API
versioning.

## Install

```bash
go get github.com/crawlsnap/crawlsnap-go
```

Requires Go 1.23+ (the paginator uses range-over-func).

## Authentication

Get an API key from your dashboard and pass it to `NewClient`, or set
`CRAWLSNAP_API_KEY` and pass an empty string.

```go
client, err := crawlsnap.NewClient("sk-cs-...")   // or NewClient("") with the env var
if err != nil {
    log.Fatal(err)
}
```

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    crawlsnap "github.com/crawlsnap/crawlsnap-go"
)

func main() {
    client, err := crawlsnap.NewClient("")
    if err != nil {
        log.Fatal(err)
    }

    ip, err := client.VectorSnap.IP(context.Background(), "8.8.8.8")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(ip.GetReputation(), ip.GetAsOwner())
}
```

Every call returns the unwrapped, typed payload on success and a typed error
otherwise — you never inspect the response envelope yourself.

## Resources

| Call | Returns |
|---|---|
| `client.VectorSnap.URL(ctx, q)` | reputation enrichment for a URL |
| `client.VectorSnap.Hash(ctx, q)` | reputation enrichment for a file hash |
| `client.VectorSnap.IP(ctx, q)` | reputation enrichment for an IP |
| `client.VectorSnap.Domain(ctx, q)` | reputation enrichment for a domain |
| `client.PulseSnap.URL / Hash / IP / Domain(ctx, q)` | threat-intelligence pulse enrichment |
| `client.SubdoSnap.Scan(ctx, q)` | one page of subdomains |
| `client.SubdoSnap.ScanIter(ctx, q)` | every subdomain across all pages |

## Errors

Failures return typed errors discoverable with `errors.As`:

```go
ip, err := client.VectorSnap.IP(ctx, "8.8.8.8")
if err != nil {
    var nf *crawlsnap.NotFoundError
    switch {
    case errors.As(err, &nf):
        // no data for this indicator
    default:
        var status *crawlsnap.APIStatusError // any HTTP status error
        if errors.As(err, &status) {
            log.Printf("status=%d request_id=%s", status.StatusCode, status.RequestID)
        }
    }
}
```

| Type | Status | Meaning |
|---|---|---|
| `BadRequestError` | 400 | invalid indicator |
| `AuthenticationError` | 401 | missing / invalid key |
| `QuotaExceededError` | 402 | out of credits or monthly quota |
| `SubscriptionInactiveError` | 403 | subscription not active |
| `NotFoundError` | 404 | no data for the indicator |
| `RateLimitError` | 429 | request limit (carries `RetryAfter`) |
| `ServerError` | 5xx | server / upstream failure |
| `APIConnectionError` | — | transport failure (DNS / connection / TLS) |
| `APITimeoutError` | — | request timed out (also an `APIConnectionError`) |

`APIStatusError` matches any status error; the `CrawlSnapError` interface
matches anything the SDK returns.

## Pagination

`ScanIter` follows the cursor for you. Range over it with Go 1.23+:

```go
for sub, err := range client.SubdoSnap.ScanIter(ctx, "example.com") {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(sub["subdomain"])
}
```

For manual paging, use `Scan` then `ScanWithCursor` with the returned `Cursor`.

## API versioning

Each product is versioned independently. A direct call uses that product's
**stable default** version (`crawlsnap.DefaultVersion`), which never moves on its
own when you upgrade the SDK. Pin one product to a specific version without
affecting the others:

```go
client.VectorSnap.IP(ctx, "8.8.8.8")        // stable default version
client.VectorSnap.V1().IP(ctx, "8.8.8.8")   // explicitly VectorSnap v1
client.PulseSnap.Domain(ctx, "example.com") // unaffected — PulseSnap default
```

## Configuration

```go
client, err := crawlsnap.NewClient("sk-cs-...",
    crawlsnap.WithBaseURL("https://api.crawlsnap.com"), // or $CRAWLSNAP_BASE_URL
    crawlsnap.WithTimeout(30*time.Second),
    crawlsnap.WithMaxRetries(3),                        // 429 / 5xx / connection
    crawlsnap.WithHTTPClient(myHTTPClient),             // advanced
)
```

Retries use exponential backoff with jitter and honor the `Retry-After` header
on 429s.

## Raw response

Capture the full envelope (status, headers, request id) alongside the typed return:

```go
var raw crawlsnap.RawResponse
ip, err := client.VectorSnap.IP(ctx, "8.8.8.8", crawlsnap.WithRawResponse(&raw))
// raw.StatusCode, raw.RequestID, raw.Headers, raw.IsSuccess, raw.Data
```

## Development

The typed models in `models/` are generated from the public OpenAPI contract;
the facade is hand-written. Regenerate the models after a contract change:

```bash
./scripts/regenerate.sh   # re-bundles the contract and regenerates models/ only
go test ./...
```

## Releasing

Releases are git tags — pushing a `vX.Y.Z` tag publishes the module to the Go
proxy (no separate registry upload). Use the helper:

```bash
./scripts/release.sh 0.2.0   # bumps version.go, runs checks, commits, tags, pushes, gh release
```

Tags are immutable: to fix a bad release, cut a new patch. A `v2`+ release
additionally requires the `/v2` module-path suffix in `go.mod`.

## License

[MIT](LICENSE)
