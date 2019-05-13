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
	"errors"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/networks/rpc"
	"github.com/ground-x/klaytn/params"
	"math/big"
	"reflect"
	"strings"
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

type GovernanceKlayAPI struct {
	governance *Governance
	chain      *blockchain.BlockChain
}

func NewGovernanceKlayAPI(gov *Governance, chain *blockchain.BlockChain) *GovernanceKlayAPI {
	return &GovernanceKlayAPI{governance: gov, chain: chain}
}

var (
	errUnknownBlock = errors.New("Unknown block")
)

func (api *GovernanceKlayAPI) GasPriceAt(num *rpc.BlockNumber) (*big.Int, error) {
	if num == nil || *num == rpc.LatestBlockNumber || *num == rpc.PendingBlockNumber {
		ret := api.governance.GetLatestGovernanceItem("governance.unitprice").(uint64)
		return big.NewInt(0).SetUint64(ret), nil
	} else {
		blockNum := num.Int64()

		if blockNum > api.chain.CurrentHeader().Number.Int64() {
			return nil, errUnknownBlock
		}

		if ret, err := api.GasPriceAtNumber(blockNum); err != nil {
			return nil, err
		} else {
			return big.NewInt(0).SetUint64(ret), nil
		}
	}
}

func (api *GovernanceKlayAPI) GasPrice() *big.Int {
	ret := api.governance.GetLatestGovernanceItem("governance.unitprice").(uint64)
	return big.NewInt(0).SetUint64(ret)
}

// Vote injects a new vote for governance targets such as unitprice and governingnode.
func (api *PublicGovernanceAPI) Vote(key string, val interface{}) interface{} {
	gMode := api.governance.ChainConfig.Governance.GovernanceMode
	gNode := api.governance.ChainConfig.Governance.GoverningNode

	if GovernanceModeMap[gMode] == params.GovernanceMode_Single && gNode != api.governance.nodeAddress {
		return "You don't have the right to vote"
	}
	if strings.ToLower(key) == "removevalidator" {
		if !api.isRemovingSelf(val) {
			return "You can't vote on removing yourself"
		}
	}
	if api.governance.AddVote(key, val) {
		return "Your vote was successfully placed."
	}
	return "Your vote couldn't be placed. Please check your vote's key and value"
}

func (api *PublicGovernanceAPI) isRemovingSelf(val interface{}) bool {
	if reflect.TypeOf(val).String() != "string" {
		return false
	}
	target := val.(string)
	if !common.IsHexAddress(target) {
		return false
	}
	if common.HexToAddress(target) == api.governance.nodeAddress {
		return false
	} else {
		return true
	}
}

func (api *PublicGovernanceAPI) ShowTally() interface{} {
	ret := []*returnTally{}

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

type VoteList struct {
	Key      string
	Value    interface{}
	Casted   bool
	BlockNum uint64
}

func (api *PublicGovernanceAPI) MyVotes() []*VoteList {

	ret := []*VoteList{}
	api.governance.voteMapLock.RLock()
	defer api.governance.voteMapLock.RUnlock()

	for k, v := range api.governance.voteMap {
		item := &VoteList{
			Key:      k,
			Value:    v.value,
			Casted:   v.casted,
			BlockNum: v.num,
		}
		ret = append(ret, item)
	}

	return ret
}

func (api *PublicGovernanceAPI) MyVotingPower() interface{} {
	if !api.isGovernanceModeBallot() {
		return "In current governance mode, voting power is not available"
	}
	return float64(atomic.LoadUint64(&api.governance.votingPower)) / 1000.0
}

func (api *PublicGovernanceAPI) ChainConfig() interface{} {
	return api.governance.ChainConfig
}

func (api *PublicGovernanceAPI) NodeAddress() interface{} {
	return api.governance.nodeAddress
}

func (api *PublicGovernanceAPI) isGovernanceModeBallot() bool {
	if GovernanceModeMap[api.governance.ChainConfig.Governance.GovernanceMode] == params.GovernanceMode_Ballot {
		return true
	}
	return false
}

func (api *GovernanceKlayAPI) GasPriceAtNumber(num int64) (uint64, error) {
	val, err := api.governance.GetGovernanceItemAtNumber(uint64(num), "governance.unitprice")
	if err != nil {
		return 0, err
	}
	return val.(uint64), nil
}
