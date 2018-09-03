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

// GXPRewardABI is the input ABI used to generate the binding from.
const GXPRewardABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"totalAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"reward\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"safeWithdrawal\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"}]"

// GXPRewardBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const GXPRewardBinRuntime = `0x6080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820ca3321999b03e59c10199e70af16e9a8ca057580ce875aa18541b976ea3fec6a0029`

// GXPRewardBin is the compiled bytecode used for deploying new contracts.
const GXPRewardBin = `0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a72305820ca3321999b03e59c10199e70af16e9a8ca057580ce875aa18541b976ea3fec6a0029`

// DeployGXPReward deploys a new GXP contract, binding an instance of GXPReward to it.
func DeployGXPReward(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *GXPReward, error) {
	parsed, err := abi.JSON(strings.NewReader(GXPRewardABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(GXPRewardBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GXPReward{GXPRewardCaller: GXPRewardCaller{contract: contract}, GXPRewardTransactor: GXPRewardTransactor{contract: contract}, GXPRewardFilterer: GXPRewardFilterer{contract: contract}}, nil
}

// GXPReward is an auto generated Go binding around an GXP contract.
type GXPReward struct {
	GXPRewardCaller     // Read-only binding to the contract
	GXPRewardTransactor // Write-only binding to the contract
	GXPRewardFilterer   // Log filterer for contract events
}

// GXPRewardCaller is an auto generated read-only Go binding around an GXP contract.
type GXPRewardCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GXPRewardTransactor is an auto generated write-only Go binding around an GXP contract.
type GXPRewardTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GXPRewardFilterer is an auto generated log filtering Go binding around an GXP contract events.
type GXPRewardFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GXPRewardSession is an auto generated Go binding around an GXP contract,
// with pre-set call and transact options.
type GXPRewardSession struct {
	Contract     *GXPReward        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GXPRewardCallerSession is an auto generated read-only Go binding around an GXP contract,
// with pre-set call options.
type GXPRewardCallerSession struct {
	Contract *GXPRewardCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// GXPRewardTransactorSession is an auto generated write-only Go binding around an GXP contract,
// with pre-set transact options.
type GXPRewardTransactorSession struct {
	Contract     *GXPRewardTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// GXPRewardRaw is an auto generated low-level Go binding around an GXP contract.
type GXPRewardRaw struct {
	Contract *GXPReward // Generic contract binding to access the raw methods on
}

// GXPRewardCallerRaw is an auto generated low-level read-only Go binding around an GXP contract.
type GXPRewardCallerRaw struct {
	Contract *GXPRewardCaller // Generic read-only contract binding to access the raw methods on
}

// GXPRewardTransactorRaw is an auto generated low-level write-only Go binding around an GXP contract.
type GXPRewardTransactorRaw struct {
	Contract *GXPRewardTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGXPReward creates a new instance of GXPReward, bound to a specific deployed contract.
func NewGXPReward(address common.Address, backend bind.ContractBackend) (*GXPReward, error) {
	contract, err := bindGXPReward(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GXPReward{GXPRewardCaller: GXPRewardCaller{contract: contract}, GXPRewardTransactor: GXPRewardTransactor{contract: contract}, GXPRewardFilterer: GXPRewardFilterer{contract: contract}}, nil
}

// NewGXPRewardCaller creates a new read-only instance of GXPReward, bound to a specific deployed contract.
func NewGXPRewardCaller(address common.Address, caller bind.ContractCaller) (*GXPRewardCaller, error) {
	contract, err := bindGXPReward(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GXPRewardCaller{contract: contract}, nil
}

// NewGXPRewardTransactor creates a new write-only instance of GXPReward, bound to a specific deployed contract.
func NewGXPRewardTransactor(address common.Address, transactor bind.ContractTransactor) (*GXPRewardTransactor, error) {
	contract, err := bindGXPReward(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GXPRewardTransactor{contract: contract}, nil
}

// NewGXPRewardFilterer creates a new log filterer instance of GXPReward, bound to a specific deployed contract.
func NewGXPRewardFilterer(address common.Address, filterer bind.ContractFilterer) (*GXPRewardFilterer, error) {
	contract, err := bindGXPReward(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GXPRewardFilterer{contract: contract}, nil
}

// bindGXPReward binds a generic wrapper to an already deployed contract.
func bindGXPReward(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(GXPRewardABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GXPReward *GXPRewardRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _GXPReward.Contract.GXPRewardCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GXPReward *GXPRewardRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GXPReward.Contract.GXPRewardTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GXPReward *GXPRewardRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GXPReward.Contract.GXPRewardTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GXPReward *GXPRewardCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _GXPReward.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GXPReward *GXPRewardTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GXPReward.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GXPReward *GXPRewardTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GXPReward.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_GXPReward *GXPRewardCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _GXPReward.contract.Call(opts, out, "balanceOf", arg0)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_GXPReward *GXPRewardSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _GXPReward.Contract.BalanceOf(&_GXPReward.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_GXPReward *GXPRewardCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _GXPReward.Contract.BalanceOf(&_GXPReward.CallOpts, arg0)
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_GXPReward *GXPRewardCaller) TotalAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _GXPReward.contract.Call(opts, out, "totalAmount")
	return *ret0, err
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_GXPReward *GXPRewardSession) TotalAmount() (*big.Int, error) {
	return _GXPReward.Contract.TotalAmount(&_GXPReward.CallOpts)
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_GXPReward *GXPRewardCallerSession) TotalAmount() (*big.Int, error) {
	return _GXPReward.Contract.TotalAmount(&_GXPReward.CallOpts)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_GXPReward *GXPRewardTransactor) Reward(opts *bind.TransactOpts, receiver common.Address) (*types.Transaction, error) {
	return _GXPReward.contract.Transact(opts, "reward", receiver)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_GXPReward *GXPRewardSession) Reward(receiver common.Address) (*types.Transaction, error) {
	return _GXPReward.Contract.Reward(&_GXPReward.TransactOpts, receiver)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_GXPReward *GXPRewardTransactorSession) Reward(receiver common.Address) (*types.Transaction, error) {
	return _GXPReward.Contract.Reward(&_GXPReward.TransactOpts, receiver)
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_GXPReward *GXPRewardTransactor) SafeWithdrawal(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GXPReward.contract.Transact(opts, "safeWithdrawal")
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_GXPReward *GXPRewardSession) SafeWithdrawal() (*types.Transaction, error) {
	return _GXPReward.Contract.SafeWithdrawal(&_GXPReward.TransactOpts)
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_GXPReward *GXPRewardTransactorSession) SafeWithdrawal() (*types.Transaction, error) {
	return _GXPReward.Contract.SafeWithdrawal(&_GXPReward.TransactOpts)
}
