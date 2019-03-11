// Copyright 2018 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
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
//
// This file is derived from params/protocol_params.go (2018/06/04).
// Modified and improved for the klaytn development.

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
)

var (
	// TODO-Klaytn-Issue1587 Decide whether to remove below three variables after discussing token economy policy for service chain and private network
	RewardContractIncentive = big.NewInt(0).Mul(big.NewInt(rewardContractIncentiveInSton), big.NewInt(Ston))
	KIRContractIncentive    = big.NewInt(0).Mul(big.NewInt(kirContractIncentiveInSton), big.NewInt(Ston))
	PoCContractIncentive    = big.NewInt(0).Mul(big.NewInt(pocContractIncentiveInSton), big.NewInt(Ston))

	DefaultMintedKLAY = big.NewInt(0).Mul(big.NewInt(defaultMintedKLAYInSton), big.NewInt(Ston))
)
