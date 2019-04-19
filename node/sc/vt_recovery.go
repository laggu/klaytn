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
	"github.com/ground-x/klaytn/contracts/bridge"
	"github.com/pkg/errors"
)

// valueTransferHint stores the last handled block number and nonce (Request or Handle).
type valueTransferHint struct {
	blockNumber  uint64
	requestNonce uint64
	handleNonce  uint64
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

// getValueTransferHint gets a hint for value transfer transactions.
// The hint includes block number of the request value transfer, request nonce and handle nonce.
func (vtr *valueTransferRecovery) getValueTransferHint(requestBridge, handleBridge *bridge.Bridge) (*valueTransferHint, error) {
	if requestBridge == nil {
		return nil, errors.New("request bridge is nil")
	}
	if handleBridge == nil {
		return nil, errors.New("handle bridge is nil")
	}
	blockNumber, err := handleBridge.LastHandledRequestBlockNumber(nil)
	if err != nil {
		return nil, err
	}
	requestNonce, err := requestBridge.RequestNonce(nil)
	if err != nil {
		return nil, err
	}
	handleNonce, err := handleBridge.HandleNonce(nil)
	if err != nil {
		return nil, err
	}
	// -1 to get a nonce in the logs.
	return &valueTransferHint{blockNumber, requestNonce - 1, handleNonce - 1}, nil
}

// getPendingEvents gets pending events by using a bridge's log filter.
// The filter uses a hint as a search range. It returns a slice of events that has log details.
func (vtr *valueTransferRecovery) getPendingEvents(br *bridge.Bridge, hint *valueTransferHint) ([]*bridge.BridgeRequestValueTransfer, error) {
	if br == nil {
		return []*bridge.BridgeRequestValueTransfer{}, errors.New("bridge is nil")
	}
	if hint.requestNonce == hint.handleNonce {
		return []*bridge.BridgeRequestValueTransfer{}, nil
	}

	vtr.pendingEvents = []*bridge.BridgeRequestValueTransfer{}
	it, err := br.FilterRequestValueTransfer(&bind.FilterOpts{Start: hint.blockNumber}) // to the current
	if err != nil {
		return []*bridge.BridgeRequestValueTransfer{}, err
	}

	for it.Next() {
		logger.Debug("pending nonce in the events", it.Event.RequestNonce)
		if it.Event.RequestNonce > hint.handleNonce {
			logger.Debug("filtered pending nonce", it.Event.RequestNonce)
			vtr.pendingEvents = append(vtr.pendingEvents, it.Event)
		}
	}
	logger.Debug("pending events", len(vtr.pendingEvents))

	return vtr.pendingEvents, nil
}

// recoverTransactions recovers all pending transactions by resending them.
// TODO-Klaytn-Servicechain: implement resending transaction
func (vtr *valueTransferRecovery) recoverTransactions() {
	vtr.done = true
}
