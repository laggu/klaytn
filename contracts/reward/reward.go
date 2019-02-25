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
//go:generate abigen --sol contract/ProposerReward.sol --pkg contract --out contract/ProposerReward.go
//go:generate abigen --sol contract/PIRerve.sol --pkg contract --out contract/PIRerve.go
//go:generate abigen --sol contract/AddressBook.sol --pkg contract --out contract/AddressBook.go

package reward

import (
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/contracts/reward/contract"
	"github.com/ground-x/klaytn/log"
	"github.com/ground-x/klaytn/params"
	"math/big"
)

var logger = log.NewModuleLogger(log.Reward)

const (
	// TODO-Klaytn-Issue1166 We use small number for testing. We have to decide staking interval for real network.
	StakingUpdateInterval uint64 = 16
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
	// TODO-Klaytn-Issue973 Developing Klaytn token economy
	b.AddBalance(common.HexToAddress(contract.RewardContractAddress), params.RewardContractIncentive)
	b.AddBalance(common.HexToAddress(contract.KIRContractAddress), params.KIRContractIncentive)
	b.AddBalance(common.HexToAddress(contract.PoCContractAddress), params.PoCContractIncentive)
}

func isEmptyAddress(addr common.Address) bool {
	return addr == common.Address{}
}

// DistributeBlockReward mints KLAY and distribute newly minted KLAY to proposer, kirAddr and pocAddr. proposer also gets totalTxFee.
func DistributeBlockReward(b BalanceAdder, validators []common.Address, totalTxFee *big.Int, kirAddr common.Address, pocAddr common.Address) {
	proposer := validators[0]

	// Mint KLAY for proposer
	mintedKLAY := params.ProposerIncentive
	b.AddBalance(proposer, mintedKLAY)
	logger.Debug("Block reward - Minted KLAY", "reward address of proposer", proposer, "Amount", mintedKLAY)

	// Transfer Tx fee for proposer
	b.AddBalance(proposer, totalTxFee)
	logger.Debug("Block reward - Tx fee", "reward address of proposer", proposer, "Amount", totalTxFee)

	// Mint KLAY for KIR
	if isEmptyAddress(kirAddr) {
		// Consider bootstrapping
		b.AddBalance(proposer, params.KIRContractIncentive)
		logger.Debug("Block reward - KIR. No KIR address.", "reward address of proposer", proposer, "Amount", params.KIRContractIncentive)
	} else {
		b.AddBalance(kirAddr, params.KIRContractIncentive)
		logger.Debug("Block reward - KIR", "KIR address", kirAddr, "Amount", params.KIRContractIncentive)
	}

	// Mint KLAY for PoC
	if isEmptyAddress(pocAddr) {
		// Consider bootstrapping
		b.AddBalance(proposer, params.KIRContractIncentive)
		logger.Debug("Block reward - PoC. No PoC address.", "reward address of proposer", proposer, "Amount", params.PoCContractIncentive)
	} else {
		b.AddBalance(pocAddr, params.PoCContractIncentive)
		logger.Debug("Block reward - PoC", "PoC address", pocAddr, "Amount", params.PoCContractIncentive)
	}
}

func IsStakingUpdateInterval(blockNum uint64) bool {
	return (blockNum % StakingUpdateInterval) == 0
}

// CalcStakingBlockNumber returns number of block which contains staking information required to make a new block with blockNum.
func CalcStakingBlockNumber(blockNum uint64) uint64 {
	if blockNum < 2*StakingUpdateInterval {
		// Bootstrapping. Just return genesis block number.
		return 0
	}
	number := blockNum - StakingUpdateInterval - (blockNum % StakingUpdateInterval)
	return number
}

// StakingCache
const (
	// TODO-Klaytn-Issue1166 Decide size of cache
	maxStakingCache = 3
)

var stakingCache common.Cache // TODO-Klaytn-Issue1166 Cache for staking information of Council

func init() {
	initStakingCache()
}

func initStakingCache() {
	stakingCache, _ = common.NewCache(common.LRUConfig{CacheSize: maxStakingCache})
}

// GetStakingInfoFromStakingCache returns corresponding staking information for a block of blockNum.
func GetStakingInfoFromStakingCache(blockNum uint64) *common.StakingInfo {
	number := CalcStakingBlockNumber(blockNum)
	stakingCacheKey := common.StakingCacheKey(number)
	value, ok := stakingCache.Get(stakingCacheKey)
	if !ok {
		logger.Error("Staking cache missed", "Block number", blockNum, "cache key", stakingCacheKey)
		return nil
	}

	stakingInfo, ok := value.(*common.StakingInfo)
	if !ok {
		logger.Error("Found staking information is invalid", "Block number", blockNum, "cache key", stakingCacheKey)
		return nil
	}

	if stakingInfo.BlockNum != number {
		logger.Error("Staking cache hit. But staking information not found", "Block number", blockNum, "cache key", stakingCacheKey)
		return nil
	}

	logger.Debug("Staking cache hit.", "Block number", blockNum, "stakingInfo", stakingInfo, "cache key", stakingCacheKey)
	return stakingInfo
}
