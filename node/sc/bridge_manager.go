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
	"context"
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/common"
	bridgecontract "github.com/ground-x/klaytn/contracts/bridge"
	"github.com/ground-x/klaytn/event"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
	"math/big"
	"path"
	"sync"
	"time"
)

const (
	TokenEventChanSize = 10000
	BridgeAddrJournal  = "bridge_addrs.rlp"
)

const (
	KLAY uint8 = iota
	TOKEN
	NFT
)

// RequestValueTransfer Event from SmartContract
type TokenReceivedEvent struct {
	TokenType    uint8
	ContractAddr common.Address
	TokenAddr    common.Address
	From         common.Address
	To           common.Address
	Amount       *big.Int // Amount is UID in NFT
	RequestNonce uint64
	BlockNumber  uint64
}

// TokenWithdraw Event from SmartContract
type TokenTransferEvent struct {
	TokenType    uint8
	ContractAddr common.Address
	TokenAddr    common.Address
	Owner        common.Address
	Amount       *big.Int // Amount is UID in NFT
	HandleNonce  uint64
}

// BridgeJournal has two types. When a single address is inserted, the Paired is disabled.
// In this case, only one of the LocalAddress or RemoteAddress is filled with the address.
// If two address in a pair is inserted, the Pared is enabled.
type BridgeJournal struct {
	LocalAddress  common.Address `json:"localAddress"`
	RemoteAddress common.Address `json:"remoteAddress"`
	Paired        bool           `json:"paired"`
}

type BridgeInfo struct {
	// TODO-Klaytn need to remove and replace subBridge in BridgeInfo
	// subBridge is used for only AddressManager().GetCounterPartToken.
	// Token pair information will be included by BridgeInfo instead of subBridge
	subBridge *SubBridge

	address        common.Address
	account        *accountInfo
	bridge         *bridgecontract.Bridge
	onServiceChain bool
	subscribed     bool

	mu                  *sync.Mutex     // mutex for pendingRequestEvent
	pendingRequestEvent *eventSortedMap // TODO-Klaytn Need to consider the nonce overflow(priority queue?) and the size overflow.
	nextHandleNonce     uint64          // This nonce will be used for getting pending request value transfer events.

	handledNonce   uint64 // the nonce from the handle value transfer event from the bridge.
	requestedNonce uint64 // the nonce from the request value transfer event from the counter part bridge.

	newEvent chan struct{}
	closed   chan struct{}
}

func NewBridgeInfo(subBridge *SubBridge, addr common.Address, bridge *bridgecontract.Bridge, account *accountInfo, local, subscribed bool) *BridgeInfo {
	bi := &BridgeInfo{
		subBridge,
		addr,
		account,
		bridge,
		local,
		subscribed,
		&sync.Mutex{},
		newEventSortedMap(),
		0,
		0,
		0,
		make(chan struct{}),
		make(chan struct{}),
	}

	handleNonce, err := bi.bridge.HandleNonce(nil)
	if err != nil {
		logger.Error("Failed to get handleNonce from the bridge", "err", err, "bridgeAddr", addr.Hex())
		return bi // return bridgeInfo with zero nonce.
		// TODO-Klaytn consider the failed case. The nonce(nextHandleNonce) should be updated later.
	}

	logger.Debug("Updated the handle nonce", "nonce", handleNonce, "bridgeAddr", addr.Hex())

	bi.nextHandleNonce = handleNonce
	bi.requestedNonce = handleNonce // This requestedNonce will be updated by counter part bridge contract's new request event.
	bi.handledNonce = handleNonce

	go bi.loop()

	return bi
}

func (bi *BridgeInfo) loop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	logger.Info("start bridge loop", "addr", bi.address.String(), "onServiceChain", bi.onServiceChain)

	for {
		select {
		case <-bi.newEvent:
			bi.processingPendingRequestEvents()

		case <-ticker.C:
			bi.processingPendingRequestEvents()

		case <-bi.closed:
			logger.Info("stop bridge loop", "addr", bi.address.String(), "onServiceChain", bi.onServiceChain)
			return
		}
	}
}

