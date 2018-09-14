package tests

import (
	"fmt"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/common/profile"
	"github.com/ground-x/go-gxplatform/storage/statedb"
	"testing"
	"time"
)

func BenchmarkDeriveSha(b *testing.B) {
	funcs := map[string]types.IDeriveSha{
		"Orig":statedb.DeriveShaOrig{},
		"Simple":types.DeriveShaSimple{},
		"Concat": types.DeriveShaConcat{} }

	NTS := []int{1000}

	for k, f := range funcs {
		for _, nt := range NTS {
			testName := fmt.Sprintf("%s,%d",k, nt)
			b.Run(testName, func(b *testing.B) {
				benchDeriveSha(b, nt, 4, f)
			})
		}
	}
}

func benchDeriveSha(b *testing.B, numTransactions, numValidators int, sha types.IDeriveSha) {
	// Initialize blockchain
	start := time.Now()
	maxAccounts := numTransactions * 2
	bcdata, err := NewBCData(maxAccounts, numValidators)
	if err != nil {
		b.Fatal(err)
	}
	profile.Prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		b.Fatal(err)
	}
	profile.Prof.Profile("main_init_accountMap", time.Now().Sub(start))

	signer := types.MakeSigner(bcdata.bc.Config(), bcdata.bc.CurrentBlock().Number())

	txs, err := makeTransactions(accountMap,
		bcdata.addrs[0:numTransactions], bcdata.privKeys[0:numTransactions],
		signer, bcdata.addrs[numTransactions:numTransactions*2], nil, 0, false)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := sha.DeriveSha(txs)
		if testing.Verbose() {
			fmt.Printf("[%d] txhash = %s\n", i, hash.Hex())
		}
	}
}
