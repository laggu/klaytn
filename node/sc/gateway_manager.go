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
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	gatewaycontract "github.com/ground-x/klaytn/contracts/gateway"
	"github.com/ground-x/klaytn/event"
	"github.com/ground-x/klaytn/ser/rlp"
	"io"
	"math/big"
	"path"
)

const (
	TokenEventChanSize = 30
)

// TokenReceived Event from SmartContract
type TokenReceivedEvent struct {
	ContractAddr common.Address
	TokenAddr    common.Address
	From         common.Address
	Amount       *big.Int
}

// TokenWithdraw Event from SmartContract
type TokenTransferEvent struct {
	ContractAddr common.Address
	TokenAddr    common.Address
	Owner        common.Address
	Amount       *big.Int
}

type GateWayJournal struct {
	Address common.Address
	IsLocal bool
}

// DecodeRLP decodes the Klaytn
func (b *GateWayJournal) DecodeRLP(s *rlp.Stream) error {
	var gatewayjournal struct {
		Address common.Address
		IsLocal bool
	}
	if err := s.Decode(&gatewayjournal); err != nil {
		return err
	}
	b.Address, b.IsLocal = gatewayjournal.Address, gatewayjournal.IsLocal
	return nil
}

// EncodeRLP serializes b into the Klaytn RLP GateWayJournal format.
func (b *GateWayJournal) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		b.Address,
		b.IsLocal,
	})
}

// GateWayManager manages GateWay SmartContracts
// for value transfer between service chain and parent chain
type GateWayManager struct {
	subBridge *SubBridge

	localGateWays  map[common.Address]*gatewaycontract.Gateway
	remoteGateWays map[common.Address]*gatewaycontract.Gateway

	erc20Received event.Feed
	erc20Withdraw event.Feed

	scope event.SubscriptionScope

	journal *gatewayAddrJournal
}

func NewGateWayManager(main *SubBridge) (*GateWayManager, error) {

	gatewayAddrJournal := newGateWayAddrJournal(path.Join(main.config.DataDir, "gateway_addrs.rlp"))

	gwm := &GateWayManager{
		subBridge:      main,
		localGateWays:  make(map[common.Address]*gatewaycontract.Gateway),
		remoteGateWays: make(map[common.Address]*gatewaycontract.Gateway),
		journal:        gatewayAddrJournal,
	}

	if err := gwm.journal.load(func(gwjournal GateWayJournal) error {
		//return gwm.load(gwjournal.Address ,nil, gwjournal.IsLocal)
		return nil
	}); err != nil {
		logger.Error("fail to load gateway address", "err", err)
	}
	if err := gwm.journal.rotate(gwm.GetAllGateway()); err != nil {
		logger.Error("fail to load gateway address", "err", err)
	}

	return gwm, nil
}

// SubscribeKRC20TokenReceived registers a subscription of TokenReceivedEvent.
func (gwm *GateWayManager) SubscribeKRC20TokenReceived(ch chan<- TokenReceivedEvent) event.Subscription {
	return gwm.scope.Track(gwm.erc20Received.Subscribe(ch))
}

// SubscribeKRC20WithDraw registers a subscription of TokenTransferEvent.
func (gwm *GateWayManager) SubscribeKRC20WithDraw(ch chan<- TokenTransferEvent) event.Subscription {
	return gwm.scope.Track(gwm.erc20Withdraw.Subscribe(ch))
}

// GetGateway get gateway smartcontract with address
// addr is smartcontract address and local is same node or not
func (gwm *GateWayManager) GetGateway(addr common.Address, local bool) *gatewaycontract.Gateway {
	if local {
		gateway, ok := gwm.localGateWays[addr]
		if !ok {
			return nil
		}
		return gateway
	} else {
		gateway, ok := gwm.remoteGateWays[addr]
		if !ok {
			return nil
		}
		return gateway
	}
}

func (gwm *GateWayManager) GetAllGateway() []*GateWayJournal {
	gwjs := []*GateWayJournal{}
	for addr := range gwm.localGateWays {
		gwjs = append(gwjs, &GateWayJournal{addr, true})
	}
	for addr := range gwm.remoteGateWays {
		gwjs = append(gwjs, &GateWayJournal{addr, false})
	}
	return gwjs
}

// Reorganize GateWayManager with smartcontract addresses
func (gwm *GateWayManager) Load(addrs []common.Address, backend bind.ContractBackend, local bool) error {
	for _, addr := range addrs {
		err := gwm.load(addr, backend, local)
		if err != nil {
			logger.Error("fail to load gateway contract ", "err", err)
		}
	}
	return nil
}

func (gwm *GateWayManager) load(addr common.Address, backend bind.ContractBackend, local bool) error {

	gateway, err := gatewaycontract.NewGateway(addr, backend)
	if err != nil {
		return err
	}
	if local {
		gwm.localGateWays[addr] = gateway
	} else {
		gwm.remoteGateWays[addr] = gateway
	}
	return nil
}

