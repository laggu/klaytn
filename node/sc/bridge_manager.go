// Copyright 2019 The klaytn Authors
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
	"github.com/klaytn/klaytn/accounts/abi/bind"
	"github.com/klaytn/klaytn/blockchain/types"
	"github.com/klaytn/klaytn/common"
	bridgecontract "github.com/klaytn/klaytn/contracts/bridge"
	"github.com/klaytn/klaytn/event"
	"github.com/klaytn/klaytn/node/sc/bridgepool"
	"github.com/klaytn/klaytn/ser/rlp"
	"github.com/klaytn/klaytn/storage/database"
	"io"
	"math/big"
	"path"
	"sync"
	"time"
)

const (
	TokenEventChanSize  = 10000
	BridgeAddrJournal   = "bridge_addrs.rlp"
	maxPendingNonceDiff = 20000 // TODO-Klaytn-ServiceChain: update this limitation. Currently, 2 * 10000 TPS.
)

const (
	KLAY uint8 = iota
	ERC20
	ERC721
)

const (
	voteTypeValueTransfer = 0
	voteTypeConfiguration = 1
)

var (
	ErrInvalidTokenPair     = errors.New("invalid token pair")
	ErrNoBridgeInfo         = errors.New("bridge information does not exist")
	ErrDuplicatedBridgeInfo = errors.New("bridge information is duplicated")
	ErrDuplicatedToken      = errors.New("token is duplicated")
	ErrNoRecovery           = errors.New("recovery does not exist")
	ErrAlreadySubscribed    = errors.New("already subscribed")
	ErrBridgeRestore        = errors.New("restoring bridges is failed")
)

// RequestValueTransferEvent from Bridge contract
type RequestValueTransferEvent struct {
	*bridgecontract.BridgeRequestValueTransfer
}

func (rEv RequestValueTransferEvent) Nonce() uint64 {
	return rEv.RequestNonce
}

// HandleValueTransferEvent from Bridge contract
type HandleValueTransferEvent struct {
	*bridgecontract.BridgeHandleValueTransfer
}

type BridgeJournal struct {
	ChildAddress  common.Address `json:"childAddress"`
	ParentAddress common.Address `json:"parentAddress"`
	Subscribed    bool           `json:"subscribed"`
}

type BridgeInfo struct {
	bridgeDB database.DBManager

	address            common.Address
	counterpartAddress common.Address // TODO-Klaytn need to set counterpart
	account            *accountInfo
	bridge             *bridgecontract.Bridge
	counterpartBridge  *bridgecontract.Bridge
	onChildChain       bool
	subscribed         bool

	counterpartToken map[common.Address]common.Address

	pendingRequestEvent *bridgepool.ItemSortedMap // TODO-Klaytn Need to consider the nonce overflow(priority queue?) and the size overflow.
	nextHandleNonce     uint64                    // This nonce will be used for getting pending request value transfer events.

	isRunning                   bool
	handleNonce                 uint64 // the nonce from the handle value transfer event from the bridge.
	requestNonceFromCounterPart uint64 // the nonce from the request value transfer event from the counter part bridge.
	requestNonce                uint64 // the nonce from the request value transfer event from the counter part bridge.

	newEvent chan struct{}
	closed   chan struct{}
}

func NewBridgeInfo(db database.DBManager, addr common.Address, bridge *bridgecontract.Bridge, cpAddr common.Address, cpBridge *bridgecontract.Bridge, account *accountInfo, local, subscribed bool) (*BridgeInfo, error) {
	bi := &BridgeInfo{
		db,
		addr,
		cpAddr,
		account,
		bridge,
		cpBridge,
		local,
		subscribed,
		make(map[common.Address]common.Address),
		bridgepool.NewItemSortedMap(),
		0,
		true,
		0,
		0,
		0,
		make(chan struct{}),
		make(chan struct{}),
	}

	if err := bi.UpdateInfo(); err != nil {
		return bi, err
	}

	bi.nextHandleNonce = bi.handleNonce
	go bi.loop()

	return bi, nil
}

