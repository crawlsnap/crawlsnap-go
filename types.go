package crawlsnap

import "github.com/crawlsnap/crawlsnap-go/models"

// Convenience aliases for the typed response payloads, so callers can refer to
// them without importing the models package directly. The canonical definitions
// (generated from the OpenAPI contract) live in package models.
type (
	IocURLScanData    = models.IocUrlScanData
	IocHashScanData   = models.IocHashScanData
	IocIPScanData     = models.IocIpScanData
	IocDomainScanData = models.IocDomainScanData

	PulseURLScanData    = models.PulseUrlScanData
	PulseHashScanData   = models.PulseHashScanData
	PulseIPScanData     = models.PulseIpScanData
	PulseDomainScanData = models.PulseDomainScanData

	SubdoSnapScanData = models.SubdoSnapScanData

	// SportSnap payloads.
	LivescoresData          = models.LivescoresData
	MatchesData             = models.MatchesData
	MatchData               = models.MatchData
	CompetitionsData        = models.CompetitionsData
	CompetitionDetailData   = models.CompetitionDetailData
	CompetitionTablesData   = models.CompetitionTablesData
	CompetitionTvRightsData = models.CompetitionTvRightsData
	CompetitionTwitterData  = models.CompetitionTwitterData
	PopularTeamsData        = models.PopularTeamsData
	AllTeamsData            = models.AllTeamsData
	TeamDetailData          = models.TeamDetailData
	ChannelsData            = models.ChannelsData
	ChannelInfoData         = models.ChannelInfoData
	ChannelRepeatsData      = models.ChannelRepeatsData
	NewsListData            = models.NewsListData
	NewsDetailData          = models.NewsDetailData
	SearchData              = models.SearchData
	PlayerData              = models.PlayerData
)
