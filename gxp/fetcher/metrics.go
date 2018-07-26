package fetcher

import "github.com/ground-x/go-gxplatform/metrics"

var (
	propAnnounceInMeter   = metrics.NewRegisteredMeter("klay-fetcher-prop-announces-in", nil)
	propAnnounceOutTimer  = metrics.NewRegisteredTimer("klay-fetcher-prop-announces-out", nil)
	propAnnounceDropMeter = metrics.NewRegisteredMeter("klay-fetcher-prop-announces-drop", nil)
	propAnnounceDOSMeter  = metrics.NewRegisteredMeter("klay-fetcher-prop-announces-dos", nil)

	propBroadcastInMeter   = metrics.NewRegisteredMeter("klay-fetcher-prop-broadcasts-in", nil)
	propBroadcastOutTimer  = metrics.NewRegisteredTimer("klay-fetcher-prop-broadcasts-out", nil)
	propBroadcastDropMeter = metrics.NewRegisteredMeter("klay-fetcher-prop-broadcasts-drop", nil)
	propBroadcastDOSMeter  = metrics.NewRegisteredMeter("klay-fetcher-prop-broadcasts-dos", nil)

	headerFetchMeter = metrics.NewRegisteredMeter("klay-fetcher-fetch-headers", nil)
	bodyFetchMeter   = metrics.NewRegisteredMeter("klay-fetcher-fetch-bodies", nil)

	headerFilterInMeter  = metrics.NewRegisteredMeter("klay-fetcher-filter-headers-in", nil)
	headerFilterOutMeter = metrics.NewRegisteredMeter("klay-fetcher-filter-headers-out", nil)
	bodyFilterInMeter    = metrics.NewRegisteredMeter("klay-fetcher-filter-bodies-in", nil)
	bodyFilterOutMeter   = metrics.NewRegisteredMeter("klay-fetcher-filter-bodies-out", nil)
)
