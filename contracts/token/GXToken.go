// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package token

import (
	"math/big"
	"strings"

	"github.com/ground-x/klaytn"
	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/event"
)

// GXTokenABI is the input ABI used to generate the binding from.
const GXTokenABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"balanceOfMine\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"INITIAL_SUPPLY\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"mintToGateway\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"depositToGateway\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"safeTransferAndCall\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_gateway\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"}]"

// GXTokenBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const GXTokenBinRuntime = `0x6080604052600436106100985763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde03811461009d5780631459cef4146101275780632ff2e9dc1461014e578063313ce56714610163578063544297f51461018e5780639267daba146101a857806395d89b41146101c0578063a9059cbb146101d5578063f7507f001461020d575b600080fd5b3480156100a957600080fd5b506100b2610231565b6040805160208082528351818301528351919283929083019185019080838360005b838110156100ec5781810151838201526020016100d4565b50505050905090810190601f1680156101195780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561013357600080fd5b5061013c6102be565b60408051918252519081900360200190f35b34801561015a57600080fd5b5061013c6102d1565b34801561016f57600080fd5b506101786102d9565b6040805160ff9092168252519081900360200190f35b34801561019a57600080fd5b506101a66004356102e2565b005b3480156101b457600080fd5b506101a6600435610354565b3480156101cc57600080fd5b506100b261036d565b3480156101e157600080fd5b506101f9600160a060020a03600435166024356103c5565b604080519115158252519081900360200190f35b34801561021957600080fd5b506101a6600160a060020a03600435166024356103db565b60018054604080516020600284861615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156102b65780601f1061028b576101008083540402835291602001916102b6565b820191906000526020600020905b81548152906001019060200180831161029957829003601f168201915b505050505081565b3360009081526005602052604090205490565b633b9aca0081565b60035460ff1681565b600054600160a060020a031633146102f957600080fd5b60045461030c908263ffffffff61040016565b60045560008054600160a060020a0316815260056020526040902054610338908263ffffffff61040016565b60008054600160a060020a031681526005602052604090205550565b60005461036a90600160a060020a0316826103db565b50565b6002805460408051602060018416156101000260001901909316849004601f810184900484028201840190925281815292918301828280156102b65780601f1061028b576101008083540402835291602001916102b6565b60006103d2338484610419565b50600192915050565b6103e582826103c5565b506103f13383836104ef565b15156103fc57600080fd5b5050565b60008282018381101561041257600080fd5b9392505050565b600160a060020a038216151561042e57600080fd5b600160a060020a038316600090815260056020526040902054610457908263ffffffff6105e716565b600160a060020a03808516600090815260056020526040808220939093559084168152205461048c908263ffffffff61040016565b600160a060020a0380841660008181526005602090815260409182902094909455805192871683529282015280820183905290517fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9181900360600190a1505050565b60008083600160a060020a031663bc04f0af86856040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083600160a060020a0316600160a060020a0316815260200182815260200192505050602060405180830381600087803b15801561056e57600080fd5b505af1158015610582573d6000803e3d6000fd5b505050506040513d602081101561059857600080fd5b50517fffffffff00000000000000000000000000000000000000000000000000000000167fbc04f0af000000000000000000000000000000000000000000000000000000001495945050505050565b600080838311156105f757600080fd5b50509003905600a165627a7a723058205cd2c596b859ed5fad2777d028a79e4de6f93c8dca90b6e464791eb463a76e6c0029`

