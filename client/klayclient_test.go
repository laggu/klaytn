// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package client

import "github.com/ground-x/go-gxplatform"

// Verify that Client implements the klaytn interfaces.
var (
	// _ = klaytn.Subscription(&Client{})
	_ = klaytn.ChainReader(&Client{})
	_ = klaytn.TransactionReader(&Client{})
	_ = klaytn.ChainStateReader(&Client{})
	_ = klaytn.ChainSyncReader(&Client{})
	_ = klaytn.ContractCaller(&Client{})
	_ = klaytn.LogFilterer(&Client{})
	_ = klaytn.TransactionSender(&Client{})
	_ = klaytn.GasPricer(&Client{})
	_ = klaytn.PendingStateReader(&Client{})
	_ = klaytn.PendingContractCaller(&Client{})
	_ = klaytn.GasEstimator(&Client{})
	// _ = klaytn.PendingStateEventer(&Client{})
)
