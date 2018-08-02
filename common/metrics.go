package common

import (
	"github.com/ground-x/go-gxplatform/metrics"
	)

var (
	cacheGetLRUTryMeter   = metrics.NewRegisteredMeter("klay/cache/get/lru/try", nil)
	cacheGetLRUHitMeter   = metrics.NewRegisteredMeter("klay/cache/get/lru/hit", nil)
)