// GXTokenBin is the compiled bytecode used for deploying new contracts.
const GXTokenBin = `0x60c0604052600760808190527f4758546f6b656e0000000000000000000000000000000000000000000000000060a090815261003e91600191906100f4565b506040805180820190915260028082527f4758000000000000000000000000000000000000000000000000000000000000602090920191825261008191816100f4565b506003805460ff1916601217905534801561009b57600080fd5b506040516020806107c8833981016040908152905160008054600160a060020a031916600160a060020a03909216918217815560035460ff16600a0a633b9aca000260048190559181526005602052919091205561018f565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061013557805160ff1916838001178555610162565b82800160010185558215610162579182015b82811115610162578251825591602001919060010190610147565b5061016e929150610172565b5090565b61018c91905b8082111561016e5760008155600101610178565b90565b61062a8061019e6000396000f3006080604052600436106100985763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde03811461009d5780631459cef4146101275780632ff2e9dc1461014e578063313ce56714610163578063544297f51461018e5780639267daba146101a857806395d89b41146101c0578063a9059cbb146101d5578063f7507f001461020d575b600080fd5b3480156100a957600080fd5b506100b2610231565b6040805160208082528351818301528351919283929083019185019080838360005b838110156100ec5781810151838201526020016100d4565b50505050905090810190601f1680156101195780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561013357600080fd5b5061013c6102be565b60408051918252519081900360200190f35b34801561015a57600080fd5b5061013c6102d1565b34801561016f57600080fd5b506101786102d9565b6040805160ff9092168252519081900360200190f35b34801561019a57600080fd5b506101a66004356102e2565b005b3480156101b457600080fd5b506101a6600435610354565b3480156101cc57600080fd5b506100b261036d565b3480156101e157600080fd5b506101f9600160a060020a03600435166024356103c5565b604080519115158252519081900360200190f35b34801561021957600080fd5b506101a6600160a060020a03600435166024356103db565b60018054604080516020600284861615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156102b65780601f1061028b576101008083540402835291602001916102b6565b820191906000526020600020905b81548152906001019060200180831161029957829003601f168201915b505050505081565b3360009081526005602052604090205490565b633b9aca0081565b60035460ff1681565b600054600160a060020a031633146102f957600080fd5b60045461030c908263ffffffff61040016565b60045560008054600160a060020a0316815260056020526040902054610338908263ffffffff61040016565b60008054600160a060020a031681526005602052604090205550565b60005461036a90600160a060020a0316826103db565b50565b6002805460408051602060018416156101000260001901909316849004601f810184900484028201840190925281815292918301828280156102b65780601f1061028b576101008083540402835291602001916102b6565b60006103d2338484610419565b50600192915050565b6103e582826103c5565b506103f13383836104ef565b15156103fc57600080fd5b5050565b60008282018381101561041257600080fd5b9392505050565b600160a060020a038216151561042e57600080fd5b600160a060020a038316600090815260056020526040902054610457908263ffffffff6105e716565b600160a060020a03808516600090815260056020526040808220939093559084168152205461048c908263ffffffff61040016565b600160a060020a0380841660008181526005602090815260409182902094909455805192871683529282015280820183905290517fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9181900360600190a1505050565b60008083600160a060020a031663bc04f0af86856040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083600160a060020a0316600160a060020a0316815260200182815260200192505050602060405180830381600087803b15801561056e57600080fd5b505af1158015610582573d6000803e3d6000fd5b505050506040513d602081101561059857600080fd5b50517fffffffff00000000000000000000000000000000000000000000000000000000167fbc04f0af000000000000000000000000000000000000000000000000000000001495945050505050565b600080838311156105f757600080fd5b50509003905600a165627a7a723058205cd2c596b859ed5fad2777d028a79e4de6f93c8dca90b6e464791eb463a76e6c0029`

