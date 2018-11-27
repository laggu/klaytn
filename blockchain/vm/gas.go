package vm

import (
	"github.com/ground-x/go-gxplatform/params"
	"math/big"
)

// Gas costs
const (
	GasZero        uint64 = 0  // G_zero
	GasQuickStep   uint64 = 2  // G_base
	GasFastestStep uint64 = 3  // G_verylow
	GasFastStep    uint64 = 5  // G_low
	GasMidStep     uint64 = 8  // G_mid
	GasSlowStep    uint64 = 10 // G_high or G_exp
	GasExtStep     uint64 = 20 // G_blockhash
)

// calcGas returns the actual gas cost of the call.
//
// The cost of gas was changed during the homestead price change HF. To allow for EIP150
// to be implemented. The returned gas is gas - base * 63 / 64.
func callGas(gasTable params.GasTable, availableGas, base uint64, callCost *big.Int) (uint64, error) {
	if gasTable.CreateBySuicide > 0 {
		availableGas = availableGas - base
		gas := availableGas - availableGas/64
		// If the bit length exceeds 64 bit we know that the newly calculated "gas" for EIP150
		// is smaller than the requested amount. Therefor we return the new gas instead
		// of returning an error.
		if callCost.BitLen() > 64 || gas < callCost.Uint64() {
			return gas, nil
		}
	}
	if callCost.BitLen() > 64 {
		return 0, errGasUintOverflow
	}

	return callCost.Uint64(), nil
}
