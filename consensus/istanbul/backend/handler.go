// Copyright 2018 The klaytn Authors
// Copyright 2017 The go-ethereum Authors
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
// This file is derived from quorum/consensus/istanbul/backend/handler.go (2018/06/04).
// Modified and improved for the klaytn development.

package backend

import (
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/consensus"
	"github.com/ground-x/klaytn/consensus/istanbul"
	"github.com/ground-x/klaytn/governance"
	"github.com/ground-x/klaytn/networks/p2p"
	"github.com/ground-x/klaytn/node"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/hashicorp/golang-lru"
)

const (
	istanbulMsg = 0x11
)

var (
	// errDecodeFailed is returned when decode message fails
	errDecodeFailed = errors.New("fail to decode istanbul message")
)

// Protocol implements consensus.Engine.Protocol
func (sb *backend) Protocol() consensus.Protocol {
	return consensus.Protocol{
		Name:     "istanbul",
		Versions: []uint{64},
		//Lengths:  []uint64{18},
		//Lengths:  []uint64{19},  // add PoRMsg
		Lengths: []uint64{21},
	}
}

// HandleMsg implements consensus.Handler.HandleMsg
func (sb *backend) HandleMsg(addr common.Address, msg p2p.Msg) (bool, error) {
	sb.coreMu.Lock()
	defer sb.coreMu.Unlock()

	if msg.Code == istanbulMsg {
		if !sb.coreStarted {
			return true, istanbul.ErrStoppedEngine
		}

		var cmsg istanbul.ConsensusMsg

		//var data []byte
		if err := msg.Decode(&cmsg); err != nil {
			return true, errDecodeFailed
		}
		data := cmsg.Payload
		hash := istanbul.RLPHash(data)

		// Mark peer's message
		ms, ok := sb.recentMessages.Get(addr)
		var m *lru.ARCCache
		if ok {
			m, _ = ms.(*lru.ARCCache)
		} else {
			m, _ = lru.NewARC(inmemoryMessages)
			sb.recentMessages.Add(addr, m)
		}
		m.Add(hash, true)

		// Mark self known message
		if _, ok := sb.knownMessages.Get(hash); ok {
			return true, nil
		}
		sb.knownMessages.Add(hash, true)

		go sb.istanbulEventMux.Post(istanbul.MessageEvent{
			Payload: data,
			Hash:    cmsg.PrevHash,
		})

		return true, nil
	}
	return false, nil
}

func (sb *backend) ValidatePeerType(addr common.Address) error {
	// istanbul.Start vs try to connect by peer
	for sb.chain == nil {
		return errors.New("sb.chain is nil! --mine option might be missing")
	}
	for _, val := range sb.getValidators(sb.chain.CurrentHeader().Number.Uint64(), sb.chain.CurrentHeader().Hash()).List() {
		if addr == val.Address() {
			return nil
		}
	}
	return errors.New("invalid address")
}

// SetBroadcaster implements consensus.Handler.SetBroadcaster
func (sb *backend) SetBroadcaster(broadcaster consensus.Broadcaster, nodetype p2p.ConnType) {
	sb.broadcaster = broadcaster
	if nodetype == node.CONSENSUSNODE {
		sb.broadcaster.RegisterValidator(node.CONSENSUSNODE, sb)
	}
}

func (sb *backend) NewChainHead() error {
	sb.coreMu.RLock()
	defer sb.coreMu.RUnlock()
	if !sb.coreStarted {
		return istanbul.ErrStoppedEngine
	}

	// Start a goroutine to update governance information
	go sb.updateGovernance(sb.chain.CurrentHeader().Number.Uint64())

	go sb.istanbulEventMux.Post(istanbul.FinalCommittedEvent{})
	return nil
}