func (bi *BridgeInfo) loop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	logger.Info("start bridge loop", "addr", bi.address.String(), "onChildChain", bi.onChildChain)

	for {
		select {
		case <-bi.newEvent:
			bi.processingPendingRequestEvents()

		case <-ticker.C:
			bi.processingPendingRequestEvents()

		case <-bi.closed:
			logger.Info("stop bridge loop", "addr", bi.address.String(), "onChildChain", bi.onChildChain)
			return
		}
	}
}

func (bi *BridgeInfo) RegisterToken(token, counterpartToken common.Address) error {
	_, exist := bi.counterpartToken[token]
	if exist {
		return ErrDuplicatedToken
	}
	bi.counterpartToken[token] = counterpartToken
	return nil
}

func (bi *BridgeInfo) DeregisterToken(token, counterpartToken common.Address) error {
	_, exist := bi.counterpartToken[token]
	if !exist {
		return ErrInvalidTokenPair
	}
	delete(bi.counterpartToken, token)
	return nil
}

func (bi *BridgeInfo) GetCounterPartToken(token common.Address) common.Address {
	cpToken, exist := bi.counterpartToken[token]
	if !exist {
		return common.Address{}
	}
	return cpToken
}

func (bi *BridgeInfo) GetPendingRequestEvents(start uint64) []*RequestValueTransferEvent {
	ready := bi.pendingRequestEvent.Ready(start)
	var readyEvent []*RequestValueTransferEvent
	for _, item := range ready {
		readyEvent = append(readyEvent, item.(*RequestValueTransferEvent))
	}

	return readyEvent
}

// processingPendingRequestEvents handles pending request value transfer events of the bridge.
func (bi *BridgeInfo) processingPendingRequestEvents() error {
	logger.Trace("Get pending request value transfer event", "nextHandleNonce", bi.nextHandleNonce, "len(pendingEvent)", bi.pendingRequestEvent.Len())

	ReadyEvent := bi.GetReadyRequestValueTransferEvents()
	if ReadyEvent == nil {
		return nil
	}

	logger.Trace("Get ready request value transfer event", "len(readyEvent)", len(ReadyEvent), "nextHandleNonce", bi.nextHandleNonce, "len(pendingEvent)", bi.pendingRequestEvent.Len())

	diff := bi.requestNonceFromCounterPart - bi.handleNonce
	if diff > errorDiffRequestHandleNonce {
		logger.Error("Value transfer requested/handled nonce gap is too much.", "toSC", bi.onChildChain, "diff", diff, "requestedNonce", bi.requestNonceFromCounterPart, "handledNonce", bi.handleNonce)
		// TODO-Klaytn need to consider starting value transfer recovery.
	}

	for idx, ev := range ReadyEvent {
		if ev.RequestNonce < bi.handleNonce {
			logger.Trace("past requests are ignored", "RequestNonce", ev.RequestNonce, "handleNonce", bi.handleNonce)
			continue
		}

		logger.Trace("handle value transfer event", "evt", ev)

		if ev.RequestNonce > bi.handleNonce+maxPendingNonceDiff {
			logger.Trace("nonce diff is too large", "limitation", maxPendingNonceDiff)
			return errors.New("nonce diff is too large")
		}

		if err := bi.handleRequestValueTransferEvent(ev); err != nil {
			bi.AddRequestValueTransferEvents(ReadyEvent[idx:])
			logger.Error("Failed handle request value transfer event", "err", err, "len(RePutEvent)", len(ReadyEvent[idx:]))
			return err
		}
		if bi.nextHandleNonce <= ev.RequestNonce {
			bi.nextHandleNonce = ev.RequestNonce + 1
		}
	}

	return nil
}

func (bi *BridgeInfo) UpdateInfo() error {
	if bi == nil {
		return ErrNoBridgeInfo
	}

	rn, err := bi.bridge.RequestNonce(nil)
	if err != nil {
		return err
	}
	bi.UpdateRequestNonce(rn)

	hn, err := bi.bridge.LowerHandleNonce(nil)
	if err != nil {
		return err
	}
	bi.UpdateHandledNonce(hn)
	bi.UpdateRequestNonceFromCounterpart(hn)

	isRunning, err := bi.bridge.IsRunning(nil)
	if err != nil {
		return err
	}
	bi.isRunning = isRunning

	return nil
}

