package crawlsnap

import (
	"context"
	"iter"
	"net/url"
	"sync"

	"github.com/crawlsnap/crawlsnap-go/models"
)

// DefaultVersion is the stable, per-release default API version used by unpinned
// calls. It does not move on its own: upgrading the SDK never silently retargets
// your calls at a newer API version. Promoting a product's default to a new
// version is a deliberate, changelogged release change.
//
// Pin a single product to a specific version with its version accessor (e.g.
// VectorSnap.V1()) — this affects only that product, leaving the others on their
// own defaults.
const DefaultVersion = "v1"

// versionCache lazily caches one resource instance per pinned API version, so a
// repeated accessor (client.VectorSnap.V1()) returns the same instance.
type versionCache[T any] struct {
	mu   sync.Mutex
	pins map[string]*T
}

func (vc *versionCache[T]) pin(version string, mk func(string) *T) *T {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	if vc.pins == nil {
		vc.pins = make(map[string]*T)
	}
	if inst, ok := vc.pins[version]; ok {
		return inst
	}
	inst := mk(version)
	vc.pins[version] = inst
	return inst
}

// --------------------------------------------------------------------------
// VectorSnap — IoC reputation enrichment for url / hash / ip / domain.
// --------------------------------------------------------------------------

// VectorSnap is the IoC reputation enrichment resource. A direct call uses the
// stable default API version; pin explicitly with V1().
type VectorSnap struct {
	client  *Client
	version string
	cache   versionCache[VectorSnap]
}

func newVectorSnap(c *Client) *VectorSnap {
	return &VectorSnap{client: c, version: DefaultVersion}
}

// V1 pins VectorSnap to API version v1.
func (r *VectorSnap) V1() *VectorSnap {
	return r.cache.pin("v1", func(v string) *VectorSnap {
		return &VectorSnap{client: r.client, version: v}
	})
}

// URL returns reputation enrichment for a URL.
func (r *VectorSnap) URL(ctx context.Context, query string, opts ...RequestOption) (models.IocUrlScanData, error) {
	return doGet[models.IocUrlScanData](ctx, r.client, "/"+r.version+"/ioc/search/url", url.Values{"query": {query}}, opts...)
}

// Hash returns reputation enrichment for a file hash.
func (r *VectorSnap) Hash(ctx context.Context, query string, opts ...RequestOption) (models.IocHashScanData, error) {
	return doGet[models.IocHashScanData](ctx, r.client, "/"+r.version+"/ioc/search/hash", url.Values{"query": {query}}, opts...)
}

// IP returns reputation enrichment for an IP address.
func (r *VectorSnap) IP(ctx context.Context, query string, opts ...RequestOption) (models.IocIpScanData, error) {
	return doGet[models.IocIpScanData](ctx, r.client, "/"+r.version+"/ioc/search/ip", url.Values{"query": {query}}, opts...)
}

// Domain returns reputation enrichment for a domain.
func (r *VectorSnap) Domain(ctx context.Context, query string, opts ...RequestOption) (models.IocDomainScanData, error) {
	return doGet[models.IocDomainScanData](ctx, r.client, "/"+r.version+"/ioc/search/domain", url.Values{"query": {query}}, opts...)
}

// --------------------------------------------------------------------------
// PulseSnap — threat-intelligence pulse enrichment for url / hash / ip / domain.
// --------------------------------------------------------------------------

// PulseSnap is the threat-intelligence pulse enrichment resource. A direct call
// uses the stable default API version; pin explicitly with V1().
type PulseSnap struct {
	client  *Client
	version string
	cache   versionCache[PulseSnap]
}

func newPulseSnap(c *Client) *PulseSnap {
	return &PulseSnap{client: c, version: DefaultVersion}
}

// V1 pins PulseSnap to API version v1.
func (r *PulseSnap) V1() *PulseSnap {
	return r.cache.pin("v1", func(v string) *PulseSnap {
		return &PulseSnap{client: r.client, version: v}
	})
}

