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

//go:generate abigen --sol contract/KlaytnReward.sol --pkg contract --out contract/KlaytnReward.go
//go:generate abigen --sol contract/CommitteeReward.sol --pkg contract --out contract/CommitteeReward.go
//go:generate abigen --sol contract/ProposerReward.sol --pkg contract --out contract/ProposerReward.go
//go:generate abigen --sol contract/PIRerve.sol --pkg contract --out contract/PIRerve.go
//go:generate abigen --sol contract/AddressBook.sol --pkg contract --out contract/AddressBook.go

package reward

import (
	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/vm"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/contracts/reward/contract"
	"github.com/ground-x/klaytn/log"
	"github.com/ground-x/klaytn/params"
	"math/big"
	"strings"
)

var logger = log.NewModuleLogger(log.Reward)

const (
	// TODO-Klaytn-Issue1166 We use small number for testing. We have to decide staking interval for real network.
	StakingUpdateInterval uint64 = 16
)

type Reward struct {
	*contract.KlaytnRewardSession
	contractBackend bind.ContractBackend
}

func NewReward(transactOpts *bind.TransactOpts, contractAddr common.Address, contractBackend bind.ContractBackend) (*Reward, error) {
	klaytnReward, err := contract.NewKlaytnReward(contractAddr, contractBackend)
	if err != nil {
		return nil, err
	}

	return &Reward{
		&contract.KlaytnRewardSession{
			Contract:     klaytnReward,
			TransactOpts: *transactOpts,
		},
		contractBackend,
	}, nil
}

func DeployReward(transactOpts *bind.TransactOpts, contractBackend bind.ContractBackend) (common.Address, *Reward, error) {

	rewardAddr, _, _, err := contract.DeployKlaytnReward(transactOpts, contractBackend)
	if err != nil {
		return rewardAddr, nil, err
	}

	reward, err := NewReward(transactOpts, rewardAddr, contractBackend)
	if err != nil {
		return rewardAddr, nil, err
	}

	return rewardAddr, reward, nil
}

type BalanceAdder interface {
	AddBalance(addr common.Address, v *big.Int)
}

// MintKLAY mints KLAY and deposits newly minted KLAY to three predefined accounts, i.e. Reward contract, KIR contract, PoC contract.
func MintKLAY(b BalanceAdder) {
	// TODO-Klaytn-Issue973 Developing Klaytn token economy
	b.AddBalance(common.HexToAddress(contract.RewardContractAddress), params.RewardContractIncentive)
	b.AddBalance(common.HexToAddress(contract.KIRContractAddress), params.KIRContractIncentive)
	b.AddBalance(common.HexToAddress(contract.PoCContractAddress), params.PoCContractIncentive)
}

func isEmptyAddress(addr common.Address) bool {
	return addr == common.Address{}
}

// DistributeBlockReward mints KLAY and distribute newly minted KLAY to proposer, kirAddr and pocAddr. proposer also gets totalTxFee.
func DistributeBlockReward(b BalanceAdder, validators []common.Address, totalTxFee *big.Int, kirAddr common.Address, pocAddr common.Address) {
	proposer := validators[0]

	// Mint KLAY for proposer
	mintedKLAY := params.ProposerIncentive
	b.AddBalance(proposer, mintedKLAY)
	logger.Debug("Block reward - Minted KLAY", "reward address of proposer", proposer, "Amount", mintedKLAY)

	// Transfer Tx fee for proposer
	b.AddBalance(proposer, totalTxFee)
	logger.Debug("Block reward - Tx fee", "reward address of proposer", proposer, "Amount", totalTxFee)

	// Mint KLAY for KIR
	if isEmptyAddress(kirAddr) {
		// Consider bootstrapping
		b.AddBalance(proposer, params.KIRContractIncentive)
		logger.Debug("Block reward - KIR. No KIR address.", "reward address of proposer", proposer, "Amount", params.KIRContractIncentive)
	} else {
		b.AddBalance(kirAddr, params.KIRContractIncentive)
		logger.Debug("Block reward - KIR", "KIR address", kirAddr, "Amount", params.KIRContractIncentive)
	}

	// Mint KLAY for PoC
	if isEmptyAddress(pocAddr) {
		// Consider bootstrapping
		b.AddBalance(proposer, params.KIRContractIncentive)
		logger.Debug("Block reward - PoC. No PoC address.", "reward address of proposer", proposer, "Amount", params.PoCContractIncentive)
	} else {
		b.AddBalance(pocAddr, params.PoCContractIncentive)
		logger.Debug("Block reward - PoC", "PoC address", pocAddr, "Amount", params.PoCContractIncentive)
	}
}

