package validator

import (
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/consensus/istanbul"
	"github.com/ground-x/go-gxplatform/log"
)

var logger = log.NewModuleLogger(log.ConsensusIstanbulValidator)

func New(addr common.Address) istanbul.Validator {
	return &defaultValidator{
		address: addr,
	}
}

func NewSet(addrs []common.Address, policy istanbul.ProposerPolicy) istanbul.ValidatorSet {
	return newDefaultSet(addrs, policy)
}

func NewSubSet(addrs []common.Address, policy istanbul.ProposerPolicy, subSize int) istanbul.ValidatorSet {
	return newDefaultSubSet(addrs, policy, subSize)
}

func ExtractValidators(extraData []byte) []common.Address {
	// get the validator addresses
	addrs := make([]common.Address, (len(extraData) / common.AddressLength))
	for i := 0; i < len(addrs); i++ {
		copy(addrs[i][:], extraData[i*common.AddressLength:])
	}

	return addrs
}
