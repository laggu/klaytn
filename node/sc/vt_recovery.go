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

// valueTransferRecovery stores status information for the value transfer recovery.
type valueTransferRecovery struct {
	config        *SCConfig
	requestBridge *bridge.Bridge
	handleBridge  *bridge.Bridge
	hint          *valueTransferHint
	pendingEvents []*bridge.BridgeRequestValueTransfer
	auth          *bind.TransactOpts
	gasLimit      uint64
}

// NewValueTransferRecovery creates a new value transfer recovery structure.
func NewValueTransferRecovery(config *SCConfig, requestBridge, handleBridge *bridge.Bridge, auth *bind.TransactOpts) *valueTransferRecovery {
	return &valueTransferRecovery{
		config:        config,
		requestBridge: requestBridge,
		handleBridge:  handleBridge,
		hint:          nil,
		pendingEvents: []*bridge.BridgeRequestValueTransfer{},
		auth:          auth,
		gasLimit:      100000,
	}
}

// Recover handles the whole recovery process of the value transfer recovery.
func (vtr *valueTransferRecovery) Recover() error {
	if !vtr.config.VTRecovery {
		logger.Debug("value transfer recovery is disabled")
		return nil
	}

	logger.Debug("start to get value transfer hint")
	_, err := vtr.getRecoveryHint()
	if err != nil {
		return err
	}

	logger.Debug("start to get pending events")
	events, err := vtr.getPendingEvents()
	if err != nil {
		return err
	}
	if len(events) == 0 {
		logger.Debug("nothing to recover value transfer")
		return nil
	}

	logger.Debug("start to recover pending transactions")
	err = vtr.recoverTransactions()
	if err != nil {
		return err
	}

	return nil
}

// getValueTransferHint gets a hint for value transfer transactions.
// The hint includes a block number to begin search, request nonce and handle nonce.
func (vtr *valueTransferRecovery) getRecoveryHint() (*valueTransferHint, error) {
	if vtr.requestBridge == nil {
		return nil, errors.New("request bridge is nil")
	}
	if vtr.handleBridge == nil {
		return nil, errors.New("handle bridge is nil")
	}
	blockNumber, err := vtr.handleBridge.LastHandledRequestBlockNumber(nil)
	if err != nil {
		return nil, err
	}
	requestNonce, err := vtr.requestBridge.RequestNonce(nil)
	if err != nil {
		return nil, err
	}
	handleNonce, err := vtr.handleBridge.HandleNonce(nil)
	if err != nil {
		return nil, err
	}
	// -1 to get a nonce in the logs.
	vtr.hint = &valueTransferHint{blockNumber, requestNonce - 1, handleNonce - 1}
	return vtr.hint, nil
}

// getPendingEvents gets pending events by using a bridge's log filter.
// The filter uses a hint as a search range. It returns a slice of events that has log details.
func (vtr *valueTransferRecovery) getPendingEvents() ([]*bridge.BridgeRequestValueTransfer, error) {
	if vtr.requestBridge == nil {
		return []*bridge.BridgeRequestValueTransfer{}, errors.New("bridge is nil")
	}
	if vtr.hint.requestNonce == vtr.hint.handleNonce {
		return []*bridge.BridgeRequestValueTransfer{}, nil
	}

	vtr.pendingEvents = []*bridge.BridgeRequestValueTransfer{}
	it, err := vtr.requestBridge.FilterRequestValueTransfer(&bind.FilterOpts{Start: vtr.hint.blockNumber}) // to the current
	if err != nil {
		return []*bridge.BridgeRequestValueTransfer{}, err
	}

	for it.Next() {
		logger.Debug("pending nonce in the events", it.Event.RequestNonce)
		if it.Event.RequestNonce > vtr.hint.handleNonce {
			logger.Debug("filtered pending nonce", it.Event.RequestNonce)
			vtr.pendingEvents = append(vtr.pendingEvents, it.Event)
		}
	}
	logger.Debug("pending events", len(vtr.pendingEvents))

	return vtr.pendingEvents, nil
}

// recoverTransactions recovers all pending transactions by resending them.
func (vtr *valueTransferRecovery) recoverTransactions() error {
	if !vtr.config.VTRecovery {
		logger.Debug("value transfer recovery is disabled")
		return nil
	}

	defer func() {
		vtr.pendingEvents = []*bridge.BridgeRequestValueTransfer{}
	}()

	for _, ev := range vtr.pendingEvents {
		logger.Warn("try to recover value transfer transaction", ev.Raw.TxHash, "nonce", ev.RequestNonce)
		opts := &bind.TransactOpts{From: vtr.auth.From, Signer: vtr.auth.Signer, GasLimit: vtr.gasLimit}
		_, err := vtr.handleBridge.HandleKLAYTransfer(opts, ev.Amount, ev.To, ev.RequestNonce, ev.Raw.BlockNumber)
		if err != nil {
			logger.Error("fail to recover handle value transfer transaction.", "nonce", ev.RequestNonce)
			return err
		}
	}

	return nil
}
