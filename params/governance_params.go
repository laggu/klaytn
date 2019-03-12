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

package params

import (
	"math/big"
)

const (
	// Because we need int64 type to allocate big.Int, define these parameters as int64 type.
	// In addition, let's define these constants in ston instead of peb, because int64 can hold
	// up to about 9*10^18 in golang.

	// TODO-Klaytn-Issue1587 Decide whether to remove below three variables after discussing token economy policy for service chain and private network
	rewardContractIncentiveInSton int64 = 3200000000 // 3.2 KLAY for Reward contract (Unit: ston)
	kirContractIncentiveInSton    int64 = 3200000000 // 3.2 KLAY for KIR contract (Unit: ston)
	pocContractIncentiveInSton    int64 = 3200000000 // 3.2 KLAY for PoC contract (Unit: ston)

	defaultMintedKLAYInSton int64 = 9600000000 // Default amount of minted KLAY. 9.6 KLAY for block reward (Unit: ston)

	DefaultCNRewardRatio  = 330 // Default CN reward ratio 33.0%
	DefaultPoCRewardRatio = 545 // Default PoC ratio 54.5%
	DefaultKIRRewardRatio = 125 // Default KIR ratio 12.5%

	stakingUpdateInterval  uint64 = 86400 // About 1 day. 86400 blocks = (24 hrs) * (3600 secs/hr) * (1 block/sec)
	proposerUpdateInterval uint64 = 3600  // About 1 hour. 3600 blocks = (1 hr) * (3600 secs/hr) * (1 block/sec)
)

var (
	// TODO-Klaytn-Issue1587 Decide whether to remove below three variables after discussing token economy policy for service chain and private network
	RewardContractIncentive = big.NewInt(0).Mul(big.NewInt(rewardContractIncentiveInSton), big.NewInt(Ston))
	KIRContractIncentive    = big.NewInt(0).Mul(big.NewInt(kirContractIncentiveInSton), big.NewInt(Ston))
	PoCContractIncentive    = big.NewInt(0).Mul(big.NewInt(pocContractIncentiveInSton), big.NewInt(Ston))

	DefaultMintedKLAY = big.NewInt(0).Mul(big.NewInt(defaultMintedKLAYInSton), big.NewInt(Ston))
)

func IsStakingUpdatePossible(blockNum uint64) bool {
	return (blockNum % stakingUpdateInterval) == 0
}

// CalcStakingBlockNumber returns number of block which contains staking information required to make a new block with blockNum.
func CalcStakingBlockNumber(blockNum uint64) uint64 {
	if blockNum <= 2*stakingUpdateInterval {
		// Just return genesis block number.
		return 0
	}

	var number uint64
	if IsStakingUpdatePossible(blockNum) {
		number = blockNum - 2*stakingUpdateInterval
	} else {
		number = blockNum - stakingUpdateInterval - (blockNum % stakingUpdateInterval)
	}
	return number
}

func IsProposerUpdateInterval(blockNum uint64) bool {
	return (blockNum % proposerUpdateInterval) == 0
}

// CalcProposerBlockNumber returns number of block where list of proposers is updated for block blockNum
func CalcProposerBlockNumber(blockNum uint64) uint64 {
	if blockNum <= proposerUpdateInterval {
		// Just return genesis block number.
		return 0
	}

	var number uint64
	if IsProposerUpdateInterval(blockNum) {
		number = blockNum - proposerUpdateInterval
	} else {
		number = blockNum - (blockNum % proposerUpdateInterval)

	}
	return number
}
