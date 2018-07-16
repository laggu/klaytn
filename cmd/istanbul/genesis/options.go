package genesis

import (
	"math/big"
	"github.com/ground-x/go-gxplatform/core"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/cmd/istanbul/extra"
	"github.com/ground-x/go-gxplatform/common/hexutil"
	"github.com/ground-x/go-gxplatform/log"
	"github.com/ground-x/go-gxplatform/contracts/reward/contract"
)

type Option func(*core.Genesis)

func Validators(addrs ...common.Address) Option {
	return func(genesis *core.Genesis) {
		extraData, err := extra.Encode("0x00", addrs)
		if err != nil {
			log.Error("Failed to encode extra data", "err", err)
			return
		}
		genesis.ExtraData = hexutil.MustDecode(extraData)
	}
}

func GasLimit(limit uint64) Option {
	return func(genesis *core.Genesis) {
		genesis.GasLimit = limit
	}
}

func Alloc(addrs []common.Address, balance *big.Int) Option {
	return func(genesis *core.Genesis) {
		alloc := make(map[common.Address]core.GenesisAccount)
		for _, addr := range addrs {
			alloc[addr] = core.GenesisAccount{Balance: balance}
		}
		genesis.Alloc = alloc
	}
}

func AllocSmartContract() Option {
	return func(genesis *core.Genesis) {
		alloc := make(map[common.Address]core.GenesisAccount)

		alloc[common.HexToAddress(contract.PIReserveAddr)]       = core.GenesisAccount{Code: common.FromHex(contract.PIRRewardBinRuntime) , Balance:big.NewInt(0)}
		alloc[common.HexToAddress(contract.CommitteeRewardAddr)] = core.GenesisAccount{Code: common.FromHex(contract.CommitteeRewardBinRuntime) , Balance:big.NewInt(0)}
		alloc[common.HexToAddress(contract.RNRewardAddr)]        = core.GenesisAccount{Code: common.FromHex(contract.RNRewardBinRuntime) , Balance:big.NewInt(0)}

		genesis.Alloc = alloc
	}
}
