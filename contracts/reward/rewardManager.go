//go:generate abigen --sol contract/KlaytnReward.sol --pkg contract --out contract/KlaytnReward.go
//go:generate abigen --sol contract/AddressBook.sol --pkg contract --out contract/AddressBook.go

package reward

import (
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/log"
	"math/big"
)

var logger = log.NewModuleLogger(log.Reward)

type BalanceAdder interface {
	AddBalance(addr common.Address, v *big.Int)
}

type ConfigManager interface {
	UnitPrice() uint64
	Epoch() uint64
	ProposerPolicy() uint64
	MintingAmount() string
	Ratio() string
	UseGiniCoeff() bool
	ChainId() uint64
	GetGovernanceItemAtNumber(num uint64, key string) (interface{}, error)
	DeferredTxFee() bool
}

func isEmptyAddress(addr common.Address) bool {
	return addr == common.Address{}
}

type RewardManager struct {
	stakingManager    *StakingManager
	rewardConfigCache *rewardConfig
	configManager     ConfigManager
}

func NewRewardManager(bc *blockchain.BlockChain, configManager ConfigManager) *RewardManager {
	stakingManager := NewStakingManager(bc)
	rewardConfig := NewRewardConfig()
	return &RewardManager{
		stakingManager:    stakingManager,
		rewardConfigCache: rewardConfig,
		configManager:     configManager,
	}
}

// MintKLAY mints KLAY and gives the KLAY to the block proposer
func (rm *RewardManager) MintKLAY(b BalanceAdder, header *types.Header) error {

	unitPrice := big.NewInt(0)
	// use key only for temporary it should be removed after changing the way of getting configure
	if r, err := rm.configManager.GetGovernanceItemAtNumber(header.Number.Uint64(), "governance.unitprice"); err == nil {
		unitPrice.SetUint64(r.(uint64))
	} else {
		logger.Error("Couldn't get UnitPrice from governance", "err", err, "received", r)
		return err
	}

	mintingAmount := big.NewInt(0)
	// use key only for temporary it should be removed after changing the way of getting configure
	if r, err := rm.configManager.GetGovernanceItemAtNumber(header.Number.Uint64(), "reward.mintingamount"); err == nil {
		mintingAmount.SetString(r.(string), 10)
	} else {
		logger.Error("Couldn't get MintingAmount from governance", "err", err, "received", r)
		return err
	}

	totalGasUsed := big.NewInt(0).SetUint64(header.GasUsed)
	totalTxFee := big.NewInt(0).Mul(totalGasUsed, unitPrice)
	blockReward := big.NewInt(0).Add(mintingAmount, totalTxFee)

	b.AddBalance(header.Rewardbase, blockReward)
	return nil
}

// DistributeBlockReward distributes block reward to proposer, kirAddr and pocAddr.
func (rm *RewardManager) DistributeBlockReward(b BalanceAdder, header *types.Header, pocAddr common.Address, kirAddr common.Address) {

	// Calculate total tx fee
	totalTxFee := common.Big0
	if rm.configManager.DeferredTxFee() {
		totalGasUsed := big.NewInt(0).SetUint64(header.GasUsed)
		unitPrice := big.NewInt(0).SetUint64(rm.configManager.UnitPrice())
		totalTxFee = big.NewInt(0).Mul(totalGasUsed, unitPrice)
	}

	rm.distributeBlockReward(b, header, totalTxFee, pocAddr, kirAddr)
}

// distributeBlockReward mints KLAY and distribute newly minted KLAY and transaction fee to proposer, kirAddr and pocAddr.
func (rm *RewardManager) distributeBlockReward(b BalanceAdder, header *types.Header, totalTxFee *big.Int, pocAddr common.Address, kirAddr common.Address) {
	proposer := header.Rewardbase
	rewardParams := rm.rewardConfigCache.getRewardConfigCache(rm.configManager, header)

	// Block reward
	blockReward := big.NewInt(0).Add(rewardParams.mintingAmount, totalTxFee)

	tmpInt := big.NewInt(0)

	tmpInt = tmpInt.Mul(blockReward, rewardParams.cnRewardRatio)
	cnReward := big.NewInt(0).Div(tmpInt, rewardParams.totalRatio)

	tmpInt = tmpInt.Mul(blockReward, rewardParams.pocRatio)
	pocIncentive := big.NewInt(0).Div(tmpInt, rewardParams.totalRatio)

	tmpInt = tmpInt.Mul(blockReward, rewardParams.kirRatio)
	kirIncentive := big.NewInt(0).Div(tmpInt, rewardParams.totalRatio)

	remaining := tmpInt.Sub(blockReward, cnReward)
	remaining = tmpInt.Sub(remaining, pocIncentive)
	remaining = tmpInt.Sub(remaining, kirIncentive)
	pocIncentive = pocIncentive.Add(pocIncentive, remaining)

	// CN reward
	b.AddBalance(proposer, cnReward)

	// Proposer gets PoC incentive and KIR incentive, if there is no PoC/KIR address.
	// PoC
	if isEmptyAddress(pocAddr) {
		pocAddr = proposer
	}
	b.AddBalance(pocAddr, pocIncentive)

	// KIR
	if isEmptyAddress(kirAddr) {
		kirAddr = proposer
	}
	b.AddBalance(kirAddr, kirIncentive)

	logger.Debug("Block reward",
		"Reward address of a proposer", proposer, "CN reward amount", cnReward,
		"PoC address", pocAddr, "Poc incentive", pocIncentive,
		"KIR address", kirAddr, "KIR incentive", kirIncentive)
}

func (rm *RewardManager) GetStakingInfo(blockNum uint64) *StakingInfo {
	return rm.stakingManager.GetStakingInfoFromStakingCache(blockNum)
}

func (rm *RewardManager) Start() {
	rm.stakingManager.Subscribe()
}

func (rm *RewardManager) Stop() {
	rm.stakingManager.Unsubscribe()
}
