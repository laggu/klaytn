// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"math/big"
	"strings"

	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
)

// RNRewardABI is the input ABI used to generate the binding from.
const RNRewardABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"totalAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"reward\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"safeWithdrawal\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"}]"

// RNRewardBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const RNRewardBinRuntime = `0x6080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a723058208578a13dea9684c89b98bed72f7f8471753d3fc504f5b6de851cf09fb001fdfc0029`

// RNRewardBin is the compiled bytecode used for deploying new contracts.
const RNRewardBin = `0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a723058208578a13dea9684c89b98bed72f7f8471753d3fc504f5b6de851cf09fb001fdfc0029`

// DeployRNReward deploys a new GXP contract, binding an instance of RNReward to it.
func DeployRNReward(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *RNReward, error) {
	parsed, err := abi.JSON(strings.NewReader(RNRewardABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(RNRewardBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &RNReward{RNRewardCaller: RNRewardCaller{contract: contract}, RNRewardTransactor: RNRewardTransactor{contract: contract}, RNRewardFilterer: RNRewardFilterer{contract: contract}}, nil
}

// RNReward is an auto generated Go binding around an GXP contract.
type RNReward struct {
	RNRewardCaller     // Read-only binding to the contract
	RNRewardTransactor // Write-only binding to the contract
	RNRewardFilterer   // Log filterer for contract events
}

// RNRewardCaller is an auto generated read-only Go binding around an GXP contract.
type RNRewardCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RNRewardTransactor is an auto generated write-only Go binding around an GXP contract.
type RNRewardTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RNRewardFilterer is an auto generated log filtering Go binding around an GXP contract events.
type RNRewardFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RNRewardSession is an auto generated Go binding around an GXP contract,
// with pre-set call and transact options.
type RNRewardSession struct {
	Contract     *RNReward         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RNRewardCallerSession is an auto generated read-only Go binding around an GXP contract,
// with pre-set call options.
type RNRewardCallerSession struct {
	Contract *RNRewardCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// RNRewardTransactorSession is an auto generated write-only Go binding around an GXP contract,
// with pre-set transact options.
type RNRewardTransactorSession struct {
	Contract     *RNRewardTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// RNRewardRaw is an auto generated low-level Go binding around an GXP contract.
type RNRewardRaw struct {
	Contract *RNReward // Generic contract binding to access the raw methods on
}

// RNRewardCallerRaw is an auto generated low-level read-only Go binding around an GXP contract.
type RNRewardCallerRaw struct {
	Contract *RNRewardCaller // Generic read-only contract binding to access the raw methods on
}

// RNRewardTransactorRaw is an auto generated low-level write-only Go binding around an GXP contract.
type RNRewardTransactorRaw struct {
	Contract *RNRewardTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRNReward creates a new instance of RNReward, bound to a specific deployed contract.
func NewRNReward(address common.Address, backend bind.ContractBackend) (*RNReward, error) {
	contract, err := bindRNReward(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &RNReward{RNRewardCaller: RNRewardCaller{contract: contract}, RNRewardTransactor: RNRewardTransactor{contract: contract}, RNRewardFilterer: RNRewardFilterer{contract: contract}}, nil
}

// NewRNRewardCaller creates a new read-only instance of RNReward, bound to a specific deployed contract.
func NewRNRewardCaller(address common.Address, caller bind.ContractCaller) (*RNRewardCaller, error) {
	contract, err := bindRNReward(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RNRewardCaller{contract: contract}, nil
}

// NewRNRewardTransactor creates a new write-only instance of RNReward, bound to a specific deployed contract.
func NewRNRewardTransactor(address common.Address, transactor bind.ContractTransactor) (*RNRewardTransactor, error) {
	contract, err := bindRNReward(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RNRewardTransactor{contract: contract}, nil
}

// NewRNRewardFilterer creates a new log filterer instance of RNReward, bound to a specific deployed contract.
func NewRNRewardFilterer(address common.Address, filterer bind.ContractFilterer) (*RNRewardFilterer, error) {
	contract, err := bindRNReward(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RNRewardFilterer{contract: contract}, nil
}

// bindRNReward binds a generic wrapper to an already deployed contract.
func bindRNReward(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(RNRewardABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RNReward *RNRewardRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _RNReward.Contract.RNRewardCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RNReward *RNRewardRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RNReward.Contract.RNRewardTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RNReward *RNRewardRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RNReward.Contract.RNRewardTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RNReward *RNRewardCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _RNReward.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RNReward *RNRewardTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RNReward.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RNReward *RNRewardTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RNReward.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_RNReward *RNRewardCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _RNReward.contract.Call(opts, out, "balanceOf", arg0)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_RNReward *RNRewardSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _RNReward.Contract.BalanceOf(&_RNReward.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_RNReward *RNRewardCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _RNReward.Contract.BalanceOf(&_RNReward.CallOpts, arg0)
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_RNReward *RNRewardCaller) TotalAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _RNReward.contract.Call(opts, out, "totalAmount")
	return *ret0, err
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_RNReward *RNRewardSession) TotalAmount() (*big.Int, error) {
	return _RNReward.Contract.TotalAmount(&_RNReward.CallOpts)
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_RNReward *RNRewardCallerSession) TotalAmount() (*big.Int, error) {
	return _RNReward.Contract.TotalAmount(&_RNReward.CallOpts)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_RNReward *RNRewardTransactor) Reward(opts *bind.TransactOpts, receiver common.Address) (*types.Transaction, error) {
	return _RNReward.contract.Transact(opts, "reward", receiver)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_RNReward *RNRewardSession) Reward(receiver common.Address) (*types.Transaction, error) {
	return _RNReward.Contract.Reward(&_RNReward.TransactOpts, receiver)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_RNReward *RNRewardTransactorSession) Reward(receiver common.Address) (*types.Transaction, error) {
	return _RNReward.Contract.Reward(&_RNReward.TransactOpts, receiver)
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_RNReward *RNRewardTransactor) SafeWithdrawal(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RNReward.contract.Transact(opts, "safeWithdrawal")
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_RNReward *RNRewardSession) SafeWithdrawal() (*types.Transaction, error) {
	return _RNReward.Contract.SafeWithdrawal(&_RNReward.TransactOpts)
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_RNReward *RNRewardTransactorSession) SafeWithdrawal() (*types.Transaction, error) {
	return _RNReward.Contract.SafeWithdrawal(&_RNReward.TransactOpts)
}
