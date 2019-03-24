// Copyright 2018 The klaytn Authors
// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package genesis

import (
	"github.com/ground-x/klaytn/cmd/homi/extra"
	"github.com/ground-x/klaytn/consensus/clique"
	"github.com/ground-x/klaytn/log"
	"github.com/ground-x/klaytn/params"
	"math/big"

	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/hexutil"
)

type Option func(*blockchain.Genesis)

var logger = log.NewModuleLogger(log.CMDIstanbul)

const (
	predefinedAddress = "0x00000000000000000400"
	predefinedCode    = "0x6080604052600436106101535763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630b81604581146101585780630d11f967146102255780631e482f8"
)

func Validators(addrs ...common.Address) Option {
	return func(genesis *blockchain.Genesis) {
		extraData, err := extra.Encode("0x00", addrs)
		if err != nil {
			logger.Error("Failed to encode extra data", "err", err)
			return
		}
		genesis.ExtraData = hexutil.MustDecode(extraData)
	}
}

func ValidatorsOfClique(signers ...common.Address) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.ExtraData = make([]byte, clique.ExtraVanity+len(signers)*common.AddressLength+clique.ExtraSeal)
		for i, signer := range signers {
			copy(genesis.ExtraData[32+i*common.AddressLength:], signer[:])
		}
	}
}

func GasLimit(limit uint64) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.GasLimit = limit
	}
}

func Alloc(addrs []common.Address, balance *big.Int) Option {
	return func(genesis *blockchain.Genesis) {
		// 어드레스 하나에도 많은 펀드를 추가한다. 이 계정은 key1, passwords.txt에 있음.
		// addrs = append(addrs, common.HexToAddress("f1112d590851764745499c855bd4a4574ffe9079"))
		alloc := make(map[common.Address]blockchain.GenesisAccount)
		for _, addr := range addrs {
			alloc[addr] = blockchain.GenesisAccount{Balance: balance}
		}

		alloc[common.HexToAddress(predefinedAddress)] = blockchain.GenesisAccount{
			Code:    common.FromHex(predefinedCode),
			Balance: big.NewInt(0),
		}
		genesis.Alloc = alloc
	}
}

func UnitPrice(price uint64) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.UnitPrice = price
	}
}

func Istanbul(config *params.IstanbulConfig) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.Istanbul = config
	}
}

func DeriveShaImpl(impl int) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.DeriveShaImpl = impl
	}
}

func Governance(config *params.GovernanceConfig) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.Governance = config
	}
}

func Clique(config *params.CliqueConfig) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.Clique = config
	}
}

func StakingInterval(interval uint64) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.StakingUpdateInterval = interval
	}
}

func ProposerInterval(interval uint64) Option {
	return func(genesis *blockchain.Genesis) {
		genesis.Config.ProposerUpdateInterval = interval
	}
}
