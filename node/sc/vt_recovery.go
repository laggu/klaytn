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
	"sync"
	"time"
)

// valueTransferHint stores the last handled block number and nonce (Request or Handle).
type valueTransferHint struct {
	blockNumber     uint64
	requestNonce    uint64
	handleNonce     uint64
	lastHandleNonce uint64
}

// valueTransferRecovery stores status information for the value transfer recovery.
type valueTransferRecovery struct {
	config             *SCConfig
	scBridgeInfo       *BridgeInfo
	mcBridgeInfo       *BridgeInfo
	service2mainHint   *valueTransferHint
	main2serviceHint   *valueTransferHint
	serviceChainEvents []*bridge.BridgeRequestValueTransfer
	mainChainEvents    []*bridge.BridgeRequestValueTransfer
	auth               *bind.TransactOpts
	gasLimit           uint64
	stopCh             chan interface{}
	status             bool // Start() or Stop() status for checking duplicated start
	wg                 sync.WaitGroup
}

// NewValueTransferRecovery creates a new value transfer recovery structure.
func NewValueTransferRecovery(config *SCConfig, scBridgeInfo, mcBridgeInfo *BridgeInfo, auth *bind.TransactOpts) *valueTransferRecovery {
	return &valueTransferRecovery{
		config:             config,
		scBridgeInfo:       scBridgeInfo,
		mcBridgeInfo:       mcBridgeInfo,
		service2mainHint:   &valueTransferHint{},
		main2serviceHint:   &valueTransferHint{},
		serviceChainEvents: []*bridge.BridgeRequestValueTransfer{},
		mainChainEvents:    []*bridge.BridgeRequestValueTransfer{},
		auth:               auth,
		gasLimit:           100000,
		stopCh:             make(chan interface{}),
		status:             false,
		wg:                 sync.WaitGroup{},
	}
}

// Start implements starting all internal goroutines used by the value transfer recovery.
func (vtr *valueTransferRecovery) Start() error {
	if !vtr.config.VTRecovery {
		logger.Debug("value transfer recovery is disabled")
		return nil
	}
	if vtr.status {
		logger.Info("value transfer recovery is already started")
		return nil
	}

	err := vtr.Recover()
	if err != nil {
		logger.Info("value transfer recovery is failed")
		return err
	}

	vtr.status = true
	vtr.wg.Add(1)

	go func() {
		ticker := time.NewTicker(time.Duration(vtr.config.VTRecoveryInterval) * time.Second)
		defer func() {
			ticker.Stop()
			vtr.wg.Done()
		}()

		for {
			select {
			case <-vtr.stopCh:
				vtr.status = false
				logger.Info("value transfer recovery is stopped")
				return
			case <-ticker.C:
				if vtr.status {
					if err := vtr.Recover(); err != nil {
						logger.Info("value transfer recovery is failed")
					}
				}
			}
		}
	}()

	return nil
}

// Stop implements terminating all internal goroutines used by the value transfer recovery.
func (vtr *valueTransferRecovery) Stop() error {
	if !vtr.status {
		logger.Info("value transfer recovery is already stopped")
		return nil
	}
	close(vtr.stopCh)
	vtr.wg.Wait()
	return nil
}

// Recover implements the whole recovery process of the value transfer recovery.
func (vtr *valueTransferRecovery) Recover() error {
	if !vtr.config.VTRecovery {
		logger.Debug("value transfer recovery is disabled")
		return nil
	}

	logger.Debug("start to get value transfer hint")
	err := vtr.getRecoveryHint()
	if err != nil {
		return err
	}

	logger.Debug("start to get pending events")
	err = vtr.getPendingEvents()
	if err != nil {
		return err
	}

	logger.Debug("start to recover pending transactions")
	err = vtr.recoverTransactions()
	if err != nil {
		return err
	}

	return nil
}

// getRecoveryHint gets hints for value transfer transactions on the both side.
// One is form service chain to main chain, the other is from main chain to service chain value transfers.
// The hint includes a block number to begin search, request nonce and handle nonce.
func (vtr *valueTransferRecovery) getRecoveryHint() error {
	if vtr.scBridgeInfo == nil {
		return errors.New("service chain bridge is nil")
	}
	if vtr.mcBridgeInfo == nil {
		return errors.New("main chain bridge is nil")
	}

	var err error
	vtr.service2mainHint, err = getRecoveryHintFromTo(vtr.scBridgeInfo.bridge, vtr.mcBridgeInfo.bridge)
	if err != nil {
		return err
	}

	vtr.main2serviceHint, err = getRecoveryHintFromTo(vtr.mcBridgeInfo.bridge, vtr.scBridgeInfo.bridge)
	if err != nil {
		return err
	}

	// Update the lastHandledNode for initial status.
	if !vtr.status {
		vtr.service2mainHint.lastHandleNonce = vtr.service2mainHint.handleNonce
		vtr.main2serviceHint.lastHandleNonce = vtr.main2serviceHint.handleNonce
	}

	return nil
}

