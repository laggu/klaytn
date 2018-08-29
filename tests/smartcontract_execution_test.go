package tests

import (
	"time"
	"math"
	"strings"
	"testing"
	"math/big"
	"encoding/json"
	"github.com/ground-x/go-gxplatform/core"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/crypto"
	"github.com/ground-x/go-gxplatform/core/vm"
	"github.com/ground-x/go-gxplatform/core/types"
	"github.com/ground-x/go-gxplatform/accounts/abi"
	"github.com/ground-x/go-gxplatform/common/profile"
	"github.com/ground-x/go-gxplatform/common/compiler"
)

type deployedContract struct {
	abi     string
	name    string
	address common.Address
}

func deployContract(filename string, bcdata *BCData, accountMap *AccountMap,
	prof *profile.Profiler) (map[string]*deployedContract, error) {
	contracts, err := compiler.CompileSolidity("", filename)
	if err != nil {
		return nil, err
	}

	cont := make(map[string]*deployedContract)
	transactions := make(types.Transactions, 0, 10)

	// create a contract tx
	for name, contract := range contracts {

		abiStr, err := json.Marshal(contract.Info.AbiDefinition)
		if err != nil {
			return nil, err
		}

		userAddr := bcdata.addrs[0]
		contractAddr := crypto.CreateAddress(*userAddr, accountMap.GetNonce(*userAddr))

		signer := types.MakeSigner(bcdata.bc.Config(), bcdata.bc.CurrentHeader().Number)
		tx := types.NewTransaction(accountMap.GetNonce(*userAddr), common.Address{},
			big.NewInt(0), 50000000, big.NewInt(0), common.FromHex(contract.Code))
		signedTx, err := types.SignTx(tx, signer, bcdata.privKeys[0])
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, signedTx)

		cont[name] = &deployedContract{
			abi: string(abiStr),
			name: name,
			address: contractAddr,
		}
	}

	bcdata.GenABlockWithTransactions(accountMap, transactions, prof)

	return cont, nil
}

func callContract(bcdata *BCData, tx *types.Transaction) ([]byte, error) {
	header := bcdata.bc.CurrentHeader()
	statedb, err := bcdata.bc.State()
	if err != nil {
		return nil, err
	}

	signer := types.MakeSigner(bcdata.bc.Config(), bcdata.bc.CurrentHeader().Number)
	msg, err := tx.AsMessage(signer)
	if err != nil {
		return nil, err
	}

	evmContext := core.NewEVMContext(msg, header, bcdata.bc, nil)
	vmenv := vm.NewEVM(evmContext, statedb, bcdata.bc.Config(), vm.Config{})
	gaspool := new(core.GasPool).AddGas(math.MaxUint64)

	ret, _, _, err := core.NewStateTransition(vmenv, msg, gaspool).TransitionDb()
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func makeRewardTransactions(c *deployedContract, accountMap *AccountMap, bcdata *BCData,
	numTransactions int) (types.Transactions, error){
	abii, err := abi.JSON(strings.NewReader(c.abi))
	if err != nil {
		return nil, err
	}

	signer := types.MakeSigner(bcdata.bc.Config(), bcdata.bc.CurrentHeader().Number)

	transactions := make(types.Transactions, numTransactions)

	numAddrs := len(bcdata.addrs)
	fromNonces := make([]uint64, numAddrs)
	for i, addr := range bcdata.addrs {
		fromNonces[i] = accountMap.GetNonce(*addr)
	}
	for i := 0; i < numTransactions; i++ {
		idx := i % numAddrs

		addr := bcdata.addrs[idx]
		data, err := abii.Pack("reward", addr)
		if err != nil {
			return nil, err
		}

		tx := types.NewTransaction(fromNonces[idx], c.address, big.NewInt(10), 5000000, big.NewInt(0), data)
		signedTx, err := types.SignTx(tx, signer, bcdata.privKeys[0])
		if err != nil {
			return nil, err
		}

		transactions[i] = signedTx
		fromNonces[idx]++
	}

	return transactions, nil
}

func executeBalanceOf(c *deployedContract, accountMap *AccountMap, bcdata *BCData,
	numTransactions int) (types.Transactions, error){
	abii, err := abi.JSON(strings.NewReader(c.abi))
	if err != nil {
		return nil, err
	}

	signer := types.MakeSigner(bcdata.bc.Config(), bcdata.bc.CurrentHeader().Number)

	numAddrs := len(bcdata.addrs)
	fromNonces := make([]uint64, numAddrs)
	for i, addr := range bcdata.addrs {
		fromNonces[i] = accountMap.GetNonce(*addr)
	}
	for i := 0; i < numTransactions; i++ {
		idx := i % numAddrs

		addr := bcdata.addrs[idx]
		data, err := abii.Pack("balanceOf", addr)
		if err != nil {
			return nil, err
		}

		tx := types.NewTransaction(fromNonces[idx], c.address, big.NewInt(0), 5000000, big.NewInt(0), data)
		signedTx, err := types.SignTx(tx, signer, bcdata.privKeys[idx])
		if err != nil {
			return nil, err
		}

		ret, err := callContract(bcdata, signedTx)
		if err != nil {
			return nil, err
		}
		balance := new(big.Int)
		abii.Unpack(&balance, "balanceOf", ret)
		//fmt.Printf("balance = %d\n", balance.Uint64())

		// This is not required because the transactions will not be inserted into the blockchain.
		//fromNonces[idx]++
	}

	return nil, nil
}

func executeSmartContract(b *testing.B, opt *ContractExecutionOption, prof *profile.Profiler) {
	// Initialize blockchain
	start := time.Now()
	bcdata, err := NewBCData(2000, 4)
	if err != nil {
		b.Fatal(err)
	}
	prof.Profile("main_init_blockchain", time.Now().Sub(start))
	defer bcdata.Shutdown()

	// Initialize address-balance map for verification
	start = time.Now()
	accountMap := NewAccountMap()
	if err := accountMap.Initialize(bcdata); err != nil {
		b.Fatal(err)
	}
	prof.Profile("main_init_accountMap", time.Now().Sub(start))

	contracts, err := deployContract(opt.filepath, bcdata, accountMap, prof)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for _, c := range contracts {
		transactions, err := opt.testFunc(c, accountMap, bcdata, b.N)
		if err != nil {
			b.Fatal(err)
		}

		if transactions != nil {
			bcdata.GenABlockWithTransactions(accountMap, transactions, prof)
		}
	}
}

type ContractExecutionOption struct {
	name string
	filepath string
	testFunc  func(c *deployedContract, accountMap *AccountMap, bcdata *BCData, numTransactions int) (types.Transactions, error)
}

func BenchmarkSmartContractExecute(b *testing.B) {
	prof := profile.NewProfiler()

	benches := []ContractExecutionOption{
		{"GXPReward:reward", "../contracts/reward/contract/GXPReward.sol", makeRewardTransactions},
		{"GXPReward:balanceOf", "../contracts/reward/contract/GXPReward.sol", executeBalanceOf},
	}

	for _, bench := range benches {
		b.Run(bench.name, func(b *testing.B) {
			executeSmartContract(b, &bench, prof)
		})
	}

	if testing.Verbose() {
		prof.PrintProfileInfo()
	}
}