// processingPendingRequestEvents handles pending request value transfer events of the bridge.
func (bi *BridgeInfo) processingPendingRequestEvents() error {
	pendingEvent := bi.GetReadyRequestValueTransferEvents()
	if pendingEvent == nil {
		return nil
	}

	logger.Debug("Get Pending request value transfer event", "len(pendingEvent)", len(pendingEvent))

	diff := bi.requestedNonce - bi.handledNonce
	if diff > errorDiffRequestHandleNonce {
		logger.Error("Value transfer requested/handled nonce gap is too much.", "toSC", bi.onServiceChain, "diff", diff, "requestedNonce", bi.requestedNonce, "handledNonce", bi.handledNonce)
		// TODO-Klaytn need to consider starting value transfer recovery.
	}

	for idx, ev := range pendingEvent {
		if err := bi.handleRequestValueTransferEvent(ev); err != nil {
			bi.AddRequestValueTransferEvents(pendingEvent[idx:])
			logger.Error("Failed handle request value transfer event", "err", err, "len(RePutEvent)", len(pendingEvent[idx:]))
			return err
		}
		bi.nextHandleNonce++
	}

	return nil
}

// handleRequestValueTransferEvent handles the given request value transfer event.
func (bi *BridgeInfo) handleRequestValueTransferEvent(ev *TokenReceivedEvent) error {
	tokenType := ev.TokenType
	tokenAddr := bi.subBridge.AddressManager().GetCounterPartToken(ev.TokenAddr)
	if tokenType != KLAY && tokenAddr == (common.Address{}) {
		logger.Error("Unregisterd counter part token address.", "addr", tokenAddr.Hex())
		// TODO-Klaytn consider the invalid token address
		// - prevent the invalid token address in the bridge contract.
		// - ignore and keep the request with the invalid token address during handling transaction.

		// Increase only handle nonce of bridge contract.
		ev.TokenType = KLAY
		ev.Amount = big.NewInt(0)
	}

	to := ev.To

	bridgeAcc := bi.account

	bridgeAcc.Lock()
	defer bridgeAcc.UnLock()

	auth := bridgeAcc.GetTransactOpts()

	switch tokenType {
	case KLAY:
		tx, err := bi.bridge.HandleKLAYTransfer(auth, ev.Amount, to, ev.RequestNonce, ev.BlockNumber)
		if err != nil {
			return err
		}
		logger.Trace("Bridge succeeded to HandleKLAYTransfer", "nonce", ev.RequestNonce, "tx", tx.Hash().Hex())

	case TOKEN:
		tx, err := bi.bridge.HandleTokenTransfer(auth, ev.Amount, to, tokenAddr, ev.RequestNonce, ev.BlockNumber)
		if err != nil {
			return err
		}
		logger.Trace("Bridge succeeded to HandleTokenTransfer", "nonce", ev.RequestNonce, "tx", tx.Hash().Hex())
	case NFT:
		tx, err := bi.bridge.HandleNFTTransfer(auth, ev.Amount, to, tokenAddr, ev.RequestNonce, ev.BlockNumber)
		if err != nil {
			return err
		}
		logger.Trace("Bridge succeeded to HandleNFTTransfer", "nonce", ev.RequestNonce, "tx", tx.Hash().Hex())
	default:
		logger.Warn("Got Unknown Token Type ReceivedEvent", "bridge", ev.ContractAddr, "nonce", ev.RequestNonce, "from", ev.From)
		return nil
	}

	bridgeAcc.IncNonce()
	return nil
}

// UpdateRequestedNonce updates the requested nonce with new nonce.
func (bi *BridgeInfo) UpdateRequestedNonce(nonce uint64) {
	if bi.requestedNonce < nonce {
		bi.requestedNonce = nonce
	}
}

// UpdateHandledNonce updates the handled nonce with new nonce.
func (bi *BridgeInfo) UpdateHandledNonce(nonce uint64) {
	if bi.handledNonce < nonce {
		bi.handledNonce = nonce
	}
}

// AddRequestValueTransferEvents adds events into the pendingRequestEvent.
func (bi *BridgeInfo) AddRequestValueTransferEvents(evs []*TokenReceivedEvent) {
	bi.mu.Lock()
	defer bi.mu.Unlock()
	// TODO-Klaytn Need to consider the nonce overflow(priority queue?) and the size overflow.
	// - If the size is full and received event has the omitted nonce, it can be allowed.
	for _, ev := range evs {
		bi.UpdateRequestedNonce(ev.RequestNonce)
		bi.pendingRequestEvent.Put(ev)
	}

	select {
	case bi.newEvent <- struct{}{}:
	default:
	}
}

