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
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"math/big"
	"reflect"
	"testing"
)

type voteValue struct {
	k string
	v interface{}
	e bool
}

var tstData = []voteValue{
	{k: "epoch", v: uint64(30000), e: true},
	{k: "epoch", v: "bad", e: false},
	{k: "epoch", v: float64(30000.00), e: true},
	{k: "Epoch", v: float64(30000.10), e: false},
	{k: "sub", v: uint64(7), e: true},
	{k: "sub", v: float64(7.0), e: true},
	{k: "sub", v: float64(7.1), e: false},
	{k: "sub", v: "7", e: false},
	{k: "policy", v: "roundrobin", e: true},
	{k: "policy", v: "RoundRobin", e: true},
	{k: "policy", v: "sticky", e: true},
	{k: "policy", v: "weightedrandom", e: true},
	{k: "policy", v: "WeightedRandom", e: true},
	{k: "policy", v: uint64(0), e: false},
	{k: "policy", v: uint64(1), e: false},
	{k: "policy", v: uint64(2), e: false},
	{k: "policy", v: uint64(3), e: false},
	{k: "policy", v: float64(1.2), e: false},
	{k: "policy", v: float64(1.0), e: false},
	{k: "governancemode", v: "none", e: true},
	{k: "governancemode", v: "single", e: true},
	{k: "governancemode", v: "ballot", e: true},
	{k: "governancemode", v: 0, e: false},
	{k: "governancemode", v: 1, e: false},
	{k: "governancemode", v: 2, e: false},
	{k: "governancemode", v: "unexpected", e: false},
	{k: "governingnode", v: "0x00000000000000000000", e: false},
	{k: "governingnode", v: "0x0000000000000000000000000000000000000000", e: true},
	{k: "governingnode", v: "0x000000000000000000000000000abcd000000000", e: true},
	{k: "governingnode", v: "000000000000000000000000000abcd000000000", e: true},
	{k: "governingnode", v: common.HexToAddress("000000000000000000000000000abcd000000000"), e: true},
	{k: "governingnode", v: "0x000000000000000000000000000xxxx000000000", e: false},
	{k: "governingnode", v: "address", e: false},
	{k: "governingnode", v: 0, e: false},
	{k: "unitprice", v: float64(0.0), e: true},
	{k: "unitprice", v: uint64(25000000000), e: true},
	{k: "unitprice", v: float64(-10), e: false},
	{k: "unitprice", v: "25000000000", e: false},
	{k: "useginicoeff", v: false, e: true},
	{k: "useginicoeff", v: true, e: true},
	{k: "useginicoeff", v: "true", e: false},
	{k: "useginicoeff", v: 0, e: false},
	{k: "useginicoeff", v: 1, e: false},
	{k: "mintingamount", v: "9600000000000000000", e: true},
	{k: "mintingamount", v: "0", e: true},
	{k: "mintingamount", v: 96000, e: false},
	{k: "mintingamount", v: "many", e: false},
	{k: "ratio", v: "30/40/30", e: true},
	{k: "ratio", v: "10/10/80", e: true},
	{k: "ratio", v: "30/70", e: false},
	{k: "ratio", v: "30.5/40/29.5", e: false},
	{k: "ratio", v: "30.5/40/30.5", e: false},
}

var goodVotes = []voteValue{
	{k: "epoch", v: uint64(20000), e: true},
	{k: "sub", v: uint64(7), e: true},
	{k: "policy", v: "sticky", e: true},
	{k: "governancemode", v: "single", e: true},
	{k: "governingnode", v: "0x0000000000000000000000000000000000000000", e: true},
	{k: "unitprice", v: uint64(25000000000), e: true},
	{k: "useginicoeff", v: false, e: true},
	{k: "mintingamount", v: "9600000000000000000", e: true},
	{k: "ratio", v: "10/10/80", e: true},
}

func getTestConfig() *params.ChainConfig {
	config := params.TestChainConfig
	config.Governance = GetDefaultGovernanceConfig(UseIstanbul)
	config.Istanbul = &params.IstanbulConfig{
		Epoch:          config.Governance.Istanbul.Epoch,
		ProposerPolicy: config.Governance.Istanbul.ProposerPolicy,
		SubGroupSize:   config.Governance.Istanbul.SubGroupSize,
	}
	config.UnitPrice = config.Governance.UnitPrice
	return config
}

func getGovernance() *Governance {
	config := getTestConfig()
	return NewGovernance(config)
}

func TestNewGovernance(t *testing.T) {
	config := getTestConfig()
	tstGovernance := NewGovernance(config)

	if !reflect.DeepEqual(tstGovernance.chainConfig, config) {
		t.Errorf("New governance's config is not same as the given one")
	}
}

