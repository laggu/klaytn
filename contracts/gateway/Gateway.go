// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package gateway

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

// GatewayABI is the input ABI used to generate the binding from.
const GatewayABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"onTokenReceived\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isChild\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"contractAddress\",\"type\":\"address\"}],\"name\":\"withdrawToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getKLAY\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"contractAddress\",\"type\":\"address\"}],\"name\":\"getToken\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"to\",\"type\":\"address\"}],\"name\":\"withdrawKLAY\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"contractAddress\",\"type\":\"address\"}],\"name\":\"depositToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"DepositWithoutEvent\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_isChild\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"kind\",\"type\":\"uint8\"},{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"contractAddress\",\"type\":\"address\"}],\"name\":\"TokenReceived\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"kind\",\"type\":\"uint8\"},{\"indexed\":false,\"name\":\"contractAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"TokenWithdrawn\",\"type\":\"event\"}]"

// GatewayBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const GatewayBinRuntime = `0x60806040526004361061008d5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631035fb5c811461010357806311f2d4941461015c578063258898b3146101855780632980c75b146101b157806359770438146101d8578063840f0145146101f9578063c310b8841461021d578063cd93ee0614610241575b610095610249565b7fdc7c212294df6bd196463076e5266b73185c4e7b642293b42de00c1c693f9450600033346000604051808560028111156100cc57fe5b60ff168152600160a060020a03948516602082015260408082019490945291909316606082015290519081900360800192509050a1005b34801561010f57600080fd5b50610127600160a060020a0360043516602435610261565b604080517fffffffff000000000000000000000000000000000000000000000000000000009092168252519081900360200190f35b34801561016857600080fd5b506101716102ff565b604080519115158252519081900360200190f35b34801561019157600080fd5b506101af600435600160a060020a0360243581169060443516610308565b005b3480156101bd57600080fd5b506101c6610455565b60408051918252519081900360200190f35b3480156101e457600080fd5b506101c6600160a060020a036004351661045b565b34801561020557600080fd5b506101af600435600160a060020a0360243516610476565b34801561022957600080fd5b506101af600435600160a060020a0360243516610517565b6101af610645565b60015461025c903463ffffffff61064f16565b600155565b600061026c82610668565b7fdc7c212294df6bd196463076e5266b73185c4e7b642293b42de00c1c693f94506001848433604051808560028111156102a257fe5b60ff168152600160a060020a03948516602082015260408082019490945291909316606082015290519081900360800192509050a1507fbc04f0af0000000000000000000000000000000000000000000000000000000092915050565b60005460ff1681565b60005460ff16151561035757600160a060020a03811660009081526002602052604090205461033d908463ffffffff61069b16565b600160a060020a0382166000908152600260205260409020555b80600160a060020a031663a9059cbb83856040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083600160a060020a0316600160a060020a0316815260200182815260200192505050602060405180830381600087803b1580156103d357600080fd5b505af11580156103e7573d6000803e3d6000fd5b505050506040513d60208110156103fd57600080fd5b505060408051600160a060020a038481168252600160208301528316818301526060810185905290517f591f2d33d85291e32c4067b5a497caf3ddb5b1830eba9909e66006ec3a0051b49181900360800190a1505050565b60015490565b600160a060020a031660009081526002602052604090205490565b600154610489908363ffffffff61069b16565b600155604051600160a060020a0382169083156108fc029084906000818181858888f193505050501580156104c2573d6000803e3d6000fd5b5060408051600160a060020a0383168152600060208201819052818301526060810184905290517f591f2d33d85291e32c4067b5a497caf3ddb5b1830eba9909e66006ec3a0051b49181900360800190a15050565b604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152306024820152604481018490529051600160a060020a038316916323b872dd9160648083019260209291908290030181600087803b15801561058557600080fd5b505af1158015610599573d6000803e3d6000fd5b505050506040513d60208110156105af57600080fd5b5050600160a060020a0381166000908152600260205260409020546105da908363ffffffff61064f16565b600160a060020a0382166000818152600260209081526040918290209390935580516001815233938101939093528281018590526060830191909152517fdc7c212294df6bd196463076e5266b73185c4e7b642293b42de00c1c693f94509181900360800190a15050565b61064d610249565b565b60008282018381101561066157600080fd5b9392505050565b33600090815260026020526040902054610688908263ffffffff61064f16565b3360009081526002602052604090205550565b600080838311156106ab57600080fd5b50509003905600a165627a7a723058208bf9ca0c01f31088618931e459847525683bd119956a0bb22851711503cf78420029`

