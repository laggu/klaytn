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
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/log"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/ground-x/klaytn/storage/database"
	"github.com/pkg/errors"
	"math/big"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	GovernanceKeyMap = map[string]int{
		"governance.governancemode":     params.GovernanceMode,
		"governance.governingnode":      params.GoverningNode,
		"istanbul.epoch":                params.Epoch,
		"istanbul.policy":               params.Policy,
		"istanbul.committeesize":        params.CommitteeSize,
		"governance.unitprice":          params.UnitPrice,
		"reward.mintingamount":          params.MintingAmount,
		"reward.ratio":                  params.Ratio,
		"reward.useginicoeff":           params.UseGiniCoeff,
		"reward.deferredtxfee":          params.DeferredTxFee,
		"reward.minimumstake":           params.MinimumStake,
		"reward.stakingupdateinterval":  params.StakeUpdateInterval,
		"reward.proposerupdateinterval": params.ProposerRefreshInterval,
		"governance.addvalidator":       params.AddValidator,
		"governance.removevalidator":    params.RemoveValidator,
	}

	GovernanceKeyMapReverse = map[int]string{
		params.GovernanceMode:          "governance.governancemode",
		params.GoverningNode:           "governance.governingnode",
		params.Epoch:                   "istanbul.epoch",
		params.Policy:                  "istanbul.policy",
		params.CommitteeSize:           "istanbul.committeesize",
		params.UnitPrice:               "governance.unitprice",
		params.MintingAmount:           "reward.mintingamount",
		params.Ratio:                   "reward.ratio",
		params.UseGiniCoeff:            "reward.useginicoeff",
		params.DeferredTxFee:           "reward.deferredtxfee",
		params.MinimumStake:            "reward.minimumstake",
		params.StakeUpdateInterval:     "reward.stakingupdateinterval",
		params.ProposerRefreshInterval: "reward.proposerupdateinterval",
		params.AddValidator:            "governance.addvalidator",
		params.RemoveValidator:         "governance.removevalidator",
	}

	ProposerPolicyMap = map[string]int{
		"roundrobin":     params.RoundRobin,
		"sticky":         params.Sticky,
		"weightedrandom": params.WeightedRandom,
	}

	ProposerPolicyMapReverse = map[int]string{
		params.RoundRobin:     "roundrobin",
		params.Sticky:         "sticky",
		params.WeightedRandom: "weightedrandom",
	}

	GovernanceModeMap = map[string]int{
		"none":   params.GovernanceMode_None,
		"single": params.GovernanceMode_Single,
		"ballot": params.GovernanceMode_Ballot,
	}
)

var logger = log.NewModuleLogger(log.Governance)

// Governance item set
type GovernanceSet map[string]interface{}

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

type voteStatus struct {
	value  interface{}
	casted bool
	num    uint64
}

type Governance struct {
	ChainConfig *params.ChainConfig

	// Map used to keep multiple types of votes
	voteMap     map[string]voteStatus
	voteMapLock sync.RWMutex

	nodeAddress      common.Address
	totalVotingPower uint64
	votingPower      uint64

	GovernanceVotes     []*GovernanceVote
	GovernanceTally     []*GovernanceTally
	GovernanceTallyLock sync.RWMutex

	db                     database.DBManager
	itemCache              common.Cache
	idxCache               []uint64
	actualGovernanceBlock  uint64
	currentGovernanceBlock uint64

	currentSet GovernanceSet
	changeSet  GovernanceSet
	mu         sync.RWMutex
}

func (gs GovernanceSet) SetValue(itemType int, value interface{}) error {
	key := GovernanceKeyMapReverse[itemType]

	if GovernanceItems[itemType].t != reflect.TypeOf(value) {
		return errors.New("Value's type mismatch")
	}
	gs[key] = value
	return nil
}

