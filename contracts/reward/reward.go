
//go:generate abigen --sol contract/GXPReward.sol --pkg contract --out contract/GXPReward.go

package reward

import (
	"github.com/ground-x/go-gxplatform/accounts/abi/bind"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/contracts/reward/contract"
)

type Reward struct {
	*contract.GXPRewardSession
	contractBackend bind.ContractBackend
}

func NewReward(transactOpts *bind.TransactOpts, contractAddr common.Address, contractBackend bind.ContractBackend) (*Reward, error){
     gxpreward, err := contract.NewGXPReward(contractAddr, contractBackend)
     if err != nil {
     	return nil, err
	 }

	 return &Reward{
	 	&contract.GXPRewardSession{
	 		Contract:     gxpreward,
	 		TransactOpts: *transactOpts,
		},
		contractBackend,
	 }, nil
}

func DeployReward(transactOpts *bind.TransactOpts, contractBackend bind.ContractBackend) (common.Address, *Reward, error) {

	rewardAddr, _, _, err := contract.DeployGXPReward(transactOpts, contractBackend)
	if err != nil {
		return rewardAddr, nil, err
	}

	reward, err := NewReward(transactOpts, rewardAddr, contractBackend)
	if err != nil {
		return rewardAddr, nil, err
	}

	return rewardAddr, reward, nil
}