// updateGovernance manages GovernanceCache and updates the backend's Governance config
// To reduce timing issues, we are using GovernanceConfig which was decided at two Epochs ago
func (sb *backend) updateGovernance(curr uint64) {
	rem := curr % sb.config.Epoch

	// Current block has new Governance information in it's header
	if rem == 0 {
		// Save current header's governance data in cache
		_ = sb.makeGovernanceCacheFromHeader(curr)

		// Retrieve the cache older by single Epoch
		num := curr - sb.config.Epoch
		if config, ok := sb.getGovernanceCache(num); ok {
			sb.replaceGovernanceConfig(config)
		} else {
			ret := sb.makeGovernanceCacheFromHeader(num)
			sb.replaceGovernanceConfig(ret)
		}

		sb.lastGovernanceBlock = num
	} else {
		// Required Governance config's block number
		var num uint64

		// To prevent underflow, compare it first
		if curr > (rem + sb.config.Epoch) {
			num = curr - rem - sb.config.Epoch
		} else {
			num = 0
		}

		// If blocks passed more than an Epoch without making cache
		if num > sb.lastGovernanceBlock {
			ret := sb.makeGovernanceCacheFromHeader(num)
			sb.replaceGovernanceConfig(ret)
			sb.lastGovernanceBlock = num
		}
	}

	// if this block has a vote from this node, remove it
	if proposer, err := ecrecover(sb.chain.CurrentHeader()); err == nil {
		if len(sb.chain.CurrentHeader().Vote) > 0 && sb.address == proposer {
			sb.removeDuplicatedVote(sb.chain.CurrentHeader().Vote, sb.governance)
		}
	}
}

func (sb *backend) removeDuplicatedVote(rawvote []byte, gov *governance.Governance) {
	vote := new(governance.GovernanceVote)
	if err := rlp.DecodeBytes(rawvote, vote); err != nil {
		// Vote data might be corrupted
		logger.Error("Failed to decode vote data from header", "vote")
	} else {
		vote = gov.ParseVoteValue(vote)
		gov.RemoveVote(vote.Key, vote.Value)
	}
}

// getGovernanceCacheKey returns cache key of the given block number
func getGovernanceCacheKey(num uint64) common.GovernanceCacheKey {
	v := fmt.Sprintf("%v", num)
	return common.GovernanceCacheKey(params.GovernanceCachePrefix + "_" + v)
}

// getGovernanceCache returns cached governance config as a byte slice
func (sb *backend) getGovernanceCache(num uint64) ([]byte, bool) {
	cKey := getGovernanceCacheKey(num)

	if ret, ok := sb.GovernanceCache.Get(cKey); ok && ret != nil {
		return ret.([]byte), true
	}
	return nil, false
}

// makeGovernanceCacheFromHeader retrieves governance information from the given number of block header
// and store it in the cache
func (sb *backend) makeGovernanceCacheFromHeader(num uint64) []byte {
	head := sb.chain.GetHeaderByNumber(num)
	cKey := getGovernanceCacheKey(num)

	sb.GovernanceCache.Add(cKey, head.Governance)
	return head.Governance
}

// replaceGovernanceConfig updates backend's governance with rlp decoded governance information
func (sb *backend) replaceGovernanceConfig(g []byte) bool {
	newGovernance := params.GovernanceConfig{}
	if err := rlp.DecodeBytes(g, &newGovernance); err != nil {
		logger.Error("Failed to replace Governance Config", "err", err)
		return false

	} else {
		// deep copy new governance
		sb.chain.Config().Governance = newGovernance.Copy()
		sb.governance.ChainConfig.Governance = newGovernance.Copy()
		// TODO-Klaytn-Governance Code for compatibility
		// Need to be cleaned up when developers use same template for genesis.json
		sb.config.Epoch = newGovernance.Istanbul.Epoch
		sb.config.SubGroupSize = newGovernance.Istanbul.SubGroupSize
		sb.config.ProposerPolicy = istanbul.ProposerPolicy(newGovernance.Istanbul.ProposerPolicy)
		sb.chain.Config().Istanbul = newGovernance.Istanbul.Copy()
		sb.chain.Config().UnitPrice = newGovernance.UnitPrice
		logger.Info("Governance config updated", "config", sb.chain.Config())

		return true
	}
}