// GetReadyRequestValueTransferEvents returns the processable events with the increasing nonce.
func (bi *BridgeInfo) GetReadyRequestValueTransferEvents() []*TokenReceivedEvent {
	bi.mu.Lock()
	defer bi.mu.Unlock()
	return bi.pendingRequestEvent.Ready(bi.nextHandleNonce)
}

// DecodeRLP decodes the Klaytn
func (b *BridgeJournal) DecodeRLP(s *rlp.Stream) error {
	var elem struct {
		LocalAddress  common.Address
		RemoteAddress common.Address
		Paired        bool
	}
	if err := s.Decode(&elem); err != nil {
		return err
	}
	b.LocalAddress, b.RemoteAddress, b.Paired = elem.LocalAddress, elem.RemoteAddress, elem.Paired
	return nil
}

// EncodeRLP serializes b into the Klaytn RLP BridgeJournal format.
func (b *BridgeJournal) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		b.LocalAddress,
		b.RemoteAddress,
		b.Paired,
	})
}

// BridgeManager manages Bridge SmartContracts
// for value transfer between service chain and parent chain
type BridgeManager struct {
	subBridge *SubBridge

	receivedEvents map[common.Address]event.Subscription
	withdrawEvents map[common.Address]event.Subscription
	bridges        map[common.Address]*BridgeInfo

	tokenReceived event.Feed
	tokenWithdraw event.Feed

	scope event.SubscriptionScope

	journal    *bridgeAddrJournal
	recoveries map[common.Address]*valueTransferRecovery
	auth       *bind.TransactOpts
}

func NewBridgeManager(main *SubBridge) (*BridgeManager, error) {
	bridgeAddrJournal := newBridgeAddrJournal(path.Join(main.config.DataDir, BridgeAddrJournal), main.config)

	bridgeManager := &BridgeManager{
		subBridge:      main,
		receivedEvents: make(map[common.Address]event.Subscription),
		withdrawEvents: make(map[common.Address]event.Subscription),
		bridges:        make(map[common.Address]*BridgeInfo),
		journal:        bridgeAddrJournal,
		recoveries:     make(map[common.Address]*valueTransferRecovery),
	}

	logger.Info("Load Bridge Address from JournalFiles ", "path", bridgeManager.journal.path)
	bridgeManager.journal.cache = []*BridgeJournal{}

	if err := bridgeManager.journal.load(func(gwjournal BridgeJournal) error {
		logger.Info("Load Bridge Address from JournalFiles ",
			"local address", gwjournal.LocalAddress.Hex(), "remote address", gwjournal.RemoteAddress.Hex())
		bridgeManager.journal.cache = append(bridgeManager.journal.cache, &gwjournal)
		return nil
	}); err != nil {
		logger.Error("fail to load bridge address", "err", err)
	}

	if err := bridgeManager.journal.rotate(bridgeManager.GetAllBridge()); err != nil {
		logger.Error("fail to rotate bridge journal", "err", err)
	}

	return bridgeManager, nil
}

// LogBridgeStatus logs the bridge contract requested/handled nonce status as an information.
func (bm *BridgeManager) LogBridgeStatus() {
	for bAddr, b := range bm.bridges {
		diffNonce := b.requestedNonce - b.handledNonce

		var headStr string
		if b.onServiceChain {
			headStr = "Bridge(Main -> Service Chain)"
		} else {
			headStr = "Bridge(Service -> Main Chain)"
		}
		logger.Info(headStr, "bridge", bAddr.String(), "requestNonce", b.requestedNonce, "handleNonce", b.handledNonce, "diffNonce", diffNonce)
	}
}

// SubscribeTokenReceived registers a subscription of TokenReceivedEvent.
func (bm *BridgeManager) SubscribeTokenReceived(ch chan<- TokenReceivedEvent) event.Subscription {
	return bm.scope.Track(bm.tokenReceived.Subscribe(ch))
}

// SubscribeTokenWithDraw registers a subscription of TokenTransferEvent.
func (bm *BridgeManager) SubscribeTokenWithDraw(ch chan<- TokenTransferEvent) event.Subscription {
	return bm.scope.Track(bm.tokenWithdraw.Subscribe(ch))
}

