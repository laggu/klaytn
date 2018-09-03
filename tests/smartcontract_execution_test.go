package tests

import (
	"encoding/json"
	"github.com/ground-x/go-gxplatform/accounts/abi"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/common/compiler"
	"github.com/ground-x/go-gxplatform/common/profile"
	"github.com/ground-x/go-gxplatform/blockchain"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/blockchain/vm"
	"github.com/ground-x/go-gxplatform/crypto"
	"math"
	"math/big"
	"strings"
	"testing"
	"time"
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
			abi:     string(abiStr),
			name:    name,
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

	evmContext := blockchain.NewEVMContext(msg, header, bcdata.bc, nil)
	vmenv := vm.NewEVM(evmContext, statedb, bcdata.bc.Config(), &vm.Config{})
	gaspool := new(blockchain.GasPool).AddGas(math.MaxUint64)

	ret, _, _, err := blockchain.NewStateTransition(vmenv, msg, gaspool).TransitionDb()
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func makeRewardTransactions(c *deployedContract, accountMap *AccountMap, bcdata *BCData,
	numTransactions int) (types.Transactions, error) {
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

func executeRewardTransactions(c *deployedContract, transactions types.Transactions, prof *profile.Profiler, bcdata *BCData,
	accountMap *AccountMap) error {
	return bcdata.GenABlockWithTransactions(accountMap, transactions, prof)
}

func makeBalanceOf(c *deployedContract, accountMap *AccountMap, bcdata *BCData,
	numTransactions int) (types.Transactions, error) {
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
		data, err := abii.Pack("balanceOf", addr)
		if err != nil {
			return nil, err
		}

		tx := types.NewTransaction(fromNonces[idx], c.address, big.NewInt(0), 5000000, big.NewInt(0), data)
		signedTx, err := types.SignTx(tx, signer, bcdata.privKeys[idx])
		if err != nil {
			return nil, err
		}

		transactions[i] = signedTx

		// This is not required because the transactions will not be inserted into the blockchain.
		//fromNonces[idx]++
	}

	return nil, nil
}

func executeBalanceOf(c *deployedContract, transactions types.Transactions, prof *profile.Profiler, bcdata *BCData,
	accountMap *AccountMap) error {
	abii, err := abi.JSON(strings.NewReader(c.abi))
	if err != nil {
		return err
	}

	for _, tx := range transactions {
		ret, err := callContract(bcdata, tx)
		if err != nil {
			return err
		}

		balance := new(big.Int)
		abii.Unpack(&balance, "balanceOf", ret)
	}

	return nil
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

	start = time.Now()
	contracts, err := deployContract(opt.filepath, bcdata, accountMap, prof)
	if err != nil {
		b.Fatal(err)
	}
	prof.Profile("main_deployContract", time.Now().Sub(start))

	b.StopTimer()
	b.ResetTimer()
	for _, c := range contracts {
		start = time.Now()
		transactions, err := opt.makeTx(c, accountMap, bcdata, b.N)
		if err != nil {
			b.Fatal(err)
		}
		prof.Profile("main_makeTx", time.Now().Sub(start))

		start = time.Now()
		b.StartTimer()
		opt.executeTx(c, transactions, prof, bcdata, accountMap)
		b.StopTimer()
		prof.Profile("main_executeTx", time.Now().Sub(start))
	}
}

type ContractExecutionOption struct {
	name      string
	filepath  string
	makeTx    func(c *deployedContract, accountMap *AccountMap, bcdata *BCData, numTransactions int) (types.Transactions, error)
	executeTx func(c *deployedContract, transactions types.Transactions, prof *profile.Profiler, bcdata *BCData, accountMap *AccountMap) error
}

func BenchmarkSmartContractExecute(b *testing.B) {
	prof := profile.NewProfiler()

	benches := []ContractExecutionOption{
		{"GXPReward:reward", "../contracts/reward/contract/GXPReward.sol", makeRewardTransactions, executeRewardTransactions},
		{"GXPReward:balanceOf", "../contracts/reward/contract/GXPReward.sol", makeBalanceOf, executeBalanceOf},
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