// handleRequestValueTransferEvent handles the given request value transfer event.
func (bi *BridgeInfo) handleRequestValueTransferEvent(ev *RequestValueTransferEvent) error {
	tokenType := ev.TokenType
	tokenAddr := bi.GetCounterPartToken(ev.TokenAddress)
	// TODO-Klaytn-Servicechain Add counterpart token address in requestValueTransferEvent
	if tokenType != KLAY && tokenAddr == (common.Address{}) {
		logger.Warn("Unregistered counter part token address.", "addr", tokenAddr.Hex())
		ctTokenAddr, err := bi.counterpartBridge.AllowedTokens(nil, ev.TokenAddress)
		if err != nil {
			return err
		}
		if ctTokenAddr == (common.Address{}) {
			return errors.New("can't get counterpart token from bridge")
		}

		if err := bi.RegisterToken(ev.TokenAddress, ctTokenAddr); err != nil {
			return err
		}
		tokenAddr = ctTokenAddr
		logger.Info("Register counter part token address.", "addr", tokenAddr.Hex(), "cpAddr", ctTokenAddr.Hex())
	}

	bridgeAcc := bi.account

	bridgeAcc.Lock()
	defer bridgeAcc.UnLock()

	auth := bridgeAcc.GetTransactOpts()

	var handleTx *types.Transaction
	var err error

	switch tokenType {
	case KLAY:
		handleTx, err = bi.bridge.HandleKLAYTransfer(auth, ev.Raw.TxHash, ev.From, ev.To, ev.ValueOrTokenId, ev.RequestNonce, ev.Raw.BlockNumber, ev.ExtraData)
		if err != nil {
			return err
		}
		logger.Trace("Bridge succeeded to HandleKLAYTransfer", "nonce", ev.RequestNonce, "tx", handleTx.Hash().String())

	case ERC20:
		handleTx, err = bi.bridge.HandleERC20Transfer(auth, ev.Raw.TxHash, ev.From, ev.To, tokenAddr, ev.ValueOrTokenId, ev.RequestNonce, ev.Raw.BlockNumber, ev.ExtraData)
		if err != nil {
			return err
		}
		logger.Trace("Bridge succeeded to HandleERC20Transfer", "nonce", ev.RequestNonce, "tx", handleTx.Hash().String())
	case ERC721:
		handleTx, err = bi.bridge.HandleERC721Transfer(auth, ev.Raw.TxHash, ev.From, ev.To, tokenAddr, ev.ValueOrTokenId, ev.RequestNonce, ev.Raw.BlockNumber, ev.Uri, ev.ExtraData)
		if err != nil {
			return err
		}
		logger.Trace("Bridge succeeded to HandleERC721Transfer", "nonce", ev.RequestNonce, "tx", handleTx.Hash().String())
	default:
		logger.Warn("Got Unknown Token Type ReceivedEvent", "bridge", ev.Raw.Address, "nonce", ev.RequestNonce, "from", ev.From)
		return nil
	}

	bridgeAcc.IncNonce()

	bi.bridgeDB.WriteHandleTxHashFromRequestTxHash(ev.Raw.TxHash, handleTx.Hash())
	return nil
}

// UpdateRequestNonceFromCounterpart updates the request nonce from counterpart bridge.
func (bi *BridgeInfo) UpdateRequestNonceFromCounterpart(nonce uint64) {
	if bi.requestNonceFromCounterPart < nonce {
		vtRequestNonceCount.Inc(int64(nonce - bi.requestNonceFromCounterPart))
		bi.requestNonceFromCounterPart = nonce
	}
}

// UpdateRequestNonce updates the request nonce of the bridge.
func (bi *BridgeInfo) UpdateRequestNonce(nonce uint64) {
	if bi.requestNonce < nonce {
		bi.requestNonce = nonce
	}
}

// UpdateHandledNonce updates the handled nonce with new nonce.
func (bi *BridgeInfo) UpdateHandledNonce(nonce uint64) {
	if bi.handleNonce < nonce {
		vtHandleNonceCount.Inc(int64(nonce - bi.handleNonce))
		bi.handleNonce = nonce
	}
}