// getRecoveryHint gets a hint for the one-way value transfer transactions.
func getRecoveryHintFromTo(from, to *bridge.Bridge) (*valueTransferHint, error) {
	var err error
	var hint valueTransferHint

	hint.blockNumber, err = to.LastHandledRequestBlockNumber(nil)
	if err != nil {
		return nil, err
	}

	hint.requestNonce, err = from.RequestNonce(nil)
	if err != nil {
		return nil, err
	}
	if hint.requestNonce > 0 {
		hint.requestNonce-- // -1 to get a nonce in the logs.
	}

	hint.handleNonce, err = to.HandleNonce(nil)
	if err != nil {
		return nil, err
	}
	if hint.handleNonce > 0 {
		hint.handleNonce-- // -1 to get a nonce in the logs.
	}

	return &hint, nil
}

// getPendingEvents gets pending events on the service chain or main chain.
// The pending event is the value transfer without processing HandleValueTransfer.
func (vtr *valueTransferRecovery) getPendingEvents() error {
	if vtr.scBridgeInfo.bridge == nil {
		return errors.New("bridge is nil")
	}

	var err error
	vtr.serviceChainEvents, err = getPendingEventsFrom(vtr.service2mainHint, vtr.scBridgeInfo.bridge)
	if err != nil {
		return err
	}
	vtr.mainChainEvents, err = getPendingEventsFrom(vtr.main2serviceHint, vtr.mcBridgeInfo.bridge)
	if err != nil {
		return err
	}

	return nil
}

// getPendingEventsFrom gets pending events from the specified bridge by using the hint provided.
// The filter uses a hint as a search range. It returns a slice of events that has log details.
func getPendingEventsFrom(hint *valueTransferHint, br *bridge.Bridge) ([]*bridge.BridgeRequestValueTransfer, error) {
	if br == nil {
		return nil, errors.New("bridge is nil")
	}
	if hint.requestNonce == hint.handleNonce {
		return nil, nil
	}
	if !checkRecoveryCondition(hint) {
		return nil, nil
	}

	var pendingEvents []*bridge.BridgeRequestValueTransfer
	it, err := br.FilterRequestValueTransfer(&bind.FilterOpts{Start: hint.blockNumber}) // to the current
	if err != nil {
		return nil, err
	}
	for it.Next() {
		logger.Debug("pending nonce in the events", it.Event.RequestNonce)
		if it.Event.RequestNonce > hint.handleNonce {
			logger.Debug("filtered pending nonce", it.Event.RequestNonce)
			pendingEvents = append(pendingEvents, it.Event)
		}
	}
	logger.Debug("pending events", len(pendingEvents))

	return pendingEvents, nil
}

// checkRecoveryCondition checks that the handle value transfer has any progress.
func checkRecoveryCondition(hint *valueTransferHint) bool {
	if hint.requestNonce != hint.handleNonce && hint.lastHandleNonce != hint.handleNonce {
		return false
	}
	return true
}

// recoverTransactions recovers all pending transactions by resending them.
func (vtr *valueTransferRecovery) recoverTransactions() error {
	if !vtr.config.VTRecovery {
		logger.Debug("value transfer recovery is disabled")
		return nil
	}

	defer func() {
		vtr.serviceChainEvents = []*bridge.BridgeRequestValueTransfer{}
		vtr.mainChainEvents = []*bridge.BridgeRequestValueTransfer{}
	}()

	var evs []*TokenReceivedEvent

	// TODO-Klaytn-ServiceChain: remove the unnecessary copy
	for _, ev := range vtr.serviceChainEvents {
		logger.Warn("try to recover service chain's value transfer transaction", ev.Raw.TxHash, "nonce", ev.RequestNonce)
		evs = append(evs, &TokenReceivedEvent{
			TokenType:    ev.Kind,
			From:         ev.From,
			To:           ev.To,
			Amount:       ev.Amount,
			RequestNonce: ev.RequestNonce,
			BlockNumber:  ev.Raw.BlockNumber,
		})
	}
	vtr.mcBridgeInfo.AddRequestValueTransferEvents(evs)

	// TODO-Klaytn-ServiceChain: remove the unnecessary copy
	for _, ev := range vtr.mainChainEvents {
		logger.Warn("try to recover main chain's value transfer transaction", ev.Raw.TxHash, "nonce", ev.RequestNonce)
		evs = append(evs, &TokenReceivedEvent{
			TokenType:    ev.Kind,
			From:         ev.From,
			To:           ev.To,
			Amount:       ev.Amount,
			RequestNonce: ev.RequestNonce,
			BlockNumber:  ev.Raw.BlockNumber,
		})
	}
	vtr.scBridgeInfo.AddRequestValueTransferEvents(evs)

	return nil
}
