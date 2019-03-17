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
	"github.com/ground-x/klaytn/params"
	"sync/atomic"
)

type PublicGovernanceAPI struct {
	governance *Governance // Node interfaced by this API
}

type returnTally struct {
	Key                string
	Value              interface{}
	ApprovalPercentage float64
}

func NewGovernanceAPI(gov *Governance) *PublicGovernanceAPI {
	return &PublicGovernanceAPI{governance: gov}
}

// Vote injects a new vote for governance targets such as unitprice and governingnode.
func (api *PublicGovernanceAPI) Vote(key string, val interface{}) interface{} {
	gMode := api.governance.chainConfig.Governance.GovernanceMode
	gNode := api.governance.chainConfig.Governance.GoverningNode

	if GovernanceModeMap[gMode] == params.GovernanceMode_Single && gNode != api.governance.nodeAddress {
		return "You don't have the right to vote"
	}
	if api.governance.AddVote(key, val) {
		return "Your vote was successfully placed."
	}
	return "Your vote couldn't be placed. Please check your vote's key and value"
}

func (api *PublicGovernanceAPI) ShowTally() interface{} {
	ret := []*returnTally{}

	if !api.isGovernanceModeBallot() {
		return "In current governance mode, the tally is not available"
	}

	api.governance.GovernanceTallyLock.RLock()
	defer api.governance.GovernanceTallyLock.RUnlock()
	for _, val := range api.governance.GovernanceTally {
		item := &returnTally{
			Key:                val.Key,
			Value:              val.Value,
			ApprovalPercentage: float64(val.Votes) / float64(atomic.LoadUint64(&api.governance.totalVotingPower)) * 100,
		}
		ret = append(ret, item)
	}

	return ret
}

func (api *PublicGovernanceAPI) TotalVotingPower() interface{} {
	if !api.isGovernanceModeBallot() {
		return "In current governance mode, voting power is not available"
	}
	return float64(atomic.LoadUint64(&api.governance.totalVotingPower)) / 1000.0
}

func (api *PublicGovernanceAPI) MyVotes() map[string]interface{} {
	return api.governance.voteMap
}

func (api *PublicGovernanceAPI) MyVotingPower() interface{} {
	if !api.isGovernanceModeBallot() {
		return "In current governance mode, voting power is not available"
	}
	return float64(atomic.LoadUint64(&api.governance.votingPower)) / 1000.0
}

func (api *PublicGovernanceAPI) ChainConfig() interface{} {
	return api.governance.chainConfig
}

func (api *PublicGovernanceAPI) NodeAddress() interface{} {
	return api.governance.nodeAddress
}

func (api *PublicGovernanceAPI) isGovernanceModeBallot() bool {
	if GovernanceModeMap[api.governance.chainConfig.Governance.GovernanceMode] == params.GovernanceMode_Ballot {
		return true
	}
	return false
}
