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

package validator

import (
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/consensus/istanbul"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sort"
	"testing"
)

var (
	testAddrs = []common.Address{
		common.HexToAddress("0x0adBC7b05Da383157200a9Fa192285898aB2CaAc"),
		common.HexToAddress("0x371F315BeBe961776AC84B29e044b01074b93E69"),
		common.HexToAddress("0x5845EAa7ac251542Dc96fBaD09E3CAd3ec105a7a"),
		common.HexToAddress("0x63805D23fC86Aa16EFB157C036F226f3aa93099d"),
		common.HexToAddress("0x68E0DEf1e6beb308eF5FdF2e19dB2884571c465c"),
		common.HexToAddress("0x72E23aAe2Cc6eE54682bD67B6093F7b7971f3D2F"),
		common.HexToAddress("0x78B898e37A45069518775972AB8155493e69A2F0"),
		common.HexToAddress("0x8704Ffb473a16638ea42c7704995d6505102a4Ca"),
		common.HexToAddress("0x93d3Ce8940c7907b0C1c3898dF7Aa797C457cD0f"),
		common.HexToAddress("0x9a049EefC01aAE911F2B6F19d724dF9d3ca5cAe6"),
		common.HexToAddress("0xC14124d61fc940c7aF29F62438D1B54fD7FFB65B"),
		common.HexToAddress("0xc4cB0B3c2682C15D96739f9a13fE26f17c893f8f"),
		common.HexToAddress("0xd4aB68fcEC8Fa23856188163B131F3E443e09EF8"),
	}
	testRewardAddrs = []common.Address{
		common.HexToAddress("0x0a6e50a28f10CD9dba36DD9D3B95BaA32F9fe77a"),
		common.HexToAddress("0x23FB6C77E069BD6456181f48a9c77f3a3812E7e7"),
		common.HexToAddress("0x43d5e084D8A6c7FbCd0EbA9a517533fF384f0577"),
		common.HexToAddress("0x4d180C12FB3B061f44E91D30d574F78D1DeCAD90"),
		common.HexToAddress("0x53094cE69ea701bfb9D06239087d4CF09F127B78"),
		common.HexToAddress("0x5F2152bf0C97f1d2c3Ffec8A98FEEB1e50798090"),
		common.HexToAddress("0x653f42fb1F3474de222F7DDa2109250218989B19"),
		common.HexToAddress("0x93eaEAa38D534B52E7DB3AB939330022330cD427"),
		common.HexToAddress("0x96a0d7f6A82B860313FF8668b858aD4930d7B2d6"),
		common.HexToAddress("0xB89ff800C21b3334f0e355A73242bB4363cf6e10"),
		common.HexToAddress("0xDEeeF6fAC16f095Fa944E481F8e6c3b42ae3Cefa"),
		common.HexToAddress("0xbDE3Ee8c01484dDBD59a425457Ab138cf3aa0E11"),
		common.HexToAddress("0xdD5572A7aC7AB7407e8e4082dB442668C02924E3"),
	}
	testStakingAddrs = []common.Address{
		common.HexToAddress("0x3776A66698babFA24F0316e4363B2E6C95B09ceF"),
		common.HexToAddress("0x4d086A88329233E00158FEcbe7b38Dd8667Dd9f9"),
		common.HexToAddress("0x5d7d13278AEF56263B7d25d51E1B2519Ac0D656B"),
		common.HexToAddress("0x60fA2326f6C1A7a90Bd1B3c31Bd1A7f9Aed61443"),
		common.HexToAddress("0x681C55B2CD831D262C785e213a70e277D0226c79"),
		common.HexToAddress("0x6EeA09FF2bB16F1cD075c748E1684f1100085541"),
		common.HexToAddress("0x817617C3f09d08a5d475bf72b4723A755CD9b8c7"),
		common.HexToAddress("0x83F28D3512dC32701F375b112d0CB0810Cb736e4"),
		common.HexToAddress("0x92D03E4998fB3F91A1E24496EDCf625136037f9e"),
		common.HexToAddress("0xA0360cDC935A9f3bFe7Ad03D1C34989427ad239f"),
		common.HexToAddress("0xE2946677DcEEDDF36F1f6EA00421635804872D49"),
		common.HexToAddress("0xF246283a57A8018085AF39bdadFCC4aaC682e6dD"),
		common.HexToAddress("0xF3c6f39e231C7363F9B5F4d71b5EE7Eb1fB265d7"),
	}
	testVotingPowers = []uint64{
		1, 1, 1, 1, 1,
		1, 1, 1, 1, 1,
		1, 1, 1,
	}
	testZeroWeights = []int{
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0,
	}
	testPrevHash = common.HexToHash("0xf99eb1626cfa6db435c0836235942d7ccaa935f1ae247d3f1c21e495685f903a")

	testExpectedProposers = []common.Address{
		common.HexToAddress("0x8704Ffb473a16638ea42c7704995d6505102a4Ca"),
		common.HexToAddress("0xC14124d61fc940c7aF29F62438D1B54fD7FFB65B"),
		common.HexToAddress("0x72E23aAe2Cc6eE54682bD67B6093F7b7971f3D2F"),
		common.HexToAddress("0xc4cB0B3c2682C15D96739f9a13fE26f17c893f8f"),
		common.HexToAddress("0x68E0DEf1e6beb308eF5FdF2e19dB2884571c465c"),
		common.HexToAddress("0x371F315BeBe961776AC84B29e044b01074b93E69"),
		common.HexToAddress("0x63805D23fC86Aa16EFB157C036F226f3aa93099d"),
		common.HexToAddress("0x93d3Ce8940c7907b0C1c3898dF7Aa797C457cD0f"),
		common.HexToAddress("0x5845EAa7ac251542Dc96fBaD09E3CAd3ec105a7a"),
		common.HexToAddress("0x78B898e37A45069518775972AB8155493e69A2F0"),
		common.HexToAddress("0x0adBC7b05Da383157200a9Fa192285898aB2CaAc"),
		common.HexToAddress("0xd4aB68fcEC8Fa23856188163B131F3E443e09EF8"),
		common.HexToAddress("0x9a049EefC01aAE911F2B6F19d724dF9d3ca5cAe6"),
	}

	testNonZeroWeights = []int{
		1, 1, 2, 1, 1,
		1, 0, 3, 2, 1,
		0, 1, 5,
	}
)