// GetAllBridge returns a journal cache while removing unnecessary address pair.
func (bm *BridgeManager) GetAllBridge() []*BridgeJournal {
	gwjs := []*BridgeJournal{}

	for _, journal := range bm.journal.cache {
		if journal.Paired {
			bridgeInfo, ok := bm.GetBridgeInfo(journal.LocalAddress)
			if ok && !bridgeInfo.subscribed {
				if bm.subBridge.AddressManager() != nil {
					bm.subBridge.addressManager.DeleteBridge(journal.LocalAddress)
				}
				continue
			}

			bridgeInfo, ok = bm.GetBridgeInfo(journal.RemoteAddress)
			if ok && !bridgeInfo.subscribed {
				if bm.subBridge.AddressManager() != nil {
					bm.subBridge.addressManager.DeleteBridge(journal.RemoteAddress)
				}
				continue
			}
		}
		gwjs = append(gwjs, journal)
	}

	bm.journal.cache = gwjs

	return bm.journal.cache
}

// GetBridge returns bridge contract of the specified address.
func (bm *BridgeManager) GetBridgeInfo(addr common.Address) (*BridgeInfo, bool) {
	bridge, ok := bm.bridges[addr]
	return bridge, ok
}

// SetBridge stores the address and bridge pair with local/remote and subscription status.
func (bm *BridgeManager) SetBridge(addr common.Address, bridge *bridgecontract.Bridge, account *accountInfo, local bool, subscribed bool) {
	bm.bridges[addr] = NewBridgeInfo(bm.subBridge, addr, bridge, account, local, subscribed)
}

