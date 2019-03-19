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
	"sync/atomic"
)

var (
	GovernanceKeyMap = map[string]int{
		"governancemode": params.GovernanceMode,
		"governingnode":  params.GoverningNode,
		"epoch":          params.Epoch,
		"policy":         params.Policy,
		"sub":            params.Sub,
		"unitprice":      params.UnitPrice,
		"mintingamount":  params.MintingAmount,
		"ratio":          params.Ratio,
		"useginicoeff":   params.UseGiniCoeff,
	}

	ProposerPolicyMap = map[string]int{
		"roundrobin":     params.RoundRobin,
		"sticky":         params.Sticky,
		"weightedrandom": params.WeightedRandom,
	}

	GovernanceModeMap = map[string]int{
		"none":   params.GovernanceMode_None,
		"single": params.GovernanceMode_Single,
		"ballot": params.GovernanceMode_Ballot,
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

	nodeAddress      common.Address
	totalVotingPower uint64
	votingPower      uint64

	GovernanceTally     []*GovernanceTally
	GovernanceTallyLock sync.RWMutex
}

func NewGovernance(chainConfig *params.ChainConfig) *Governance {
	ret := Governance{
		chainConfig: chainConfig,
		voteMap:     make(map[string]interface{}),
	}
	return &ret
}

func (g *Governance) SetNodeAddress(addr common.Address) {
	g.nodeAddress = addr
}

func (g *Governance) SetTotalVotingPower(t uint64) {
	atomic.StoreUint64(&g.totalVotingPower, t)
}

func (g *Governance) SetMyVotingPower(t uint64) {
	atomic.StoreUint64(&g.votingPower, t)
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

	if v, ok := g.CheckVoteValidity(key, val); ok {
		g.voteMap[key] = v
		return true
	}
	return false
}

func (g *Governance) checkValueType(key string, val interface{}) (interface{}, bool) {
	keyIdx := GovernanceKeyMap[key]
	switch t := val.(type) {
	case uint64:
		if keyIdx == params.Epoch || keyIdx == params.Sub || keyIdx == params.UnitPrice {
			return val, true
		}
	case string:
		if keyIdx == params.GovernanceMode || keyIdx == params.MintingAmount || keyIdx == params.Ratio || keyIdx == params.Policy {
			return strings.ToLower(val.(string)), true
		} else if keyIdx == params.GoverningNode {
			if common.IsHexAddress(val.(string)) {
				return val, true
			}
		}
	case bool:
		if keyIdx == params.UseGiniCoeff {
			return val, true
		}
	case common.Address:
		if keyIdx == params.GoverningNode {
			return val, true
		}
	case float64:
		// When value comes from JS console, all numbers come in a form of float64
		if keyIdx == params.Epoch || keyIdx == params.Sub || keyIdx == params.UnitPrice {
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

func (g *Governance) ClearVotes(num uint64) {
	g.voteMapLock.Lock()
	defer g.voteMapLock.Unlock()

	g.voteMap = make(map[string]interface{})
	logger.Info("Governance votes are cleared", "num", num)
}

// CheckVoteValidity checks if the given key and value are appropriate for governance vote
func (g *Governance) CheckVoteValidity(key string, val interface{}) (interface{}, bool) {
	lowerKey := g.getKey(key)

	// Check if the val's type meets type requirements
	var passed bool
	if val, passed = g.checkValueType(lowerKey, val); !passed {
		logger.Warn("New vote couldn't pass the validity check", "key", key, "val", val)
		return val, false
	}

	return g.checkValue(key, val)
}

// checkValue checks if the given value is appropriate
func (g *Governance) checkValue(key string, val interface{}) (interface{}, bool) {
	k := GovernanceKeyMap[key]

	// Using type assertion is okay below, because type check was done before calling this method
	switch k {
	case params.GoverningNode:
		if reflect.TypeOf(val).String() == "common.Address" {
			return val, true
		} else if common.IsHexAddress(val.(string)) {
			return val, true
		}

	case params.GovernanceMode:
		if _, ok := GovernanceModeMap[val.(string)]; ok {
			return val, true
		}

	case params.Epoch, params.Sub, params.UnitPrice, params.UseGiniCoeff:
		// For Uint64 and bool types, no more check is needed
		return val, true

	case params.Policy:
		if _, ok := ProposerPolicyMap[val.(string)]; ok {
			return val, true
		}

	case params.MintingAmount:
		x := new(big.Int)
		if _, ok := x.SetString(val.(string), 10); ok {
			return val, true
		}

	case params.Ratio:
		x := strings.Split(val.(string), "/")
		if len(x) != params.RewardSliceCount {
			return val, false
		}
		var sum uint64
		for _, item := range x {
			v, err := strconv.ParseUint(item, 10, 64)
			if err != nil {
				return val, false
			}
			sum += v
		}
		if sum == 100 {
			return val, true
		}
	default:
		logger.Warn("Unknown vote key was given", "key", k)
	}
	return val, false
}

// parseVoteValue parse vote.Value from []uint8 to appropriate type
func (g *Governance) ParseVoteValue(gVote *GovernanceVote) *GovernanceVote {
	var val interface{}
	k := GovernanceKeyMap[gVote.Key]

	switch k {
	case params.GovernanceMode, params.GoverningNode, params.MintingAmount, params.Ratio, params.Policy:
		val = string(gVote.Value.([]uint8))
	case params.Epoch, params.Sub, params.UnitPrice:
		gVote.Value = append(make([]byte, 8-len(gVote.Value.([]uint8))), gVote.Value.([]uint8)...)
		val = binary.BigEndian.Uint64(gVote.Value.([]uint8))
	case params.UseGiniCoeff:
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
	payload, err := rlp.EncodeToBytes(governance)

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
	case params.GoverningNode:
		// CAUTION: governingnode can be changed at any current mode
		// If it passed, a mode change have to be followed after setting governingnode
		governance.GoverningNode = common.HexToAddress(vote.Value.(string))
		return true
	case params.GovernanceMode:
		governance.GovernanceMode = vote.Value.(string)
		return true
	case params.Epoch:
		governance.Istanbul.Epoch = vote.Value.(uint64)
		return true
	case params.Policy:
		governance.Istanbul.ProposerPolicy = uint64(ProposerPolicyMap[vote.Value.(string)])
		return true
	case params.UnitPrice:
		governance.UnitPrice = vote.Value.(uint64)
		return true
	case params.Sub:
		governance.Istanbul.SubGroupSize = vote.Value.(uint64)
		return true
	case params.MintingAmount:
		governance.Reward.MintingAmount, _ = governance.Reward.MintingAmount.SetString(vote.Value.(string), 10)
		return true
	case params.Ratio:
		governance.Reward.Ratio = vote.Value.(string)
		return true
	case params.UseGiniCoeff:
		governance.Reward.UseGiniCoeff = vote.Value.(bool)
		return true
	default:
		logger.Warn("Unknown vote key was given", "key", vote.Key)
	}
	return false
}

func GetDefaultGovernanceConfig(engine params.EngineType) *params.GovernanceConfig {
	gov := &params.GovernanceConfig{
		GovernanceMode: params.DefaultGovernanceMode,
		GoverningNode:  common.HexToAddress(params.DefaultGoverningNode),
		Reward:         GetDefaultRewardConfig(),
		UnitPrice:      params.DefaultUnitPrice,
	}

	if engine == params.UseIstanbul {
		gov.Istanbul = GetDefaultIstanbulConfig()
	}

	return gov
}

func GetDefaultIstanbulConfig() *params.IstanbulConfig {
	return &params.IstanbulConfig{
		Epoch:          params.DefaultEpoch,
		ProposerPolicy: params.DefaultProposerPolicy,
		SubGroupSize:   params.DefaultSubGroupSize,
	}
}

func GetDefaultRewardConfig() *params.RewardConfig {
	return &params.RewardConfig{
		MintingAmount: big.NewInt(params.DefaultMintingAmount),
		Ratio:         params.DefaultRatio,
		UseGiniCoeff:  params.DefaultUseGiniCoeff,
		DeferredTxFee: params.DefaultDefferedTxFee,
	}
}

func GetDefaultCliqueConfig() *params.CliqueConfig {
	return &params.CliqueConfig{
		Epoch:  params.DefaultEpoch,
		Period: params.DefaultPeriod,
	}
}
