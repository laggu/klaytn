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
	"errors"
	"github.com/ground-x/klaytn/common"
)

// AddressManager manages mapping addresses for gateway,contract,user
// to exchange value between parent and child chain
type AddressManager struct {
	gatewayContracts map[common.Address]common.Address
	tokenContracts   map[common.Address]common.Address
}

func NewAddressManager() (*AddressManager, error) {
	return &AddressManager{
		gatewayContracts: make(map[common.Address]common.Address),
		tokenContracts:   make(map[common.Address]common.Address),
	}, nil
}

func (am *AddressManager) AddGateway(gateway1 common.Address, gateway2 common.Address) error {
	_, ok1 := am.gatewayContracts[gateway1]
	_, ok2 := am.gatewayContracts[gateway2]

	if ok1 || ok2 {
		return errors.New("gateway already exists")
	}

	am.gatewayContracts[gateway1] = gateway2
	am.gatewayContracts[gateway2] = gateway1
	return nil
}

func (am *AddressManager) DeleteGateway(gateway1 common.Address) (common.Address, common.Address, error) {
	gateway2, ok1 := am.gatewayContracts[gateway1]
	if !ok1 {
		return common.Address{}, common.Address{}, errors.New("gateway does not exist")
	}

	delete(am.gatewayContracts, gateway1)
	delete(am.gatewayContracts, gateway2)

	return gateway1, gateway2, nil
}

func (am *AddressManager) AddToken(token1 common.Address, token2 common.Address) error {
	_, ok1 := am.tokenContracts[token1]
	_, ok2 := am.tokenContracts[token2]

	if ok1 || ok2 {
		return errors.New("token already exists")
	}

	am.tokenContracts[token2] = token1
	am.tokenContracts[token1] = token2
	return nil
}

func (am *AddressManager) DeleteToken(token1 common.Address) (common.Address, common.Address, error) {
	token2, ok1 := am.tokenContracts[token1]
	if !ok1 {
		return common.Address{}, common.Address{}, errors.New("token does not exist")
	}

	delete(am.tokenContracts, token1)
	delete(am.tokenContracts, token2)

	return token1, token2, nil
}

func (am *AddressManager) GetCounterPartGateway(addr common.Address) common.Address {
	gateway, ok := am.gatewayContracts[addr]
	if !ok {
		return common.Address{}
	}
	return gateway
}

func (am *AddressManager) GetCounterPartToken(addr common.Address) common.Address {
	token, ok := am.tokenContracts[addr]
	if !ok {
		return common.Address{}
	}
	return token
}
