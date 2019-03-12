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
	"encoding/binary"
	"encoding/json"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/log"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/pkg/errors"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	// block interval for propagating governance information.
	// This value shouldn't be changed after a network's launch
	GovernanceRefreshInterval = 3600 // block interval. Default is about 1 hour (3600 blocks)
	// Block reward will be separated by three pieces and distributed
	RewardSliceCount = 3
	// GovernanceConfig is stored in a cache which has below capacity
	GovernanceCacheLimit = 3
	// The prefix for governance cache
	GovernanceCachePrefix = "governance"
)

type EngineType int

const (
	// Engine type
	UseIstanbul EngineType = iota
	UseClique
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
	DefaultGoverningNode  = "0x0000000000000000000000000000000000000000"
	DefaultEpoch          = 30000
	DefaultProposerPolicy = 0
	DefaultSubGroupSize   = 21
	DefaultMintingAmount  = 0
	DefaultRatio          = "100/0/0"
	DefaultUseGiniCoeff   = false
	DefaultDefferedTxFee  = false
	DefaultUnitPrice      = 250000000000
	DefaultPeriod         = 1
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
		"roundrobin":     RoundRobin,
		"sticky":         Sticky,
		"weightedrandom": WeightedRandom,
	}

	GovernanceModeMap = map[string]int{
		"none":   GovernanceMode_None,
		"single": GovernanceMode_Single,
		"ballot": GovernanceMode_Ballot,
	}
)

var logger = log.NewModuleLogger(log.Governance)

// Governance represents vote information given from istanbul.vote()
type GovernanceVote struct {
	Validator common.Address `json:"validator"`
	Key       string         `json:"key"`
	Value     interface{}    `json:"value"`
}

// GovernanceTally represents a tally for each governance item
type GovernanceTally struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Votes uint64      `json:"votes"`
}

type Governance struct {
	chainConfig *params.ChainConfig

	// Map used to keep multiple types of votes
	voteMap     map[string]interface{}
	voteMapLock sync.RWMutex
}

func NewGovernance(chainConfig *params.ChainConfig) *Governance {
	ret := Governance{
		chainConfig: chainConfig,
		voteMap:     make(map[string]interface{}),
	}
	return &ret
}

func (g *Governance) GetEncodedVote(addr common.Address) []byte {
	// TODO-Klaytn-Governance Change this part to add all votes to the header at once
	g.voteMapLock.RLock()
	defer g.voteMapLock.RUnlock()

	if len(g.voteMap) > 0 {
		for key, val := range g.voteMap {
			vote := new(GovernanceVote)
			vote.Validator = addr
			vote.Key = key
			vote.Value = val
			encoded, err := rlp.EncodeToBytes(vote)
			if err != nil {
				logger.Error("Failed to RLP Encode a vote", "vote", vote)
				g.RemoveVote(key, val)
				continue
			}
			return encoded
		}
	}
	return nil
}

func (g *Governance) getKey(k string) string {
	return strings.Trim(strings.ToLower(k), " ")
}

// AddVote adds a vote to the voteMap
func (g *Governance) AddVote(key string, val interface{}) bool {
	g.voteMapLock.Lock()
	defer g.voteMapLock.Unlock()

	key = g.getKey(key)

	if ok := g.CheckVoteValidity(key, val); ok {
		g.voteMap[key] = val
		return true
	}
	return false
}

func (g *Governance) checkValueType(key string, val interface{}) (interface{}, bool) {
	keyIdx := GovernanceKeyMap[key]
	switch t := val.(type) {
	case uint64:
		if keyIdx == Epoch || keyIdx == Sub || keyIdx == UnitPrice {
			return val, true
		}
	case string:
		if keyIdx == GovernanceMode || keyIdx == MintingAmount || keyIdx == Ratio || keyIdx == Policy {
			return strings.ToLower(val.(string)), true
		} else if keyIdx == GoverningNode {
			if common.IsHexAddress(val.(string)) {
				return val, true
			}
		}
	case bool:
		if keyIdx == UseGiniCoeff {
			return val, true
		}
	case common.Address:
		if keyIdx == GoverningNode {
			return val, true
		}
	case float64:
		// When value comes from JS console, all numbers come in a form of float64
		if keyIdx == Epoch || keyIdx == Sub || keyIdx == UnitPrice {
			if val.(float64) >= 0 && val.(float64) == float64(uint64(val.(float64))) {
				val = uint64(val.(float64))
				return val, true
			}
		}
	default:
		logger.Warn("Unknown value type of the given vote", "key", key, "value", val, "type", t)

	}
	return val, false
}

// RemoveVote remove a vote from the voteMap to prevent repetitive addition of same vote
func (g *Governance) RemoveVote(key string, value interface{}) {
	g.voteMapLock.Lock()
	defer g.voteMapLock.Unlock()

	key = g.getKey(key)
	if g.voteMap[key] == value {
		delete(g.voteMap, key)
	}
}

func (g *Governance) ClearVotes() {
	g.voteMapLock.Lock()
	defer g.voteMapLock.Unlock()

	g.voteMap = make(map[string]interface{})
	logger.Info("Governance votes are cleared")
}

// CheckVoteValidity checks if the given key and value are appropriate for governance vote
func (g *Governance) CheckVoteValidity(key string, val interface{}) bool {
	lowerKey := g.getKey(key)

	// Check if the val's type meets type requirements
	var passed bool
	if val, passed = g.checkValueType(lowerKey, val); !passed {
		logger.Warn("New vote couldn't pass the validity check", "key", key, "val", val)
		return false
	}

	return g.checkValue(key, val)
}

