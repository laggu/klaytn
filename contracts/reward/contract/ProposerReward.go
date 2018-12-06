// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"math/big"
	"strings"

	"github.com/ground-x/go-gxplatform/accounts/abi"
	"github.com/ground-x/go-gxplatform/accounts/abi/bind"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/common"
)

// ProposerRewardABI is the input ABI used to generate the binding from.
const ProposerRewardABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"totalAmount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"reward\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"safeWithdrawal\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"}]"

// ProposerRewardBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const ProposerRewardBinRuntime = `0x6080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a7230582012439e6e3258a1af6d8e24672a37d4391b32cbdcc445aea62442b2f96048aa4d0029`

// ProposerRewardBin is the compiled bytecode used for deploying new contracts.
const ProposerRewardBin = `0x608060405234801561001057600080fd5b506101de806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631a39d8ef81146100805780636353586b146100a757806370a08231146100ca578063fd6b7ef8146100f8575b3360009081526001602052604081208054349081019091558154019055005b34801561008c57600080fd5b5061009561010d565b60408051918252519081900360200190f35b6100c873ffffffffffffffffffffffffffffffffffffffff60043516610113565b005b3480156100d657600080fd5b5061009573ffffffffffffffffffffffffffffffffffffffff60043516610147565b34801561010457600080fd5b506100c8610159565b60005481565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604081208054349081019091558154019055565b60016020526000908152604090205481565b336000908152600160205260408120805490829055908111156101af57604051339082156108fc029083906000818181858888f193505050501561019c576101af565b3360009081526001602052604090208190555b505600a165627a7a7230582012439e6e3258a1af6d8e24672a37d4391b32cbdcc445aea62442b2f96048aa4d0029`