// URL returns pulse enrichment for a URL.
func (r *PulseSnap) URL(ctx context.Context, query string, opts ...RequestOption) (models.PulseUrlScanData, error) {
	return doGet[models.PulseUrlScanData](ctx, r.client, "/"+r.version+"/pulse-snap/scan/url", url.Values{"query": {query}}, opts...)
}

// Hash returns pulse enrichment for a file hash.
func (r *PulseSnap) Hash(ctx context.Context, query string, opts ...RequestOption) (models.PulseHashScanData, error) {
	return doGet[models.PulseHashScanData](ctx, r.client, "/"+r.version+"/pulse-snap/scan/hash", url.Values{"query": {query}}, opts...)
}

// IP returns pulse enrichment for an IP address.
func (r *PulseSnap) IP(ctx context.Context, query string, opts ...RequestOption) (models.PulseIpScanData, error) {
	return doGet[models.PulseIpScanData](ctx, r.client, "/"+r.version+"/pulse-snap/scan/ip", url.Values{"query": {query}}, opts...)
}

// Domain returns pulse enrichment for a domain.
func (r *PulseSnap) Domain(ctx context.Context, query string, opts ...RequestOption) (models.PulseDomainScanData, error) {
	return doGet[models.PulseDomainScanData](ctx, r.client, "/"+r.version+"/pulse-snap/scan/domain", url.Values{"query": {query}}, opts...)
}

// --------------------------------------------------------------------------
// SubdoSnap — paginated subdomain enumeration for a domain.
// --------------------------------------------------------------------------

// SubdoSnap is the subdomain enumeration resource. A direct call uses the stable
// default API version; pin explicitly with V1().
type SubdoSnap struct {
	client  *Client
	version string
	cache   versionCache[SubdoSnap]
}

func newSubdoSnap(c *Client) *SubdoSnap {
	return &SubdoSnap{client: c, version: DefaultVersion}
}

// V1 pins SubdoSnap to API version v1.
func (r *SubdoSnap) V1() *SubdoSnap {
	return r.cache.pin("v1", func(v string) *SubdoSnap {
		return &SubdoSnap{client: r.client, version: v}
	})
}

// Scan fetches the first page of subdomains for a domain. Use the returned
// Cursor with ScanWithCursor to page manually, or ScanIter to stream every
// subdomain automatically.
func (r *SubdoSnap) Scan(ctx context.Context, query string, opts ...RequestOption) (models.SubdoSnapScanData, error) {
	return r.scan(ctx, query, "", opts...)
}

// ScanWithCursor fetches one page of subdomains starting at the given cursor.
func (r *SubdoSnap) ScanWithCursor(ctx context.Context, query, cursor string, opts ...RequestOption) (models.SubdoSnapScanData, error) {
	return r.scan(ctx, query, cursor, opts...)
}

func (r *SubdoSnap) scan(ctx context.Context, query, cursor string, opts ...RequestOption) (models.SubdoSnapScanData, error) {
	params := url.Values{"query": {query}}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	return doGet[models.SubdoSnapScanData](ctx, r.client, "/"+r.version+"/subdo-snap/scan", params, opts...)
}

// ScanIter streams every subdomain across all pages, following the cursor for
// you. Range over it with Go 1.23+ range-over-func:
//
//	for sub, err := range client.SubdoSnap.ScanIter(ctx, "example.com") {
//	    if err != nil { return err }
//	    fmt.Println(sub["subdomain"])
//	}
//
// Iteration stops at the first error (yielded once with a nil subdomain) or when
// the cursor is exhausted.
func (r *SubdoSnap) ScanIter(ctx context.Context, query string) iter.Seq2[map[string]any, error] {
	return func(yield func(map[string]any, error) bool) {
		cursor := ""
		for {
			page, err := r.scan(ctx, query, cursor)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, sub := range page.Subdomains {
				if !yield(sub, nil) {
					return
				}
			}
			cursor = ""
			if page.Cursor != nil {
				cursor = *page.Cursor
			}
			if cursor == "" {
				return
			}
		}
	}
}