// checkValue checks if the given value is appropriate
func (g *Governance) checkValue(key string, val interface{}) bool {
	k := GovernanceKeyMap[key]

	// Using type assertion is okay below, because type check was done before calling this method
	switch k {
	case GoverningNode:
		if reflect.TypeOf(val).String() == "common.Address" {
			return true
		} else if common.IsHexAddress(val.(string)) {
			return true
		}

	case GovernanceMode:
		if _, ok := GovernanceModeMap[val.(string)]; ok {
			return true
		}

	case Epoch, Sub, UnitPrice, UseGiniCoeff:
		// For Uint64 and bool types, no more check is needed
		return true

	case Policy:
		if _, ok := ProposerPolicyMap[val.(string)]; ok {
			return true
		}

	case MintingAmount:
		x := new(big.Int)
		if _, ok := x.SetString(val.(string), 10); ok {
			return true
		}

	case Ratio:
		x := strings.Split(val.(string), "/")
		if len(x) != RewardSliceCount {
			return false
		}
		var sum uint64
		for _, item := range x {
			v, err := strconv.ParseUint(item, 10, 64)
			if err != nil {
				return false
			}
			sum += v
		}
		if sum == 100 {
			return true
		}
	default:
		logger.Warn("Unknown vote key was given", "key", k)
	}
	return false
}

// parseVoteValue parse vote.Value from []uint8 to appropriate type
func (g *Governance) ParseVoteValue(gVote *GovernanceVote) *GovernanceVote {
	var val interface{}
	k := GovernanceKeyMap[gVote.Key]

	switch k {
	case GovernanceMode, GoverningNode, MintingAmount, Ratio, Policy:
		val = string(gVote.Value.([]uint8))
	case Epoch, Sub, UnitPrice:
		gVote.Value = append(make([]byte, 8-len(gVote.Value.([]uint8))), gVote.Value.([]uint8)...)
		val = binary.BigEndian.Uint64(gVote.Value.([]uint8))
	case UseGiniCoeff:
		gVote.Value = append(make([]byte, 8-len(gVote.Value.([]uint8))), gVote.Value.([]uint8)...)
		if binary.BigEndian.Uint64(gVote.Value.([]uint8)) != uint64(0) {
			val = true
		} else {
			val = false
		}
	default:
		logger.Warn("Unknown vote key was given", "key", k)
	}
	gVote.Value = val
	return gVote
}

// MakeGovernanceData returns rlp encoded json retrieved from GovernanceConfig
func MakeGovernanceData(governance *params.GovernanceConfig) ([]byte, error) {
	j, _ := json.Marshal(governance)
	payload, err := rlp.EncodeToBytes(j)
	if err != nil {
		return nil, errors.New("Failed to encode governance data")
	}
	return payload, nil
}

func ReflectVotes(vote GovernanceVote, governance *params.GovernanceConfig) {
	if ok := updateGovernanceConfig(vote, governance); !ok {
		logger.Error("Failed to reflect Governance Config", "Key", vote.Key, "Value", vote.Value)
	}
}

func updateGovernanceConfig(vote GovernanceVote, governance *params.GovernanceConfig) bool {
	// Error check had been done when vote was injected. So no error check is required here.
	switch GovernanceKeyMap[vote.Key] {
	case GoverningNode:
		// CAUTION: governingnode can be changed at any current mode
		// If it passed, a mode change have to be followed after setting governingnode
		governance.GoverningNode = common.HexToAddress(vote.Value.(string))
		return true
	case GovernanceMode:
		governance.GovernanceMode = vote.Value.(string)
		return true
	case Epoch:
		governance.Istanbul.Epoch = vote.Value.(uint64)
		return true
	case Policy:
		governance.Istanbul.ProposerPolicy = uint64(ProposerPolicyMap[vote.Value.(string)])
		return true
	case UnitPrice:
		governance.UnitPrice = vote.Value.(uint64)
		return true
	case Sub:
		governance.Istanbul.SubGroupSize = int(vote.Value.(uint64))
		return true
	case MintingAmount:
		governance.Reward.MintingAmount, _ = governance.Reward.MintingAmount.SetString(vote.Value.(string), 10)
		return true
	case Ratio:
		governance.Reward.Ratio = vote.Value.(string)
		return true
	case UseGiniCoeff:
		governance.Reward.UseGiniCoeff = vote.Value.(bool)
		return true
	default:
		logger.Warn("Unknown vote key was given", "key", vote.Key)
	}
	return false
}

func GetDefaultGovernanceConfig(engine EngineType) *params.GovernanceConfig {
	gov := &params.GovernanceConfig{
		GovernanceMode: DefaultGovernanceMode,
		GoverningNode:  common.HexToAddress(DefaultGoverningNode),
		Reward:         GetDefaultRewardConfig(),
		UnitPrice:      DefaultUnitPrice,
	}

	if engine == UseIstanbul {
		gov.Istanbul = GetDefaultIstanbulConfig()
	}

	return gov
}

func GetDefaultIstanbulConfig() *params.IstanbulConfig {
	return &params.IstanbulConfig{
		Epoch:          DefaultEpoch,
		ProposerPolicy: DefaultProposerPolicy,
		SubGroupSize:   DefaultSubGroupSize,
	}
}

func GetDefaultRewardConfig() *params.RewardConfig {
	return &params.RewardConfig{
		MintingAmount: big.NewInt(DefaultMintingAmount),
		Ratio:         DefaultRatio,
		UseGiniCoeff:  DefaultUseGiniCoeff,
		DeferredTxFee: DefaultDefferedTxFee,
	}
}

func GetDefaultCliqueConfig() *params.CliqueConfig {
	return &params.CliqueConfig{
		Epoch:  DefaultEpoch,
		Period: DefaultPeriod,
	}
}