// Deploy Gateway SmartContract on same node or remote node
func (gwm *GateWayManager) DeployGateway(backend bind.ContractBackend, local bool) (common.Address, error) {

	if local {
		addr, gateway, err := gwm.deployGateway(gwm.subBridge.getChainID(), big.NewInt((int64)(gwm.subBridge.handler.getNodeAccountNonce())), gwm.subBridge.handler.nodeKey, backend)
		gwm.localGateWays[addr] = gateway
		if err := gwm.journal.insert(addr, local); err != nil {
			logger.Error("fail to journal address", "err", err)
		}
		return addr, err
	} else {
		gwm.subBridge.handler.LockChainAccount()
		defer gwm.subBridge.handler.UnLockChainAccount()

		addr, gateway, err := gwm.deployGateway(gwm.subBridge.handler.parentChainID, big.NewInt((int64)(gwm.subBridge.handler.getChainAccountNonce())), gwm.subBridge.handler.chainKey, backend)
		gwm.remoteGateWays[addr] = gateway
		if err := gwm.journal.insert(addr, local); err != nil {
			logger.Error("fail to journal address", "err", err)
		}
		return addr, err
	}
}

func (gwm *GateWayManager) deployGateway(chainID *big.Int, nonce *big.Int, accountKey *ecdsa.PrivateKey, backend bind.ContractBackend) (common.Address, *gatewaycontract.Gateway, error) {

	// TODO-Klaytn change config
	auth := bind.NewKeyedTransactor(accountKey)
	auth.GasLimit = 1000000
	auth.GasPrice = big.NewInt(0)
	auth.Nonce = nonce
	auth.Signer = func(signer types.Signer, addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
		return types.SignTx(tx, types.NewEIP155Signer(chainID), accountKey)
	}

	addr, tx, contract, err := gatewaycontract.DeployGateway(auth, backend, true)
	if err != nil {
		logger.Error("", "err", err)
		return common.Address{}, nil, err
	}
	logger.Info("Gateway is deploying on CurrentChain", "addr", addr, "txHash", tx.Hash().String())

	return addr, contract, nil
}

// SubscribeEvent registers a subscription of GatewayERC20Received and GatewayTokenWithdrawn
func (gwm *GateWayManager) SubscribeEvent(addr common.Address, local bool) error {

	if local {
		gateway, ok := gwm.localGateWays[addr]
		if !ok {
			return fmt.Errorf("there is no gateway contract which address %v", addr)
		}
		gwm.subscribeEvent(addr, gateway)
	} else {
		gateway, ok := gwm.remoteGateWays[addr]
		if !ok {
			return fmt.Errorf("there is no gateway contract which address %v", addr)
		}
		gwm.subscribeEvent(addr, gateway)
	}
	return nil
}

func (gwm *GateWayManager) subscribeEvent(addr common.Address, gateway *gatewaycontract.Gateway) {

	receivedCh := make(chan *gatewaycontract.GatewayERC20Received, TokenEventChanSize)
	tokenwithdrawCh := make(chan *gatewaycontract.GatewayTokenWithdrawn, TokenEventChanSize)

	receivedSub, err := gateway.WatchERC20Received(nil, receivedCh)
	if err != nil {
		logger.Error("Failed to pGateway.WatchERC20Received", "err", err)
	}
	withdrawnSub, err := gateway.WatchTokenWithdrawn(nil, tokenwithdrawCh)
	if err != nil {
		logger.Error("Failed to pGateway.WatchTokenWithdrawn", "err", err)
	}

	go gwm.loop(addr, receivedCh, tokenwithdrawCh, gwm.scope.Track(receivedSub), gwm.scope.Track(withdrawnSub))
}

func (gwm *GateWayManager) loop(addr common.Address, receivedCh <-chan *gatewaycontract.GatewayERC20Received,
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
				ContractAddr: addr,
				TokenAddr:    ev.ContractAddress,
				From:         ev.From,
				Amount:       ev.Amount,
			}
			gwm.erc20Received.Send(receiveEvent)
		case ev := <-withdrawCh:
			withdrawEvent := TokenTransferEvent{
				ContractAddr: addr,
				TokenAddr:    ev.ContractAddress,
				Owner:        ev.Owner,
				Amount:       ev.Value,
			}
			gwm.erc20Withdraw.Send(withdrawEvent)
		case err := <-receivedSub.Err():
			logger.Info("Contract Event Loop Running Stop", "err", err)
			return
		case err := <-withdrawSub.Err():
			logger.Info("Contract Event Loop Running Stop", "err", err)
			return
		}
	}
}

func (gwm *GateWayManager) Stop() {
	gwm.scope.Close()
}
