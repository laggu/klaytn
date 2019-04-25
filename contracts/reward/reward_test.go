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

package reward

import (
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/governance"
	"github.com/ground-x/klaytn/params"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

type testAccount struct {
	balance *big.Int
}

type testAccounts struct {
	accounts map[common.Address]*testAccount
}

func (ta *testAccounts) AddBalance(addr common.Address, v *big.Int) {
	if account, ok := ta.accounts[addr]; ok {
		account.balance.Add(account.balance, v)
	} else {
		ta.accounts[addr] = &testAccount{new(big.Int).Set(v)}
	}
}

func (ta *testAccounts) GetBalance(addr common.Address) *big.Int {
	account := ta.accounts[addr]
	if account != nil {
		return account.balance
	} else {
		return nil
	}
}

func newTestAccounts() *testAccounts {
	return &testAccounts{
		accounts: make(map[common.Address]*testAccount),
	}
}

var (
	addr1 = common.HexToAddress("0xac5e047d39692be8c81d0724543d5de721d0dd54")
)

// TestBlockRewardWithDefaultGovernance1 tests DistributeBlockReward with DefaultGovernanceConfig.
func TestBlockRewardWithDefaultGovernance(t *testing.T) {
	// 1. DefaultGovernance
	allocBlockRewardCache()
	accounts := newTestAccounts()

	// header
	header := &types.Header{Number: big.NewInt(0)}
	proposerAddr := addr1
	header.Rewardbase = proposerAddr

	// chain config
	config := &params.ChainConfig{Istanbul: governance.GetDefaultIstanbulConfig(), Governance: governance.GetDefaultGovernanceConfig(params.UseIstanbul)}
	DistributeBlockReward(accounts, header, common.Address{}, common.Address{}, config)

	balance := accounts.GetBalance(proposerAddr)
	if balance == nil {
		t.Errorf("Fail to get balance from addr(%s)", proposerAddr.String())
	} else {
		assert.Equal(t, balance, config.Governance.Reward.MintingAmount)
	}

	// 2. DefaultGovernance and when there is used gas in block
	allocBlockRewardCache()
	accounts = newTestAccounts()

	// header
	header = &types.Header{Number: big.NewInt(0)}
	proposerAddr = addr1
	header.Rewardbase = proposerAddr
	header.GasUsed = uint64(100000)

	// chain config
	config = &params.ChainConfig{}
	config.Governance = governance.GetDefaultGovernanceConfig(params.UseIstanbul)
	config.Istanbul = governance.GetDefaultIstanbulConfig()

	DistributeBlockReward(accounts, header, common.Address{}, common.Address{}, config)

	balance = accounts.GetBalance(proposerAddr)
	if balance == nil {
		t.Errorf("Fail to get balance from addr(%s)", proposerAddr.String())
	} else {
		expectedBalance := config.Governance.Reward.MintingAmount
		assert.Equal(t, expectedBalance, balance)
	}
}

// TestBlockRewardWithDeferredTxFeeEnabled tests DistributeBlockReward when DeferredTxFee is true
func TestBlockRewardWithDeferredTxFeeEnabled(t *testing.T) {
	// 1. DefaultGovernance + header.GasUsed + DeferredTxFee True
	allocBlockRewardCache()
	accounts := newTestAccounts()

	// header
	header := &types.Header{Number: big.NewInt(0)}
	proposerAddr := addr1
	header.Rewardbase = proposerAddr
	header.GasUsed = uint64(100000)

	// chain config
	config := &params.ChainConfig{Istanbul: governance.GetDefaultIstanbulConfig(), Governance: governance.GetDefaultGovernanceConfig(params.UseIstanbul)}

	config.Governance.Reward.DeferredTxFee = true

	DistributeBlockReward(accounts, header, common.Address{}, common.Address{}, config)

	balance := accounts.GetBalance(proposerAddr)
	if balance == nil {
		t.Errorf("Fail to get balance from addr(%s)", proposerAddr.String())
	} else {
		gasUsed := new(big.Int).SetUint64(header.GasUsed)
		unitPrice := new(big.Int).SetUint64(config.Governance.UnitPrice)
		tmpInt := new(big.Int).Mul(gasUsed, unitPrice)
		expectedBalance := tmpInt.Add(tmpInt, config.Governance.Reward.MintingAmount)

		assert.Equal(t, expectedBalance, balance)
	}

	// 2. DefaultGovernance + header.GasUsed + DeferredTxFee True + params.DefaultMintedKLAY
	accounts = newTestAccounts()
	allocBlockRewardCache()

	// header
	header = &types.Header{Number: big.NewInt(0)}
	proposerAddr = addr1
	header.Rewardbase = proposerAddr
	header.GasUsed = uint64(100000)

	// chain config
	config = &params.ChainConfig{Istanbul: governance.GetDefaultIstanbulConfig(), Governance: governance.GetDefaultGovernanceConfig(params.UseIstanbul)}

	config.Governance.Reward.DeferredTxFee = true
	config.Governance.Reward.MintingAmount = params.DefaultMintedKLAY

	DistributeBlockReward(accounts, header, common.Address{}, common.Address{}, config)

	balance = accounts.GetBalance(proposerAddr)
	if balance == nil {
		t.Errorf("Fail to get balance from addr(%s)", proposerAddr.String())
	} else {
		gasUsed := new(big.Int).SetUint64(header.GasUsed)
		unitPrice := new(big.Int).SetUint64(config.Governance.UnitPrice)
		tmpInt := new(big.Int).Mul(gasUsed, unitPrice)
		expectedBalance := tmpInt.Add(tmpInt, config.Governance.Reward.MintingAmount)

		assert.Equal(t, expectedBalance, balance)
	}
}

func TestPocKirRewardDistribute(t *testing.T) {
	allocBlockRewardCache()

	accounts := newTestAccounts()
	header := &types.Header{Number: big.NewInt(0)}
	proposerAddr := addr1
	header.Rewardbase = proposerAddr
	header.GasUsed = uint64(100000)
	mintingAmount := big.NewInt(int64(1000000000))

	// chain config
	config := &params.ChainConfig{Istanbul: governance.GetDefaultIstanbulConfig(), Governance: governance.GetDefaultGovernanceConfig(params.UseIstanbul)}
	config.Governance.Reward.MintingAmount = mintingAmount
	config.Governance.Reward.Ratio = "70/20/10"

	pocAddr := common.StringToAddress("1111111111111111111111111111111111111111")
	kirAddr := common.StringToAddress("2222222222222222222222222222222222222222")

	distributeBlockReward(accounts, header, big.NewInt(0), pocAddr, kirAddr, config)

	cnBalance := accounts.GetBalance(proposerAddr)
	pocBalance := accounts.GetBalance(pocAddr)
	kirBalance := accounts.GetBalance(kirAddr)

	expectedKirBalance := big.NewInt(0).Div(mintingAmount, big.NewInt(10))     // 10%
	expectedPocBalance := big.NewInt(0).Mul(expectedKirBalance, big.NewInt(2)) // 20%
	expectedCnBalance := big.NewInt(0).Mul(expectedKirBalance, big.NewInt(7))  // 70%

	if expectedCnBalance.Cmp(cnBalance) != 0 || pocBalance.Cmp(expectedPocBalance) != 0 || kirBalance.Cmp(expectedKirBalance) != 0 {
		t.Errorf("balances are calculated incorrectly. CN Balance : %v, PoC Balance : %v, KIR Balance : %v, ratio : %v",
			cnBalance, pocBalance, kirBalance, config.Governance.Reward.Ratio)
	}

	totalBalance := big.NewInt(0).Add(cnBalance, pocBalance)
	totalBalance = big.NewInt(0).Add(totalBalance, kirBalance)

	if mintingAmount.Cmp(totalBalance) != 0 {
		t.Errorf("The sum of balance is diffrent from mintingAmount. totalBalance : %v, mintingAmount : %v", totalBalance, mintingAmount)
	}
}

// TestBlockRewardWithCustomRewardRatio tests DistributeBlockReward with reward ratio defined in params package.
func TestBlockRewardWithCustomRewardRatio(t *testing.T) {
	// 1. DefaultGovernance + header.GasUsed + DeferredTxFee True + params.DefaultMintedKLAY + DefaultCNRatio/DefaultKIRRewardRatio/DefaultPocRewardRatio
	accounts := newTestAccounts()
	allocBlockRewardCache()

	// header
	header := &types.Header{Number: big.NewInt(0)}
	proposerAddr := addr1
	header.Rewardbase = proposerAddr
	header.GasUsed = uint64(100000)

	// chain config
	config := &params.ChainConfig{Istanbul: governance.GetDefaultIstanbulConfig(), Governance: governance.GetDefaultGovernanceConfig(params.UseIstanbul)}

	config.Governance.Reward.DeferredTxFee = true
	config.Governance.Reward.MintingAmount = params.DefaultMintedKLAY
	config.Governance.Reward.Ratio = fmt.Sprintf("%d/%d/%d", params.DefaultCNRewardRatio, params.DefaultKIRRewardRatio, params.DefaultPoCRewardRatio)

	DistributeBlockReward(accounts, header, common.Address{}, common.Address{}, config)

	balance := accounts.GetBalance(proposerAddr)
	if balance == nil {
		t.Errorf("Fail to get balance from addr(%s)", proposerAddr.String())
	} else {
		gasUsed := new(big.Int).SetUint64(header.GasUsed)
		unitPrice := new(big.Int).SetUint64(config.Governance.UnitPrice)
		tmpInt := new(big.Int).Mul(gasUsed, unitPrice)
		expectedBalance := tmpInt.Add(tmpInt, config.Governance.Reward.MintingAmount)

		assert.Equal(t, expectedBalance, balance)
	}
}

func TestStakingInfoCache_Add(t *testing.T) {
	initStakingCache()

	// test cache limit
	for i := 1; i <= 10; i++ {
		testStakingInfo, _ := newEmptyStakingInfo(nil, uint64(i))
		stakingCache.add(testStakingInfo)

		if len(stakingCache.cells) > maxStakingCache {
			t.Errorf("over the max limit of staking cache. Current Len : %v, MaxStakingCache : %v", len(stakingCache.cells), maxStakingCache)
		}
	}

	// test adding same block number
	initStakingCache()
	testStakingInfo1, _ := newEmptyStakingInfo(nil, uint64(1))
	testStakingInfo2, _ := newEmptyStakingInfo(nil, uint64(1))
	stakingCache.add(testStakingInfo1)
	stakingCache.add(testStakingInfo2)

	if len(stakingCache.cells) > 1 {
		t.Errorf("StakingInfo with Same block number is saved to the cache stakingCache. result : %v, expected : %v ", len(stakingCache.cells), maxStakingCache)
	}

	// test minBlockNum
	initStakingCache()
	for i := 1; i < 5; i++ {
		testStakingInfo, _ := newEmptyStakingInfo(nil, uint64(i))
		stakingCache.add(testStakingInfo)
	}

	testStakingInfo, _ := newEmptyStakingInfo(nil, uint64(5))
	stakingCache.add(testStakingInfo) // blockNum 1 should be deleted
	if stakingCache.minBlockNum != 2 {
		t.Errorf("minBlockNum of staking cache is diffrent from expected blocknum. result : %v, expected : %v", stakingCache.minBlockNum, 2)
	}

	testStakingInfo, _ = newEmptyStakingInfo(nil, uint64(6))
	stakingCache.add(testStakingInfo) // blockNum 2 should be deleted
	if stakingCache.minBlockNum != 3 {
		t.Errorf("minBlockNum of staking cache is diffrent from expected blocknum. result : %v, expected : %v", stakingCache.minBlockNum, 3)
	}
}

func TestStakingInfoCache_Get(t *testing.T) {
	initStakingCache()

	for i := 1; i <= 4; i++ {
		testStakingInfo, _ := newEmptyStakingInfo(nil, uint64(i))
		stakingCache.add(testStakingInfo)
	}

	// should find correct stakingInfo with given block number
	for i := uint64(1); i <= 4; i++ {
		testStakingInfo := stakingCache.get(i)

		if testStakingInfo.BlockNum != i {
			t.Errorf("The block number of gotten staking info is diffrent. result : %v, expected : %v", testStakingInfo.BlockNum, i)
		}
	}

	// nothing should be found as no matched block number is in cache
	for i := uint64(5); i < 10; i++ {
		testStakingInfo := stakingCache.get(i)

		if testStakingInfo != nil {
			t.Errorf("The result should be nil. result : %v", testStakingInfo)
		}
	}
}
