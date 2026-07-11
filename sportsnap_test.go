package crawlsnap

import (
	"context"
	"errors"
	"testing"
)

func TestSportSnapLivescores(t *testing.T) {
	c, _ := newTestClient(t, 2)
	board, err := c.SportSnap.Livescores(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if board.GetSport() != "soccer" {
		t.Errorf("sport = %q, want soccer", board.GetSport())
	}
	matches := board.GetMatches()
	if len(matches) != 1 || matches[0].GetGame() != "Brazil vs Norway" {
		t.Fatalf("matches = %+v, want one Brazil vs Norway entry", matches)
	}
	if matches[0].GetResult() != "2 - 1" {
		t.Errorf("result = %q, want 2 - 1", matches[0].GetResult())
	}
}

func TestSportSnapMatches(t *testing.T) {
	c, _ := newTestClient(t, 2)
	fx, err := c.SportSnap.Matches(context.Background(), "TR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comps := fx.GetCompetitions()
	if len(comps) != 1 || comps[0].GetCompetition() != "Friendly" {
		t.Fatalf("competitions = %+v, want one Friendly entry", comps)
	}
	if f := comps[0].GetFixtures(); len(f) != 1 || f[0].GetTeam1Name() != "Brazil" {
		t.Errorf("fixtures = %+v, want Brazil vs Norway", f)
	}
}

func TestSportSnapMatch(t *testing.T) {
	c, _ := newTestClient(t, 2)
	match, err := c.SportSnap.Match(context.Background(), 5542814)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comp := match.GetCompetition()
	if comp.GetCompetition() != "Friendly" {
		t.Errorf("competition = %q, want Friendly", comp.GetCompetition())
	}
	fixture := match.GetFixture()
	if fixture.GetFixtureId() != "5542814" || fixture.GetResult() != "2 - 1" {
		t.Errorf("fixture = %+v, want id 5542814 result 2 - 1", fixture)
	}
}

func TestSportSnapMatchNotFound(t *testing.T) {
	c, _ := newTestClient(t, 2)
	_, err := c.SportSnap.Match(context.Background(), 404)
	var nf *NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("error = %v, want *NotFoundError", err)
	}
}

func TestSportSnapCompetitions(t *testing.T) {
	c, _ := newTestClient(t, 2)
	cat, err := c.SportSnap.Competitions(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comps := cat.GetCompetitions()
	popular := comps.GetCompPopular()
	if len(popular) != 1 || popular[0].GetName() != "Premier League" {
		t.Errorf("comp_popular = %+v, want one Premier League entry", popular)
	}
}

func TestSportSnapCompetition(t *testing.T) {
	c, _ := newTestClient(t, 2)
	comp, err := c.SportSnap.Competition(context.Background(), "england", "premier-league")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	header := comp.GetCompetition()
	if header.GetSlug() != "premier-league" {
		t.Errorf("slug = %q, want premier-league", header.GetSlug())
	}
	if f := comp.GetFixtures(); len(f) != 1 || f[0].GetTeam1Name() != "Arsenal" {
		t.Errorf("fixtures = %+v, want Arsenal vs Chelsea", f)
	}
}

func TestSportSnapNationalTeam(t *testing.T) {
	c, _ := newTestClient(t, 2)
	team, err := c.SportSnap.NationalTeam(context.Background(), "brazil")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tm := team.GetTeam()
	if tm.GetTitle() != "Brazil" {
		t.Errorf("title = %q, want Brazil", tm.GetTitle())
	}
}

func TestSportSnapChannels(t *testing.T) {
	c, _ := newTestClient(t, 2)
	list, err := c.SportSnap.Channels(context.Background(), "TR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	chs := list.GetChannels()
	if len(chs) != 1 || chs[0].GetSlug() != "bein-connect-turkey" {
		t.Errorf("channels = %+v, want bein-connect-turkey", chs)
	}
}

func TestSportSnapChannelInfo(t *testing.T) {
	c, _ := newTestClient(t, 2)
	info, err := c.SportSnap.ChannelInfo(context.Background(), "bein-connect-turkey", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	channel := info.GetChannel()
	if channel.GetName() != "beIN CONNECT Turkey" {
		t.Errorf("name = %q, want beIN CONNECT Turkey", channel.GetName())
	}
	if rights := info.GetTvRights(); len(rights) != 1 || rights[0].GetName() != "beIN Sports" {
		t.Errorf("tv_rights = %+v, want one beIN Sports entry", rights)
	}
}

func TestSportSnapNews(t *testing.T) {
	c, _ := newTestClient(t, 2)
	feed, err := c.SportSnap.News(context.Background(), 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	articles := feed.GetArticles()
	if len(articles) != 1 || articles[0].GetTitle() != "Transfer news" {
		t.Errorf("articles = %+v, want one Transfer news entry", articles)
	}
}

func TestSportSnapSearchAll(t *testing.T) {
	c, _ := newTestClient(t, 2)
	res, err := c.SportSnap.SearchAll(context.Background(), "barcelona")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	results := res.GetResults()
	if len(results) != 1 || results[0].GetTitle() != "Barcelona" {
		t.Errorf("results = %+v, want one Barcelona entry", results)
	}
	if results[0].GetType() != "team" {
		t.Errorf("type = %q, want team", results[0].GetType())
	}
}

func TestSportSnapPlayer(t *testing.T) {
	c, _ := newTestClient(t, 2)
	player, err := c.SportSnap.Player(context.Background(), "messi", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	profile := player.GetProfile()
	if profile.GetName() != "Lionel Messi" {
		t.Errorf("name = %q, want Lionel Messi", profile.GetName())
	}
}

func TestSportSnapVersionPinning(t *testing.T) {
	c, _ := newTestClient(t, 2)
	if c.SportSnap.V1() != c.SportSnap.V1() {
		t.Error("V1() should return the same cached instance")
	}
	board, err := c.SportSnap.V1().Livescores(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if board.GetSport() != "soccer" {
		t.Errorf("sport = %q, want soccer", board.GetSport())
	}
}
