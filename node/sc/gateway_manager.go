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
}

// TokenWithdraw Event from SmartContract
type TokenTransferEvent struct {
	TokenType    uint8
	ContractAddr common.Address
	TokenAddr    common.Address
	Owner        common.Address
	Amount       *big.Int // Amount is UID in NFT
}

type GateWayJournal struct {
	Address common.Address `json:"address"`
	IsLocal bool           `json:"local"`
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

	all            map[common.Address]bool
	localGateWays  map[common.Address]*gatewaycontract.Gateway
	remoteGateWays map[common.Address]*gatewaycontract.Gateway

	tokenReceived event.Feed
	tokenWithdraw event.Feed

	scope event.SubscriptionScope

	journal *gatewayAddrJournal
}

func NewGateWayManager(main *SubBridge) (*GateWayManager, error) {

	gatewayAddrJournal := newGateWayAddrJournal(path.Join(main.config.DataDir, GatewayAddrJournal))

	gwm := &GateWayManager{
		subBridge:      main,
		all:            make(map[common.Address]bool),
		localGateWays:  make(map[common.Address]*gatewaycontract.Gateway),
		remoteGateWays: make(map[common.Address]*gatewaycontract.Gateway),
		journal:        gatewayAddrJournal,
	}

	logger.Info("Load Gateway Address from JournalFiles ", "path", gwm.journal.path)
	if err := gwm.journal.load(func(gwjournal GateWayJournal) error {
		logger.Info("Load Gateway Address from JournalFiles ", "Address", gwjournal.Address.Hex(), "Local", gwjournal.IsLocal)
		if gwjournal.IsLocal {
			return gwm.load(gwjournal.Address, main.localBackend, gwjournal.IsLocal)
		} else {
			return gwm.load(gwjournal.Address, main.remoteBackend, gwjournal.IsLocal)
		}
	}); err != nil {
		logger.Error("fail to load gateway address", "err", err)
	}
	if err := gwm.journal.rotate(gwm.GetAllGateway()); err != nil {
		logger.Error("fail to load gateway address", "err", err)
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

// GetGateway get gateway smartcontract with address
// addr is smartcontract address and local is same node or not
func (gwm *GateWayManager) GetGateway(addr common.Address) *gatewaycontract.Gateway {
	local, ok := gwm.all[addr]
	if !ok {
		return nil
	}
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
		gwm.all[addr] = true
	} else {
		gwm.remoteGateWays[addr] = gateway
		gwm.all[addr] = false
	}
	return nil
}

func (gwm *GateWayManager) IsLocal(addr common.Address) (bool, bool) {
	local, ok := gwm.all[addr]
	if !ok {
		return false, false
	}
	return local, true
}

// Deploy Gateway SmartContract on same node or remote node
func (gwm *GateWayManager) DeployGateway(backend bind.ContractBackend, local bool) (common.Address, error) {

	if local {
		addr, gateway, err := gwm.deployGateway(gwm.subBridge.getChainID(), big.NewInt((int64)(gwm.subBridge.handler.getNodeAccountNonce())), gwm.subBridge.handler.nodeKey, backend, gwm.subBridge.txPool.GasPrice())
		gwm.localGateWays[addr] = gateway
		gwm.all[addr] = true
		if err := gwm.journal.insert(addr, local); err != nil {
			logger.Error("fail to journal address", "err", err)
		}
		return addr, err
	} else {
		gwm.subBridge.handler.LockChainAccount()
		defer gwm.subBridge.handler.UnLockChainAccount()
		addr, gateway, err := gwm.deployGateway(gwm.subBridge.handler.parentChainID, big.NewInt((int64)(gwm.subBridge.handler.getChainAccountNonce())), gwm.subBridge.handler.chainKey, backend, new(big.Int).SetUint64(gwm.subBridge.handler.remoteGasPrice))
		if err != nil {
			logger.Error("fail to deploy gateway", "err", err)
			return common.Address{}, err
		}
		gwm.remoteGateWays[addr] = gateway
		gwm.all[addr] = false
		if err := gwm.journal.insert(addr, local); err != nil {
			logger.Error("fail to journal address", "err", err)
			return common.Address{}, err
		}
		gwm.subBridge.handler.addChainAccountNonce(1)
		return addr, err
	}
}

func (gwm *GateWayManager) deployGateway(chainID *big.Int, nonce *big.Int, accountKey *ecdsa.PrivateKey, backend bind.ContractBackend, gasPrice *big.Int) (common.Address, *gatewaycontract.Gateway, error) {

	// TODO-Klaytn change config

	if accountKey == nil {
		return common.Address{}, nil, errors.New("nil accountKey")
	}

	auth := MakeTransactOpts(accountKey, nonce, chainID, gasPrice)

	addr, tx, contract, err := gatewaycontract.DeployGateway(auth, backend, true)
	if err != nil {
		return common.Address{}, nil, err
	}
	logger.Info("Gateway is deploying on CurrentChain", "addr", addr, "txHash", tx.Hash().String())

	return addr, contract, nil
}

// SubscribeEvent registers a subscription of GatewayERC20Received and GatewayTokenWithdrawn
func (gwm *GateWayManager) SubscribeEvent(addr common.Address) error {

	local, ok := gwm.all[addr]
	if !ok {
		return fmt.Errorf("there is no gateway contract which address %v", addr)
	}
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

	tokenReceivedCh := make(chan *gatewaycontract.GatewayTokenReceived, TokenEventChanSize)
	tokenWithdrawCh := make(chan *gatewaycontract.GatewayTokenWithdrawn, TokenEventChanSize)

	receivedSub, err := gateway.WatchTokenReceived(nil, tokenReceivedCh)
	if err != nil {
		logger.Error("Failed to pGateway.WatchERC20Received", "err", err)
	}
	withdrawnSub, err := gateway.WatchTokenWithdrawn(nil, tokenWithdrawCh)
	if err != nil {
		logger.Error("Failed to pGateway.WatchTokenWithdrawn", "err", err)
	}

	go gwm.loop(addr, tokenReceivedCh, tokenWithdrawCh, gwm.scope.Track(receivedSub), gwm.scope.Track(withdrawnSub))
}

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
			}
			gwm.tokenReceived.Send(receiveEvent)
		case ev := <-withdrawCh:
			withdrawEvent := TokenTransferEvent{
				TokenType:    ev.Kind,
				ContractAddr: addr,
				TokenAddr:    ev.ContractAddress,
				Owner:        ev.Owner,
				Amount:       ev.Value,
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

func (gwm *GateWayManager) Stop() {
	gwm.scope.Close()
}
