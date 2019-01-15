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

// PIRRewardABI is the input ABI used to generate the binding from.
const PIRRewardABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"totalAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"reward\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"safeWithdrawal\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"}]"

// PIRRewardBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const PIRRewardBinRuntime = `0x6080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a723058201307c3756f4e627009187dcdbc0b3e286c13b98ba9279a25bfcc18dd8bcd73e40029`

// PIRRewardBin is the compiled bytecode used for deploying new contracts.
const PIRRewardBin = `0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a723058201307c3756f4e627009187dcdbc0b3e286c13b98ba9279a25bfcc18dd8bcd73e40029`

// DeployPIRReward deploys a new GXP contract, binding an instance of PIRReward to it.
func DeployPIRReward(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *PIRReward, error) {
	parsed, err := abi.JSON(strings.NewReader(PIRRewardABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(PIRRewardBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &PIRReward{PIRRewardCaller: PIRRewardCaller{contract: contract}, PIRRewardTransactor: PIRRewardTransactor{contract: contract}, PIRRewardFilterer: PIRRewardFilterer{contract: contract}}, nil
}

// PIRReward is an auto generated Go binding around an GXP contract.
type PIRReward struct {
	PIRRewardCaller     // Read-only binding to the contract
	PIRRewardTransactor // Write-only binding to the contract
	PIRRewardFilterer   // Log filterer for contract events
}

// PIRRewardCaller is an auto generated read-only Go binding around an GXP contract.
type PIRRewardCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PIRRewardTransactor is an auto generated write-only Go binding around an GXP contract.
type PIRRewardTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PIRRewardFilterer is an auto generated log filtering Go binding around an GXP contract events.
type PIRRewardFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PIRRewardSession is an auto generated Go binding around an GXP contract,
// with pre-set call and transact options.
type PIRRewardSession struct {
	Contract     *PIRReward        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PIRRewardCallerSession is an auto generated read-only Go binding around an GXP contract,
// with pre-set call options.
type PIRRewardCallerSession struct {
	Contract *PIRRewardCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// PIRRewardTransactorSession is an auto generated write-only Go binding around an GXP contract,
// with pre-set transact options.
type PIRRewardTransactorSession struct {
	Contract     *PIRRewardTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// PIRRewardRaw is an auto generated low-level Go binding around an GXP contract.
type PIRRewardRaw struct {
	Contract *PIRReward // Generic contract binding to access the raw methods on
}

// PIRRewardCallerRaw is an auto generated low-level read-only Go binding around an GXP contract.
type PIRRewardCallerRaw struct {
	Contract *PIRRewardCaller // Generic read-only contract binding to access the raw methods on
}

// PIRRewardTransactorRaw is an auto generated low-level write-only Go binding around an GXP contract.
type PIRRewardTransactorRaw struct {
	Contract *PIRRewardTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPIRReward creates a new instance of PIRReward, bound to a specific deployed contract.
func NewPIRReward(address common.Address, backend bind.ContractBackend) (*PIRReward, error) {
	contract, err := bindPIRReward(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PIRReward{PIRRewardCaller: PIRRewardCaller{contract: contract}, PIRRewardTransactor: PIRRewardTransactor{contract: contract}, PIRRewardFilterer: PIRRewardFilterer{contract: contract}}, nil
}

// NewPIRRewardCaller creates a new read-only instance of PIRReward, bound to a specific deployed contract.
func NewPIRRewardCaller(address common.Address, caller bind.ContractCaller) (*PIRRewardCaller, error) {
	contract, err := bindPIRReward(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PIRRewardCaller{contract: contract}, nil
}

// NewPIRRewardTransactor creates a new write-only instance of PIRReward, bound to a specific deployed contract.
func NewPIRRewardTransactor(address common.Address, transactor bind.ContractTransactor) (*PIRRewardTransactor, error) {
	contract, err := bindPIRReward(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PIRRewardTransactor{contract: contract}, nil
}

// NewPIRRewardFilterer creates a new log filterer instance of PIRReward, bound to a specific deployed contract.
func NewPIRRewardFilterer(address common.Address, filterer bind.ContractFilterer) (*PIRRewardFilterer, error) {
	contract, err := bindPIRReward(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PIRRewardFilterer{contract: contract}, nil
}

// bindPIRReward binds a generic wrapper to an already deployed contract.
func bindPIRReward(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PIRRewardABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PIRReward *PIRRewardRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PIRReward.Contract.PIRRewardCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PIRReward *PIRRewardRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PIRReward.Contract.PIRRewardTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PIRReward *PIRRewardRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PIRReward.Contract.PIRRewardTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PIRReward *PIRRewardCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PIRReward.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PIRReward *PIRRewardTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PIRReward.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PIRReward *PIRRewardTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PIRReward.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_PIRReward *PIRRewardCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PIRReward.contract.Call(opts, out, "balanceOf", arg0)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_PIRReward *PIRRewardSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _PIRReward.Contract.BalanceOf(&_PIRReward.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_PIRReward *PIRRewardCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _PIRReward.Contract.BalanceOf(&_PIRReward.CallOpts, arg0)
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_PIRReward *PIRRewardCaller) TotalAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PIRReward.contract.Call(opts, out, "totalAmount")
	return *ret0, err
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_PIRReward *PIRRewardSession) TotalAmount() (*big.Int, error) {
	return _PIRReward.Contract.TotalAmount(&_PIRReward.CallOpts)
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_PIRReward *PIRRewardCallerSession) TotalAmount() (*big.Int, error) {
	return _PIRReward.Contract.TotalAmount(&_PIRReward.CallOpts)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_PIRReward *PIRRewardTransactor) Reward(opts *bind.TransactOpts, receiver common.Address) (*types.Transaction, error) {
	return _PIRReward.contract.Transact(opts, "reward", receiver)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_PIRReward *PIRRewardSession) Reward(receiver common.Address) (*types.Transaction, error) {
	return _PIRReward.Contract.Reward(&_PIRReward.TransactOpts, receiver)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_PIRReward *PIRRewardTransactorSession) Reward(receiver common.Address) (*types.Transaction, error) {
	return _PIRReward.Contract.Reward(&_PIRReward.TransactOpts, receiver)
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_PIRReward *PIRRewardTransactor) SafeWithdrawal(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PIRReward.contract.Transact(opts, "safeWithdrawal")
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_PIRReward *PIRRewardSession) SafeWithdrawal() (*types.Transaction, error) {
	return _PIRReward.Contract.SafeWithdrawal(&_PIRReward.TransactOpts)
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_PIRReward *PIRRewardTransactorSession) SafeWithdrawal() (*types.Transaction, error) {
	return _PIRReward.Contract.SafeWithdrawal(&_PIRReward.TransactOpts)
}
