package crawlsnap

import (
	"context"
	"iter"
	"net/url"
	"strconv"
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

// --------------------------------------------------------------------------
// SportSnap — football (soccer) data: live scores, fixtures, match detail,
// competitions, teams, TV channels, news, search, and player profiles.
// --------------------------------------------------------------------------

// SportSnap is the football data resource. It exposes the full source surface:
// live scores with in-match events, fixture lists, single-match views (detail,
// stats, commentary, broadcasts), competition catalog and detail (fixtures,
// standings, top scorers, TV rights), national and club teams, TV channel
// directories and schedules, a news feed, full-text search across every entity
// type, and player profiles. A direct call uses the stable default API version;
// pin explicitly with V1().
//
// Path segments (competition country/slug, team country/team, player slug/id)
// come from the `url` fields in list, search, and detail payloads. iso_code is
// a two-letter region code that resolves region-specific broadcast channels;
// pass "" to use the server default.
type SportSnap struct {
	client  *Client
	version string
	cache   versionCache[SportSnap]
}

func newSportSnap(c *Client) *SportSnap {
	return &SportSnap{client: c, version: DefaultVersion}
}

// V1 pins SportSnap to API version v1.
func (r *SportSnap) V1() *SportSnap {
	return r.cache.pin("v1", func(v string) *SportSnap {
		return &SportSnap{client: r.client, version: v}
	})
}

// base returns the versioned "/vN/sport-snap" prefix.
func (r *SportSnap) base() string { return "/" + r.version + "/sport-snap" }

// isoParams returns query values carrying iso_code only when it is non-empty.
func isoParams(isoCode string) url.Values {
	if isoCode == "" {
		return nil
	}
	return url.Values{"iso_code": {isoCode}}
}

// startParams returns query values carrying start only when it is > 0 (start=0
// is the server default first page).
func startParams(start int) url.Values {
	if start <= 0 {
		return nil
	}
	return url.Values{"start": {strconv.Itoa(start)}}
}

// --- Live scores & fixtures ------------------------------------------------

// Livescores returns the live-score board for every tracked competition: score
// line, match status, and structured in-match events.
func (r *SportSnap) Livescores(ctx context.Context, opts ...RequestOption) (models.LivescoresData, error) {
	return doGet[models.LivescoresData](ctx, r.client, r.base()+"/livescores", nil, opts...)
}

// Matches returns upcoming/current fixtures grouped by competition, with the
// broadcast channels available in the requested iso_code region. Pass "" for
// the server-default region.
func (r *SportSnap) Matches(ctx context.Context, isoCode string, opts ...RequestOption) (models.MatchesData, error) {
	return doGet[models.MatchesData](ctx, r.client, r.base()+"/matches", isoParams(isoCode), opts...)
}

// MatchesExtended is the extended fixture list: it adds per-locale team-name
// translations and omits per-fixture channels. timestamp is a unix-second
// watermark for incremental fetches (0 returns everything).
func (r *SportSnap) MatchesExtended(ctx context.Context, isoCode string, timestamp int64, opts ...RequestOption) (models.MatchesData, error) {
	params := isoParams(isoCode)
	if timestamp > 0 {
		if params == nil {
			params = url.Values{}
		}
		params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	}
	return doGet[models.MatchesData](ctx, r.client, r.base()+"/xmatches", params, opts...)
}

// --- Single match ----------------------------------------------------------

// Match returns the full match view keyed by the numeric match id (fixture_id
// from fixture lists, live scores, or search): teams, kickoff, venue, lineups,
// structured events, statistics, and broadcast channels.
func (r *SportSnap) Match(ctx context.Context, id int64, opts ...RequestOption) (models.MatchData, error) {
	return doGet[models.MatchData](ctx, r.client, r.base()+"/match/"+strconv.FormatInt(id, 10), nil, opts...)
}

// MatchExtended is the extended match view: it adds *_translations maps for the
// competition and team names.
func (r *SportSnap) MatchExtended(ctx context.Context, id int64, opts ...RequestOption) (models.MatchData, error) {
	return doGet[models.MatchData](ctx, r.client, r.base()+"/xmatch/"+strconv.FormatInt(id, 10), nil, opts...)
}

// MatchStats returns the match view with statistics fields populated when the
// source provides them.
func (r *SportSnap) MatchStats(ctx context.Context, id int64, opts ...RequestOption) (models.MatchData, error) {
	return doGet[models.MatchData](ctx, r.client, r.base()+"/match/"+strconv.FormatInt(id, 10)+"/stats", nil, opts...)
}

// MatchCommentaries returns the match view with the structured event feed
// populated when the source provides it.
func (r *SportSnap) MatchCommentaries(ctx context.Context, id int64, opts ...RequestOption) (models.MatchData, error) {
	return doGet[models.MatchData](ctx, r.client, r.base()+"/match/"+strconv.FormatInt(id, 10)+"/commentaries", nil, opts...)
}

// MatchChannels returns the match view with broadcast channels resolved for the
// requested iso_code region.
func (r *SportSnap) MatchChannels(ctx context.Context, id int64, isoCode string, opts ...RequestOption) (models.MatchData, error) {
	return doGet[models.MatchData](ctx, r.client, r.base()+"/match/"+strconv.FormatInt(id, 10)+"/channels", isoParams(isoCode), opts...)
}

// MatchExtraBroadcasts returns the extended match view carrying additional
// broadcast options (repeats, on-demand listings).
func (r *SportSnap) MatchExtraBroadcasts(ctx context.Context, id int64, opts ...RequestOption) (models.MatchData, error) {
	return doGet[models.MatchData](ctx, r.client, r.base()+"/xmatch/"+strconv.FormatInt(id, 10)+"/extra_broadcasts", nil, opts...)
}

// --- Competitions ----------------------------------------------------------

// Competitions returns the competition catalog: popular competitions, domestic
// leagues grouped by country, international club competitions, and tournaments.
func (r *SportSnap) Competitions(ctx context.Context, isoCode string, opts ...RequestOption) (models.CompetitionsData, error) {
	return doGet[models.CompetitionsData](ctx, r.client, r.base()+"/competitions", isoParams(isoCode), opts...)
}

// Competition returns the competition detail: fixtures around the current date,
// standings, top scorers, and broadcast-rights holders. country/slug come from
// competition_url values in other payloads.
func (r *SportSnap) Competition(ctx context.Context, country, slug string, opts ...RequestOption) (models.CompetitionDetailData, error) {
	return doGet[models.CompetitionDetailData](ctx, r.client, r.base()+"/competitions/"+url.PathEscape(country)+"/"+url.PathEscape(slug), nil, opts...)
}

// CompetitionTables returns the competition standings, grouped by stage.
func (r *SportSnap) CompetitionTables(ctx context.Context, country, slug string, opts ...RequestOption) (models.CompetitionTablesData, error) {
	return doGet[models.CompetitionTablesData](ctx, r.client, r.base()+"/competitions/"+url.PathEscape(country)+"/"+url.PathEscape(slug)+"/tables", nil, opts...)
}

// CompetitionTvRights returns the channels holding broadcast rights for the
// competition.
func (r *SportSnap) CompetitionTvRights(ctx context.Context, country, slug string, opts ...RequestOption) (models.CompetitionTvRightsData, error) {
	return doGet[models.CompetitionTvRightsData](ctx, r.client, r.base()+"/competitions/"+url.PathEscape(country)+"/"+url.PathEscape(slug)+"/tv_rights", nil, opts...)
}

// CompetitionTwitter returns social feed items for the competition (often empty).
func (r *SportSnap) CompetitionTwitter(ctx context.Context, country, slug string, opts ...RequestOption) (models.CompetitionTwitterData, error) {
	return doGet[models.CompetitionTwitterData](ctx, r.client, r.base()+"/competitions/"+url.PathEscape(country)+"/"+url.PathEscape(slug)+"/twitter", nil, opts...)
}

// --- Teams -----------------------------------------------------------------

// PopularTeams returns the popular-teams list.
func (r *SportSnap) PopularTeams(ctx context.Context, opts ...RequestOption) (models.PopularTeamsData, error) {
	return doGet[models.PopularTeamsData](ctx, r.client, r.base()+"/teams/popular", nil, opts...)
}

// AllTeams returns the full national-team catalog: per-country men's and women's
// team names with the slugs used by NationalTeam.
func (r *SportSnap) AllTeams(ctx context.Context, opts ...RequestOption) (models.AllTeamsData, error) {
	return doGet[models.AllTeamsData](ctx, r.client, r.base()+"/teams/all", nil, opts...)
}

// NationalTeam returns the national-team view (profile, squad, competitions,
// fixtures) for a country slug.
func (r *SportSnap) NationalTeam(ctx context.Context, slug string, opts ...RequestOption) (models.TeamDetailData, error) {
	return doGet[models.TeamDetailData](ctx, r.client, r.base()+"/countries/"+url.PathEscape(slug), nil, opts...)
}

// ClubTeam returns the club view (profile, squad, competitions, fixtures).
// country/team come from team url values in other payloads (e.g. spain/barcelona).
func (r *SportSnap) ClubTeam(ctx context.Context, country, team string, opts ...RequestOption) (models.TeamDetailData, error) {
	return doGet[models.TeamDetailData](ctx, r.client, r.base()+"/teams/"+url.PathEscape(country)+"/"+url.PathEscape(team), nil, opts...)
}

// --- Channels --------------------------------------------------------------

// AllChannels returns the full TV channel catalog for the iso_code region.
func (r *SportSnap) AllChannels(ctx context.Context, isoCode string, opts ...RequestOption) (models.ChannelsData, error) {
	return doGet[models.ChannelsData](ctx, r.client, r.base()+"/all_channels", isoParams(isoCode), opts...)
}

// Channels returns the curated (football-relevant) channel list for the iso_code
// region.
func (r *SportSnap) Channels(ctx context.Context, isoCode string, opts ...RequestOption) (models.ChannelsData, error) {
	return doGet[models.ChannelsData](ctx, r.client, r.base()+"/channels", isoParams(isoCode), opts...)
}

// ChannelInfo returns channel metadata (name, platform, website, coverage) and
// the channel's broadcast rights for a channel slug.
func (r *SportSnap) ChannelInfo(ctx context.Context, slug, isoCode string, opts ...RequestOption) (models.ChannelInfoData, error) {
	return doGet[models.ChannelInfoData](ctx, r.client, r.base()+"/channels/"+url.PathEscape(slug)+"/info", isoParams(isoCode), opts...)
}

// ChannelRepeats returns the channel's repeat/upcoming broadcast schedule with
// paging cursors. An empty fixtures array is a valid result, not an error.
func (r *SportSnap) ChannelRepeats(ctx context.Context, slug, isoCode string, opts ...RequestOption) (models.ChannelRepeatsData, error) {
	return doGet[models.ChannelRepeatsData](ctx, r.client, r.base()+"/channels/"+url.PathEscape(slug)+"/repeat", isoParams(isoCode), opts...)
}

// --- News ------------------------------------------------------------------

// News returns the news feed, paginated via start (0 for the first page), for
// the iso_code region.
func (r *SportSnap) News(ctx context.Context, start int, isoCode string, opts ...RequestOption) (models.NewsListData, error) {
	params := startParams(start)
	if isoCode != "" {
		if params == nil {
			params = url.Values{}
		}
		params.Set("iso_code", isoCode)
	}
	return doGet[models.NewsListData](ctx, r.client, r.base()+"/news", params, opts...)
}

// NewsByTag returns news carrying the given tag (e.g. "messi", "world-cup"),
// paginated via start.
func (r *SportSnap) NewsByTag(ctx context.Context, tag string, start int, opts ...RequestOption) (models.NewsListData, error) {
	params := url.Values{"tag": {tag}}
	if start > 0 {
		params.Set("start", strconv.Itoa(start))
	}
	return doGet[models.NewsListData](ctx, r.client, r.base()+"/news/tags", params, opts...)
}

// NewsArticle returns a single article (body HTML, tags, byline) plus related
// articles.
func (r *SportSnap) NewsArticle(ctx context.Context, id int64, opts ...RequestOption) (models.NewsDetailData, error) {
	return doGet[models.NewsDetailData](ctx, r.client, r.base()+"/news/"+strconv.FormatInt(id, 10), nil, opts...)
}

// CompetitionNews returns news about a competition, paginated via start.
func (r *SportSnap) CompetitionNews(ctx context.Context, country, slug string, start int, opts ...RequestOption) (models.NewsListData, error) {
	return doGet[models.NewsListData](ctx, r.client, r.base()+"/news_about/competitions/"+url.PathEscape(country)+"/"+url.PathEscape(slug), startParams(start), opts...)
}

// NationalTeamNews returns news about a national team, paginated via start.
func (r *SportSnap) NationalTeamNews(ctx context.Context, slug string, start int, opts ...RequestOption) (models.NewsListData, error) {
	return doGet[models.NewsListData](ctx, r.client, r.base()+"/news_about/countries/"+url.PathEscape(slug), startParams(start), opts...)
}

// ClubTeamNews returns news about a club team, paginated via start.
func (r *SportSnap) ClubTeamNews(ctx context.Context, country, team string, start int, opts ...RequestOption) (models.NewsListData, error) {
	return doGet[models.NewsListData](ctx, r.client, r.base()+"/news_about/teams/"+url.PathEscape(country)+"/"+url.PathEscape(team), startParams(start), opts...)
}

// --- Search ----------------------------------------------------------------

// SearchAll runs full-text search across teams, competitions, matches, and
// players. Result url values feed the corresponding endpoints of this API.
func (r *SportSnap) SearchAll(ctx context.Context, query string, opts ...RequestOption) (models.SearchData, error) {
	return doGet[models.SearchData](ctx, r.client, r.base()+"/search/all", url.Values{"q": {query}}, opts...)
}

// SearchTeams runs full-text search over teams.
func (r *SportSnap) SearchTeams(ctx context.Context, query string, opts ...RequestOption) (models.SearchData, error) {
	return doGet[models.SearchData](ctx, r.client, r.base()+"/search/teams", url.Values{"q": {query}}, opts...)
}

// SearchCompetitions runs full-text search over competitions.
func (r *SportSnap) SearchCompetitions(ctx context.Context, query string, opts ...RequestOption) (models.SearchData, error) {
	return doGet[models.SearchData](ctx, r.client, r.base()+"/search/competitions", url.Values{"q": {query}}, opts...)
}

// SearchMatches runs full-text search over matches.
func (r *SportSnap) SearchMatches(ctx context.Context, query string, opts ...RequestOption) (models.SearchData, error) {
	return doGet[models.SearchData](ctx, r.client, r.base()+"/search/matches", url.Values{"q": {query}}, opts...)
}

// SearchPlayers runs full-text search over players. Result url values carry the
// slug/id segments used by Player.
func (r *SportSnap) SearchPlayers(ctx context.Context, query string, opts ...RequestOption) (models.SearchData, error) {
	return doGet[models.SearchData](ctx, r.client, r.base()+"/search/players", url.Values{"q": {query}}, opts...)
}

// PopularSearches returns the currently popular search results.
func (r *SportSnap) PopularSearches(ctx context.Context, opts ...RequestOption) (models.SearchData, error) {
	return doGet[models.SearchData](ctx, r.client, r.base()+"/search/popular", nil, opts...)
}

// --- Players ---------------------------------------------------------------

// Player returns the player profile, current team, and per-season statistics.
// The slug/id pair comes from player url values in search results, lineups, and
// squads.
func (r *SportSnap) Player(ctx context.Context, slug string, id int64, opts ...RequestOption) (models.PlayerData, error) {
	return doGet[models.PlayerData](ctx, r.client, r.base()+"/player/"+url.PathEscape(slug)+"/"+strconv.FormatInt(id, 10), nil, opts...)
}