// DeployProposerReward deploys a new GXP contract, binding an instance of ProposerReward to it.
func DeployProposerReward(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ProposerReward, error) {
	parsed, err := abi.JSON(strings.NewReader(ProposerRewardABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ProposerRewardBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ProposerReward{ProposerRewardCaller: ProposerRewardCaller{contract: contract}, ProposerRewardTransactor: ProposerRewardTransactor{contract: contract}, ProposerRewardFilterer: ProposerRewardFilterer{contract: contract}}, nil
}

// ProposerReward is an auto generated Go binding around an GXP contract.
type ProposerReward struct {
	ProposerRewardCaller     // Read-only binding to the contract
	ProposerRewardTransactor // Write-only binding to the contract
	ProposerRewardFilterer   // Log filterer for contract events
}

// ProposerRewardCaller is an auto generated read-only Go binding around an GXP contract.
type ProposerRewardCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProposerRewardTransactor is an auto generated write-only Go binding around an GXP contract.
type ProposerRewardTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProposerRewardFilterer is an auto generated log filtering Go binding around an GXP contract events.
type ProposerRewardFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProposerRewardSession is an auto generated Go binding around an GXP contract,
// with pre-set call and transact options.
type ProposerRewardSession struct {
	Contract     *ProposerReward   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ProposerRewardCallerSession is an auto generated read-only Go binding around an GXP contract,
// with pre-set call options.
type ProposerRewardCallerSession struct {
	Contract *ProposerRewardCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// ProposerRewardTransactorSession is an auto generated write-only Go binding around an GXP contract,
// with pre-set transact options.
type ProposerRewardTransactorSession struct {
	Contract     *ProposerRewardTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// ProposerRewardRaw is an auto generated low-level Go binding around an GXP contract.
type ProposerRewardRaw struct {
	Contract *ProposerReward // Generic contract binding to access the raw methods on
}

// ProposerRewardCallerRaw is an auto generated low-level read-only Go binding around an GXP contract.
type ProposerRewardCallerRaw struct {
	Contract *ProposerRewardCaller // Generic read-only contract binding to access the raw methods on
}

// ProposerRewardTransactorRaw is an auto generated low-level write-only Go binding around an GXP contract.
type ProposerRewardTransactorRaw struct {
	Contract *ProposerRewardTransactor // Generic write-only contract binding to access the raw methods on
}

// NewProposerReward creates a new instance of ProposerReward, bound to a specific deployed contract.
func NewProposerReward(address common.Address, backend bind.ContractBackend) (*ProposerReward, error) {
	contract, err := bindProposerReward(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ProposerReward{ProposerRewardCaller: ProposerRewardCaller{contract: contract}, ProposerRewardTransactor: ProposerRewardTransactor{contract: contract}, ProposerRewardFilterer: ProposerRewardFilterer{contract: contract}}, nil
}

// NewProposerRewardCaller creates a new read-only instance of ProposerReward, bound to a specific deployed contract.
func NewProposerRewardCaller(address common.Address, caller bind.ContractCaller) (*ProposerRewardCaller, error) {
	contract, err := bindProposerReward(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ProposerRewardCaller{contract: contract}, nil
}

// NewProposerRewardTransactor creates a new write-only instance of ProposerReward, bound to a specific deployed contract.
func NewProposerRewardTransactor(address common.Address, transactor bind.ContractTransactor) (*ProposerRewardTransactor, error) {
	contract, err := bindProposerReward(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ProposerRewardTransactor{contract: contract}, nil
}

// NewProposerRewardFilterer creates a new log filterer instance of ProposerReward, bound to a specific deployed contract.
func NewProposerRewardFilterer(address common.Address, filterer bind.ContractFilterer) (*ProposerRewardFilterer, error) {
	contract, err := bindProposerReward(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ProposerRewardFilterer{contract: contract}, nil
}

// bindProposerReward binds a generic wrapper to an already deployed contract.
func bindProposerReward(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ProposerRewardABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProposerReward *ProposerRewardRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ProposerReward.Contract.ProposerRewardCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProposerReward *ProposerRewardRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProposerReward.Contract.ProposerRewardTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProposerReward *ProposerRewardRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProposerReward.Contract.ProposerRewardTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProposerReward *ProposerRewardCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ProposerReward.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProposerReward *ProposerRewardTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProposerReward.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProposerReward *ProposerRewardTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProposerReward.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_ProposerReward *ProposerRewardCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ProposerReward.contract.Call(opts, out, "balanceOf", arg0)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_ProposerReward *ProposerRewardSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _ProposerReward.Contract.BalanceOf(&_ProposerReward.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf( address) constant returns(uint256)
func (_ProposerReward *ProposerRewardCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _ProposerReward.Contract.BalanceOf(&_ProposerReward.CallOpts, arg0)
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_ProposerReward *ProposerRewardCaller) TotalAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ProposerReward.contract.Call(opts, out, "totalAmount")
	return *ret0, err
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_ProposerReward *ProposerRewardSession) TotalAmount() (*big.Int, error) {
	return _ProposerReward.Contract.TotalAmount(&_ProposerReward.CallOpts)
}

// TotalAmount is a free data retrieval call binding the contract method 0x1a39d8ef.
//
// Solidity: function totalAmount() constant returns(uint256)
func (_ProposerReward *ProposerRewardCallerSession) TotalAmount() (*big.Int, error) {
	return _ProposerReward.Contract.TotalAmount(&_ProposerReward.CallOpts)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_ProposerReward *ProposerRewardTransactor) Reward(opts *bind.TransactOpts, receiver common.Address) (*types.Transaction, error) {
	return _ProposerReward.contract.Transact(opts, "reward", receiver)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_ProposerReward *ProposerRewardSession) Reward(receiver common.Address) (*types.Transaction, error) {
	return _ProposerReward.Contract.Reward(&_ProposerReward.TransactOpts, receiver)
}

// Reward is a paid mutator transaction binding the contract method 0x6353586b.
//
// Solidity: function reward(receiver address) returns()
func (_ProposerReward *ProposerRewardTransactorSession) Reward(receiver common.Address) (*types.Transaction, error) {
	return _ProposerReward.Contract.Reward(&_ProposerReward.TransactOpts, receiver)
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_ProposerReward *ProposerRewardTransactor) SafeWithdrawal(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProposerReward.contract.Transact(opts, "safeWithdrawal")
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_ProposerReward *ProposerRewardSession) SafeWithdrawal() (*types.Transaction, error) {
	return _ProposerReward.Contract.SafeWithdrawal(&_ProposerReward.TransactOpts)
}

// SafeWithdrawal is a paid mutator transaction binding the contract method 0xfd6b7ef8.
//
// Solidity: function safeWithdrawal() returns()
func (_ProposerReward *ProposerRewardTransactorSession) SafeWithdrawal() (*types.Transaction, error) {
	return _ProposerReward.Contract.SafeWithdrawal(&_ProposerReward.TransactOpts)
}
