package ranger

import (
	"github.com/ground-x/go-gxplatform/core/types"
	"github.com/ground-x/go-gxplatform/common"
)

type NewProofEvent struct{
	addr  common.Address
	proof *types.Proof
}
