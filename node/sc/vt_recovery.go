// Copyright 2018 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.

package sc

import (
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/contracts/bridge"
)

// valueTransferHint stores the last handled block number, tx index and bridge address.
type valueTransferHint struct {
	blockNumber   uint64
	txIndex       uint64
	bridgeAddress common.Address
}

// valueTransferRecovery stores status information for the value transfer recovery process.
type valueTransferRecovery struct {
	config        *SCConfig
	done          bool
	pendingEvents []*bridge.BridgeRequestValueTransfer
}

// NewValueTransferRecovery creates a new value transfer recovery structure.
func NewValueTransferRecovery(config *SCConfig) *valueTransferRecovery {
	return &valueTransferRecovery{
		config:        config,
		pendingEvents: []*bridge.BridgeRequestValueTransfer{},
	}
}

// getRecoveryHint loads hint information for value transfer recovery.
// TODO-Klaytn-Servicechain: load block number, tx index, bridge addresses from a journal
func (vtr *valueTransferRecovery) getRecoveryHint(contract common.Address) *valueTransferHint {
	return &valueTransferHint{3, 1, contract}
}

// getPendingEvents gets pending events by using a bridge's log filter.
// The filter uses a hint as a search range. It returns a slice of events that has log details.
// TODO-Klaytn-Servicechain: check if pending or not
func (vtr *valueTransferRecovery) getPendingEvents(br *bridge.Bridge, hint *valueTransferHint) ([]*bridge.BridgeRequestValueTransfer, error) {
	vtr.pendingEvents = []*bridge.BridgeRequestValueTransfer{}
	it, err := br.FilterRequestValueTransfer(&bind.FilterOpts{Start: hint.blockNumber})
	if err != nil {
		return []*bridge.BridgeRequestValueTransfer{}, err
	}
	for it.Next() {
		vtr.pendingEvents = append(vtr.pendingEvents, it.Event)
	}
	return vtr.pendingEvents, nil
}

// recoverTransactions recovers all pending transactions by resending them.
// TODO-Klaytn-Servicechain: implement resending transaction
func (vtr *valueTransferRecovery) recoverTransactions() {
	vtr.done = true
}