// GatewayBin is the compiled bytecode used for deploying new contracts.
const GatewayBin = `0x608060405260405160208061071583398101604052516000805491151560ff199092169190911790556106de806100376000396000f30060806040526004361061008d5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631035fb5c811461010357806311f2d4941461015c578063258898b3146101855780632980c75b146101b157806359770438146101d8578063840f0145146101f9578063c310b8841461021d578063cd93ee0614610241575b610095610249565b7fdc7c212294df6bd196463076e5266b73185c4e7b642293b42de00c1c693f9450600033346000604051808560028111156100cc57fe5b60ff168152600160a060020a03948516602082015260408082019490945291909316606082015290519081900360800192509050a1005b34801561010f57600080fd5b50610127600160a060020a0360043516602435610261565b604080517fffffffff000000000000000000000000000000000000000000000000000000009092168252519081900360200190f35b34801561016857600080fd5b506101716102ff565b604080519115158252519081900360200190f35b34801561019157600080fd5b506101af600435600160a060020a0360243581169060443516610308565b005b3480156101bd57600080fd5b506101c6610455565b60408051918252519081900360200190f35b3480156101e457600080fd5b506101c6600160a060020a036004351661045b565b34801561020557600080fd5b506101af600435600160a060020a0360243516610476565b34801561022957600080fd5b506101af600435600160a060020a0360243516610517565b6101af610645565b60015461025c903463ffffffff61064f16565b600155565b600061026c82610668565b7fdc7c212294df6bd196463076e5266b73185c4e7b642293b42de00c1c693f94506001848433604051808560028111156102a257fe5b60ff168152600160a060020a03948516602082015260408082019490945291909316606082015290519081900360800192509050a1507fbc04f0af0000000000000000000000000000000000000000000000000000000092915050565b60005460ff1681565b60005460ff16151561035757600160a060020a03811660009081526002602052604090205461033d908463ffffffff61069b16565b600160a060020a0382166000908152600260205260409020555b80600160a060020a031663a9059cbb83856040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083600160a060020a0316600160a060020a0316815260200182815260200192505050602060405180830381600087803b1580156103d357600080fd5b505af11580156103e7573d6000803e3d6000fd5b505050506040513d60208110156103fd57600080fd5b505060408051600160a060020a038481168252600160208301528316818301526060810185905290517f591f2d33d85291e32c4067b5a497caf3ddb5b1830eba9909e66006ec3a0051b49181900360800190a1505050565b60015490565b600160a060020a031660009081526002602052604090205490565b600154610489908363ffffffff61069b16565b600155604051600160a060020a0382169083156108fc029084906000818181858888f193505050501580156104c2573d6000803e3d6000fd5b5060408051600160a060020a0383168152600060208201819052818301526060810184905290517f591f2d33d85291e32c4067b5a497caf3ddb5b1830eba9909e66006ec3a0051b49181900360800190a15050565b604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152306024820152604481018490529051600160a060020a038316916323b872dd9160648083019260209291908290030181600087803b15801561058557600080fd5b505af1158015610599573d6000803e3d6000fd5b505050506040513d60208110156105af57600080fd5b5050600160a060020a0381166000908152600260205260409020546105da908363ffffffff61064f16565b600160a060020a0382166000818152600260209081526040918290209390935580516001815233938101939093528281018590526060830191909152517fdc7c212294df6bd196463076e5266b73185c4e7b642293b42de00c1c693f94509181900360800190a15050565b61064d610249565b565b60008282018381101561066157600080fd5b9392505050565b33600090815260026020526040902054610688908263ffffffff61064f16565b3360009081526002602052604090205550565b600080838311156106ab57600080fd5b50509003905600a165627a7a723058208bf9ca0c01f31088618931e459847525683bd119956a0bb22851711503cf78420029`

