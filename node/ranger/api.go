// Copyright 2018 The go-klaytn Authors
//
// This file is part of the go-klaytn library.
//
// The go-klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-klaytn library. If not, see <http://www.gnu.org/licenses/>.

package ranger

import "github.com/ground-x/go-gxplatform/common"

// PublicGXPAPI provides an API to access GXPlatform full node-related
// information.
type PublicRangerAPI struct {
	ranger *Ranger
}

// NewPublicGXPAPI creates a new GXP protocol API for full nodes.
func NewPublicRangerAPI(e *Ranger) *PublicRangerAPI {
	return &PublicRangerAPI{e}
}

func (pr *PublicRangerAPI) Accounts() []common.Address {
	addresses := make([]common.Address, 0) // return [] instead of nil if empty
	for _, wallet := range pr.ranger.accountManager.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// Coinbase is the address that mining rewards will be send to
func (pr *PublicRangerAPI) Coinbase() (common.Address, error) {
	return pr.ranger.Coinbase()
}