// AddRequestValueTransferEvents adds events into the pendingRequestEvent.
func (bi *BridgeInfo) AddRequestValueTransferEvents(evs []*RequestValueTransferEvent) {
	// TODO-Klaytn Need to consider the nonce overflow(priority queue?) and the size overflow.
	// - If the size is full and received event has the omitted nonce, it can be allowed.
	for _, ev := range evs {
		bi.UpdateRequestNonceFromCounterpart(ev.RequestNonce + 1)
		bi.pendingRequestEvent.Put(ev)
	}
	logger.Trace("added pending request events to the bridge info:", "bi.pendingRequestEvent", bi.pendingRequestEvent.Len())

	vtPendingRequestEventMeter.Inc(int64(len(evs)))

	select {
	case bi.newEvent <- struct{}{}:
	default:
	}
}

// GetReadyRequestValueTransferEvents returns the processable events with the increasing nonce.
func (bi *BridgeInfo) GetReadyRequestValueTransferEvents() []*RequestValueTransferEvent {
	return bi.GetPendingRequestEvents(bi.nextHandleNonce)
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
	b.ChildAddress, b.ParentAddress, b.Subscribed = elem.LocalAddress, elem.RemoteAddress, elem.Paired
	return nil
}

// EncodeRLP serializes a BridgeJournal into the Klaytn RLP BridgeJournal format.
func (b *BridgeJournal) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		b.ChildAddress,
		b.ParentAddress,
		b.Subscribed,
	})
}

// BridgeManager manages Bridge SmartContracts
// for value transfer between child chain and parent chain
type BridgeManager struct {
	subBridge *SubBridge

	receivedEvents map[common.Address]event.Subscription
	withdrawEvents map[common.Address]event.Subscription
	bridges        map[common.Address]*BridgeInfo
	mu             sync.RWMutex

	requestEventFeeder event.Feed
	handleEventFeeder  event.Feed

	scope event.SubscriptionScope

	journal    *bridgeAddrJournal
	recoveries map[common.Address]*valueTransferRecovery
	auth       *bind.TransactOpts
}

func NewBridgeManager(main *SubBridge) (*BridgeManager, error) {
	bridgeAddrJournal := newBridgeAddrJournal(path.Join(main.config.DataDir, BridgeAddrJournal))

	bridgeManager := &BridgeManager{
		subBridge:      main,
		receivedEvents: make(map[common.Address]event.Subscription),
		withdrawEvents: make(map[common.Address]event.Subscription),
		bridges:        make(map[common.Address]*BridgeInfo),
		journal:        bridgeAddrJournal,
		recoveries:     make(map[common.Address]*valueTransferRecovery),
	}

	logger.Info("Load Bridge Address from JournalFiles ", "path", bridgeManager.journal.path)
	bridgeManager.journal.cache = make(map[common.Address]*BridgeJournal)

	if err := bridgeManager.journal.load(func(gwjournal BridgeJournal) error {
		logger.Info("Load Bridge Address from JournalFiles ",
			"local address", gwjournal.ChildAddress.Hex(), "remote address", gwjournal.ParentAddress.Hex())
		bridgeManager.journal.cache[gwjournal.ChildAddress] = &gwjournal
		return nil
	}); err != nil {
		logger.Error("fail to load bridge address", "err", err)
	}

	if err := bridgeManager.journal.rotate(bridgeManager.GetAllBridge()); err != nil {
		logger.Error("fail to rotate bridge journal", "err", err)
	}

	return bridgeManager, nil
}

func (bm *BridgeManager) IsValidBridgePair(bridge1, bridge2 common.Address) bool {
	b1, ok1 := bm.GetBridgeInfo(bridge1)
	b2, ok2 := bm.GetBridgeInfo(bridge2)

	if ok1 && ok2 {
		if bridge1 == b2.counterpartAddress && bridge2 == b1.counterpartAddress {
			return true
		}
	}

	return false
}

func (bm *BridgeManager) GetCounterPartBridgeAddr(bridgeAddr common.Address) common.Address {
	bridge, ok := bm.GetBridgeInfo(bridgeAddr)
	if ok {
		return bridge.counterpartAddress
	}
	return common.Address{}
}