func NewGovernance(chainConfig *params.ChainConfig, dbm database.DBManager) *Governance {
	ret := Governance{
		ChainConfig: chainConfig,
		voteMap:     make(map[string]voteStatus),
		db:          dbm,
		itemCache:   newGovernanceCache(),
		currentSet:  GovernanceSet{},
		changeSet:   GovernanceSet{},
	}
	// nil is for testing or simple function usage
	if dbm != nil {
		ret.initializeCache()
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

func (g *Governance) GetEncodedVote(addr common.Address, number uint64) []byte {
	// TODO-Klaytn-Governance Change this part to add all votes to the header at once
	g.voteMapLock.RLock()
	defer g.voteMapLock.RUnlock()

	if len(g.voteMap) > 0 {
		for key, val := range g.voteMap {
			if val.casted == false {
				vote := new(GovernanceVote)
				vote.Validator = addr
				vote.Key = key
				vote.Value = val.value
				encoded, err := rlp.EncodeToBytes(vote)
				if err != nil {
					logger.Error("Failed to RLP Encode a vote", "vote", vote)
					g.RemoveVote(key, val, number)
					continue
				}
				return encoded
			}
		}
	}
	return nil
}

func (g *Governance) getKey(k string) string {
	return strings.Trim(strings.ToLower(k), " ")
}

// RemoveVote remove a vote from the voteMap to prevent repetitive addition of same vote
func (g *Governance) RemoveVote(key string, value interface{}, number uint64) {
	g.voteMapLock.Lock()
	defer g.voteMapLock.Unlock()

	key = g.getKey(key)
	if g.voteMap[key].value == value {
		g.voteMap[key] = voteStatus{
			value:  value,
			casted: true,
			num:    number,
		}
	}
}

func (g *Governance) ClearVotes(num uint64) {
	g.voteMapLock.Lock()
	defer g.voteMapLock.Unlock()

	g.GovernanceVotes = nil
	g.GovernanceTally = nil
	g.changeSet = GovernanceSet{}
	g.voteMap = make(map[string]voteStatus)
	logger.Info("Governance votes are cleared", "num", num)
}

// parseVoteValue parse vote.Value from []uint8 to appropriate type
func (g *Governance) ParseVoteValue(gVote *GovernanceVote) *GovernanceVote {
	var val interface{}
	k := GovernanceKeyMap[gVote.Key]

	switch k {
	case params.GovernanceMode, params.MintingAmount, params.MinimumStake, params.Ratio, params.Policy:
		val = string(gVote.Value.([]uint8))
	case params.GoverningNode, params.AddValidator, params.RemoveValidator:
		val = common.BytesToAddress(gVote.Value.([]uint8))
	case params.Epoch, params.CommitteeSize, params.UnitPrice, params.StakeUpdateInterval, params.ProposerRefreshInterval:
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
		logger.Warn("Unknown key was given", "key", k)
	}
	gVote.Value = val
	return gVote
}

func (gov *Governance) ReflectVotes(vote GovernanceVote) {
	if ok := gov.updateChangeSet(vote); !ok {
		logger.Error("Failed to reflect Governance Config", "Key", vote.Key, "Value", vote.Value)
	}
}

func (gov *Governance) updateChangeSet(vote GovernanceVote) bool {
	gov.mu.Lock()
	defer gov.mu.Unlock()

	switch GovernanceKeyMap[vote.Key] {
	case params.GoverningNode:
		gov.changeSet[vote.Key] = vote.Value.(common.Address)
		return true
	case params.GovernanceMode, params.Ratio:
		gov.changeSet[vote.Key] = vote.Value.(string)
		return true
	case params.Epoch, params.StakeUpdateInterval, params.ProposerRefreshInterval, params.CommitteeSize, params.UnitPrice:
		gov.changeSet[vote.Key] = vote.Value.(uint64)
		return true
	case params.Policy:
		gov.changeSet[vote.Key] = uint64(ProposerPolicyMap[vote.Value.(string)])
		return true
	case params.MintingAmount, params.MinimumStake:
		gov.changeSet[vote.Key], _ = new(big.Int).SetString(vote.Value.(string), 10)
		return true
	case params.UseGiniCoeff, params.DeferredTxFee:
		gov.changeSet[vote.Key] = vote.Value.(bool)
		return true
	default:
		logger.Warn("Unknown key was given", "key", vote.Key)
	}
	return false
}

func GetDefaultGovernanceConfig(engine params.EngineType) *params.GovernanceConfig {
	gov := &params.GovernanceConfig{
		GovernanceMode: params.DefaultGovernanceMode,
		GoverningNode:  common.HexToAddress(params.DefaultGoverningNode),
		Reward:         GetDefaultRewardConfig(),
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
		MintingAmount:          big.NewInt(params.DefaultMintingAmount),
		Ratio:                  params.DefaultRatio,
		UseGiniCoeff:           params.DefaultUseGiniCoeff,
		DeferredTxFee:          params.DefaultDefferedTxFee,
		StakingUpdateInterval:  uint64(86400),
		ProposerUpdateInterval: uint64(3600),
		MinimumStake:           big.NewInt(2000000),
	}
}

func GetDefaultCliqueConfig() *params.CliqueConfig {
	return &params.CliqueConfig{
		Epoch:  params.DefaultEpoch,
		Period: params.DefaultPeriod,
	}
}

func CheckGenesisValues(c *params.ChainConfig) error {
	gov := NewGovernance(c, nil)

	var tstMap = map[string]interface{}{
		"istanbul.epoch":                c.Istanbul.Epoch,
		"istanbul.committeesize":        c.Istanbul.SubGroupSize,
		"istanbul.policy":               ProposerPolicyMapReverse[int(c.Istanbul.ProposerPolicy)],
		"governance.governancemode":     c.Governance.GovernanceMode,
		"governance.governingnode":      c.Governance.GoverningNode,
		"governance.unitprice":          c.UnitPrice,
		"reward.ratio":                  c.Governance.Reward.Ratio,
		"reward.useginicoeff":           c.Governance.Reward.UseGiniCoeff,
		"reward.deferredtxfee":          c.Governance.Reward.DeferredTxFee,
		"reward.mintingamount":          c.Governance.Reward.MintingAmount.String(),
		"reward.minimumstake":           c.Governance.Reward.MinimumStake.String(),
		"reward.stakingupdateinterval":  c.Governance.Reward.StakingUpdateInterval,
		"reward.proposerupdateinterval": c.Governance.Reward.ProposerUpdateInterval,
	}

	for k, v := range tstMap {
		if _, ok := gov.ValidateVote(&GovernanceVote{Key: k, Value: v}); !ok {
			return errors.New(k + " value is wrong")
		}
	}
	return nil
}

func newGovernanceCache() common.Cache {
	cache := common.NewCache(common.LRUConfig{CacheSize: params.GovernanceCacheLimit})
	return cache
}

func (g *Governance) initializeCache() {
	// get last n governance change block number
	indices, err := g.db.ReadRecentGovernanceIdx(params.GovernanceCacheLimit)
	if err != nil {
		logger.Warn("Failed to retrieve recent governance indices", "err", err)
		return
	}
	g.idxCache = indices
	// Put governance items into the itemCache
	for _, v := range indices {
		if num, data, err := g.ReadGovernance(v); err == nil {
			g.itemCache.Add(getGovernanceCacheKey(num), data)
			g.actualGovernanceBlock = num
		} else {
			logger.Crit("Couldn't read governance cache from database. Check database consistency")
		}
	}

	// the last one is the one to be used now
	ret, _ := g.itemCache.Get(getGovernanceCacheKey(g.actualGovernanceBlock))
	g.currentSet = ret.(GovernanceSet)
}

// getGovernanceCache returns cached governance config as a byte slice
func (g *Governance) getGovernanceCache(num uint64) (GovernanceSet, bool) {
	cKey := getGovernanceCacheKey(num)

	if ret, ok := g.itemCache.Get(cKey); ok && ret != nil {
		return ret.(GovernanceSet), true
	}
	return nil, false
}

func (g *Governance) addGovernanceCache(num uint64, data GovernanceSet) {
	cKey := getGovernanceCacheKey(num)
	g.itemCache.Add(cKey, data)
	g.addIdxCache(num)
}

// getGovernanceCacheKey returns cache key of the given block number
func getGovernanceCacheKey(num uint64) common.GovernanceCacheKey {
	v := fmt.Sprintf("%v", num)
	return common.GovernanceCacheKey(params.GovernanceCachePrefix + "_" + v)
}

func (g *Governance) addIdxCache(num uint64) {
	g.idxCache = append(g.idxCache, num)
	if len(g.idxCache) > params.GovernanceCacheLimit {
		g.idxCache = g.idxCache[len(g.idxCache)-params.GovernanceCacheLimit:]
	}
}

// Store new governance data on DB. This updates Governance cache too.
func (g *Governance) WriteGovernance(num uint64, data GovernanceSet, delta GovernanceSet) error {

	new := make(GovernanceSet)
	new = CopyGovernanceSet(new, data)

	// merge delta into data
	if delta != nil {
		new = CopyGovernanceSet(new, delta)
	}
	g.addGovernanceCache(num, new)
	return g.db.WriteGovernance(new, num)
}

func CopyGovernanceSet(dst GovernanceSet, src GovernanceSet) GovernanceSet {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (g *Governance) searchCache(num uint64) (uint64, bool) {
	for i := len(g.idxCache) - 1; i >= 0; i-- {
		if g.idxCache[i] <= num {
			return g.idxCache[i], true
		}
	}
	return 0, false
}

func (g *Governance) ReadGovernance(num uint64) (uint64, GovernanceSet, error) {
	blockNum := CalcGovernanceInfoBlock(num, g.ChainConfig.Istanbul.Epoch)
	// Check cache first
	if gBlockNum, ok := g.searchCache(blockNum); ok {
		if data, okay := g.getGovernanceCache(gBlockNum); okay {
			return gBlockNum, data, nil
		}
	}
	return g.db.ReadGovernanceAtNumber(num, g.ChainConfig.Istanbul.Epoch)
}

func CalcGovernanceInfoBlock(num uint64, epoch uint64) uint64 {
	governanceInfoBlock := num - (num % epoch)
	if governanceInfoBlock >= epoch {
		governanceInfoBlock -= epoch
	}
	return governanceInfoBlock
}

func (g *Governance) GetGovernanceChange() GovernanceSet {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.changeSet) > 0 {
		return g.changeSet
	}
	return nil
}

func (gov *Governance) UpdateGovernance(header *types.Header, proposer common.Address, self common.Address) {
	var epoch uint64
	var ok bool

	if epoch, ok = gov.currentSet["istanbul.epoch"].(uint64); !ok {
		if epoch, ok = gov.currentSet["clique.epoch"].(uint64); !ok {
			logger.Error("Couldn't find epoch from governance items")
			return
		}
	}

	// Store updated governance information if exist
	number := header.Number.Uint64()
	if number%epoch == 0 {
		if len(header.Governance) > 0 {
			tempData := []byte("")
			tempSet := GovernanceSet{}
			if err := rlp.DecodeBytes(header.Governance, &tempData); err != nil {
				logger.Error("Failed to decode governance data", "err", err, "data", header.Governance)
				return
			}
			if err := json.Unmarshal(tempData, &tempSet); err != nil {
				logger.Error("Failed to unmarshal governance data", "err", err, "data", tempData)
				return

			}
			tempSet = adjustDecodedSet(tempSet)

			// Store new currentSet to governance database
			if err := gov.WriteGovernance(number, gov.currentSet, tempSet); err != nil {
				logger.Error("Failed to store new governance data", "err", err)
			}
		}
		gov.ClearVotes(number)
		gov.updateCurrentGovernance(number)
	}
}

func (gov *Governance) removeDuplicatedVote(vote *GovernanceVote, number uint64) {
	gov.RemoveVote(vote.Key, vote.Value, number)
}

func (gov *Governance) updateCurrentGovernance(num uint64) {
	newNumber, newGovernanceSet, _ := gov.ReadGovernance(num)

	gov.actualGovernanceBlock = newNumber
	gov.currentSet = newGovernanceSet
	gov.triggerChange(newGovernanceSet)
	gov.currentGovernanceBlock = CalcGovernanceInfoBlock(num, gov.currentSet["istanbul.epoch"].(uint64))
}

func (gov *Governance) triggerChange(set GovernanceSet) {
	for k, v := range set {
		GovernanceItems[GovernanceKeyMap[k]].trigger(gov, k, v)
	}
}

func adjustDecodedSet(set GovernanceSet) GovernanceSet {
	for k, v := range set {
		x := reflect.ValueOf(v)
		if x.Kind() == reflect.Float64 {
			set[k] = uint64(v.(float64))
		}
	}
	return set
}

func (gov *Governance) GetGovernanceValue(key string) interface{} {
	if v, ok := gov.currentSet[key]; !ok {
		return nil
	} else {
		return v
	}
}

func (gov *Governance) VerifyGovernance(received []byte) error {
	change := []byte{}
	if rlp.DecodeBytes(received, &change) != nil {
		return errors.New("Failed to decode received governance changes")
	}

	rChangeSet := make(GovernanceSet)
	if json.Unmarshal(change, &rChangeSet) != nil {
		return errors.New("Failed to unmarshal received governance changes")
	}
	rChangeSet = adjustDecodedSet(rChangeSet)

	if len(rChangeSet) == len(gov.changeSet) {
		for k, v := range rChangeSet {
			if GovernanceKeyMap[k] == params.GoverningNode {
				if reflect.TypeOf(v) == stringT {
					v = common.HexToAddress(v.(string))
				}
			}
			if gov.changeSet[k] != v {
				logger.Error("Verification Error", "key", k, "received", rChangeSet[k], "have", gov.changeSet[k], "receivedType", reflect.TypeOf(rChangeSet[k]), "haveType", reflect.TypeOf(gov.changeSet[k]))
				return errors.New("Received change mismatches with the value this node has!!")
			}
		}
	}
	return nil
}
