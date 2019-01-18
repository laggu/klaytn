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

//go:generate abigen --sol contract/KlaytnReward.sol --pkg contract --out contract/KlaytnReward.go
//go:generate abigen --sol contract/CommitteeReward.sol --pkg contract --out contract/CommitteeReward.go
//go:generate abigen --sol contract/RNReward.sol --pkg contract --out contract/RNReward.go
//go:generate abigen --sol contract/ProposerReward.sol --pkg contract --out contract/ProposerReward.go
//go:generate abigen --sol contract/PIRerve.sol --pkg contract --out contract/PIRerve.go

package reward

import (
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/contracts/reward/contract"
	"github.com/ground-x/klaytn/params"
	"math/big"
)

type Reward struct {
	*contract.KlaytnRewardSession
	contractBackend bind.ContractBackend
}

func NewReward(transactOpts *bind.TransactOpts, contractAddr common.Address, contractBackend bind.ContractBackend) (*Reward, error) {
	klaytnReward, err := contract.NewKlaytnReward(contractAddr, contractBackend)
	if err != nil {
		return nil, err
	}

	return &Reward{
		&contract.KlaytnRewardSession{
			Contract:     klaytnReward,
			TransactOpts: *transactOpts,
		},
		contractBackend,
	}, nil
}

func DeployReward(transactOpts *bind.TransactOpts, contractBackend bind.ContractBackend) (common.Address, *Reward, error) {

	rewardAddr, _, _, err := contract.DeployKlaytnReward(transactOpts, contractBackend)
	if err != nil {
		return rewardAddr, nil, err
	}

	reward, err := NewReward(transactOpts, rewardAddr, contractBackend)
	if err != nil {
		return rewardAddr, nil, err
	}

	return rewardAddr, reward, nil
}

type BalanceAdder interface {
	AddBalance(addr common.Address, v *big.Int)
}

// MintKLAY mints KLAY and deposits newly minted KLAY to three predefined accounts, i.e. Reward contract, KIR contract, PoC contract.
func MintKLAY(b BalanceAdder) {
	// TODO-GX-issue973 Developing Klaytn token economy
	b.AddBalance(common.HexToAddress(contract.RewardContractAddress), params.RewardContractIncentive)
	b.AddBalance(common.HexToAddress(contract.KIRContractAddress), params.KIRContractIncentive)
	b.AddBalance(common.HexToAddress(contract.PoCContractAddress), params.PoCContractIncentive)
}
