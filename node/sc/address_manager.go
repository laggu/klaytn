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

import "github.com/ground-x/klaytn/common"

// AddressManager manages mapping addresses for gateway,contract,user
// to exchange value between parent and child chain
type AddressManager struct {
	gatewayContracts map[common.Address]common.Address
	tokenContracts   map[common.Address]common.Address
	//TODO-Klaytn consider too many user mapping
	users map[common.Address]common.Address
}

func NewAddressManager() (*AddressManager, error) {
	return &AddressManager{
		gatewayContracts: make(map[common.Address]common.Address),
		tokenContracts:   make(map[common.Address]common.Address),
		users:            make(map[common.Address]common.Address),
	}, nil
}

func (am *AddressManager) AddGateway(gateway1 common.Address, gateway2 common.Address) {
	am.gatewayContracts[gateway1] = gateway2
	am.gatewayContracts[gateway2] = gateway1
}

func (am *AddressManager) DeleteGateway(addr common.Address) {
	gateway := am.GetCounterPartGateway(addr)
	if gateway != (common.Address{}) {
		delete(am.gatewayContracts, addr)
		delete(am.gatewayContracts, gateway)
	}
}

func (am *AddressManager) AddUser(user1 common.Address, user2 common.Address) {
	am.users[user1] = user2
	am.users[user2] = user1
}

func (am *AddressManager) DeleteUser(addr common.Address) {
	user := am.GetCounterPartUser(addr)
	if user != (common.Address{}) {
		delete(am.gatewayContracts, addr)
		delete(am.gatewayContracts, user)
	}
}

func (am *AddressManager) AddToken(token1 common.Address, token2 common.Address) {
	am.tokenContracts[token1] = token2
	am.tokenContracts[token2] = token1
}

func (am *AddressManager) DeleteToken(addr common.Address) {
	token := am.GetCounterPartToken(addr)
	if token != (common.Address{}) {
		delete(am.gatewayContracts, addr)
		delete(am.gatewayContracts, token)
	}
}

func (am *AddressManager) GetCounterPartGateway(addr common.Address) common.Address {
	gateway, ok := am.gatewayContracts[addr]
	if !ok {
		return common.Address{}
	}
	return gateway
}

func (am *AddressManager) GetCounterPartUser(addr common.Address) common.Address {
	user, ok := am.users[addr]
	if !ok {
		// if there is no specific counter part user, it can be its own address.
		return addr
	}
	return user
}

func (am *AddressManager) GetCounterPartToken(addr common.Address) common.Address {
	token, ok := am.tokenContracts[addr]
	if !ok {
		return common.Address{}
	}
	return token
}