func (bm *BridgeManager) GetCounterPartBridge(bridgeAddr common.Address) *bridgecontract.Bridge {
	bridge, ok := bm.GetBridgeInfo(bridgeAddr)
	if ok {
		return bridge.counterpartBridge
	}
	return nil
}

// LogBridgeStatus logs the bridge contract requested/handled nonce status as an information.
func (bm *BridgeManager) LogBridgeStatus() {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if len(bm.bridges) == 0 {
		return
	}

	p2cTotalRequestNonce, p2cTotalHandleNonce := uint64(0), uint64(0)
	c2pTotalRequestNonce, c2pTotalHandleNonce := uint64(0), uint64(0)

	for bAddr, b := range bm.bridges {
		diffNonce := b.requestNonceFromCounterPart - b.handleNonce

		if b.subscribed {
			var headStr string
			if b.onChildChain {
				headStr = "Bridge(Parent -> Child Chain)"
				p2cTotalRequestNonce += b.requestNonceFromCounterPart
				p2cTotalHandleNonce += b.handleNonce
			} else {
				headStr = "Bridge(Child -> Parent Chain)"
				c2pTotalRequestNonce += b.requestNonceFromCounterPart
				c2pTotalHandleNonce += b.handleNonce
			}
			logger.Debug(headStr, "bridge", bAddr.String(), "requestNonce", b.requestNonceFromCounterPart, "handleNonce", b.handleNonce, "pending", diffNonce)
		}
	}

	logger.Info("VT : Parent -> Child Chain", "request", p2cTotalRequestNonce, "handle", p2cTotalHandleNonce, "pending", p2cTotalRequestNonce-p2cTotalHandleNonce)
	logger.Info("VT : Child -> Parent Chain", "request", c2pTotalRequestNonce, "handle", c2pTotalHandleNonce, "pending", c2pTotalRequestNonce-c2pTotalHandleNonce)
}

// SubscribeRequestEvent registers a subscription of RequestValueTransferEvent.
func (bm *BridgeManager) SubscribeRequestEvent(ch chan<- *RequestValueTransferEvent) event.Subscription {
	return bm.scope.Track(bm.requestEventFeeder.Subscribe(ch))
}

// SubscribeHandleEvent registers a subscription of RequestValueTransferEvent.
func (bm *BridgeManager) SubscribeHandleEvent(ch chan<- *HandleValueTransferEvent) event.Subscription {
	return bm.scope.Track(bm.handleEventFeeder.Subscribe(ch))
}

// GetAllBridge returns a slice of journal cache.
func (bm *BridgeManager) GetAllBridge() []*BridgeJournal {
	var gwjs []*BridgeJournal

	for _, journal := range bm.journal.cache {
		gwjs = append(gwjs, journal)
	}
	return gwjs
}

// GetBridge returns bridge contract of the specified address.
func (bm *BridgeManager) GetBridgeInfo(addr common.Address) (*BridgeInfo, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bridge, ok := bm.bridges[addr]
	return bridge, ok
}

// DeleteBridgeInfo deletes the bridge info of the specified address.
func (bm *BridgeManager) DeleteBridgeInfo(addr common.Address) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bi := bm.bridges[addr]
	if bi == nil {
		return ErrNoBridgeInfo
	}

	close(bi.closed)

	delete(bm.bridges, addr)
	return nil
}

// SetBridgeInfo stores the address and bridge pair with local/remote and subscription status.
func (bm *BridgeManager) SetBridgeInfo(addr common.Address, bridge *bridgecontract.Bridge, cpAddr common.Address, cpBridge *bridgecontract.Bridge, account *accountInfo, local bool, subscribed bool) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.bridges[addr] != nil {
		return ErrDuplicatedBridgeInfo
	}
	var err error
	bm.bridges[addr], err = NewBridgeInfo(bm.subBridge.chainDB, addr, bridge, cpAddr, cpBridge, account, local, subscribed)
	return err
}

