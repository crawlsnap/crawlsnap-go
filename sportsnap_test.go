package crawlsnap

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSportSnapChannel(t *testing.T) {
	c, _ := newTestClient(t, 2)
	ch, err := c.SportSnap.Channel(context.Background(), "bein-connect-turkey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.GetSlug() != "bein-connect-turkey" {
		t.Errorf("slug = %q, want bein-connect-turkey", ch.GetSlug())
	}
	rights := ch.GetBroadcastRights()
	if len(rights) != 1 || rights[0].GetCompetition() != "England - Premier League" {
		t.Errorf("broadcast_rights = %+v, want one Premier League entry", rights)
	}
	if rights[0].GetYearEnd() != 2027 {
		t.Errorf("year_end = %d, want 2027", rights[0].GetYearEnd())
	}
}

func TestSportSnapChannelSchedule(t *testing.T) {
	c, _ := newTestClient(t, 2)
	sched, err := c.SportSnap.ChannelSchedule(context.Background(), "bein-connect-turkey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := sched.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0].GetMatchId() != 5542814 {
		t.Errorf("match_id = %d, want 5542814", entries[0].GetMatchId())
	}
	if entries[0].GetMatchTitle() != "Brazil vs Norway" {
		t.Errorf("match_title = %q, want Brazil vs Norway", entries[0].GetMatchTitle())
	}
	if entries[0].GetIsPlaceholder() {
		t.Error("is_placeholder = true, want false")
	}
	if entries[0].GetKickoffLocal() != "10:00pm" || entries[0].GetKickoffRaw() != "10:00pm" {
		t.Errorf("kickoff_local/kickoff_raw = %q/%q, want 10:00pm/10:00pm", entries[0].GetKickoffLocal(), entries[0].GetKickoffRaw())
	}
}

func TestSportSnapMatch(t *testing.T) {
	c, _ := newTestClient(t, 2)
	match, err := c.SportSnap.Match(context.Background(), 5542814)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(match.GetStatus()) != "finished" {
		t.Errorf("status = %q, want finished", match.GetStatus())
	}
	score := match.GetScore()
	if score.GetHome() != 2 || score.GetAway() != 1 {
		t.Errorf("score = %d-%d, want 2-1", score.GetHome(), score.GetAway())
	}
	broadcasts := match.GetBroadcasts()
	if len(broadcasts) != 1 || broadcasts[0].GetCountrySlug() != "turkey" {
		t.Fatalf("broadcasts = %+v, want one turkey entry", broadcasts)
	}
	if chs := broadcasts[0].GetChannels(); len(chs) != 1 || chs[0].GetSlug() != "bein-connect-turkey" {
		t.Errorf("channels = %+v, want bein-connect-turkey", chs)
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

func TestSportSnapCountryChannels(t *testing.T) {
	c, _ := newTestClient(t, 2)
	cc, err := c.SportSnap.CountryChannels(context.Background(), "turkey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc.GetCountry() != "Turkey" {
		t.Errorf("country = %q, want Turkey", cc.GetCountry())
	}
	if chs := cc.GetChannels(); len(chs) != 1 || chs[0].GetSlug() != "bein-connect-turkey" {
		t.Errorf("channels = %+v, want bein-connect-turkey", chs)
	}
}

func TestSportSnapDailySchedule(t *testing.T) {
	c, _ := newTestClient(t, 2)
	day, err := c.SportSnap.DailySchedule(context.Background(), "2026-07-05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comps := day.GetCompetitions()
	if len(comps) != 1 || comps[0].GetCompetition() != "Friendly" {
		t.Fatalf("competitions = %+v, want one Friendly entry", comps)
	}
	if m := comps[0].GetMatches(); len(m) != 1 || m[0].GetTitle() != "Brazil vs Norway" {
		t.Errorf("matches = %+v, want Brazil vs Norway", m)
	}
}

func TestSportSnapDailyScheduleTime(t *testing.T) {
	c, _ := newTestClient(t, 2)
	// The time.Time variant formats to YYYY-MM-DD and hits the same endpoint.
	day, err := c.SportSnap.DailyScheduleTime(context.Background(), time.Date(2026, 7, 5, 15, 4, 5, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if day.GetDate() != "2026-07-05" {
		t.Errorf("date = %v, want 2026-07-05", day.GetDate())
	}
}

func TestSportSnapVersionPinning(t *testing.T) {
	c, _ := newTestClient(t, 2)
	if c.SportSnap.V1() != c.SportSnap.V1() {
		t.Error("V1() should return the same cached instance")
	}
	ch, err := c.SportSnap.V1().Channel(context.Background(), "bein-connect-turkey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.GetSlug() != "bein-connect-turkey" {
		t.Errorf("slug = %q, want bein-connect-turkey", ch.GetSlug())
	}
}