func makeTestValidators(weights []int) (validators istanbul.Validators) {
	validators = make([]istanbul.Validator, len(testAddrs))
	for i := range testAddrs {
		validators[i] = newWeightedValidator(testAddrs[i], testRewardAddrs[i], testVotingPowers[i], weights[i])
	}
	sort.Sort(validators)

	return
}

func makeTestWeightedCouncil(weights []int) (valSet *weightedCouncil) {
	// prepare weighted council
	valSet = NewWeightedCouncil(testAddrs, testRewardAddrs, testVotingPowers, weights, istanbul.WeightedRandom, 21, 0, 0, nil)
	return
}

func TestWeightedCouncil_RefreshWithZeroWeight(t *testing.T) {

	validators := makeTestValidators(testZeroWeights)

	valSet := makeTestWeightedCouncil(testZeroWeights)
	valSet.Refresh(testPrevHash, 1)

	// Run tests

	// 1. check all validators are chosen for proposers
	var sortedProposers istanbul.Validators
	sortedProposers = make([]istanbul.Validator, len(testAddrs))
	copy(sortedProposers, valSet.proposers)
	sort.Sort(sortedProposers)
	if !reflect.DeepEqual(sortedProposers, validators) {
		t.Errorf("All validators are not in proposers: sorted proposers %v, validators %v", sortedProposers, validators)
	}

	// 2. check proposers
	for i, val := range valSet.proposers {
		if !reflect.DeepEqual(val.Address(), testExpectedProposers[i]) {
			t.Errorf("proposer mismatch: have %v, want %v", val.Address().String(), testExpectedProposers[i].String())
		}
	}

	// 3. test calculate proposer different round
	checkCalcProposerWithRound(t, valSet, testAddrs[0], 0)
	checkCalcProposerWithRound(t, valSet, testAddrs[0], 1)
	checkCalcProposerWithRound(t, valSet, testAddrs[0], 5)
	checkCalcProposerWithRound(t, valSet, testAddrs[0], 13)
	checkCalcProposerWithRound(t, valSet, testAddrs[0], 1000)
}

