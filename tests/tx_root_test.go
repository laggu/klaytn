package tests

import (
	"fmt"
	"time"
	"testing"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/core/types"
	"github.com/ground-x/go-gxplatform/common/profile"
)

type HashFunc func(list types.DerivableList) common.Hash

func BenchmarkDeriveSha(b *testing.B) {
	funcs := map[string]HashFunc{
		"Orig":types.DeriveShaOrig,
		"Simple":types.DeriveShaSimple,
		"Concat": types.DeriveShaConcat }

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

func benchDeriveSha(b *testing.B, numTransactions, numValidators int, sha HashFunc) {
	// Initialize blockchain
	start := time.Now()
	maxAccounts := numTransactions * 2
	bcdata, err := initializeBC(maxAccounts, numValidators)
	if err != nil {
		b.Fatal(err)
	}
	profile.Prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer shutdown(bcdata)

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := make(AccountMap)
	if err := accountMap.Initialize(bcdata); err != nil {
		b.Fatal(err)
	}
	profile.Prof.Profile("main_init_accountMap", time.Now().Sub(start))

	signer := types.MakeSigner(bcdata.bc.Config(), bcdata.bc.CurrentBlock().Number())

	txs, err := makeTransactions(&accountMap,
		bcdata.addrs[0:numTransactions], bcdata.privKeys[0:numTransactions],
		signer, bcdata.addrs[numTransactions:numTransactions*2], nil, 0, false)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := sha(txs)
		if testing.Verbose() {
			fmt.Printf("[%d] txhash = %s\n", i, hash.Hex())
		}
	}
}
