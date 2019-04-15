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
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/common"
	gatewaycontract "github.com/ground-x/klaytn/contracts/gateway"
	"github.com/ground-x/klaytn/event"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
	"math/big"
	"path"
	"time"
)

const (
	TokenEventChanSize = 30
	GatewayAddrJournal = "gateway_addrs.rlp"
)

const (
	KLAY uint8 = iota
	TOKEN
	NFT
)

// TokenReceived Event from SmartContract
type TokenReceivedEvent struct {
	TokenType    uint8
	ContractAddr common.Address
	TokenAddr    common.Address
	From         common.Address
	To           common.Address
	Amount       *big.Int // Amount is UID in NFT
	RequestNonce uint64
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

// GateWayJournal has two types. When a single address is inserted, the Paired is disabled.
// In this case, only one of the LocalAddress or RemoteAddress is filled with the address.
// If two address in a pair is inserted, the Pared is enabled.
type GateWayJournal struct {
	LocalAddress  common.Address `json:"localAddress"`
	RemoteAddress common.Address `json:"remoteAddress"`
	Paired        bool           `json:"paired"`
}

type GateWayInfo struct {
	gateway        *gatewaycontract.Gateway
	onServiceChain bool
	subscribed     bool
}

// DecodeRLP decodes the Klaytn
func (b *GateWayJournal) DecodeRLP(s *rlp.Stream) error {
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

// EncodeRLP serializes b into the Klaytn RLP GateWayJournal format.
func (b *GateWayJournal) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		b.LocalAddress,
		b.RemoteAddress,
		b.Paired,
	})
}

// GateWayManager manages GateWay SmartContracts
// for value transfer between service chain and parent chain
type GateWayManager struct {
	subBridge *SubBridge

	receivedEvents map[common.Address]event.Subscription
	withdrawEvents map[common.Address]event.Subscription
	gateways       map[common.Address]*GateWayInfo

	tokenReceived event.Feed
	tokenWithdraw event.Feed

	scope event.SubscriptionScope

	journal *gatewayAddrJournal
}

func NewGateWayManager(main *SubBridge) (*GateWayManager, error) {
	gatewayAddrJournal := newGateWayAddrJournal(path.Join(main.config.DataDir, GatewayAddrJournal), main.config)

	gwm := &GateWayManager{
		subBridge:      main,
		receivedEvents: make(map[common.Address]event.Subscription),
		withdrawEvents: make(map[common.Address]event.Subscription),
		gateways:       make(map[common.Address]*GateWayInfo),
		journal:        gatewayAddrJournal,
	}

	logger.Info("Load Gateway Address from JournalFiles ", "path", gwm.journal.path)
	gwm.journal.cache = []*GateWayJournal{}

	if err := gwm.journal.load(func(gwjournal GateWayJournal) error {
		logger.Info("Load Gateway Address from JournalFiles ",
			"local address", gwjournal.LocalAddress.Hex(), "remote address", gwjournal.RemoteAddress.Hex())
		gwm.journal.cache = append(gwm.journal.cache, &gwjournal)
		return nil
	}); err != nil {
		logger.Error("fail to load gateway address", "err", err)
	}

	if err := gwm.journal.rotate(gwm.GetAllGateway()); err != nil {
		logger.Error("fail to rotate gateway journal", "err", err)
	}

	return gwm, nil
}

// SubscribeTokenReceived registers a subscription of TokenReceivedEvent.
func (gwm *GateWayManager) SubscribeTokenReceived(ch chan<- TokenReceivedEvent) event.Subscription {
	return gwm.scope.Track(gwm.tokenReceived.Subscribe(ch))
}

// SubscribeTokenWithDraw registers a subscription of TokenTransferEvent.
func (gwm *GateWayManager) SubscribeTokenWithDraw(ch chan<- TokenTransferEvent) event.Subscription {
	return gwm.scope.Track(gwm.tokenWithdraw.Subscribe(ch))
}

// GetAllGateway returns a journal cache while removing unnecessary address pair.
func (gwm *GateWayManager) GetAllGateway() []*GateWayJournal {
	gwjs := []*GateWayJournal{}

	for _, journal := range gwm.journal.cache {
		if journal.Paired {
			gatewayInfo, ok := gwm.gateways[journal.LocalAddress]
			if ok && !gatewayInfo.subscribed {
				continue
			}
			if gwm.subBridge.AddressManager() != nil {
				gwm.subBridge.addressManager.DeleteGateway(journal.LocalAddress)
			}

			gatewayInfo, ok = gwm.gateways[journal.RemoteAddress]
			if ok && !gatewayInfo.subscribed {
				continue
			}
			if gwm.subBridge.AddressManager() != nil {
				gwm.subBridge.addressManager.DeleteGateway(journal.RemoteAddress)
			}
		}
		gwjs = append(gwjs, journal)
	}

	gwm.journal.cache = gwjs

	return gwm.journal.cache
}

