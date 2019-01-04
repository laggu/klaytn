// Copyright 2018 The go-klaytn Authors
//
// This file is derived from metrics/exp/exp.go (2018/06/04).
// See LICENSE in the metrics directory for the original copyright and license.

// Hook go-metrics into expvar
// on any /debug/metrics request, load all vars from the registry into expvar, and execute regular expvar handler
package exp
