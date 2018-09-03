// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"math/big"
	"strings"

	"github.com/ground-x/go-gxplatform/accounts/abi"
	"github.com/ground-x/go-gxplatform/accounts/abi/bind"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/blockchain/types"
)

// CommitteeRewardABI is the input ABI used to generate the binding from.
const CommitteeRewardABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"totalAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"reward\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"safeWithdrawal\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"}]"

// CommitteeRewardBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const CommitteeRewardBinRuntime = `0x6080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820938a7476ba3d648a301aa0f94c1a1173b649c5325a0e0433c22e50d86a38d7990029`

// CommitteeRewardBin is the compiled bytecode used for deploying new contracts.
const CommitteeRewardBin = `0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820938a7476ba3d648a301aa0f94c1a1173b649c5325a0e0433c22e50d86a38d7990029`

// DeployCommitteeReward deploys a new GXP contract, binding an instance of CommitteeReward to it.
func DeployCommitteeReward(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *CommitteeReward, error) {
	parsed, err := abi.JSON(strings.NewReader(CommitteeRewardABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(CommitteeRewardBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &CommitteeReward{CommitteeRewardCaller: CommitteeRewardCaller{contract: contract}, CommitteeRewardTransactor: CommitteeRewardTransactor{contract: contract}, CommitteeRewardFilterer: CommitteeRewardFilterer{contract: contract}}, nil
}

// CommitteeReward is an auto generated Go binding around an GXP contract.
type CommitteeReward struct {
	CommitteeRewardCaller     // Read-only binding to the contract
	CommitteeRewardTransactor // Write-only binding to the contract
	CommitteeRewardFilterer   // Log filterer for contract events
}

// CommitteeRewardCaller is an auto generated read-only Go binding around an GXP contract.
type CommitteeRewardCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CommitteeRewardTransactor is an auto generated write-only Go binding around an GXP contract.
type CommitteeRewardTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CommitteeRewardFilterer is an auto generated log filtering Go binding around an GXP contract events.
type CommitteeRewardFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CommitteeRewardSession is an auto generated Go binding around an GXP contract,
// with pre-set call and transact options.
type CommitteeRewardSession struct {
	Contract     *CommitteeReward  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CommitteeRewardCallerSession is an auto generated read-only Go binding around an GXP contract,
// with pre-set call options.
type CommitteeRewardCallerSession struct {
	Contract *CommitteeRewardCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// CommitteeRewardTransactorSession is an auto generated write-only Go binding around an GXP contract,
// with pre-set transact options.
type CommitteeRewardTransactorSession struct {
	Contract     *CommitteeRewardTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// CommitteeRewardRaw is an auto generated low-level Go binding around an GXP contract.
type CommitteeRewardRaw struct {
	Contract *CommitteeReward // Generic contract binding to access the raw methods on
}

// CommitteeRewardCallerRaw is an auto generated low-level read-only Go binding around an GXP contract.
type CommitteeRewardCallerRaw struct {
	Contract *CommitteeRewardCaller // Generic read-only contract binding to access the raw methods on
}

// CommitteeRewardTransactorRaw is an auto generated low-level write-only Go binding around an GXP contract.
type CommitteeRewardTransactorRaw struct {
	Contract *CommitteeRewardTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCommitteeReward creates a new instance of CommitteeReward, bound to a specific deployed contract.
func NewCommitteeReward(address common.Address, backend bind.ContractBackend) (*CommitteeReward, error) {
	contract, err := bindCommitteeReward(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CommitteeReward{CommitteeRewardCaller: CommitteeRewardCaller{contract: contract}, CommitteeRewardTransactor: CommitteeRewardTransactor{contract: contract}, CommitteeRewardFilterer: CommitteeRewardFilterer{contract: contract}}, nil
}

// NewCommitteeRewardCaller creates a new read-only instance of CommitteeReward, bound to a specific deployed contract.
func NewCommitteeRewardCaller(address common.Address, caller bind.ContractCaller) (*CommitteeRewardCaller, error) {
	contract, err := bindCommitteeReward(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CommitteeRewardCaller{contract: contract}, nil
}

// NewCommitteeRewardTransactor creates a new write-only instance of CommitteeReward, bound to a specific deployed contract.
func NewCommitteeRewardTransactor(address common.Address, transactor bind.ContractTransactor) (*CommitteeRewardTransactor, error) {
	contract, err := bindCommitteeReward(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CommitteeRewardTransactor{contract: contract}, nil
}

// NewCommitteeRewardFilterer creates a new log filterer instance of CommitteeReward, bound to a specific deployed contract.
func NewCommitteeRewardFilterer(address common.Address, filterer bind.ContractFilterer) (*CommitteeRewardFilterer, error) {
	contract, err := bindCommitteeReward(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CommitteeRewardFilterer{contract: contract}, nil
}

// bindCommitteeReward binds a generic wrapper to an already deployed contract.
func bindCommitteeReward(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(CommitteeRewardABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CommitteeReward *CommitteeRewardRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _CommitteeReward.Contract.CommitteeRewardCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CommitteeReward *CommitteeRewardRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CommitteeReward.Contract.CommitteeRewardTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CommitteeReward *CommitteeRewardRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CommitteeReward.Contract.CommitteeRewardTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CommitteeReward *CommitteeRewardCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _CommitteeReward.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CommitteeReward *CommitteeRewardTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CommitteeReward.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CommitteeReward *CommitteeRewardTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CommitteeReward.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_CommitteeReward *CommitteeRewardCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _CommitteeReward.contract.Call(opts, out, "balanceOf", arg0)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_CommitteeReward *CommitteeRewardSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _CommitteeReward.Contract.BalanceOf(&_CommitteeReward.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_CommitteeReward *CommitteeRewardCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _CommitteeReward.Contract.BalanceOf(&_CommitteeReward.CallOpts, arg0)
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_CommitteeReward *CommitteeRewardCaller) TotalAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _CommitteeReward.contract.Call(opts, out, "totalAmount")
	return *ret0, err
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_CommitteeReward *CommitteeRewardSession) TotalAmount() (*big.Int, error) {
	return _CommitteeReward.Contract.TotalAmount(&_CommitteeReward.CallOpts)
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_CommitteeReward *CommitteeRewardCallerSession) TotalAmount() (*big.Int, error) {
	return _CommitteeReward.Contract.TotalAmount(&_CommitteeReward.CallOpts)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_CommitteeReward *CommitteeRewardTransactor) Reward(opts *bind.TransactOpts, receiver common.Address) (*types.Transaction, error) {
	return _CommitteeReward.contract.Transact(opts, "reward", receiver)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_CommitteeReward *CommitteeRewardSession) Reward(receiver common.Address) (*types.Transaction, error) {
	return _CommitteeReward.Contract.Reward(&_CommitteeReward.TransactOpts, receiver)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_CommitteeReward *CommitteeRewardTransactorSession) Reward(receiver common.Address) (*types.Transaction, error) {
	return _CommitteeReward.Contract.Reward(&_CommitteeReward.TransactOpts, receiver)
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_CommitteeReward *CommitteeRewardTransactor) SafeWithdrawal(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CommitteeReward.contract.Transact(opts, "safeWithdrawal")
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_CommitteeReward *CommitteeRewardSession) SafeWithdrawal() (*types.Transaction, error) {
	return _CommitteeReward.Contract.SafeWithdrawal(&_CommitteeReward.TransactOpts)
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_CommitteeReward *CommitteeRewardTransactorSession) SafeWithdrawal() (*types.Transaction, error) {
	return _CommitteeReward.Contract.SafeWithdrawal(&_CommitteeReward.TransactOpts)
}