// RestoreBridges setups bridge subscription by using the journal cache.
func (bm *BridgeManager) RestoreBridges() error {
	if bm.subBridge.peers.Len() < 1 {
		logger.Error("check peer connections to restore bridges")
		return ErrBridgeRestore
	}

	var counter = 0
	bm.stopAllRecoveries()

	for _, journal := range bm.journal.cache {
		cBridgeAddr := journal.ChildAddress
		pBridgeAddr := journal.ParentAddress
		bacc := bm.subBridge.bridgeAccounts

		// Set bridge info
		cBridgeInfo, cOk := bm.GetBridgeInfo(cBridgeAddr)
		pBridgeInfo, pOk := bm.GetBridgeInfo(pBridgeAddr)

		cBridge, err := bridgecontract.NewBridge(cBridgeAddr, bm.subBridge.localBackend)
		if err != nil {
			logger.Error("local bridge creation is failed", "err", err, "bridge", cBridge)
			break
		}

		pBridge, err := bridgecontract.NewBridge(pBridgeAddr, bm.subBridge.remoteBackend)
		if err != nil {
			logger.Error("remote bridge creation is failed", "err", err, "bridge", pBridge)
			break
		}

		if !cOk {
			err = bm.SetBridgeInfo(cBridgeAddr, cBridge, pBridgeAddr, pBridge, bacc.cAccount, true, false)
			if err != nil {
				logger.Error("setting local bridge info is failed", "err", err)
				bm.DeleteBridgeInfo(cBridgeAddr)
				break
			}
			cBridgeInfo, _ = bm.GetBridgeInfo(cBridgeAddr)
		}

		if !pOk {
			err = bm.SetBridgeInfo(pBridgeAddr, pBridge, cBridgeAddr, cBridgeInfo.bridge, bacc.pAccount, false, false)
			if err != nil {
				logger.Error("setting remote bridge info is failed", "err", err)
				bm.DeleteBridgeInfo(pBridgeAddr)
				break
			}
			pBridgeInfo, _ = bm.GetBridgeInfo(pBridgeAddr)
		}

		// Subscribe bridge events
		if journal.Subscribed {
			bm.UnsubscribeEvent(cBridgeAddr)
			bm.UnsubscribeEvent(pBridgeAddr)

			if !cBridgeInfo.subscribed {
				logger.Info("automatic local bridge subscription", "info", cBridgeInfo, "address", cBridgeInfo.address.String())
				if err := bm.subscribeEvent(cBridgeAddr, cBridgeInfo.bridge); err != nil {
					logger.Error("local bridge subscription is failed", "err", err)
					break
				}
			}
			if !pBridgeInfo.subscribed {
				logger.Info("automatic remote bridge subscription", "info", pBridgeInfo, "address", pBridgeInfo.address.String())
				if err := bm.subscribeEvent(pBridgeAddr, pBridgeInfo.bridge); err != nil {
					logger.Error("remote bridge subscription is failed", "err", err)
					bm.DeleteBridgeInfo(pBridgeAddr)
					break
				}
			}
			recovery := bm.recoveries[cBridgeAddr]
			if recovery == nil {
				bm.AddRecovery(cBridgeAddr, pBridgeAddr)
			}
		}

		counter++
	}

	if len(bm.journal.cache) == counter {
		logger.Info("succeeded to restore bridges", "pairs", counter)
		return nil
	}
	return ErrBridgeRestore
}

// SetJournal inserts or updates journal for a given addresses pair.
func (bm *BridgeManager) SetJournal(localAddress, remoteAddress common.Address) error {
	err := bm.journal.insert(localAddress, remoteAddress)
	return err
}

// AddRecovery starts value transfer recovery for a given addresses pair.
func (bm *BridgeManager) AddRecovery(localAddress, remoteAddress common.Address) error {
	if !bm.subBridge.config.VTRecovery {
		logger.Info("value transfer recovery is disabled")
		return nil
	}

	// Check if bridge information is exist.
	localBridgeInfo, ok := bm.GetBridgeInfo(localAddress)
	if !ok {
		return ErrNoBridgeInfo
	}
	remoteBridgeInfo, ok := bm.GetBridgeInfo(remoteAddress)
	if !ok {
		return ErrNoBridgeInfo
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
		return ErrNoRecovery
	}
	recovery.Stop()
	delete(bm.recoveries, localAddress)

	return nil
}

// stopAllRecoveries stops the internal value transfer recoveries.
func (bm *BridgeManager) stopAllRecoveries() {
	for _, recovery := range bm.recoveries {
		recovery.Stop()
	}
	bm.recoveries = make(map[common.Address]*valueTransferRecovery)
}