// LoadAllBridge reloads bridge and handles subscription by using the journal cache.
func (bm *BridgeManager) LoadAllBridge() error {
	bm.stopAllRecoveries()

	for _, journal := range bm.journal.cache {
		if journal.Paired {
			if bm.subBridge.AddressManager() == nil {
				return errors.New("address manager is not exist")
			}

			logger.Info("automatic bridge subscription", "local address", journal.LocalAddress, "remote address", journal.RemoteAddress)

			// 1. Register bridge.
			localBridge, err := bridgecontract.NewBridge(journal.LocalAddress, bm.subBridge.localBackend)
			if err != nil {
				return err
			}
			remoteBridge, err := bridgecontract.NewBridge(journal.RemoteAddress, bm.subBridge.remoteBackend)
			if err != nil {
				return err
			}
			bm.SetBridge(journal.LocalAddress, localBridge, bm.subBridge.bridgeAccountManager.scAccount, true, false)
			bm.SetBridge(journal.RemoteAddress, remoteBridge, bm.subBridge.bridgeAccountManager.mcAccount, false, false)

			// 2. Set the address manager.
			bm.subBridge.AddressManager().AddBridge(journal.LocalAddress, journal.RemoteAddress)

			// 3. Subscribe event.
			err = bm.subscribeEvent(journal.LocalAddress, localBridge)
			if err != nil {
				return err
			}
			err = bm.subscribeEvent(journal.RemoteAddress, remoteBridge)
			if err != nil {
				bm.subBridge.AddressManager().DeleteBridge(journal.LocalAddress)
				bm.UnsubscribeEvent(journal.LocalAddress)
				return err
			}

			// 4. Add recovery.
			bm.AddRecovery(journal.LocalAddress, journal.RemoteAddress)
		} else {
			err := bm.loadBridge(journal.LocalAddress, bm.subBridge.localBackend, true, false)
			if err != nil {
				return err
			}
			err = bm.loadBridge(journal.RemoteAddress, bm.subBridge.remoteBackend, false, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddRecovery writes journal and creates value transfer recovery for a given addresses pair.
func (bm *BridgeManager) AddRecovery(localAddress, remoteAddress common.Address) error {
	// Write journal information.
	bm.journal.insert(localAddress, remoteAddress, true)
	bm.journal.rotate(bm.GetAllBridge())

	// Check if bridge information is exist.
	localBridgeInfo, ok := bm.GetBridgeInfo(localAddress)
	if !ok {
		return errors.New("local bridge is not exist to create value transfer recovery")
	}
	remoteBridgeInfo, ok := bm.GetBridgeInfo(remoteAddress)
	if !ok {
		return errors.New("remote bridge is not exist to create value transfer recovery")
	}

	// Create and start value transfer recovery.
	recovery := NewValueTransferRecovery(bm.subBridge.config, localBridgeInfo, remoteBridgeInfo)
	recovery.Start()
	bm.recoveries[localAddress] = recovery // suppose local/remote is always a pair.
	return nil
}

// DeleteRecovery deletes the journal and stop the value transfer recovery for a given address pair.
func (bm *BridgeManager) DeleteRecovery(localAddress, remoteAddress common.Address) error {
	// Stop the recovery.
	recovery, ok := bm.recoveries[localAddress]
	if !ok {
		return errors.New("recovery is not exist")
	}
	recovery.Stop()
	delete(bm.recoveries, localAddress)

	// Delete the journal.
	bm.journal.rotate(bm.GetAllBridge())
	return nil
}

// stopAllRecoveries stops the internal value transfer recoveries.
func (bm *BridgeManager) stopAllRecoveries() {
	for _, recovery := range bm.recoveries {
		recovery.Stop()
	}
	bm.recoveries = make(map[common.Address]*valueTransferRecovery)
}

// LoadBridge creates new bridge contract for a given address and subscribes an event if needed.
func (bm *BridgeManager) loadBridge(addr common.Address, backend bind.ContractBackend, local bool, subscribed bool) error {
	var bridgeInfo *BridgeInfo

	defer func() {
		if bridgeInfo != nil && subscribed && !bridgeInfo.subscribed {
			logger.Info("bridge subscription is enabled by journal", "address", addr)
			bm.subscribeEvent(addr, bridgeInfo.bridge)
		}
	}()

	bridgeInfo, ok := bm.GetBridgeInfo(addr)
	if ok {
		return nil
	}

	bridge, err := bridgecontract.NewBridge(addr, backend)
	if err != nil {
		return err
	}
	logger.Info("bridge ", "address", addr)
	if local {
		bm.SetBridge(addr, bridge, bm.subBridge.bridgeAccountManager.scAccount, local, false)
	} else {
		bm.SetBridge(addr, bridge, bm.subBridge.bridgeAccountManager.mcAccount, local, false)
	}

	bridgeInfo, _ = bm.GetBridgeInfo(addr)

	return nil
}

// Deploy Bridge SmartContract on same node or remote node
func (bm *BridgeManager) DeployBridge(backend bind.ContractBackend, local bool) (common.Address, error) {
	var acc *accountInfo
	if local {
		acc = bm.subBridge.bridgeAccountManager.scAccount
	} else {
		acc = bm.subBridge.bridgeAccountManager.mcAccount
	}

	addr, bridge, err := bm.deployBridge(acc, backend)
	if err != nil {
		return common.Address{}, err
	}

	bm.SetBridge(addr, bridge, acc, local, false)
	return addr, err
}

// DeployBridge handles actual smart contract deployment.
// To create contract, the chain ID, nonce, account key, private key, contract binding and gas price are used.
// The deployed contract address, transaction are returned. An error is also returned if any.
func (bm *BridgeManager) deployBridge(acc *accountInfo, backend bind.ContractBackend) (common.Address, *bridgecontract.Bridge, error) {
	acc.Lock()
	defer acc.UnLock()
	auth := acc.GetTransactOpts()
	addr, tx, contract, err := bridgecontract.DeployBridge(auth, backend, true)
	if err != nil {
		logger.Error("Failed to deploy contract.", "err", err)
		return common.Address{}, nil, err
	}
	acc.IncNonce()

	logger.Info("Bridge is deploying...", "addr", addr, "txHash", tx.Hash().String())

	back, ok := backend.(bind.DeployBackend)
	if !ok {
		logger.Warn("DeployBacked type assertion is failed. Skip WaitDeployed.")
		return addr, contract, nil
	}

	timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelTimeout()

	addr, err = bind.WaitDeployed(timeoutContext, back, tx)
	if err != nil {
		logger.Error("Failed to WaitDeployed.", "err", err, "txHash", tx.Hash().String())
		return common.Address{}, nil, err
	}
	logger.Info("Bridge is deployed.", "addr", addr, "txHash", tx.Hash().String())
	return addr, contract, nil
}

// SubscribeEvent registers a subscription of BridgeERC20Received and BridgeTokenWithdrawn
func (bm *BridgeManager) SubscribeEvent(addr common.Address) error {
	bridgeInfo, ok := bm.GetBridgeInfo(addr)
	if !ok {
		return fmt.Errorf("there is no bridge contract which address %v", addr)
	}
	err := bm.subscribeEvent(addr, bridgeInfo.bridge)
	if err != nil {
		return err
	}

	return nil
}

// SubscribeEvent sets watch logs and creates a goroutine loop to handle event messages.
func (bm *BridgeManager) subscribeEvent(addr common.Address, bridge *bridgecontract.Bridge) error {
	tokenReceivedCh := make(chan *bridgecontract.BridgeRequestValueTransfer, TokenEventChanSize)
	tokenWithdrawCh := make(chan *bridgecontract.BridgeHandleValueTransfer, TokenEventChanSize)

	receivedSub, err := bridge.WatchRequestValueTransfer(nil, tokenReceivedCh)
	if err != nil {
		logger.Error("Failed to pBridge.WatchERC20Received", "err", err)
		return err
	}
	bm.receivedEvents[addr] = receivedSub
	withdrawnSub, err := bridge.WatchHandleValueTransfer(nil, tokenWithdrawCh)
	if err != nil {
		logger.Error("Failed to pBridge.WatchTokenWithdrawn", "err", err)
		receivedSub.Unsubscribe()
		delete(bm.receivedEvents, addr)
		return err
	}
	bm.withdrawEvents[addr] = withdrawnSub
	bridgeInfo, ok := bm.GetBridgeInfo(addr)
	if !ok {
		receivedSub.Unsubscribe()
		withdrawnSub.Unsubscribe()
		delete(bm.receivedEvents, addr)
		delete(bm.withdrawEvents, addr)
		return errors.New("fail to get bridge info")
	}
	bridgeInfo.subscribed = true

	go bm.loop(addr, tokenReceivedCh, tokenWithdrawCh, bm.scope.Track(receivedSub), bm.scope.Track(withdrawnSub))

	return nil
}

// UnsubscribeEvent cancels the contract's watch logs and initializes the status.
func (bm *BridgeManager) UnsubscribeEvent(addr common.Address) {
	receivedSub := bm.receivedEvents[addr]
	if receivedSub != nil {
		receivedSub.Unsubscribe()
		delete(bm.receivedEvents, addr)
	}

	withdrawSub := bm.withdrawEvents[addr]
	if withdrawSub != nil {
		withdrawSub.Unsubscribe()
		delete(bm.withdrawEvents, addr)
	}

	bridgeInfo, ok := bm.GetBridgeInfo(addr)
	if ok {
		bridgeInfo.subscribed = false
	}
}

// Loop handles subscribed event messages.
func (bm *BridgeManager) loop(
	addr common.Address,
	receivedCh <-chan *bridgecontract.BridgeRequestValueTransfer,
	withdrawCh <-chan *bridgecontract.BridgeHandleValueTransfer,
	receivedSub event.Subscription,
	withdrawSub event.Subscription) {

	defer receivedSub.Unsubscribe()
	defer withdrawSub.Unsubscribe()

	// TODO-Klaytn change goroutine logic for performance
	for {
		select {
		case ev := <-receivedCh:
			receiveEvent := TokenReceivedEvent{
				TokenType:    ev.Kind,
				ContractAddr: addr,
				TokenAddr:    ev.ContractAddress,
				From:         ev.From,
				To:           ev.To,
				Amount:       ev.Amount,
				RequestNonce: ev.RequestNonce,
				BlockNumber:  ev.Raw.BlockNumber,
			}
			bm.tokenReceived.Send(receiveEvent)
		case ev := <-withdrawCh:
			withdrawEvent := TokenTransferEvent{
				TokenType:    ev.Kind,
				ContractAddr: addr,
				TokenAddr:    ev.ContractAddress,
				Owner:        ev.Owner,
				Amount:       ev.Value,
				HandleNonce:  ev.HandleNonce,
			}
			bm.tokenWithdraw.Send(withdrawEvent)
		case err := <-receivedSub.Err():
			logger.Info("Contract Event Loop Running Stop by receivedSub.Err()", "err", err)
			return
		case err := <-withdrawSub.Err():
			logger.Info("Contract Event Loop Running Stop by withdrawSub.Err()", "err", err)
			return
		}
	}
}

// Stop closes a subscribed event scope of the bridge manager.
func (bm *BridgeManager) Stop() {
	for _, bi := range bm.bridges {
		close(bi.closed)
	}

	bm.scope.Close()
}
