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
	"fmt"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/math"
	"github.com/ground-x/klaytn/log"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/pkg/errors"
	"math/big"
	"math/rand"
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

var logger = log.NewModuleLogger(log.Governance)

// Governance represents vote information given from istanbul.vote()
type GovernanceVote struct {
	Validator common.Address `json:"validator"`
	Key       string         `json:"key"`
	Value     string         `json:"value"`
}

// GovernanceTally represents a tally for each governance item
type GovernanceTally struct {
	Key   string  `json:"key"`
	Value string  `json:"value"`
	Votes float64 `json:"votes"`
}

type Governance struct {
	chainConfig *params.ChainConfig

	// Map used to keep multiple types of votes
	voteMap     map[string]string
	voteMapLock sync.RWMutex
}

func NewGovernance(chainConfig *params.ChainConfig) *Governance {
	ret := Governance{
		chainConfig: chainConfig,
		voteMap:     make(map[string]string),
	}
	return &ret
}

func (g *Governance) GetEncodedVote(addr common.Address) []byte {
	// TODO-Klaytn-Governance Change this part to add all votes to the header at once
	// TODO-Klaytn-Governance Random selection can make side effects
	//  (e.g: Not all votes may not be included).
	//  Make it more deterministic or remove the vote when it is written in a block.
	g.voteMapLock.RLock()
	defer g.voteMapLock.RUnlock()

	mapSize := len(g.voteMap)
	if mapSize > 0 {
		index := rand.Intn(mapSize)
		i := 0
		for key, val := range g.voteMap {
			if i == index {
				vote := new(GovernanceVote)
				vote.Validator = addr
				vote.Key = key
				vote.Value = val
				encoded, err := rlp.EncodeToBytes(vote)
				if err != nil {
					logger.Error("Failed to RLP Encode a vote", "vote", vote)
					return nil
				}
				return encoded
			}
			i++
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

	// value from JS console can be float64, bool or string
	var newVal string
	switch val.(type) {
	case float64:
		newVal = fmt.Sprint(newVal, val)
	case bool:
		newVal = strconv.FormatBool(val.(bool))
	case string:
		newVal = strings.ToLower(val.(string))
	default:
		newVal = fmt.Sprint(newVal, val)
	}

	if v, ok := g.CheckVoteValidity(key, newVal); ok {
		g.voteMap[key] = v
		logger.Info("New governance vote added", "key", key, "val", v)
		return true
	}
	return false
}

// RemoveVote remove a vote from the voteMap to prevent repetitive addition of same vote
func (g *Governance) RemoveVote(key string, value string) {
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

	g.voteMap = make(map[string]string)
	logger.Info("Governance votes are cleared")
}

// CheckVoteValidity checks if the given key and value are appropriate for governance vote
func (g *Governance) CheckVoteValidity(key string, val string) (string, bool) {
	lowerKey := g.getKey(key)

	if k, ok := GovernanceKeyMap[lowerKey]; ok {
		if ret, passed := g.checkValue(k, val); passed {
			return ret, true
		}
	}

	logger.Warn("New vote couldn't pass the validity check", "key", key, "val", val)
	return "", false
}

// checkValue checks if the given value is appropriate
func (g *Governance) checkValue(key int, val string) (string, bool) {
	switch key {
	case GoverningNode:
		if common.IsHexAddress(val) {
			return val, true
		}

	case GovernanceMode:
		if _, ok := GovernanceModeMap[val]; ok {
			return val, true
		}

	case Epoch:
		return g.checkUint(val, 64)

	case Policy:
		if _, ok := g.checkUint(val, 64); ok {
			if _, ok := ProposerPolicyMap[val]; ok {
				return val, true
			}
		}

	case Sub:
		return g.checkUint(val, 8)

	case UnitPrice:
		return g.checkUint(val, 64)

	case MintingAmount:
		x := new(big.Int)
		x, ok := x.SetString(val, 10)
		if ok {
			return val, true
		}

	case Ratio:
		x := strings.Split(val, "/")
		if len(x) != RewardSliceCount {
			return "", false
		}
		var sum float64
		for _, item := range x {
			v, err := strconv.ParseFloat(item, 64)
			if err != nil {
				return "", false
			}
			sum += v
		}
		if sum == 100.0 {
			return val, true
		}

	case UseGiniCoeff:
		_, err := strconv.ParseBool(val)
		if err == nil {
			return val, true // send string because RLP can't support boolean,
		}
	}
	return "", false
}

// checkUint checks if the given string can be casted into uint8 or uint64
func (g *Governance) checkUint(val string, bitsize int) (string, bool) {
	tmp, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		logger.Error("Checking uint value failed", "value", val, "err", err)
		return "", false
	}

	x := uint64(tmp)

	// Check if it meets uint8 limits
	if bitsize == 64 {
		return val, true
	} else if bitsize == 8 && x <= math.MaxUint8 { // checking overflow
		return val, true
	} else {
		return "", false
	}
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
		governance.GoverningNode = common.HexToAddress(vote.Value)
		return true
	case GovernanceMode:
		governance.GovernanceMode = vote.Value
		return true
	case Epoch:
		v, _ := strconv.ParseUint(vote.Value, 10, 64)
		governance.Istanbul.Epoch = v
		return true
	case Policy:
		v, _ := strconv.ParseUint(vote.Value, 10, 64)
		governance.Istanbul.ProposerPolicy = v
		return true
	case UnitPrice:
		v, _ := strconv.ParseUint(vote.Value, 10, 64)
		governance.UnitPrice = v
		return true
	case Sub:
		v, _ := strconv.ParseInt(vote.Value, 10, 64)
		governance.Istanbul.SubGroupSize = int(v)
		return true
	case MintingAmount:
		governance.Reward.MintingAmount, _ = governance.Reward.MintingAmount.SetString(vote.Value, 10)
		return true
	case Ratio:
		governance.Reward.Ratio = vote.Value
		return true
	case UseGiniCoeff:
		governance.Reward.UseGiniCoeff, _ = strconv.ParseBool(vote.Value)
		return true
	}
	return false
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
