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
const GXTokenABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"balanceOfMine\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"INITIAL_SUPPLY\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"mintToGateway\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_gateway\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"},{\"name\":\"_to\",\"type\":\"address\"}],\"name\":\"safeTransferAndCall\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"to\",\"type\":\"address\"}],\"name\":\"depositToGateway\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_gateway\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"}]"

// GXTokenBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const GXTokenBinRuntime = `0x6080604052600436106100a35763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde0381146100a85780631459cef4146101325780632ff2e9dc14610159578063313ce5671461016e578063544297f5146101995780635a7df164146101b35780636d1a473a146101de57806370a082311461020257806395d89b4114610223578063a9059cbb14610238575b600080fd5b3480156100b457600080fd5b506100bd610270565b6040805160208082528351818301528351919283929083019185019080838360005b838110156100f75781810151838201526020016100df565b50505050905090810190601f1680156101245780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561013e57600080fd5b506101476102fd565b60408051918252519081900360200190f35b34801561016557600080fd5b50610147610310565b34801561017a57600080fd5b50610183610318565b6040805160ff9092168252519081900360200190f35b3480156101a557600080fd5b506101b1600435610321565b005b3480156101bf57600080fd5b506101b1600160a060020a036004358116906024359060443516610393565b3480156101ea57600080fd5b506101b1600435600160a060020a03602435166103ba565b34801561020e57600080fd5b50610147600160a060020a03600435166103d5565b34801561022f57600080fd5b506100bd6103f0565b34801561024457600080fd5b5061025c600160a060020a0360043516602435610448565b604080519115158252519081900360200190f35b60018054604080516020600284861615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156102f55780601f106102ca576101008083540402835291602001916102f5565b820191906000526020600020905b8154815290600101906020018083116102d857829003601f168201915b505050505081565b3360009081526005602052604090205490565b633b9aca0081565b60035460ff1681565b600054600160a060020a0316331461033857600080fd5b60045461034b908263ffffffff61045e16565b60045560008054600160a060020a0316815260056020526040902054610377908263ffffffff61045e16565b60008054600160a060020a031681526005602052604090205550565b61039d8383610448565b506103aa33848484610477565b15156103b557600080fd5b505050565b6000546103d190600160a060020a03168383610393565b5050565b600160a060020a031660009081526005602052604090205490565b6002805460408051602060018416156101000260001901909316849004601f810184900484028201840190925281815292918301828280156102f55780601f106102ca576101008083540402835291602001916102f5565b6000610455338484610569565b50600192915050565b60008282018381101561047057600080fd5b9392505050565b604080517ff099d9bd000000000000000000000000000000000000000000000000000000008152600160a060020a038681166004830152602482018590528381166044830152915160009283929087169163f099d9bd9160648082019260209290919082900301818787803b1580156104ef57600080fd5b505af1158015610503573d6000803e3d6000fd5b505050506040513d602081101561051957600080fd5b50517fffffffff00000000000000000000000000000000000000000000000000000000167fbc04f0af00000000000000000000000000000000000000000000000000000000149695505050505050565b600160a060020a038216151561057e57600080fd5b600160a060020a0383166000908152600560205260409020546105a7908263ffffffff61063f16565b600160a060020a0380851660009081526005602052604080822093909355908416815220546105dc908263ffffffff61045e16565b600160a060020a0380841660008181526005602090815260409182902094909455805192871683529282015280820183905290517fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9181900360600190a1505050565b6000808383111561064f57600080fd5b50509003905600a165627a7a72305820882b182aee5602d0727661b68503af71e567c24db47de0d68540b878e3850f0f0029`