func IsStakingUpdateInterval(blockNum uint64) bool {
	return (blockNum % StakingUpdateInterval) == 0
}

// CalcStakingBlockNumber returns number of block which contains staking information required to make a new block with blockNum.
func CalcStakingBlockNumber(blockNum uint64) uint64 {
	if blockNum < 2*StakingUpdateInterval {
		// Bootstrapping. Just return genesis block number.
		return 0
	}
	number := blockNum - StakingUpdateInterval - (blockNum % StakingUpdateInterval)
	return number
}

func CalcProposerBlockNumber(blockNum uint64) uint64 {
	number := blockNum - (blockNum % StakingUpdateInterval)
	return number
}

// StakingCache
const (
	// TODO-Klaytn-Issue1166 Decide size of cache
	maxStakingCache = 3
)

var StakingCache common.Cache // TODO-Klaytn-Issue1166 Cache for staking information of Council

func init() {
	initStakingCache()
}

func initStakingCache() {
	StakingCache, _ = common.NewCache(common.LRUConfig{CacheSize: maxStakingCache})
}

// GetStakingInfoFromStakingCache returns corresponding staking information for a block of blockNum.
func GetStakingInfoFromStakingCache(blockNum uint64) *common.StakingInfo {
	number := CalcStakingBlockNumber(blockNum)
	stakingCacheKey := common.StakingCacheKey(number)
	value, ok := StakingCache.Get(stakingCacheKey)
	if !ok {
		logger.Error("Staking cache missed", "Block number", blockNum, "cache key", stakingCacheKey)
		return nil
	}

	stakingInfo, ok := value.(*common.StakingInfo)
	if !ok {
		logger.Error("Found staking information is invalid", "Block number", blockNum, "cache key", stakingCacheKey)
		return nil
	}

	if stakingInfo.BlockNum != number {
		logger.Error("Staking cache hit. But staking information not found", "Block number", blockNum, "cache key", stakingCacheKey)
		return nil
	}

	logger.Debug("Staking cache hit.", "Block number", blockNum, "stakingInfo", stakingInfo, "cache key", stakingCacheKey)
	return stakingInfo
}

func MakeGetAllAddressInfoMsg() (*types.Transaction, error) {
	abiStr := contract.AddressBookABI
	abii, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, err
	}

	data, err := abii.Pack("getAllAddressInfo")
	if err != nil {
		return nil, err
	}

	intrinsicGas, err := types.IntrinsicGas(data, false, true)
	if err != nil {
		return nil, err
	}

	addr := common.HexToAddress(contract.AddressBookAddress)

	// Create new call message
	// TODO-Klaytn-Issue1166 Decide who will be sender(i.e. from)
	msg := types.NewMessage(common.Address{}, &addr, 0, big.NewInt(0), 10000000, big.NewInt(0), data, false, intrinsicGas)

	return msg, nil
}

func ParseGetAllAddressInfo(result []byte) ([]common.Address, []common.Address, []common.Address, common.Address, common.Address, error) {

	// TODO-Klaytn-Issue1166 Disable all below Trace log later after all block reward implementation merged

	if result == nil {
		logger.Debug("ParseGetAllAddressInfo() Got empty result", "result", result)
		return nil, nil, nil, common.Address{}, common.Address{}, nil
	}

	abiStr := contract.AddressBookABI
	abii, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		logger.Trace("ParseGetAllAddressInfo() failed to make ABI interface.")
		return nil, nil, nil, common.Address{}, common.Address{}, err
	}

	var (
		ret0 = new([]common.Address)
		ret1 = new([]common.Address)
		ret2 = new([]common.Address)
		ret3 = new(common.Address)
		ret4 = new(common.Address)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
		ret4,
	}

	err = abii.Unpack(out, "getAllAddressInfo", result)
	if err != nil {
		logger.Trace("ParseGetAllAddressInfo() abii.Unpack failed")
		return nil, nil, nil, common.Address{}, common.Address{}, err
	}

	return *ret0, *ret1, *ret2, *ret3, *ret4, nil
}

