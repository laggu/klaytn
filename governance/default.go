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

package governance

import (
	"encoding/json"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/pkg/errors"
	"math/big"
)

const (
	// block interval for propagating governance information.
	// This value shouldn't be changed after a network's launch
	GovernanceRefreshInterval = 3600
	// Block reward will be separated by three pieces and distributed
	RewardSliceCount = 3
	// GovernanceConfig is stored in a cache which has below capacity
	GovernanceCacheLimit = 3
	// The prefix for governance cache
	GovernanceCachePrefix = "governance"
)

const (
	// Governance Key
	GovernanceMode = iota
	GoverningNode
	Epoch
	Policy
	Sub
	UnitPrice
	MintingAmount
	Ratio
	UseGiniCoeff
)

const (
	GovernanceMode_None = iota
	GovernanceMode_Single
	GovernanceMode_Ballot
)

const (
	// Proposer policy
	// At the moment this is duplicated in istanbul/config.go, not to make a cross reference
	// TODO-Klatn-Governance: Find a way to manage below constants at single location
	RoundRobin = iota
	Sticky
	WeightedRandom
)

const (
	// Default Values: Constants used for getting default values for configuration
	DefaultGovernanceMode = "none"
	DefaultGoverningNode  = "0x00000000000000000000"
	DefaultEpoch          = 30000
	DefaultProposerPolicy = 0
	DefaultSubGroupSize   = 21
	DefaultMintingAmount  = 0
	DefaultRatio          = "100/0/0"
	DefaultUseGiniCoeff   = false
	DefaultDefferedTxFee  = false
	DefaultUnitPrice      = 250000000000
)

var (
	GovernanceKeyMap = map[string]int{
		"governancemode": GovernanceMode,
		"governingnode":  GoverningNode,
		"epoch":          Epoch,
		"policy":         Policy,
		"sub":            Sub,
		"unitprice":      UnitPrice,
		"mintingamount":  MintingAmount,
		"ratio":          Ratio,
		"useginicoeff":   UseGiniCoeff,
	}

	ProposerPolicyMap = map[string]int{
		"0": RoundRobin,
		"1": Sticky,
		"2": WeightedRandom,
	}

	GovernanceModeMap = map[string]int{
		"none":   0,
		"single": 1,
		"ballot": 2,
	}
)

// MakeGovernanceData returns rlp encoded json retrieved from GovernanceConfig
func MakeGovernanceData(governance *params.GovernanceConfig) ([]byte, error) {
	j, _ := json.Marshal(governance)
	payload, err := rlp.EncodeToBytes(j)
	if err != nil {
		return nil, errors.New("Failed to encode governance data")
	}
	return payload, nil
}

func GetDefaultGovernanceConfig() *params.GovernanceConfig {
	return &params.GovernanceConfig{
		GovernanceMode: DefaultGovernanceMode,
		GoverningNode:  common.HexToAddress(DefaultGoverningNode),
		Istanbul: &params.IstanbulConfig{
			Epoch:          DefaultEpoch,
			ProposerPolicy: DefaultProposerPolicy,
			SubGroupSize:   DefaultSubGroupSize,
		},
		Reward: &params.RewardConfig{
			MintingAmount: big.NewInt(DefaultMintingAmount),
			Ratio:         DefaultRatio,
			UseGiniCoeff:  DefaultUseGiniCoeff,
			DeferredTxFee: DefaultDefferedTxFee,
		},
		UnitPrice: DefaultUnitPrice,
	}
}
