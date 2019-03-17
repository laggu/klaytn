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
