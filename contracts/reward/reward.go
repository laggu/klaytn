
//go:generate abigen --sol contract/GXPReward.sol --pkg contract --out contract/reward.go

package reward

import (
	"github.com/ground-x/go-gxplatform/accounts/abi/bind"
	"github.com/ground-x/go-gxplatform/common"
)

type Reward struct {
	contractBackend bind.ContractBackend
}

func NewReward(transactOpts *bind.TransactOpts, contractAddr common.Address, contractBackend bind.ContractBackend) {

}
