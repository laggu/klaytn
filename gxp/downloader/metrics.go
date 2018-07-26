package downloader

import "github.com/ground-x/go-gxplatform/metrics"

var (
	headerInMeter      = metrics.NewRegisteredMeter("klay-downloader-headers-in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("klay-downloader-headers-req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("klay-downloader-headers-drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("klay-downloader-headers-timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("klay-downloader-bodies-in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("klay-downloader-bodies-req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("klay-downloader-bodies-drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("klay-downloader-bodies-timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("klay-downloader-receipts-in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("klay-downloader-receipts-req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("klay-downloader-receipts-drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("klay-downloader-receipts-timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("klay-downloader-states-in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("klay-downloader-states-drop", nil)
)