// GXTokenBin is the compiled bytecode used for deploying new contracts.
const GXTokenBin = `0x60c0604052600760808190527f4758546f6b656e0000000000000000000000000000000000000000000000000060a090815261003e91600191906100f4565b506040805180820190915260028082527f4758000000000000000000000000000000000000000000000000000000000000602090920191825261008191816100f4565b506003805460ff1916601217905534801561009b57600080fd5b50604051602080610820833981016040908152905160008054600160a060020a031916600160a060020a03909216918217815560035460ff16600a0a633b9aca000260048190559181526005602052919091205561018f565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061013557805160ff1916838001178555610162565b82800160010185558215610162579182015b82811115610162578251825591602001919060010190610147565b5061016e929150610172565b5090565b61018c91905b8082111561016e5760008155600101610178565b90565b6106828061019e6000396000f3006080604052600436106100a35763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde0381146100a85780631459cef4146101325780632ff2e9dc14610159578063313ce5671461016e578063544297f5146101995780635a7df164146101b35780636d1a473a146101de57806370a082311461020257806395d89b4114610223578063a9059cbb14610238575b600080fd5b3480156100b457600080fd5b506100bd610270565b6040805160208082528351818301528351919283929083019185019080838360005b838110156100f75781810151838201526020016100df565b50505050905090810190601f1680156101245780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561013e57600080fd5b506101476102fd565b60408051918252519081900360200190f35b34801561016557600080fd5b50610147610310565b34801561017a57600080fd5b50610183610318565b6040805160ff9092168252519081900360200190f35b3480156101a557600080fd5b506101b1600435610321565b005b3480156101bf57600080fd5b506101b1600160a060020a036004358116906024359060443516610393565b3480156101ea57600080fd5b506101b1600435600160a060020a03602435166103ba565b34801561020e57600080fd5b50610147600160a060020a03600435166103d5565b34801561022f57600080fd5b506100bd6103f0565b34801561024457600080fd5b5061025c600160a060020a0360043516602435610448565b604080519115158252519081900360200190f35b60018054604080516020600284861615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156102f55780601f106102ca576101008083540402835291602001916102f5565b820191906000526020600020905b8154815290600101906020018083116102d857829003601f168201915b505050505081565b3360009081526005602052604090205490565b633b9aca0081565b60035460ff1681565b600054600160a060020a0316331461033857600080fd5b60045461034b908263ffffffff61045e16565b60045560008054600160a060020a0316815260056020526040902054610377908263ffffffff61045e16565b60008054600160a060020a031681526005602052604090205550565b61039d8383610448565b506103aa33848484610477565b15156103b557600080fd5b505050565b6000546103d190600160a060020a03168383610393565b5050565b600160a060020a031660009081526005602052604090205490565b6002805460408051602060018416156101000260001901909316849004601f810184900484028201840190925281815292918301828280156102f55780601f106102ca576101008083540402835291602001916102f5565b6000610455338484610569565b50600192915050565b60008282018381101561047057600080fd5b9392505050565b604080517ff099d9bd000000000000000000000000000000000000000000000000000000008152600160a060020a038681166004830152602482018590528381166044830152915160009283929087169163f099d9bd9160648082019260209290919082900301818787803b1580156104ef57600080fd5b505af1158015610503573d6000803e3d6000fd5b505050506040513d602081101561051957600080fd5b50517fffffffff00000000000000000000000000000000000000000000000000000000167fbc04f0af00000000000000000000000000000000000000000000000000000000149695505050505050565b600160a060020a038216151561057e57600080fd5b600160a060020a0383166000908152600560205260409020546105a7908263ffffffff61063f16565b600160a060020a0380851660009081526005602052604080822093909355908416815220546105dc908263ffffffff61045e16565b600160a060020a0380841660008181526005602090815260409182902094909455805192871683529282015280820183905290517fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9181900360600190a1505050565b6000808383111561064f57600080fd5b50509003905600a165627a7a72305820882b182aee5602d0727661b68503af71e567c24db47de0d68540b878e3850f0f0029`

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

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(owner address) constant returns(uint256)
func (_GXToken *GXTokenCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _GXToken.contract.Call(opts, out, "balanceOf", owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(owner address) constant returns(uint256)
func (_GXToken *GXTokenSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _GXToken.Contract.BalanceOf(&_GXToken.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(owner address) constant returns(uint256)
func (_GXToken *GXTokenCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _GXToken.Contract.BalanceOf(&_GXToken.CallOpts, owner)
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

// DepositToGateway is a paid mutator transaction binding the contract method 0x6d1a473a.
//
// Solidity: function depositToGateway(amount uint256, to address) returns()
func (_GXToken *GXTokenTransactor) DepositToGateway(opts *bind.TransactOpts, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _GXToken.contract.Transact(opts, "depositToGateway", amount, to)
}

// DepositToGateway is a paid mutator transaction binding the contract method 0x6d1a473a.
//
// Solidity: function depositToGateway(amount uint256, to address) returns()
func (_GXToken *GXTokenSession) DepositToGateway(amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _GXToken.Contract.DepositToGateway(&_GXToken.TransactOpts, amount, to)
}

// DepositToGateway is a paid mutator transaction binding the contract method 0x6d1a473a.
//
// Solidity: function depositToGateway(amount uint256, to address) returns()
func (_GXToken *GXTokenTransactorSession) DepositToGateway(amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _GXToken.Contract.DepositToGateway(&_GXToken.TransactOpts, amount, to)
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

// SafeTransferAndCall is a paid mutator transaction binding the contract method 0x5a7df164.
//
// Solidity: function safeTransferAndCall(_gateway address, _amount uint256, _to address) returns()
func (_GXToken *GXTokenTransactor) SafeTransferAndCall(opts *bind.TransactOpts, _gateway common.Address, _amount *big.Int, _to common.Address) (*types.Transaction, error) {
	return _GXToken.contract.Transact(opts, "safeTransferAndCall", _gateway, _amount, _to)
}

// SafeTransferAndCall is a paid mutator transaction binding the contract method 0x5a7df164.
//
// Solidity: function safeTransferAndCall(_gateway address, _amount uint256, _to address) returns()
func (_GXToken *GXTokenSession) SafeTransferAndCall(_gateway common.Address, _amount *big.Int, _to common.Address) (*types.Transaction, error) {
	return _GXToken.Contract.SafeTransferAndCall(&_GXToken.TransactOpts, _gateway, _amount, _to)
}

// SafeTransferAndCall is a paid mutator transaction binding the contract method 0x5a7df164.
//
// Solidity: function safeTransferAndCall(_gateway address, _amount uint256, _to address) returns()
func (_GXToken *GXTokenTransactorSession) SafeTransferAndCall(_gateway common.Address, _amount *big.Int, _to common.Address) (*types.Transaction, error) {
	return _GXToken.Contract.SafeTransferAndCall(&_GXToken.TransactOpts, _gateway, _amount, _to)
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

// ITokenReceiverABI is the input ABI used to generate the binding from.
const ITokenReceiverABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"_to\",\"type\":\"address\"}],\"name\":\"onTokenReceived\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ITokenReceiverBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const ITokenReceiverBinRuntime = `0x`

// ITokenReceiverBin is the compiled bytecode used for deploying new contracts.
const ITokenReceiverBin = `0x`

// DeployITokenReceiver deploys a new klaytn contract, binding an instance of ITokenReceiver to it.
func DeployITokenReceiver(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ITokenReceiver, error) {
	parsed, err := abi.JSON(strings.NewReader(ITokenReceiverABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ITokenReceiverBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ITokenReceiver{ITokenReceiverCaller: ITokenReceiverCaller{contract: contract}, ITokenReceiverTransactor: ITokenReceiverTransactor{contract: contract}, ITokenReceiverFilterer: ITokenReceiverFilterer{contract: contract}}, nil
}

// ITokenReceiver is an auto generated Go binding around a klaytn contract.
type ITokenReceiver struct {
	ITokenReceiverCaller     // Read-only binding to the contract
	ITokenReceiverTransactor // Write-only binding to the contract
	ITokenReceiverFilterer   // Log filterer for contract events
}

// ITokenReceiverCaller is an auto generated read-only Go binding around a klaytn contract.
type ITokenReceiverCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITokenReceiverTransactor is an auto generated write-only Go binding around a klaytn contract.
type ITokenReceiverTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITokenReceiverFilterer is an auto generated log filtering Go binding around a klaytn contract events.
type ITokenReceiverFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITokenReceiverSession is an auto generated Go binding around a klaytn contract,
// with pre-set call and transact options.
type ITokenReceiverSession struct {
	Contract     *ITokenReceiver   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ITokenReceiverCallerSession is an auto generated read-only Go binding around a klaytn contract,
// with pre-set call options.
type ITokenReceiverCallerSession struct {
	Contract *ITokenReceiverCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// ITokenReceiverTransactorSession is an auto generated write-only Go binding around a klaytn contract,
// with pre-set transact options.
type ITokenReceiverTransactorSession struct {
	Contract     *ITokenReceiverTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// ITokenReceiverRaw is an auto generated low-level Go binding around a klaytn contract.
type ITokenReceiverRaw struct {
	Contract *ITokenReceiver // Generic contract binding to access the raw methods on
}

// ITokenReceiverCallerRaw is an auto generated low-level read-only Go binding around a klaytn contract.
type ITokenReceiverCallerRaw struct {
	Contract *ITokenReceiverCaller // Generic read-only contract binding to access the raw methods on
}

// ITokenReceiverTransactorRaw is an auto generated low-level write-only Go binding around a klaytn contract.
type ITokenReceiverTransactorRaw struct {
	Contract *ITokenReceiverTransactor // Generic write-only contract binding to access the raw methods on
}

// NewITokenReceiver creates a new instance of ITokenReceiver, bound to a specific deployed contract.
func NewITokenReceiver(address common.Address, backend bind.ContractBackend) (*ITokenReceiver, error) {
	contract, err := bindITokenReceiver(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ITokenReceiver{ITokenReceiverCaller: ITokenReceiverCaller{contract: contract}, ITokenReceiverTransactor: ITokenReceiverTransactor{contract: contract}, ITokenReceiverFilterer: ITokenReceiverFilterer{contract: contract}}, nil
}

// NewITokenReceiverCaller creates a new read-only instance of ITokenReceiver, bound to a specific deployed contract.
func NewITokenReceiverCaller(address common.Address, caller bind.ContractCaller) (*ITokenReceiverCaller, error) {
	contract, err := bindITokenReceiver(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ITokenReceiverCaller{contract: contract}, nil
}

// NewITokenReceiverTransactor creates a new write-only instance of ITokenReceiver, bound to a specific deployed contract.
func NewITokenReceiverTransactor(address common.Address, transactor bind.ContractTransactor) (*ITokenReceiverTransactor, error) {
	contract, err := bindITokenReceiver(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ITokenReceiverTransactor{contract: contract}, nil
}

// NewITokenReceiverFilterer creates a new log filterer instance of ITokenReceiver, bound to a specific deployed contract.
func NewITokenReceiverFilterer(address common.Address, filterer bind.ContractFilterer) (*ITokenReceiverFilterer, error) {
	contract, err := bindITokenReceiver(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ITokenReceiverFilterer{contract: contract}, nil
}

// bindITokenReceiver binds a generic wrapper to an already deployed contract.
func bindITokenReceiver(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ITokenReceiverABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ITokenReceiver *ITokenReceiverRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ITokenReceiver.Contract.ITokenReceiverCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ITokenReceiver *ITokenReceiverRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ITokenReceiver.Contract.ITokenReceiverTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ITokenReceiver *ITokenReceiverRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ITokenReceiver.Contract.ITokenReceiverTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ITokenReceiver *ITokenReceiverCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ITokenReceiver.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ITokenReceiver *ITokenReceiverTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ITokenReceiver.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ITokenReceiver *ITokenReceiverTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ITokenReceiver.Contract.contract.Transact(opts, method, params...)
}

// OnTokenReceived is a paid mutator transaction binding the contract method 0xf099d9bd.
//
// Solidity: function onTokenReceived(_from address, amount uint256, _to address) returns(bytes4)
func (_ITokenReceiver *ITokenReceiverTransactor) OnTokenReceived(opts *bind.TransactOpts, _from common.Address, amount *big.Int, _to common.Address) (*types.Transaction, error) {
	return _ITokenReceiver.contract.Transact(opts, "onTokenReceived", _from, amount, _to)
}

// OnTokenReceived is a paid mutator transaction binding the contract method 0xf099d9bd.
//
// Solidity: function onTokenReceived(_from address, amount uint256, _to address) returns(bytes4)
func (_ITokenReceiver *ITokenReceiverSession) OnTokenReceived(_from common.Address, amount *big.Int, _to common.Address) (*types.Transaction, error) {
	return _ITokenReceiver.Contract.OnTokenReceived(&_ITokenReceiver.TransactOpts, _from, amount, _to)
}

// OnTokenReceived is a paid mutator transaction binding the contract method 0xf099d9bd.
//
// Solidity: function onTokenReceived(_from address, amount uint256, _to address) returns(bytes4)
func (_ITokenReceiver *ITokenReceiverTransactorSession) OnTokenReceived(_from common.Address, amount *big.Int, _to common.Address) (*types.Transaction, error) {
	return _ITokenReceiver.Contract.OnTokenReceived(&_ITokenReceiver.TransactOpts, _from, amount, _to)
}

// SafeMathABI is the input ABI used to generate the binding from.
const SafeMathABI = "[]"

// SafeMathBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const SafeMathBinRuntime = `0x73000000000000000000000000000000000000000030146080604052600080fd00a165627a7a72305820adaf8068b306227c96e8741baa1bab1e1c8101a2b215b5aa8f7d65d64d696d590029`

// SafeMathBin is the compiled bytecode used for deploying new contracts.
const SafeMathBin = `0x604c602c600b82828239805160001a60731460008114601c57601e565bfe5b5030600052607381538281f30073000000000000000000000000000000000000000030146080604052600080fd00a165627a7a72305820adaf8068b306227c96e8741baa1bab1e1c8101a2b215b5aa8f7d65d64d696d590029`

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
