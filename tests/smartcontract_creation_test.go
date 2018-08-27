package tests

import (
	"os"
	"io"
	"time"
	"testing"
	"path/filepath"
	"github.com/mattn/go-colorable"
	"github.com/ground-x/go-gxplatform/log"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/log/term"
	"github.com/ground-x/go-gxplatform/common/profile"
	"github.com/ground-x/go-gxplatform/common/compiler"
)

type testData struct{
	name string
	opt testOption
}

// TODO-GX: To enable logging in the test code, we can use the following function.
// This function will be moved to somewhere utility functions are located.
func enableLog() {
	usecolor := term.IsTty(os.Stderr.Fd()) && os.Getenv("TERM") != "dumb"
	output := io.Writer(os.Stderr)
	if usecolor {
		output = colorable.NewColorableStderr()
	}
	glogger := log.NewGlogHandler(log.StreamHandler(output, log.TerminalFormat(usecolor)))
	log.PrintOrigins(true)
	glogger.Verbosity(log.Lvl(5))
	glogger.Vmodule("")
	glogger.BacktraceAt("")
	log.Root().SetHandler(glogger)
}

func genOptions(b *testing.B) ([]testData, error) {
	solFiles := []string{"../contracts/reward/contract/GXPReward.sol"}

	opts := make([]testData, len(solFiles))
	for i, filename := range solFiles {
		contracts, err := compiler.CompileSolidity("", filename)
		if err != nil {
			return nil, err
		}

		for name, contract := range contracts {
			testName := filepath.Base(name)
			opts[i] = testData{testName, testOption{
				b.N, 2000, 4, 1, common.FromHex(contract.Code), makeContractCreationTransactions}}
		}
	}

	return opts, nil
}

func deploySmartContract(b *testing.B, opt *testOption, prof *profile.Profiler) {
	// Initialize blockchain
	start := time.Now()
	bcdata, err := initializeBC(opt.numMaxAccounts, opt.numValidators)
	if err != nil {
		b.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer shutdown(bcdata)

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := make(AccountMap)
	if err := accountMap.Initialize(bcdata); err != nil {
		b.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	b.ResetTimer()
	for i := 0; i < b.N/txPerBlock; i++ {
		//fmt.Printf("iteration %d tx %d\n", i, opt.numTransactions)
		genABlock(b, bcdata, &accountMap, opt, txPerBlock, prof)
	}

	genBlocks := b.N / txPerBlock
	remainTxs := b.N % txPerBlock
	if remainTxs != 0 {
		genABlock(b, bcdata, &accountMap, opt, remainTxs, prof)
		genBlocks++
	}

	bcHeight := int(bcdata.bc.CurrentHeader().Number.Uint64())
	if bcHeight != genBlocks {
		b.Fatalf("generated blocks should be %d, but %d.\n", genBlocks, bcHeight)
	}
}

func BenchmarkSmartContractDeploy(b *testing.B) {
	prof := profile.NewProfiler()

	benches, err := genOptions(b)
	if err != nil {
		b.Fatal(err)
	}

	for _, bench := range benches {
		b.Run(bench.name, func(b *testing.B) {
			bench.opt.numTransactions = b.N
			deploySmartContract(b, &bench.opt, prof)
		})
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}