// SetGateway stores the address and gateway pair with local/remote and subscription status.
func (gwm *GateWayManager) SetGateway(addr common.Address, gateway *gatewaycontract.Gateway, local bool, subscribed bool) {
	gwm.gateways[addr] = &GateWayInfo{gateway, local, subscribed}
}

// LoadAllGateway reloads gateway and handles subscription by using the the journal cache.
func (gwm *GateWayManager) LoadAllGateway() error {
	for _, journal := range gwm.journal.cache {
		if journal.Paired {
			if gwm.subBridge.AddressManager() == nil {
				return errors.New("address manager is not exist")
			}
			logger.Info("Add gateway pair in address manager")
			// Step 1: register gateway
			localGateway, err := gatewaycontract.NewGateway(journal.LocalAddress, gwm.subBridge.localBackend)
			if err != nil {
				return err
			}
			remoteGateway, err := gatewaycontract.NewGateway(journal.RemoteAddress, gwm.subBridge.remoteBackend)
			if err != nil {
				return err
			}
			gwm.SetGateway(journal.LocalAddress, localGateway, true, false)
			gwm.SetGateway(journal.RemoteAddress, remoteGateway, false, false)

			// Step 2: set address manager
			gwm.subBridge.AddressManager().AddGateway(journal.LocalAddress, journal.RemoteAddress)

			// Step 3: subscribe event
			gwm.subscribeEvent(journal.LocalAddress, localGateway)
			gwm.subscribeEvent(journal.RemoteAddress, remoteGateway)

		} else {
			err := gwm.loadGateway(journal.LocalAddress, gwm.subBridge.localBackend, true, false)
			if err != nil {
				return err
			}
			err = gwm.loadGateway(journal.RemoteAddress, gwm.subBridge.remoteBackend, false, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// LoadGateway creates new gateway contract for a given address and subscribes an event if needed.
func (gwm *GateWayManager) loadGateway(addr common.Address, backend bind.ContractBackend, local bool, subscribed bool) error {
	var gatewayInfo *GateWayInfo

	defer func() {
		if gatewayInfo != nil && subscribed && !gwm.gateways[addr].subscribed {
			logger.Info("gateway subscription is enabled by journal", "address", addr)
			gwm.subscribeEvent(addr, gatewayInfo.gateway)
		}
	}()

	gatewayInfo = gwm.gateways[addr]
	if gatewayInfo != nil {
		return nil
	}

	gateway, err := gatewaycontract.NewGateway(addr, backend)
	if err != nil {
		return err
	}
	logger.Info("gateway ", "address", addr)
	gwm.SetGateway(addr, gateway, local, false)
	gatewayInfo = gwm.gateways[addr]

	return nil
}

// Deploy Gateway SmartContract on same node or remote node
func (gwm *GateWayManager) DeployGateway(backend bind.ContractBackend, local bool) (common.Address, error) {
	if local {
		addr, gateway, err := gwm.deployGateway(gwm.subBridge.getChainID(), big.NewInt((int64)(gwm.subBridge.handler.getNodeAccountNonce())), gwm.subBridge.handler.nodeKey, backend, gwm.subBridge.txPool.GasPrice())
		if err != nil {
			logger.Error("fail to deploy gateway", "err", err)
			return common.Address{}, err
		}
		gwm.SetGateway(addr, gateway, local, false)
		gwm.journal.insert(addr, common.Address{}, false)

		return addr, err
	} else {
		gwm.subBridge.handler.LockChainAccount()
		defer gwm.subBridge.handler.UnLockChainAccount()
		addr, gateway, err := gwm.deployGateway(gwm.subBridge.handler.parentChainID, big.NewInt((int64)(gwm.subBridge.handler.getChainAccountNonce())), gwm.subBridge.handler.chainKey, backend, new(big.Int).SetUint64(gwm.subBridge.handler.remoteGasPrice))
		if err != nil {
			logger.Error("fail to deploy gateway", "err", err)
			return common.Address{}, err
		}
		gwm.SetGateway(addr, gateway, local, false)
		gwm.journal.insert(common.Address{}, addr, false)
		gwm.subBridge.handler.addChainAccountNonce(1)
		return addr, err
	}
}

// DeployGateway handles actual smart contract deployment.
// To create contract, the chain ID, nonce, account key, private key, contract binding and gas price are used.
// The deployed contract address, transaction are returned. An error is also returned if any.
func (gwm *GateWayManager) deployGateway(chainID *big.Int, nonce *big.Int, accountKey *ecdsa.PrivateKey, backend bind.ContractBackend, gasPrice *big.Int) (common.Address, *gatewaycontract.Gateway, error) {
	// TODO-Klaytn change config
	if accountKey == nil {
		// Only for unit test
		return common.Address{}, nil, errors.New("nil accountKey")
	}

	auth := MakeTransactOpts(accountKey, nonce, chainID, gasPrice)

	addr, tx, contract, err := gatewaycontract.DeployGateway(auth, backend, true)
	if err != nil {
		logger.Error("Failed to deploy contract.", "err", err)
		return common.Address{}, nil, err
	}
	logger.Info("Gateway is deploying...", "addr", addr, "txHash", tx.Hash().String())

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
	logger.Info("Gateway is deployed.", "addr", addr, "txHash", tx.Hash().String())
	return addr, contract, nil
}

// SubscribeEvent registers a subscription of GatewayERC20Received and GatewayTokenWithdrawn
func (gwm *GateWayManager) SubscribeEvent(addr common.Address) error {
	gatewayInfo, ok := gwm.gateways[addr]
	if !ok {
		return fmt.Errorf("there is no gateway contract which address %v", addr)
	}
	gwm.subscribeEvent(addr, gatewayInfo.gateway)

	return nil
}

// SubscribeEvent sets watch logs and creates a goroutine loop to handle event messages.
func (gwm *GateWayManager) subscribeEvent(addr common.Address, gateway *gatewaycontract.Gateway) {
	tokenReceivedCh := make(chan *gatewaycontract.GatewayTokenReceived, TokenEventChanSize)
	tokenWithdrawCh := make(chan *gatewaycontract.GatewayTokenWithdrawn, TokenEventChanSize)

	receivedSub, err := gateway.WatchTokenReceived(nil, tokenReceivedCh)
	if err != nil {
		logger.Error("Failed to pGateway.WatchERC20Received", "err", err)
	}
	gwm.receivedEvents[addr] = receivedSub
	withdrawnSub, err := gateway.WatchTokenWithdrawn(nil, tokenWithdrawCh)
	if err != nil {
		logger.Error("Failed to pGateway.WatchTokenWithdrawn", "err", err)
	}
	gwm.withdrawEvents[addr] = withdrawnSub
	gwm.gateways[addr].subscribed = true

	go gwm.loop(addr, tokenReceivedCh, tokenWithdrawCh, gwm.scope.Track(receivedSub), gwm.scope.Track(withdrawnSub))
}

// UnsubscribeEvent cancels the contract's watch logs and initializes the status.
func (gwm *GateWayManager) unsubscribeEvent(addr common.Address) {
	receivedSub := gwm.receivedEvents[addr]
	receivedSub.Unsubscribe()

	withdrawSub := gwm.withdrawEvents[addr]
	withdrawSub.Unsubscribe()

	gwm.gateways[addr].subscribed = false
}

// Loop handles subscribed event messages.
func (gwm *GateWayManager) loop(
	addr common.Address,
	receivedCh <-chan *gatewaycontract.GatewayTokenReceived,
	withdrawCh <-chan *gatewaycontract.GatewayTokenWithdrawn,
	receivedSub event.Subscription,
	withdrawSub event.Subscription) {

	defer receivedSub.Unsubscribe()
	defer withdrawSub.Unsubscribe()

	// TODO-klaytn change goroutine logic for performance
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
				RequestNonce: ev.ReqeustNonce,
			}
			gwm.tokenReceived.Send(receiveEvent)
		case ev := <-withdrawCh:
			withdrawEvent := TokenTransferEvent{
				TokenType:    ev.Kind,
				ContractAddr: addr,
				TokenAddr:    ev.ContractAddress,
				Owner:        ev.Owner,
				Amount:       ev.Value,
				HandleNonce:  ev.HandleNonce,
			}
			gwm.tokenWithdraw.Send(withdrawEvent)
		case err := <-receivedSub.Err():
			logger.Info("Contract Event Loop Running Stop by receivedSub.Err()", "err", err)
			return
		case err := <-withdrawSub.Err():
			logger.Info("Contract Event Loop Running Stop by withdrawSub.Err()", "err", err)
			return
		}
	}
}

// Stop closes a subscribed event scope of the gateway manager.
func (gwm *GateWayManager) Stop() {
	gwm.scope.Close()
}