// updateStakingCache updates staking cache with staking information of given block number.
func updateStakingCache(bc *blockchain.BlockChain, blockNum uint64) error {

	// TODO-Klaytn-Issue1166 Disable all below Trace log later after all block reward implementation merged

	stakingInfo, err := getAddressBookInfo(bc, blockNum)
	if err != nil {
		logger.Trace("Failed to get staking info", "blockNum", blockNum, "err", err)
		return err
	}

	stakingCacheKey := common.StakingCacheKey(blockNum)
	evicted := StakingCache.Add(stakingCacheKey, stakingInfo)
	logger.Trace("updateStakingCache() -  add new staking info", "stakingInfo", stakingInfo, "evicted", evicted)

	return nil
}

func getAddressBookInfo(bc *blockchain.BlockChain, blockNum uint64) (*common.StakingInfo, error) {

	// TODO-Klaytn-Issue1166 Disable all below Trace log later after all block reward implementation merged

	var nodeIds []common.Address
	var stakingAddrs []common.Address
	var rewardAddrs []common.Address
	var KIRAddr = common.Address{}
	var PoCAddr = common.Address{}
	var err error

	if !IsStakingUpdateInterval(blockNum) {
		logger.Trace("Invalid block number.", "blockNum", blockNum)
		return nil, err
	}

	// Prepare a message
	msg, err := MakeGetAllAddressInfoMsg()
	if err != nil {
		logger.Trace("Failed to make message for AddressBook Contract", "err", err)
		return nil, err
	}

	// Prepare
	chainConfig := bc.Config()
	intervalBlock := bc.GetBlockByNumber(blockNum)
	gaspool := new(blockchain.GasPool).AddGas(intervalBlock.GasLimit())
	statedb, err := bc.StateAt(intervalBlock.Root())
	if err != nil {
		logger.Trace("Failed to make a state for interval block", "interval blockNum", blockNum, "err", err)
		return nil, err
	}

	// Create a new context to be used in the EVM environment
	context := blockchain.NewEVMContext(msg, intervalBlock.Header(), bc, nil)
	evm := vm.NewEVM(context, statedb, chainConfig, &vm.Config{})

	res, gas, kerr := blockchain.ApplyMessage(evm, msg, gaspool)
	logger.Trace("Call AddressBook contract", "result", res, "used gas", gas, "kerr", kerr)
	err = kerr.Err
	if err != nil {
		logger.Trace("Failed to call AddressBook contract", "err", err)
		return nil, err
	}

	nodeIds, stakingAddrs, rewardAddrs, PoCAddr, KIRAddr, err = ParseGetAllAddressInfo(res)
	if err != nil {
		logger.Trace("Failed to parse result from AddressBook contract", "err", err)
		return nil, err
	}

	// TODO-Klaytn-Issue1166 Disable Trace log later
	logger.Trace("Result from AddressBook contract", "nodeIds", nodeIds)
	logger.Trace("Result from AddressBook contract", "stakingAddrs", stakingAddrs)
	logger.Trace("Result from AddressBook contract", "rewardAddrs", rewardAddrs)
	logger.Trace("Result from AddressBook contract", "KIRAddr", KIRAddr, "PoCAddr", PoCAddr)

	return newStakingInfo(bc, blockNum, nodeIds, stakingAddrs, rewardAddrs, KIRAddr, PoCAddr)
}

func newStakingInfo(bc *blockchain.BlockChain, blockNum uint64, nodeIds []common.Address, stakingAddrs []common.Address, rewardAddrs []common.Address, KIRAddr common.Address, PoCAddr common.Address) (*common.StakingInfo, error) {

	// TODO-Klaytn-Issue1166 Disable all below Trace log later after all block reward implementation merged

	// Prepare
	intervalBlock := bc.GetBlockByNumber(blockNum)
	statedb, err := bc.StateAt(intervalBlock.Root())
	if err != nil {
		logger.Trace("Failed to make a state for interval block", "interval blockNum", blockNum, "err", err)
		return nil, err
	}

	// Get balance of rewardAddrs
	var stakingAmounts []*big.Int
	stakingAmounts = make([]*big.Int, len(stakingAddrs))
	for i, stakingAddr := range stakingAddrs {
		stakingAmounts[i] = statedb.GetBalance(stakingAddr)
		logger.Trace("Get staking amounts", "i", i, "stakingAddr", stakingAddr.String(), "stakingAmount", stakingAmounts[i])
	}

	stakingInfo := &common.StakingInfo{
		BlockNum:              blockNum,
		CouncilNodeIds:        nodeIds,
		CouncilStakingdAddrs:  stakingAddrs,
		CouncilRewardAddrs:    rewardAddrs,
		KIRAddr:               KIRAddr,
		PoCAddr:               PoCAddr,
		CouncilStakingAmounts: stakingAmounts,
	}
	return stakingInfo, nil
}
