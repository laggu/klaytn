package fetcher

import "github.com/ground-x/go-gxplatform/metrics"

var (
	propAnnounceInMeter   = metrics.NewRegisteredMeter("cn/fetcher/prop/announces/in", nil)
	propAnnounceOutTimer  = metrics.NewRegisteredTimer("cn/fetcher/prop/announces/out", nil)
	propAnnounceDropMeter = metrics.NewRegisteredMeter("cn/fetcher/prop/announces/drop", nil)
	propAnnounceDOSMeter  = metrics.NewRegisteredMeter("cn/fetcher/prop/announces/dos", nil)

	propBroadcastInMeter   = metrics.NewRegisteredMeter("cn/fetcher/prop/broadcasts/in", nil)
	propBroadcastOutTimer  = metrics.NewRegisteredTimer("cn/fetcher/prop/broadcasts/out", nil)
	propBroadcastDropMeter = metrics.NewRegisteredMeter("cn/fetcher/prop/broadcasts/drop", nil)
	propBroadcastDOSMeter  = metrics.NewRegisteredMeter("cn/fetcher/prop/broadcasts/dos", nil)

	headerFetchMeter = metrics.NewRegisteredMeter("cn/fetcher/fetch/headers", nil)
	bodyFetchMeter   = metrics.NewRegisteredMeter("cn/fetcher/fetch/bodies", nil)

	headerFilterInMeter  = metrics.NewRegisteredMeter("cn/fetcher/filter/headers/in", nil)
	headerFilterOutMeter = metrics.NewRegisteredMeter("cn/fetcher/filter/headers/out", nil)
	bodyFilterInMeter    = metrics.NewRegisteredMeter("cn/fetcher/filter/bodies/in", nil)
	bodyFilterOutMeter   = metrics.NewRegisteredMeter("cn/fetcher/filter/bodies/out", nil)
)
