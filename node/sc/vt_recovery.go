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
	blockNumber     uint64 // block number to start searching event logs
	requestNonce    uint64
	handleNonce     uint64
	prevHandleNonce uint64 // previous handleNonce between recovery interval
	candidate       bool   // to check recovery candidate between recovery interval
}

// valueTransferRecovery stores status information for the value transfer recovery.
type valueTransferRecovery struct {
	stopCh    chan interface{}
	isRunning bool           // to check duplicated start
	wg        sync.WaitGroup // wait group to handle the Stop() sync

	service2mainHint   *valueTransferHint
	main2serviceHint   *valueTransferHint
	serviceChainEvents []*bridge.BridgeRequestValueTransfer
	mainChainEvents    []*bridge.BridgeRequestValueTransfer

	config       *SCConfig
	scBridgeInfo *BridgeInfo
	mcBridgeInfo *BridgeInfo
}

// NewValueTransferRecovery creates a new value transfer recovery structure.
func NewValueTransferRecovery(config *SCConfig, scBridgeInfo, mcBridgeInfo *BridgeInfo) *valueTransferRecovery {
	return &valueTransferRecovery{
		stopCh:             make(chan interface{}),
		isRunning:          false,
		wg:                 sync.WaitGroup{},
		service2mainHint:   &valueTransferHint{},
		main2serviceHint:   &valueTransferHint{},
		serviceChainEvents: []*bridge.BridgeRequestValueTransfer{},
		mainChainEvents:    []*bridge.BridgeRequestValueTransfer{},
		config:             config,
		scBridgeInfo:       scBridgeInfo,
		mcBridgeInfo:       mcBridgeInfo,
	}
}

// Start implements starting all internal goroutines used by the value transfer recovery.
func (vtr *valueTransferRecovery) Start() error {
	if !vtr.config.VTRecovery {
		logger.Debug("value transfer recovery is disabled")
		return nil
	}
	if vtr.isRunning {
		logger.Info("value transfer recovery is already started")
		return nil
	}

	err := vtr.Recover()
	if err != nil {
		logger.Info("value transfer recovery is failed")
		return err
	}

	vtr.isRunning = true
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
				logger.Info("value transfer recovery is stopped")
				return
			case <-ticker.C:
				if vtr.isRunning {
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
	if !vtr.isRunning {
		logger.Info("value transfer recovery is already stopped")
		return nil
	}
	close(vtr.stopCh)
	vtr.wg.Wait()
	vtr.isRunning = false
	return nil
}

// Recover implements the whole recovery process of the value transfer recovery.
func (vtr *valueTransferRecovery) Recover() error {
	if !vtr.config.VTRecovery {
		logger.Debug("value transfer recovery is disabled")
		return nil
	}

	logger.Debug("update value transfer hint")
	err := vtr.updateRecoveryHint()
	if err != nil {
		return err
	}

	logger.Debug("retrieve pending events")
	err = vtr.retrievePendingEvents()
	if err != nil {
		return err
	}

	logger.Debug("recover pending events")
	err = vtr.recoverPendingEvents()
	if err != nil {
		return err
	}

	return nil
}

// updateRecoveryHint updates hints for value transfers on the both side.
// One is from service chain to main chain, the other is from main chain to service chain value transfers.
// The hint includes a block number to begin search, request nonce and handle nonce.
func (vtr *valueTransferRecovery) updateRecoveryHint() error {
	if vtr.scBridgeInfo == nil {
		return errors.New("service chain bridge is nil")
	}
	if vtr.mcBridgeInfo == nil {
		return errors.New("main chain bridge is nil")
	}

	var err error
	vtr.service2mainHint, err = updateRecoveryHintFromTo(vtr.scBridgeInfo.bridge, vtr.mcBridgeInfo.bridge)
	if err != nil {
		return err
	}

	vtr.main2serviceHint, err = updateRecoveryHintFromTo(vtr.mcBridgeInfo.bridge, vtr.scBridgeInfo.bridge)
	if err != nil {
		return err
	}

	// Update the hint for the initial status.
	if !vtr.isRunning {
		vtr.service2mainHint.prevHandleNonce = vtr.service2mainHint.handleNonce
		vtr.main2serviceHint.prevHandleNonce = vtr.main2serviceHint.handleNonce
		vtr.service2mainHint.candidate = true
		vtr.main2serviceHint.candidate = true
	}

	return nil
}

// updateRecoveryHint updates a hint for the one-way value transfers.
func updateRecoveryHintFromTo(from, to *bridge.Bridge) (*valueTransferHint, error) {
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

// retrievePendingEvents retrieves pending events on the service chain or main chain.
// The pending event is the value transfer without processing HandleValueTransfer.
func (vtr *valueTransferRecovery) retrievePendingEvents() error {
	if vtr.scBridgeInfo.bridge == nil {
		return errors.New("bridge is nil")
	}

	var err error
	vtr.serviceChainEvents, err = retrievePendingEventsFrom(vtr.service2mainHint, vtr.scBridgeInfo.bridge)
	if err != nil {
		return err
	}
	vtr.mainChainEvents, err = retrievePendingEventsFrom(vtr.main2serviceHint, vtr.mcBridgeInfo.bridge)
	if err != nil {
		return err
	}

	return nil
}

// retrievePendingEventsFrom retrieves pending events from the specified bridge by using the hint provided.
// The filter uses a hint as a search range. It returns a slice of events that has log details.
func retrievePendingEventsFrom(hint *valueTransferHint, br *bridge.Bridge) ([]*bridge.BridgeRequestValueTransfer, error) {
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

// checkRecoveryCandidateCondition checks if vtr is recovery candidate or not.
// candidate is introduced to check any normal request just before checking start.
//
// For example,
//
// ======== ======== ======== ========
// Round    R Nonce  H Nonce  Result
// ======== ======== ======== ========
// 1        10       10       false
// <burst requests just before checking>
// 2        1000     10       ? (it can be normal but candidate)
// 3        2000     10       true
func checkRecoveryCandidateCondition(hint *valueTransferHint) bool {
	return hint.requestNonce != hint.handleNonce && hint.prevHandleNonce == hint.handleNonce
}

// checkRecoveryCondition checks if recovery for the handle value transfers is needed or not.
func checkRecoveryCondition(hint *valueTransferHint) bool {
	if checkRecoveryCandidateCondition(hint) && hint.candidate {
		hint.candidate = false
		return true
	}
	if checkRecoveryCandidateCondition(hint) && !hint.candidate {
		hint.candidate = true
		return false
	}
	hint.candidate = false
	return false
}

// recoverPendingEvents recovers all pending events by resending them.
func (vtr *valueTransferRecovery) recoverPendingEvents() error {
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
		logger.Warn("try to recover service chain's value transfer events", ev.Raw.TxHash, "nonce", ev.RequestNonce)
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
		logger.Warn("try to recover main chain's value transfer events", ev.Raw.TxHash, "nonce", ev.RequestNonce)
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