// Deploy Bridge SmartContract on same node or remote node
func (bm *BridgeManager) DeployBridge(backend bind.ContractBackend, local bool) (*bridgecontract.Bridge, common.Address, error) {
	var acc *accountInfo
	var modeMintBurn bool

	if local {
		acc = bm.subBridge.bridgeAccounts.cAccount
		modeMintBurn = true
	} else {
		acc = bm.subBridge.bridgeAccounts.pAccount
		modeMintBurn = false
	}

	addr, bridge, err := bm.deployBridge(acc, backend, modeMintBurn)
	if err != nil {
		return nil, common.Address{}, err
	}

	return bridge, addr, err
}

// DeployBridge handles actual smart contract deployment.
// To create contract, the chain ID, nonce, account key, private key, contract binding and gas price are used.
// The deployed contract address, transaction are returned. An error is also returned if any.
func (bm *BridgeManager) deployBridge(acc *accountInfo, backend bind.ContractBackend, modeMintBurn bool) (common.Address, *bridgecontract.Bridge, error) {
	acc.Lock()
	auth := acc.GetTransactOpts()
	addr, tx, contract, err := bridgecontract.DeployBridge(auth, backend, modeMintBurn)
	if err != nil {
		logger.Error("Failed to deploy contract.", "err", err)
		acc.UnLock()
		return common.Address{}, nil, err
	}
	acc.IncNonce()
	acc.UnLock()

	logger.Info("Bridge is deploying...", "addr", addr, "txHash", tx.Hash().String())

	back, ok := backend.(bind.DeployBackend)
	if !ok {
		logger.Warn("DeployBacked type assertion is failed. Skip WaitDeployed.")
		return addr, contract, nil
	}

	timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), 60*time.Second)
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
		return ErrNoBridgeInfo
	}
	if bridgeInfo.subscribed {
		return ErrAlreadySubscribed
	}
	err := bm.subscribeEvent(addr, bridgeInfo.bridge)
	if err != nil {
		return err
	}

	return nil
}

// resetAllSubscribedEvents resets watch logs and recreates a goroutine loop to handle event messages.
func (bm *BridgeManager) ResetAllSubscribedEvents() error {
	logger.Info("ResetAllSubscribedEvents is called.")

	for _, journal := range bm.journal.cache {
		if journal.Subscribed {
			bm.UnsubscribeEvent(journal.ChildAddress)
			bm.UnsubscribeEvent(journal.ParentAddress)

			bridgeInfo, ok := bm.GetBridgeInfo(journal.ChildAddress)
			if !ok {
				logger.Error("ResetAllSubscribedEvents failed to GetBridgeInfo", "localBridge", journal.ChildAddress.String())
				return ErrNoBridgeInfo
			}
			err := bm.subscribeEvent(journal.ChildAddress, bridgeInfo.bridge)
			if err != nil {
				return err
			}

			bridgeInfo, ok = bm.GetBridgeInfo(journal.ParentAddress)
			if !ok {
				logger.Error("ResetAllSubscribedEvents failed to GetBridgeInfo", "remoteBridge", journal.ParentAddress.String())
				bm.UnsubscribeEvent(journal.ChildAddress)
				return ErrNoBridgeInfo
			}
			err = bm.subscribeEvent(journal.ParentAddress, bridgeInfo.bridge)
			if err != nil {
				return err
			}
		}
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
		return ErrNoBridgeInfo
	}
	bridgeInfo.subscribed = true

	go bm.loop(addr, tokenReceivedCh, tokenWithdrawCh, receivedSub, withdrawnSub)

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
	requestEventCh <-chan *bridgecontract.BridgeRequestValueTransfer,
	handleEventCh <-chan *bridgecontract.BridgeHandleValueTransfer,
	requestEventSub event.Subscription,
	handleEventSub event.Subscription) {

	defer requestEventSub.Unsubscribe()
	defer handleEventSub.Unsubscribe()

	bi, ok := bm.GetBridgeInfo(addr)
	if !ok {
		logger.Error("bridge information is missing")
		return
	}

	// TODO-Klaytn change goroutine logic for performance
	for {
		select {
		case <-bi.closed:
			return
		case ev := <-requestEventCh:
			bm.requestEventFeeder.Send(&RequestValueTransferEvent{ev})
		case ev := <-handleEventCh:
			bm.handleEventFeeder.Send(&HandleValueTransferEvent{ev})
		case err := <-requestEventSub.Err():
			logger.Info("Contract Event Loop Running Stop by receivedSub.Err()", "err", err)
			return
		case err := <-handleEventSub.Err():
			logger.Info("Contract Event Loop Running Stop by withdrawSub.Err()", "err", err)
			return
		}
	}
}

