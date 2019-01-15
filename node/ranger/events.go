package ranger

import (
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
)

type NewProofEvent struct {
	addr  common.Address
	proof *types.Proof
}