// DeployGXToken deploys a new klaytn contract, binding an instance of GXToken to it.
func DeployGXToken(auth *bind.TransactOpts, backend bind.ContractBackend, _gateway common.Address) (common.Address, *types.Transaction, *GXToken, error) {
	parsed, err := abi.JSON(strings.NewReader(GXTokenABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(GXTokenBin), backend, _gateway)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GXToken{GXTokenCaller: GXTokenCaller{contract: contract}, GXTokenTransactor: GXTokenTransactor{contract: contract}, GXTokenFilterer: GXTokenFilterer{contract: contract}}, nil
}

// GXToken is an auto generated Go binding around a klaytn contract.
type GXToken struct {
	GXTokenCaller     // Read-only binding to the contract
	GXTokenTransactor // Write-only binding to the contract
	GXTokenFilterer   // Log filterer for contract events
}

// GXTokenCaller is an auto generated read-only Go binding around a klaytn contract.
type GXTokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GXTokenTransactor is an auto generated write-only Go binding around a klaytn contract.
type GXTokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GXTokenFilterer is an auto generated log filtering Go binding around a klaytn contract events.
type GXTokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GXTokenSession is an auto generated Go binding around a klaytn contract,
// with pre-set call and transact options.
type GXTokenSession struct {
	Contract     *GXToken          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GXTokenCallerSession is an auto generated read-only Go binding around a klaytn contract,
// with pre-set call options.
type GXTokenCallerSession struct {
	Contract *GXTokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// GXTokenTransactorSession is an auto generated write-only Go binding around a klaytn contract,
// with pre-set transact options.
type GXTokenTransactorSession struct {
	Contract     *GXTokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// GXTokenRaw is an auto generated low-level Go binding around a klaytn contract.
type GXTokenRaw struct {
	Contract *GXToken // Generic contract binding to access the raw methods on
}

// GXTokenCallerRaw is an auto generated low-level read-only Go binding around a klaytn contract.
type GXTokenCallerRaw struct {
	Contract *GXTokenCaller // Generic read-only contract binding to access the raw methods on
}

// GXTokenTransactorRaw is an auto generated low-level write-only Go binding around a klaytn contract.
type GXTokenTransactorRaw struct {
	Contract *GXTokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGXToken creates a new instance of GXToken, bound to a specific deployed contract.
func NewGXToken(address common.Address, backend bind.ContractBackend) (*GXToken, error) {
	contract, err := bindGXToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GXToken{GXTokenCaller: GXTokenCaller{contract: contract}, GXTokenTransactor: GXTokenTransactor{contract: contract}, GXTokenFilterer: GXTokenFilterer{contract: contract}}, nil
}

// NewGXTokenCaller creates a new read-only instance of GXToken, bound to a specific deployed contract.
func NewGXTokenCaller(address common.Address, caller bind.ContractCaller) (*GXTokenCaller, error) {
	contract, err := bindGXToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GXTokenCaller{contract: contract}, nil
}

// NewGXTokenTransactor creates a new write-only instance of GXToken, bound to a specific deployed contract.
func NewGXTokenTransactor(address common.Address, transactor bind.ContractTransactor) (*GXTokenTransactor, error) {
	contract, err := bindGXToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GXTokenTransactor{contract: contract}, nil
}

// NewGXTokenFilterer creates a new log filterer instance of GXToken, bound to a specific deployed contract.
func NewGXTokenFilterer(address common.Address, filterer bind.ContractFilterer) (*GXTokenFilterer, error) {
	contract, err := bindGXToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GXTokenFilterer{contract: contract}, nil
}

// bindGXToken binds a generic wrapper to an already deployed contract.
func bindGXToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(GXTokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GXToken *GXTokenRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _GXToken.Contract.GXTokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GXToken *GXTokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GXToken.Contract.GXTokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GXToken *GXTokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GXToken.Contract.GXTokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GXToken *GXTokenCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _GXToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GXToken *GXTokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GXToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GXToken *GXTokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GXToken.Contract.contract.Transact(opts, method, params...)
}

// INITIALSUPPLY is a free data retrieval call binding the contract method 0x2ff2e9dc.
//
// Solidity: function INITIAL_SUPPLY() constant returns(uint256)
func (_GXToken *GXTokenCaller) INITIALSUPPLY(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _GXToken.contract.Call(opts, out, "INITIAL_SUPPLY")
	return *ret0, err
}

// INITIALSUPPLY is a free data retrieval call binding the contract method 0x2ff2e9dc.
//
// Solidity: function INITIAL_SUPPLY() constant returns(uint256)
func (_GXToken *GXTokenSession) INITIALSUPPLY() (*big.Int, error) {
	return _GXToken.Contract.INITIALSUPPLY(&_GXToken.CallOpts)
}

// INITIALSUPPLY is a free data retrieval call binding the contract method 0x2ff2e9dc.
//
// Solidity: function INITIAL_SUPPLY() constant returns(uint256)
func (_GXToken *GXTokenCallerSession) INITIALSUPPLY() (*big.Int, error) {
	return _GXToken.Contract.INITIALSUPPLY(&_GXToken.CallOpts)
}

// BalanceOfMine is a free data retrieval call binding the contract method 0x1459cef4.
//
// Solidity: function balanceOfMine() constant returns(uint256)
func (_GXToken *GXTokenCaller) BalanceOfMine(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _GXToken.contract.Call(opts, out, "balanceOfMine")
	return *ret0, err
}

// BalanceOfMine is a free data retrieval call binding the contract method 0x1459cef4.
//
// Solidity: function balanceOfMine() constant returns(uint256)
func (_GXToken *GXTokenSession) BalanceOfMine() (*big.Int, error) {
	return _GXToken.Contract.BalanceOfMine(&_GXToken.CallOpts)
}

// BalanceOfMine is a free data retrieval call binding the contract method 0x1459cef4.
//
// Solidity: function balanceOfMine() constant returns(uint256)
func (_GXToken *GXTokenCallerSession) BalanceOfMine() (*big.Int, error) {
	return _GXToken.Contract.BalanceOfMine(&_GXToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_GXToken *GXTokenCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _GXToken.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_GXToken *GXTokenSession) Decimals() (uint8, error) {
	return _GXToken.Contract.Decimals(&_GXToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_GXToken *GXTokenCallerSession) Decimals() (uint8, error) {
	return _GXToken.Contract.Decimals(&_GXToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_GXToken *GXTokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _GXToken.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_GXToken *GXTokenSession) Name() (string, error) {
	return _GXToken.Contract.Name(&_GXToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_GXToken *GXTokenCallerSession) Name() (string, error) {
	return _GXToken.Contract.Name(&_GXToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_GXToken *GXTokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _GXToken.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_GXToken *GXTokenSession) Symbol() (string, error) {
	return _GXToken.Contract.Symbol(&_GXToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_GXToken *GXTokenCallerSession) Symbol() (string, error) {
	return _GXToken.Contract.Symbol(&_GXToken.CallOpts)
}

// DepositToGateway is a paid mutator transaction binding the contract method 0x9267daba.
//
// Solidity: function depositToGateway(amount uint256) returns()
func (_GXToken *GXTokenTransactor) DepositToGateway(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _GXToken.contract.Transact(opts, "depositToGateway", amount)
}

// DepositToGateway is a paid mutator transaction binding the contract method 0x9267daba.
//
// Solidity: function depositToGateway(amount uint256) returns()
func (_GXToken *GXTokenSession) DepositToGateway(amount *big.Int) (*types.Transaction, error) {
	return _GXToken.Contract.DepositToGateway(&_GXToken.TransactOpts, amount)
}

// DepositToGateway is a paid mutator transaction binding the contract method 0x9267daba.
//
// Solidity: function depositToGateway(amount uint256) returns()
func (_GXToken *GXTokenTransactorSession) DepositToGateway(amount *big.Int) (*types.Transaction, error) {
	return _GXToken.Contract.DepositToGateway(&_GXToken.TransactOpts, amount)
}

// MintToGateway is a paid mutator transaction binding the contract method 0x544297f5.
//
// Solidity: function mintToGateway(_amount uint256) returns()
func (_GXToken *GXTokenTransactor) MintToGateway(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _GXToken.contract.Transact(opts, "mintToGateway", _amount)
}

// MintToGateway is a paid mutator transaction binding the contract method 0x544297f5.
//
// Solidity: function mintToGateway(_amount uint256) returns()
func (_GXToken *GXTokenSession) MintToGateway(_amount *big.Int) (*types.Transaction, error) {
	return _GXToken.Contract.MintToGateway(&_GXToken.TransactOpts, _amount)
}

// MintToGateway is a paid mutator transaction binding the contract method 0x544297f5.
//
// Solidity: function mintToGateway(_amount uint256) returns()
func (_GXToken *GXTokenTransactorSession) MintToGateway(_amount *big.Int) (*types.Transaction, error) {
	return _GXToken.Contract.MintToGateway(&_GXToken.TransactOpts, _amount)
}

// SafeTransferAndCall is a paid mutator transaction binding the contract method 0xf7507f00.
//
// Solidity: function safeTransferAndCall(_to address, amount uint256) returns()
func (_GXToken *GXTokenTransactor) SafeTransferAndCall(opts *bind.TransactOpts, _to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _GXToken.contract.Transact(opts, "safeTransferAndCall", _to, amount)
}

// SafeTransferAndCall is a paid mutator transaction binding the contract method 0xf7507f00.
//
// Solidity: function safeTransferAndCall(_to address, amount uint256) returns()
func (_GXToken *GXTokenSession) SafeTransferAndCall(_to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _GXToken.Contract.SafeTransferAndCall(&_GXToken.TransactOpts, _to, amount)
}

// SafeTransferAndCall is a paid mutator transaction binding the contract method 0xf7507f00.
//
// Solidity: function safeTransferAndCall(_to address, amount uint256) returns()
func (_GXToken *GXTokenTransactorSession) SafeTransferAndCall(_to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _GXToken.Contract.SafeTransferAndCall(&_GXToken.TransactOpts, _to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, value uint256) returns(bool)
func (_GXToken *GXTokenTransactor) Transfer(opts *bind.TransactOpts, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _GXToken.contract.Transact(opts, "transfer", to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, value uint256) returns(bool)
func (_GXToken *GXTokenSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _GXToken.Contract.Transfer(&_GXToken.TransactOpts, to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, value uint256) returns(bool)
func (_GXToken *GXTokenTransactorSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _GXToken.Contract.Transfer(&_GXToken.TransactOpts, to, value)
}

// GXTokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the GXToken contract.
type GXTokenTransferIterator struct {
	Event *GXTokenTransfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log      // Log channel receiving the found contract events
	sub  klaytn.Subscription // Subscription for errors, completion and termination
	done bool                // Whether the subscription completed delivering logs
	fail error               // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *GXTokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GXTokenTransfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(GXTokenTransfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *GXTokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GXTokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GXTokenTransfer represents a Transfer event raised by the GXToken contract.
type GXTokenTransfer struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from address, to address, amount uint256)
func (_GXToken *GXTokenFilterer) FilterTransfer(opts *bind.FilterOpts) (*GXTokenTransferIterator, error) {

	logs, sub, err := _GXToken.contract.FilterLogs(opts, "Transfer")
	if err != nil {
		return nil, err
	}
	return &GXTokenTransferIterator{contract: _GXToken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from address, to address, amount uint256)
func (_GXToken *GXTokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *GXTokenTransfer) (event.Subscription, error) {

	logs, sub, err := _GXToken.contract.WatchLogs(opts, "Transfer")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GXTokenTransfer)
				if err := _GXToken.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// IKRC20ReceiverABI is the input ABI used to generate the binding from.
const IKRC20ReceiverABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"onERC20Received\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// IKRC20ReceiverBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const IKRC20ReceiverBinRuntime = `0x`

// IKRC20ReceiverBin is the compiled bytecode used for deploying new contracts.
const IKRC20ReceiverBin = `0x`

// DeployIKRC20Receiver deploys a new klaytn contract, binding an instance of IKRC20Receiver to it.
func DeployIKRC20Receiver(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *IKRC20Receiver, error) {
	parsed, err := abi.JSON(strings.NewReader(IKRC20ReceiverABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(IKRC20ReceiverBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &IKRC20Receiver{IKRC20ReceiverCaller: IKRC20ReceiverCaller{contract: contract}, IKRC20ReceiverTransactor: IKRC20ReceiverTransactor{contract: contract}, IKRC20ReceiverFilterer: IKRC20ReceiverFilterer{contract: contract}}, nil
}

// IKRC20Receiver is an auto generated Go binding around a klaytn contract.
type IKRC20Receiver struct {
	IKRC20ReceiverCaller     // Read-only binding to the contract
	IKRC20ReceiverTransactor // Write-only binding to the contract
	IKRC20ReceiverFilterer   // Log filterer for contract events
}

// IKRC20ReceiverCaller is an auto generated read-only Go binding around a klaytn contract.
type IKRC20ReceiverCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IKRC20ReceiverTransactor is an auto generated write-only Go binding around a klaytn contract.
type IKRC20ReceiverTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IKRC20ReceiverFilterer is an auto generated log filtering Go binding around a klaytn contract events.
type IKRC20ReceiverFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IKRC20ReceiverSession is an auto generated Go binding around a klaytn contract,
// with pre-set call and transact options.
type IKRC20ReceiverSession struct {
	Contract     *IKRC20Receiver   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IKRC20ReceiverCallerSession is an auto generated read-only Go binding around a klaytn contract,
// with pre-set call options.
type IKRC20ReceiverCallerSession struct {
	Contract *IKRC20ReceiverCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// IKRC20ReceiverTransactorSession is an auto generated write-only Go binding around a klaytn contract,
// with pre-set transact options.
type IKRC20ReceiverTransactorSession struct {
	Contract     *IKRC20ReceiverTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// IKRC20ReceiverRaw is an auto generated low-level Go binding around a klaytn contract.
type IKRC20ReceiverRaw struct {
	Contract *IKRC20Receiver // Generic contract binding to access the raw methods on
}

// IKRC20ReceiverCallerRaw is an auto generated low-level read-only Go binding around a klaytn contract.
type IKRC20ReceiverCallerRaw struct {
	Contract *IKRC20ReceiverCaller // Generic read-only contract binding to access the raw methods on
}

// IKRC20ReceiverTransactorRaw is an auto generated low-level write-only Go binding around a klaytn contract.
type IKRC20ReceiverTransactorRaw struct {
	Contract *IKRC20ReceiverTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIKRC20Receiver creates a new instance of IKRC20Receiver, bound to a specific deployed contract.
func NewIKRC20Receiver(address common.Address, backend bind.ContractBackend) (*IKRC20Receiver, error) {
	contract, err := bindIKRC20Receiver(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IKRC20Receiver{IKRC20ReceiverCaller: IKRC20ReceiverCaller{contract: contract}, IKRC20ReceiverTransactor: IKRC20ReceiverTransactor{contract: contract}, IKRC20ReceiverFilterer: IKRC20ReceiverFilterer{contract: contract}}, nil
}

// NewIKRC20ReceiverCaller creates a new read-only instance of IKRC20Receiver, bound to a specific deployed contract.
func NewIKRC20ReceiverCaller(address common.Address, caller bind.ContractCaller) (*IKRC20ReceiverCaller, error) {
	contract, err := bindIKRC20Receiver(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IKRC20ReceiverCaller{contract: contract}, nil
}

// NewIKRC20ReceiverTransactor creates a new write-only instance of IKRC20Receiver, bound to a specific deployed contract.
func NewIKRC20ReceiverTransactor(address common.Address, transactor bind.ContractTransactor) (*IKRC20ReceiverTransactor, error) {
	contract, err := bindIKRC20Receiver(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IKRC20ReceiverTransactor{contract: contract}, nil
}

// NewIKRC20ReceiverFilterer creates a new log filterer instance of IKRC20Receiver, bound to a specific deployed contract.
func NewIKRC20ReceiverFilterer(address common.Address, filterer bind.ContractFilterer) (*IKRC20ReceiverFilterer, error) {
	contract, err := bindIKRC20Receiver(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IKRC20ReceiverFilterer{contract: contract}, nil
}

// bindIKRC20Receiver binds a generic wrapper to an already deployed contract.
func bindIKRC20Receiver(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IKRC20ReceiverABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IKRC20Receiver *IKRC20ReceiverRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IKRC20Receiver.Contract.IKRC20ReceiverCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IKRC20Receiver *IKRC20ReceiverRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IKRC20Receiver.Contract.IKRC20ReceiverTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IKRC20Receiver *IKRC20ReceiverRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IKRC20Receiver.Contract.IKRC20ReceiverTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IKRC20Receiver *IKRC20ReceiverCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IKRC20Receiver.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IKRC20Receiver *IKRC20ReceiverTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IKRC20Receiver.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IKRC20Receiver *IKRC20ReceiverTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IKRC20Receiver.Contract.contract.Transact(opts, method, params...)
}

// OnERC20Received is a paid mutator transaction binding the contract method 0xbc04f0af.
//
// Solidity: function onERC20Received(_from address, amount uint256) returns(bytes4)
func (_IKRC20Receiver *IKRC20ReceiverTransactor) OnERC20Received(opts *bind.TransactOpts, _from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IKRC20Receiver.contract.Transact(opts, "onERC20Received", _from, amount)
}

// OnERC20Received is a paid mutator transaction binding the contract method 0xbc04f0af.
//
// Solidity: function onERC20Received(_from address, amount uint256) returns(bytes4)
func (_IKRC20Receiver *IKRC20ReceiverSession) OnERC20Received(_from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IKRC20Receiver.Contract.OnERC20Received(&_IKRC20Receiver.TransactOpts, _from, amount)
}

// OnERC20Received is a paid mutator transaction binding the contract method 0xbc04f0af.
//
// Solidity: function onERC20Received(_from address, amount uint256) returns(bytes4)
func (_IKRC20Receiver *IKRC20ReceiverTransactorSession) OnERC20Received(_from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IKRC20Receiver.Contract.OnERC20Received(&_IKRC20Receiver.TransactOpts, _from, amount)
}

// SafeMathABI is the input ABI used to generate the binding from.
const SafeMathABI = "[]"

// SafeMathBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const SafeMathBinRuntime = `0x73000000000000000000000000000000000000000030146080604052600080fd00a165627a7a723058200a0aedb1774c8011e97b0a42c92cc3ce45ff001e6a704c353642b7762c8d60030029`

// SafeMathBin is the compiled bytecode used for deploying new contracts.
const SafeMathBin = `0x604c602c600b82828239805160001a60731460008114601c57601e565bfe5b5030600052607381538281f30073000000000000000000000000000000000000000030146080604052600080fd00a165627a7a723058200a0aedb1774c8011e97b0a42c92cc3ce45ff001e6a704c353642b7762c8d60030029`

// DeploySafeMath deploys a new klaytn contract, binding an instance of SafeMath to it.
func DeploySafeMath(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SafeMath, error) {
	parsed, err := abi.JSON(strings.NewReader(SafeMathABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SafeMathBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SafeMath{SafeMathCaller: SafeMathCaller{contract: contract}, SafeMathTransactor: SafeMathTransactor{contract: contract}, SafeMathFilterer: SafeMathFilterer{contract: contract}}, nil
}

// SafeMath is an auto generated Go binding around a klaytn contract.
type SafeMath struct {
	SafeMathCaller     // Read-only binding to the contract
	SafeMathTransactor // Write-only binding to the contract
	SafeMathFilterer   // Log filterer for contract events
}

// SafeMathCaller is an auto generated read-only Go binding around a klaytn contract.
type SafeMathCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathTransactor is an auto generated write-only Go binding around a klaytn contract.
type SafeMathTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathFilterer is an auto generated log filtering Go binding around a klaytn contract events.
type SafeMathFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathSession is an auto generated Go binding around a klaytn contract,
// with pre-set call and transact options.
type SafeMathSession struct {
	Contract     *SafeMath         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SafeMathCallerSession is an auto generated read-only Go binding around a klaytn contract,
// with pre-set call options.
type SafeMathCallerSession struct {
	Contract *SafeMathCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// SafeMathTransactorSession is an auto generated write-only Go binding around a klaytn contract,
// with pre-set transact options.
type SafeMathTransactorSession struct {
	Contract     *SafeMathTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// SafeMathRaw is an auto generated low-level Go binding around a klaytn contract.
type SafeMathRaw struct {
	Contract *SafeMath // Generic contract binding to access the raw methods on
}

// SafeMathCallerRaw is an auto generated low-level read-only Go binding around a klaytn contract.
type SafeMathCallerRaw struct {
	Contract *SafeMathCaller // Generic read-only contract binding to access the raw methods on
}

// SafeMathTransactorRaw is an auto generated low-level write-only Go binding around a klaytn contract.
type SafeMathTransactorRaw struct {
	Contract *SafeMathTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSafeMath creates a new instance of SafeMath, bound to a specific deployed contract.
func NewSafeMath(address common.Address, backend bind.ContractBackend) (*SafeMath, error) {
	contract, err := bindSafeMath(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SafeMath{SafeMathCaller: SafeMathCaller{contract: contract}, SafeMathTransactor: SafeMathTransactor{contract: contract}, SafeMathFilterer: SafeMathFilterer{contract: contract}}, nil
}

// NewSafeMathCaller creates a new read-only instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathCaller(address common.Address, caller bind.ContractCaller) (*SafeMathCaller, error) {
	contract, err := bindSafeMath(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SafeMathCaller{contract: contract}, nil
}

// NewSafeMathTransactor creates a new write-only instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathTransactor(address common.Address, transactor bind.ContractTransactor) (*SafeMathTransactor, error) {
	contract, err := bindSafeMath(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SafeMathTransactor{contract: contract}, nil
}

// NewSafeMathFilterer creates a new log filterer instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathFilterer(address common.Address, filterer bind.ContractFilterer) (*SafeMathFilterer, error) {
	contract, err := bindSafeMath(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SafeMathFilterer{contract: contract}, nil
}

// bindSafeMath binds a generic wrapper to an already deployed contract.
func bindSafeMath(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SafeMathABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SafeMath *SafeMathRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SafeMath.Contract.SafeMathCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SafeMath *SafeMathRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SafeMath.Contract.SafeMathTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SafeMath *SafeMathRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SafeMath.Contract.SafeMathTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SafeMath *SafeMathCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SafeMath.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SafeMath *SafeMathTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SafeMath.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SafeMath *SafeMathTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SafeMath.Contract.contract.Transact(opts, method, params...)
}