// Stop closes a subscribed event scope of the bridge manager.
func (bm *BridgeManager) Stop() {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	for _, bi := range bm.bridges {
		close(bi.closed)
	}

	bm.scope.Close()
}

// SetERC20Fee set the ERC20 transfer fee on the bridge contract.
func (bm *BridgeManager) SetERC20Fee(bridgeAddr, tokenAddr common.Address, fee *big.Int) (common.Hash, error) {
	bi, ok := bm.GetBridgeInfo(bridgeAddr)
	if !ok {
		return common.Hash{}, ErrNoBridgeInfo
	}

	auth := bi.account
	auth.Lock()
	defer auth.UnLock()

	rn, err := bi.bridge.ConfigurationNonce(nil)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := bi.bridge.SetERC20Fee(auth.GetTransactOpts(), tokenAddr, fee, rn)
	if err != nil {
		return common.Hash{}, err
	}

	auth.IncNonce()

	return tx.Hash(), nil
}

// SetKLAYFee set the KLAY transfer fee on the bridge contract.
func (bm *BridgeManager) SetKLAYFee(bridgeAddr common.Address, fee *big.Int) (common.Hash, error) {
	bi, ok := bm.GetBridgeInfo(bridgeAddr)
	if !ok {
		return common.Hash{}, ErrNoBridgeInfo
	}

	auth := bi.account
	auth.Lock()
	defer auth.UnLock()

	rn, err := bi.bridge.ConfigurationNonce(nil)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := bi.bridge.SetKLAYFee(auth.GetTransactOpts(), fee, rn)
	if err != nil {
		return common.Hash{}, err
	}

	auth.IncNonce()

	return tx.Hash(), nil
}

// SetFeeReceiver set the fee receiver which can get the fee of value transfer request.
func (bm *BridgeManager) SetFeeReceiver(bridgeAddr, receiver common.Address) (common.Hash, error) {
	bi, ok := bm.GetBridgeInfo(bridgeAddr)
	if !ok {
		return common.Hash{}, ErrNoBridgeInfo
	}

	auth := bi.account
	auth.Lock()
	defer auth.UnLock()

	tx, err := bi.bridge.SetFeeReceiver(auth.GetTransactOpts(), receiver)
	if err != nil {
		return common.Hash{}, err
	}

	auth.IncNonce()

	return tx.Hash(), nil
}

// GetERC20Fee returns the ERC20 transfer fee on the bridge contract.
func (bm *BridgeManager) GetERC20Fee(bridgeAddr, tokenAddr common.Address) (*big.Int, error) {
	bi, ok := bm.GetBridgeInfo(bridgeAddr)
	if !ok {
		return nil, ErrNoBridgeInfo
	}

	return bi.bridge.FeeOfERC20(nil, tokenAddr)
}

// GetKLAYFee returns the KLAY transfer fee on the bridge contract.
func (bm *BridgeManager) GetKLAYFee(bridgeAddr common.Address) (*big.Int, error) {
	bi, ok := bm.GetBridgeInfo(bridgeAddr)
	if !ok {
		return nil, ErrNoBridgeInfo
	}

	return bi.bridge.FeeOfKLAY(nil)
}

// GetFeeReceiver returns the receiver which can get fee of value transfer request.
func (bm *BridgeManager) GetFeeReceiver(bridgeAddr common.Address) (common.Address, error) {
	bi, ok := bm.GetBridgeInfo(bridgeAddr)
	if !ok {
		return common.Address{}, ErrNoBridgeInfo
	}

	return bi.bridge.FeeReceiver(nil)
}