func checkCalcProposerWithRound(t *testing.T, valSet *weightedCouncil, lastProposer common.Address, round uint64) {
	valSet.CalcProposer(lastProposer, round)
	_, expectedVal := valSet.GetByAddress(testExpectedProposers[round%uint64(len(valSet.proposers))])
	if val := valSet.GetProposer(); !reflect.DeepEqual(val, expectedVal) {
		t.Errorf("proposer mismatch: have %v, want %v", val.String(), expectedVal.Address().String())
	}
}

func TestWeightedCouncil_RefreshWithNonZeroWeight(t *testing.T) {

	validators := makeTestValidators(testNonZeroWeights)

	valSet := makeTestWeightedCouncil(testNonZeroWeights)
	valSet.Refresh(testPrevHash, 1)

	// Run tests

	// 1. number of proposers
	totalWeights := int(0)
	for _, v := range validators {
		totalWeights += v.Weight()
	}
	assert.Equal(t, totalWeights, len(valSet.proposers))

	// 2. weight and appearance frequency
	for _, v := range validators {
		weight := v.Weight()
		appearance := 0
		for _, p := range valSet.proposers {
			if v.Address() == p.Address() {
				appearance++
			}
		}
		assert.Equal(t, weight, appearance)
	}
}

func TestWeightedCouncil_RemoveValidator(t *testing.T) {

	validators := makeTestValidators(testNonZeroWeights)

	valSet := makeTestWeightedCouncil(testNonZeroWeights)
	valSet.Refresh(testPrevHash, 1)

	for _, val := range validators {

		_, removedVal := valSet.GetByAddress(val.Address())
		if removedVal == nil {
			t.Errorf("Fail to find validator with address %v", removedVal.Address().String())
		}

		if !valSet.RemoveValidator(removedVal.Address()) {
			t.Errorf("Fail to remove validator %v", removedVal.String())
		}

		// check whether removedVal is really removed from validators
		for _, v := range valSet.validators {
			if removedVal.Address() == v.Address() {
				t.Errorf("Validator(%v) does not removed from validators", removedVal.Address().String())
			}
		}

		// check whether removedVal is also removed from proposers immediately
		for _, p := range valSet.proposers {
			if removedVal.Address() == p.Address() {
				t.Errorf("Validator(%v) does not removed from proposers", removedVal.Address().String())
			}
		}
	}

	assert.Equal(t, 0, valSet.Size())
	assert.Equal(t, 0, len(valSet.Proposers()))
}

func TestWeightedCouncil_RefreshAfterRemoveValidator(t *testing.T) {

	validators := makeTestValidators(testNonZeroWeights)

	valSet := makeTestWeightedCouncil(testNonZeroWeights)
	valSet.Refresh(testPrevHash, 1)

	for _, val := range validators {

		_, removedVal := valSet.GetByAddress(val.Address())
		if removedVal == nil {
			t.Errorf("Fail to find validator with address %v", removedVal.Address().String())
		}

		if !valSet.RemoveValidator(removedVal.Address()) {
			t.Errorf("Fail to remove validator %v", removedVal.String())
		}

		// check whether removedVal is really removed from validators
		for _, v := range valSet.validators {
			if removedVal.Address() == v.Address() {
				t.Errorf("Validator(%v) does not removed from validators", removedVal.Address().String())
			}
		}

		valSet.Refresh(testPrevHash, 1)

		// check whether removedVal is excluded as expected when refreshing proposers
		for _, p := range valSet.proposers {
			if removedVal.Address() == p.Address() {
				t.Errorf("Validator(%v) does not removed from proposers", removedVal.Address().String())
			}
		}
	}

	assert.Equal(t, 0, valSet.Size())
	assert.Equal(t, 0, len(valSet.Proposers()))
}