func TestGetDefaultGovernanceConfig(t *testing.T) {
	tstGovernance := GetDefaultGovernanceConfig(UseIstanbul)

	want := []interface{}{
		uint64(DefaultUnitPrice),
		DefaultUseGiniCoeff,
		DefaultRatio,
		DefaultSubGroupSize,
		uint64(DefaultProposerPolicy),
		uint64(DefaultEpoch),
		common.HexToAddress(DefaultGoverningNode),
		DefaultGovernanceMode,
		DefaultDefferedTxFee,
	}

	got := []interface{}{
		tstGovernance.UnitPrice,
		tstGovernance.Reward.UseGiniCoeff,
		tstGovernance.Reward.Ratio,
		tstGovernance.Istanbul.SubGroupSize,
		tstGovernance.Istanbul.ProposerPolicy,
		tstGovernance.Istanbul.Epoch,
		tstGovernance.GoverningNode,
		tstGovernance.GovernanceMode,
		tstGovernance.DeferredTxFee(),
	}

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Want %v, got %v", want, got)
	}

	if tstGovernance.Reward.MintingAmount.Cmp(big.NewInt(DefaultMintingAmount)) != 0 {
		t.Errorf("Default minting amount is not equal")
	}
}

func TestGovernance_CheckVoteValidity(t *testing.T) {
	gov := getGovernance()

	for _, val := range tstData {
		ret := gov.CheckVoteValidity(val.k, val.v)
		if ret != val.e {
			t.Errorf("Want %v, got %v for %v and %v", val.e, ret, val.k, val.v)
		}
	}
}

func TestGovernance_AddVote(t *testing.T) {
	gov := getGovernance()

	for _, val := range tstData {
		ret := gov.AddVote(val.k, val.v)
		if ret != val.e {
			t.Errorf("Want %v, got %v for %v and %v", val.e, ret, val.k, val.v)
		}
	}

	// Added 9 types of various votes and at least one of those are right value
	if len(gov.voteMap) != 9 {
		t.Errorf("Want 9, got %v", len(gov.voteMap))
	}
}

func TestGovernance_RemoveVote(t *testing.T) {
	gov := getGovernance()

	for _, val := range goodVotes {
		ret := gov.AddVote(val.k, val.v)
		if ret != val.e {
			t.Errorf("Want %v, got %v for %v and %v", val.e, ret, val.k, val.v)
		}
	}

	// Length check. Because []votes has all valid votes, length of voteMap and votes should be equal
	if len(gov.voteMap) != len(goodVotes) {
		t.Errorf("Length of voteMap should be %d, but %d\n", len(goodVotes), len(gov.voteMap))
	}

	// Remove unvoted vote. Length should still be same
	gov.RemoveVote("Epoch", uint64(10000))
	if len(gov.voteMap) != len(goodVotes) {
		t.Errorf("Length of voteMap should be %d, but %d\n", len(goodVotes), len(gov.voteMap))
	}

	// Remove vote with wrong key. Length should still be same
	gov.RemoveVote("EpochEpoch", uint64(10000))
	if len(gov.voteMap) != len(goodVotes) {
		t.Errorf("Length of voteMap should be %d, but %d\n", len(goodVotes), len(gov.voteMap))
	}

	// Removed a vote. Length should be len(goodVotes) -1
	gov.RemoveVote("epoch", uint64(20000))
	if len(gov.voteMap) != len(goodVotes)-1 {
		t.Errorf("Length of voteMap should be %d, but %d\n", len(goodVotes), len(gov.voteMap))
	}
}

func TestGovernance_ClearVotes(t *testing.T) {
	gov := getGovernance()

	for _, val := range tstData {
		ret := gov.AddVote(val.k, val.v)
		if ret != val.e {
			t.Errorf("Want %v, got %v for %v and %v", val.e, ret, val.k, val.v)
		}
	}
	gov.ClearVotes()
	if len(gov.voteMap) != 0 {
		t.Errorf("Want 0, got %v after clearing votes", len(gov.voteMap))
	}
}

func TestGovernance_GetEncodedVote(t *testing.T) {
	gov := getGovernance()

	for _, val := range goodVotes {
		_ = gov.AddVote(val.k, val.v)
	}

	for len(gov.voteMap) > 0 {
		voteData := gov.GetEncodedVote(common.HexToAddress("0x1234567890123456789012345678901234567890"))
		v := new(GovernanceVote)
		rlp.DecodeBytes(voteData, &v)
		v = gov.ParseVoteValue(v)

		if v.Value != gov.voteMap[v.Key] {
			t.Errorf("Encoded vote and Decoded vote are different! Encoded: %v, Decoded: %v\n", gov.voteMap[v.Key], v.Value)
		}
		gov.RemoveVote(v.Key, v.Value)
	}
}

func TestGovernance_ParseVoteValue(t *testing.T) {
	gov := getGovernance()

	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	for _, val := range goodVotes {
		v := &GovernanceVote{
			Key:       val.k,
			Value:     val.v,
			Validator: addr,
		}

		b, _ := rlp.EncodeToBytes(v)

		d := new(GovernanceVote)
		rlp.DecodeBytes(b, d)
		d = gov.ParseVoteValue(d)

		if !reflect.DeepEqual(v, d) {
			t.Errorf("Parse was not successful! %v %v \n", v, d)
		}
	}
}