// DeployGateway deploys a new klaytn contract, binding an instance of Gateway to it.
func DeployGateway(auth *bind.TransactOpts, backend bind.ContractBackend, _isChild bool) (common.Address, *types.Transaction, *Gateway, error) {
	parsed, err := abi.JSON(strings.NewReader(GatewayABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(GatewayBin), backend, _isChild)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Gateway{GatewayCaller: GatewayCaller{contract: contract}, GatewayTransactor: GatewayTransactor{contract: contract}, GatewayFilterer: GatewayFilterer{contract: contract}}, nil
}

// Gateway is an auto generated Go binding around a klaytn contract.
type Gateway struct {
	GatewayCaller     // Read-only binding to the contract
	GatewayTransactor // Write-only binding to the contract
	GatewayFilterer   // Log filterer for contract events
}

// GatewayCaller is an auto generated read-only Go binding around a klaytn contract.
type GatewayCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatewayTransactor is an auto generated write-only Go binding around a klaytn contract.
type GatewayTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatewayFilterer is an auto generated log filtering Go binding around a klaytn contract events.
type GatewayFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GatewaySession is an auto generated Go binding around a klaytn contract,
// with pre-set call and transact options.
type GatewaySession struct {
	Contract     *Gateway          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GatewayCallerSession is an auto generated read-only Go binding around a klaytn contract,
// with pre-set call options.
type GatewayCallerSession struct {
	Contract *GatewayCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// GatewayTransactorSession is an auto generated write-only Go binding around a klaytn contract,
// with pre-set transact options.
type GatewayTransactorSession struct {
	Contract     *GatewayTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// GatewayRaw is an auto generated low-level Go binding around a klaytn contract.
type GatewayRaw struct {
	Contract *Gateway // Generic contract binding to access the raw methods on
}

// GatewayCallerRaw is an auto generated low-level read-only Go binding around a klaytn contract.
type GatewayCallerRaw struct {
	Contract *GatewayCaller // Generic read-only contract binding to access the raw methods on
}

// GatewayTransactorRaw is an auto generated low-level write-only Go binding around a klaytn contract.
type GatewayTransactorRaw struct {
	Contract *GatewayTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGateway creates a new instance of Gateway, bound to a specific deployed contract.
func NewGateway(address common.Address, backend bind.ContractBackend) (*Gateway, error) {
	contract, err := bindGateway(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Gateway{GatewayCaller: GatewayCaller{contract: contract}, GatewayTransactor: GatewayTransactor{contract: contract}, GatewayFilterer: GatewayFilterer{contract: contract}}, nil
}

// NewGatewayCaller creates a new read-only instance of Gateway, bound to a specific deployed contract.
func NewGatewayCaller(address common.Address, caller bind.ContractCaller) (*GatewayCaller, error) {
	contract, err := bindGateway(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GatewayCaller{contract: contract}, nil
}

// NewGatewayTransactor creates a new write-only instance of Gateway, bound to a specific deployed contract.
func NewGatewayTransactor(address common.Address, transactor bind.ContractTransactor) (*GatewayTransactor, error) {
	contract, err := bindGateway(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GatewayTransactor{contract: contract}, nil
}

// NewGatewayFilterer creates a new log filterer instance of Gateway, bound to a specific deployed contract.
func NewGatewayFilterer(address common.Address, filterer bind.ContractFilterer) (*GatewayFilterer, error) {
	contract, err := bindGateway(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GatewayFilterer{contract: contract}, nil
}

// bindGateway binds a generic wrapper to an already deployed contract.
func bindGateway(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(GatewayABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Gateway *GatewayRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Gateway.Contract.GatewayCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Gateway *GatewayRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gateway.Contract.GatewayTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Gateway *GatewayRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Gateway.Contract.GatewayTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Gateway *GatewayCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Gateway.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Gateway *GatewayTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gateway.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Gateway *GatewayTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Gateway.Contract.contract.Transact(opts, method, params...)
}

// GetKLAY is a free data retrieval call binding the contract method 0x2980c75b.
//
// Solidity: function getKLAY() constant returns(uint256)
func (_Gateway *GatewayCaller) GetKLAY(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Gateway.contract.Call(opts, out, "getKLAY")
	return *ret0, err
}

// GetKLAY is a free data retrieval call binding the contract method 0x2980c75b.
//
// Solidity: function getKLAY() constant returns(uint256)
func (_Gateway *GatewaySession) GetKLAY() (*big.Int, error) {
	return _Gateway.Contract.GetKLAY(&_Gateway.CallOpts)
}

// GetKLAY is a free data retrieval call binding the contract method 0x2980c75b.
//
// Solidity: function getKLAY() constant returns(uint256)
func (_Gateway *GatewayCallerSession) GetKLAY() (*big.Int, error) {
	return _Gateway.Contract.GetKLAY(&_Gateway.CallOpts)
}

// GetToken is a free data retrieval call binding the contract method 0x59770438.
//
// Solidity: function getToken(contractAddress address) constant returns(uint256)
func (_Gateway *GatewayCaller) GetToken(opts *bind.CallOpts, contractAddress common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Gateway.contract.Call(opts, out, "getToken", contractAddress)
	return *ret0, err
}

// GetToken is a free data retrieval call binding the contract method 0x59770438.
//
// Solidity: function getToken(contractAddress address) constant returns(uint256)
func (_Gateway *GatewaySession) GetToken(contractAddress common.Address) (*big.Int, error) {
	return _Gateway.Contract.GetToken(&_Gateway.CallOpts, contractAddress)
}

// GetToken is a free data retrieval call binding the contract method 0x59770438.
//
// Solidity: function getToken(contractAddress address) constant returns(uint256)
func (_Gateway *GatewayCallerSession) GetToken(contractAddress common.Address) (*big.Int, error) {
	return _Gateway.Contract.GetToken(&_Gateway.CallOpts, contractAddress)
}

// IsChild is a free data retrieval call binding the contract method 0x11f2d494.
//
// Solidity: function isChild() constant returns(bool)
func (_Gateway *GatewayCaller) IsChild(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Gateway.contract.Call(opts, out, "isChild")
	return *ret0, err
}

// IsChild is a free data retrieval call binding the contract method 0x11f2d494.
//
// Solidity: function isChild() constant returns(bool)
func (_Gateway *GatewaySession) IsChild() (bool, error) {
	return _Gateway.Contract.IsChild(&_Gateway.CallOpts)
}

// IsChild is a free data retrieval call binding the contract method 0x11f2d494.
//
// Solidity: function isChild() constant returns(bool)
func (_Gateway *GatewayCallerSession) IsChild() (bool, error) {
	return _Gateway.Contract.IsChild(&_Gateway.CallOpts)
}

// DepositWithoutEvent is a paid mutator transaction binding the contract method 0xcd93ee06.
//
// Solidity: function DepositWithoutEvent() returns()
func (_Gateway *GatewayTransactor) DepositWithoutEvent(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "DepositWithoutEvent")
}

// DepositWithoutEvent is a paid mutator transaction binding the contract method 0xcd93ee06.
//
// Solidity: function DepositWithoutEvent() returns()
func (_Gateway *GatewaySession) DepositWithoutEvent() (*types.Transaction, error) {
	return _Gateway.Contract.DepositWithoutEvent(&_Gateway.TransactOpts)
}

// DepositWithoutEvent is a paid mutator transaction binding the contract method 0xcd93ee06.
//
// Solidity: function DepositWithoutEvent() returns()
func (_Gateway *GatewayTransactorSession) DepositWithoutEvent() (*types.Transaction, error) {
	return _Gateway.Contract.DepositWithoutEvent(&_Gateway.TransactOpts)
}

// DepositToken is a paid mutator transaction binding the contract method 0xc310b884.
//
// Solidity: function depositToken(amount uint256, contractAddress address) returns()
func (_Gateway *GatewayTransactor) DepositToken(opts *bind.TransactOpts, amount *big.Int, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "depositToken", amount, contractAddress)
}

// DepositToken is a paid mutator transaction binding the contract method 0xc310b884.
//
// Solidity: function depositToken(amount uint256, contractAddress address) returns()
func (_Gateway *GatewaySession) DepositToken(amount *big.Int, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.DepositToken(&_Gateway.TransactOpts, amount, contractAddress)
}

// DepositToken is a paid mutator transaction binding the contract method 0xc310b884.
//
// Solidity: function depositToken(amount uint256, contractAddress address) returns()
func (_Gateway *GatewayTransactorSession) DepositToken(amount *big.Int, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.DepositToken(&_Gateway.TransactOpts, amount, contractAddress)
}

// OnTokenReceived is a paid mutator transaction binding the contract method 0x1035fb5c.
//
// Solidity: function onTokenReceived(_from address, amount uint256) returns(bytes4)
func (_Gateway *GatewayTransactor) OnTokenReceived(opts *bind.TransactOpts, _from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "onTokenReceived", _from, amount)
}

// OnTokenReceived is a paid mutator transaction binding the contract method 0x1035fb5c.
//
// Solidity: function onTokenReceived(_from address, amount uint256) returns(bytes4)
func (_Gateway *GatewaySession) OnTokenReceived(_from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.OnTokenReceived(&_Gateway.TransactOpts, _from, amount)
}

// OnTokenReceived is a paid mutator transaction binding the contract method 0x1035fb5c.
//
// Solidity: function onTokenReceived(_from address, amount uint256) returns(bytes4)
func (_Gateway *GatewayTransactorSession) OnTokenReceived(_from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.OnTokenReceived(&_Gateway.TransactOpts, _from, amount)
}

// WithdrawKLAY is a paid mutator transaction binding the contract method 0x840f0145.
//
// Solidity: function withdrawKLAY(amount uint256, to address) returns()
func (_Gateway *GatewayTransactor) WithdrawKLAY(opts *bind.TransactOpts, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "withdrawKLAY", amount, to)
}

// WithdrawKLAY is a paid mutator transaction binding the contract method 0x840f0145.
//
// Solidity: function withdrawKLAY(amount uint256, to address) returns()
func (_Gateway *GatewaySession) WithdrawKLAY(amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.WithdrawKLAY(&_Gateway.TransactOpts, amount, to)
}

// WithdrawKLAY is a paid mutator transaction binding the contract method 0x840f0145.
//
// Solidity: function withdrawKLAY(amount uint256, to address) returns()
func (_Gateway *GatewayTransactorSession) WithdrawKLAY(amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.WithdrawKLAY(&_Gateway.TransactOpts, amount, to)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x258898b3.
//
// Solidity: function withdrawToken(amount uint256, to address, contractAddress address) returns()
func (_Gateway *GatewayTransactor) WithdrawToken(opts *bind.TransactOpts, amount *big.Int, to common.Address, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "withdrawToken", amount, to, contractAddress)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x258898b3.
//
// Solidity: function withdrawToken(amount uint256, to address, contractAddress address) returns()
func (_Gateway *GatewaySession) WithdrawToken(amount *big.Int, to common.Address, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.WithdrawToken(&_Gateway.TransactOpts, amount, to, contractAddress)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x258898b3.
//
// Solidity: function withdrawToken(amount uint256, to address, contractAddress address) returns()
func (_Gateway *GatewayTransactorSession) WithdrawToken(amount *big.Int, to common.Address, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.WithdrawToken(&_Gateway.TransactOpts, amount, to, contractAddress)
}

// GatewayTokenReceivedIterator is returned from FilterTokenReceived and is used to iterate over the raw logs and unpacked data for TokenReceived events raised by the Gateway contract.
type GatewayTokenReceivedIterator struct {
	Event *GatewayTokenReceived // Event containing the contract specifics and raw log

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
func (it *GatewayTokenReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayTokenReceived)
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
		it.Event = new(GatewayTokenReceived)
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
func (it *GatewayTokenReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayTokenReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayTokenReceived represents a TokenReceived event raised by the Gateway contract.
type GatewayTokenReceived struct {
	Kind            uint8
	From            common.Address
	Amount          *big.Int
	ContractAddress common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterTokenReceived is a free log retrieval operation binding the contract event 0xdc7c212294df6bd196463076e5266b73185c4e7b642293b42de00c1c693f9450.
//
// Solidity: e TokenReceived(kind uint8, from address, amount uint256, contractAddress address)
func (_Gateway *GatewayFilterer) FilterTokenReceived(opts *bind.FilterOpts) (*GatewayTokenReceivedIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "TokenReceived")
	if err != nil {
		return nil, err
	}
	return &GatewayTokenReceivedIterator{contract: _Gateway.contract, event: "TokenReceived", logs: logs, sub: sub}, nil
}

// WatchTokenReceived is a free log subscription operation binding the contract event 0xdc7c212294df6bd196463076e5266b73185c4e7b642293b42de00c1c693f9450.
//
// Solidity: e TokenReceived(kind uint8, from address, amount uint256, contractAddress address)
func (_Gateway *GatewayFilterer) WatchTokenReceived(opts *bind.WatchOpts, sink chan<- *GatewayTokenReceived) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "TokenReceived")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayTokenReceived)
				if err := _Gateway.contract.UnpackLog(event, "TokenReceived", log); err != nil {
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

// GatewayTokenWithdrawnIterator is returned from FilterTokenWithdrawn and is used to iterate over the raw logs and unpacked data for TokenWithdrawn events raised by the Gateway contract.
type GatewayTokenWithdrawnIterator struct {
	Event *GatewayTokenWithdrawn // Event containing the contract specifics and raw log

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
func (it *GatewayTokenWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayTokenWithdrawn)
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
		it.Event = new(GatewayTokenWithdrawn)
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
func (it *GatewayTokenWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayTokenWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayTokenWithdrawn represents a TokenWithdrawn event raised by the Gateway contract.
type GatewayTokenWithdrawn struct {
	Owner           common.Address
	Kind            uint8
	ContractAddress common.Address
	Value           *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterTokenWithdrawn is a free log retrieval operation binding the contract event 0x591f2d33d85291e32c4067b5a497caf3ddb5b1830eba9909e66006ec3a0051b4.
//
// Solidity: e TokenWithdrawn(owner address, kind uint8, contractAddress address, value uint256)
func (_Gateway *GatewayFilterer) FilterTokenWithdrawn(opts *bind.FilterOpts) (*GatewayTokenWithdrawnIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "TokenWithdrawn")
	if err != nil {
		return nil, err
	}
	return &GatewayTokenWithdrawnIterator{contract: _Gateway.contract, event: "TokenWithdrawn", logs: logs, sub: sub}, nil
}

// WatchTokenWithdrawn is a free log subscription operation binding the contract event 0x591f2d33d85291e32c4067b5a497caf3ddb5b1830eba9909e66006ec3a0051b4.
//
// Solidity: e TokenWithdrawn(owner address, kind uint8, contractAddress address, value uint256)
func (_Gateway *GatewayFilterer) WatchTokenWithdrawn(opts *bind.WatchOpts, sink chan<- *GatewayTokenWithdrawn) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "TokenWithdrawn")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayTokenWithdrawn)
				if err := _Gateway.contract.UnpackLog(event, "TokenWithdrawn", log); err != nil {
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

// ITokenABI is the input ABI used to generate the binding from.
const ITokenABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"tokens\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokens\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenOwner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokens\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenOwner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"tokens\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"tokenOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"tokens\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

// ITokenBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const ITokenBinRuntime = `0x`

// ITokenBin is the compiled bytecode used for deploying new contracts.
const ITokenBin = `0x`

// DeployIToken deploys a new klaytn contract, binding an instance of IToken to it.
func DeployIToken(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *IToken, error) {
	parsed, err := abi.JSON(strings.NewReader(ITokenABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ITokenBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &IToken{ITokenCaller: ITokenCaller{contract: contract}, ITokenTransactor: ITokenTransactor{contract: contract}, ITokenFilterer: ITokenFilterer{contract: contract}}, nil
}

// IToken is an auto generated Go binding around a klaytn contract.
type IToken struct {
	ITokenCaller     // Read-only binding to the contract
	ITokenTransactor // Write-only binding to the contract
	ITokenFilterer   // Log filterer for contract events
}

// ITokenCaller is an auto generated read-only Go binding around a klaytn contract.
type ITokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITokenTransactor is an auto generated write-only Go binding around a klaytn contract.
type ITokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITokenFilterer is an auto generated log filtering Go binding around a klaytn contract events.
type ITokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITokenSession is an auto generated Go binding around a klaytn contract,
// with pre-set call and transact options.
type ITokenSession struct {
	Contract     *IToken           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ITokenCallerSession is an auto generated read-only Go binding around a klaytn contract,
// with pre-set call options.
type ITokenCallerSession struct {
	Contract *ITokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ITokenTransactorSession is an auto generated write-only Go binding around a klaytn contract,
// with pre-set transact options.
type ITokenTransactorSession struct {
	Contract     *ITokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ITokenRaw is an auto generated low-level Go binding around a klaytn contract.
type ITokenRaw struct {
	Contract *IToken // Generic contract binding to access the raw methods on
}

// ITokenCallerRaw is an auto generated low-level read-only Go binding around a klaytn contract.
type ITokenCallerRaw struct {
	Contract *ITokenCaller // Generic read-only contract binding to access the raw methods on
}

// ITokenTransactorRaw is an auto generated low-level write-only Go binding around a klaytn contract.
type ITokenTransactorRaw struct {
	Contract *ITokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIToken creates a new instance of IToken, bound to a specific deployed contract.
func NewIToken(address common.Address, backend bind.ContractBackend) (*IToken, error) {
	contract, err := bindIToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IToken{ITokenCaller: ITokenCaller{contract: contract}, ITokenTransactor: ITokenTransactor{contract: contract}, ITokenFilterer: ITokenFilterer{contract: contract}}, nil
}

// NewITokenCaller creates a new read-only instance of IToken, bound to a specific deployed contract.
func NewITokenCaller(address common.Address, caller bind.ContractCaller) (*ITokenCaller, error) {
	contract, err := bindIToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ITokenCaller{contract: contract}, nil
}

// NewITokenTransactor creates a new write-only instance of IToken, bound to a specific deployed contract.
func NewITokenTransactor(address common.Address, transactor bind.ContractTransactor) (*ITokenTransactor, error) {
	contract, err := bindIToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ITokenTransactor{contract: contract}, nil
}

// NewITokenFilterer creates a new log filterer instance of IToken, bound to a specific deployed contract.
func NewITokenFilterer(address common.Address, filterer bind.ContractFilterer) (*ITokenFilterer, error) {
	contract, err := bindIToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ITokenFilterer{contract: contract}, nil
}

// bindIToken binds a generic wrapper to an already deployed contract.
func bindIToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ITokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IToken *ITokenRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IToken.Contract.ITokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IToken *ITokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IToken.Contract.ITokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IToken *ITokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IToken.Contract.ITokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IToken *ITokenCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IToken *ITokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IToken *ITokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IToken.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(tokenOwner address, spender address) constant returns(remaining uint256)
func (_IToken *ITokenCaller) Allowance(opts *bind.CallOpts, tokenOwner common.Address, spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IToken.contract.Call(opts, out, "allowance", tokenOwner, spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(tokenOwner address, spender address) constant returns(remaining uint256)
func (_IToken *ITokenSession) Allowance(tokenOwner common.Address, spender common.Address) (*big.Int, error) {
	return _IToken.Contract.Allowance(&_IToken.CallOpts, tokenOwner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(tokenOwner address, spender address) constant returns(remaining uint256)
func (_IToken *ITokenCallerSession) Allowance(tokenOwner common.Address, spender common.Address) (*big.Int, error) {
	return _IToken.Contract.Allowance(&_IToken.CallOpts, tokenOwner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(tokenOwner address) constant returns(balance uint256)
func (_IToken *ITokenCaller) BalanceOf(opts *bind.CallOpts, tokenOwner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IToken.contract.Call(opts, out, "balanceOf", tokenOwner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(tokenOwner address) constant returns(balance uint256)
func (_IToken *ITokenSession) BalanceOf(tokenOwner common.Address) (*big.Int, error) {
	return _IToken.Contract.BalanceOf(&_IToken.CallOpts, tokenOwner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(tokenOwner address) constant returns(balance uint256)
func (_IToken *ITokenCallerSession) BalanceOf(tokenOwner common.Address) (*big.Int, error) {
	return _IToken.Contract.BalanceOf(&_IToken.CallOpts, tokenOwner)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_IToken *ITokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IToken.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_IToken *ITokenSession) TotalSupply() (*big.Int, error) {
	return _IToken.Contract.TotalSupply(&_IToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_IToken *ITokenCallerSession) TotalSupply() (*big.Int, error) {
	return _IToken.Contract.TotalSupply(&_IToken.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, tokens uint256) returns(success bool)
func (_IToken *ITokenTransactor) Approve(opts *bind.TransactOpts, spender common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _IToken.contract.Transact(opts, "approve", spender, tokens)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, tokens uint256) returns(success bool)
func (_IToken *ITokenSession) Approve(spender common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.Approve(&_IToken.TransactOpts, spender, tokens)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, tokens uint256) returns(success bool)
func (_IToken *ITokenTransactorSession) Approve(spender common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.Approve(&_IToken.TransactOpts, spender, tokens)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, tokens uint256) returns(success bool)
func (_IToken *ITokenTransactor) Transfer(opts *bind.TransactOpts, to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _IToken.contract.Transact(opts, "transfer", to, tokens)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, tokens uint256) returns(success bool)
func (_IToken *ITokenSession) Transfer(to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.Transfer(&_IToken.TransactOpts, to, tokens)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, tokens uint256) returns(success bool)
func (_IToken *ITokenTransactorSession) Transfer(to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.Transfer(&_IToken.TransactOpts, to, tokens)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, tokens uint256) returns(success bool)
func (_IToken *ITokenTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _IToken.contract.Transact(opts, "transferFrom", from, to, tokens)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, tokens uint256) returns(success bool)
func (_IToken *ITokenSession) TransferFrom(from common.Address, to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.TransferFrom(&_IToken.TransactOpts, from, to, tokens)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, tokens uint256) returns(success bool)
func (_IToken *ITokenTransactorSession) TransferFrom(from common.Address, to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.TransferFrom(&_IToken.TransactOpts, from, to, tokens)
}

// ITokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the IToken contract.
type ITokenApprovalIterator struct {
	Event *ITokenApproval // Event containing the contract specifics and raw log

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
func (it *ITokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITokenApproval)
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
		it.Event = new(ITokenApproval)
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
func (it *ITokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITokenApproval represents a Approval event raised by the IToken contract.
type ITokenApproval struct {
	TokenOwner common.Address
	Spender    common.Address
	Tokens     *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(tokenOwner indexed address, spender indexed address, tokens uint256)
func (_IToken *ITokenFilterer) FilterApproval(opts *bind.FilterOpts, tokenOwner []common.Address, spender []common.Address) (*ITokenApprovalIterator, error) {

	var tokenOwnerRule []interface{}
	for _, tokenOwnerItem := range tokenOwner {
		tokenOwnerRule = append(tokenOwnerRule, tokenOwnerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _IToken.contract.FilterLogs(opts, "Approval", tokenOwnerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &ITokenApprovalIterator{contract: _IToken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(tokenOwner indexed address, spender indexed address, tokens uint256)
func (_IToken *ITokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ITokenApproval, tokenOwner []common.Address, spender []common.Address) (event.Subscription, error) {

	var tokenOwnerRule []interface{}
	for _, tokenOwnerItem := range tokenOwner {
		tokenOwnerRule = append(tokenOwnerRule, tokenOwnerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _IToken.contract.WatchLogs(opts, "Approval", tokenOwnerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITokenApproval)
				if err := _IToken.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ITokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the IToken contract.
type ITokenTransferIterator struct {
	Event *ITokenTransfer // Event containing the contract specifics and raw log

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
func (it *ITokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITokenTransfer)
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
		it.Event = new(ITokenTransfer)
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
func (it *ITokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITokenTransfer represents a Transfer event raised by the IToken contract.
type ITokenTransfer struct {
	From   common.Address
	To     common.Address
	Tokens *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, tokens uint256)
func (_IToken *ITokenFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*ITokenTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IToken.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &ITokenTransferIterator{contract: _IToken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, tokens uint256)
func (_IToken *ITokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ITokenTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IToken.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITokenTransfer)
				if err := _IToken.contract.UnpackLog(event, "Transfer", log); err != nil {
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
const ITokenReceiverABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"onTokenReceived\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

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

// OnTokenReceived is a paid mutator transaction binding the contract method 0x1035fb5c.
//
// Solidity: function onTokenReceived(_from address, amount uint256) returns(bytes4)
func (_ITokenReceiver *ITokenReceiverTransactor) OnTokenReceived(opts *bind.TransactOpts, _from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ITokenReceiver.contract.Transact(opts, "onTokenReceived", _from, amount)
}

// OnTokenReceived is a paid mutator transaction binding the contract method 0x1035fb5c.
//
// Solidity: function onTokenReceived(_from address, amount uint256) returns(bytes4)
func (_ITokenReceiver *ITokenReceiverSession) OnTokenReceived(_from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ITokenReceiver.Contract.OnTokenReceived(&_ITokenReceiver.TransactOpts, _from, amount)
}

// OnTokenReceived is a paid mutator transaction binding the contract method 0x1035fb5c.
//
// Solidity: function onTokenReceived(_from address, amount uint256) returns(bytes4)
func (_ITokenReceiver *ITokenReceiverTransactorSession) OnTokenReceived(_from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ITokenReceiver.Contract.OnTokenReceived(&_ITokenReceiver.TransactOpts, _from, amount)
}

// SafeMathABI is the input ABI used to generate the binding from.
const SafeMathABI = "[]"

// SafeMathBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const SafeMathBinRuntime = `0x73000000000000000000000000000000000000000030146080604052600080fd00a165627a7a7230582044bbc6c278982806a1ee9e1138fd41534482b17ec97082e740df6f4465f832bd0029`

// SafeMathBin is the compiled bytecode used for deploying new contracts.
const SafeMathBin = `0x604c602c600b82828239805160001a60731460008114601c57601e565bfe5b5030600052607381538281f30073000000000000000000000000000000000000000030146080604052600080fd00a165627a7a7230582044bbc6c278982806a1ee9e1138fd41534482b17ec97082e740df6f4465f832bd0029`

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
