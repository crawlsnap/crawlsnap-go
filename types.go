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
)
