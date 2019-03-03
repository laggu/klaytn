// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

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

// ERC20ABI is the input ABI used to generate the binding from.
const ERC20ABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"tokens\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokens\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenOwner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"tokens\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"tokenOwner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"tokens\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"tokenOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"tokens\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

// ERC20BinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const ERC20BinRuntime = `0x`

// ERC20Bin is the compiled bytecode used for deploying new contracts.
const ERC20Bin = `0x`

// DeployERC20 deploys a new klaytn contract, binding an instance of ERC20 to it.
func DeployERC20(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ERC20, error) {
	parsed, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ERC20Bin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ERC20{ERC20Caller: ERC20Caller{contract: contract}, ERC20Transactor: ERC20Transactor{contract: contract}, ERC20Filterer: ERC20Filterer{contract: contract}}, nil
}

// ERC20 is an auto generated Go binding around a klaytn contract.
type ERC20 struct {
	ERC20Caller     // Read-only binding to the contract
	ERC20Transactor // Write-only binding to the contract
	ERC20Filterer   // Log filterer for contract events
}

// ERC20Caller is an auto generated read-only Go binding around a klaytn contract.
type ERC20Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20Transactor is an auto generated write-only Go binding around a klaytn contract.
type ERC20Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20Filterer is an auto generated log filtering Go binding around a klaytn contract events.
type ERC20Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20Session is an auto generated Go binding around a klaytn contract,
// with pre-set call and transact options.
type ERC20Session struct {
	Contract     *ERC20            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ERC20CallerSession is an auto generated read-only Go binding around a klaytn contract,
// with pre-set call options.
type ERC20CallerSession struct {
	Contract *ERC20Caller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ERC20TransactorSession is an auto generated write-only Go binding around a klaytn contract,
// with pre-set transact options.
type ERC20TransactorSession struct {
	Contract     *ERC20Transactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ERC20Raw is an auto generated low-level Go binding around a klaytn contract.
type ERC20Raw struct {
	Contract *ERC20 // Generic contract binding to access the raw methods on
}

// ERC20CallerRaw is an auto generated low-level read-only Go binding around a klaytn contract.
type ERC20CallerRaw struct {
	Contract *ERC20Caller // Generic read-only contract binding to access the raw methods on
}

// ERC20TransactorRaw is an auto generated low-level write-only Go binding around a klaytn contract.
type ERC20TransactorRaw struct {
	Contract *ERC20Transactor // Generic write-only contract binding to access the raw methods on
}

// NewERC20 creates a new instance of ERC20, bound to a specific deployed contract.
func NewERC20(address common.Address, backend bind.ContractBackend) (*ERC20, error) {
	contract, err := bindERC20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ERC20{ERC20Caller: ERC20Caller{contract: contract}, ERC20Transactor: ERC20Transactor{contract: contract}, ERC20Filterer: ERC20Filterer{contract: contract}}, nil
}

// NewERC20Caller creates a new read-only instance of ERC20, bound to a specific deployed contract.
func NewERC20Caller(address common.Address, caller bind.ContractCaller) (*ERC20Caller, error) {
	contract, err := bindERC20(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20Caller{contract: contract}, nil
}

// NewERC20Transactor creates a new write-only instance of ERC20, bound to a specific deployed contract.
func NewERC20Transactor(address common.Address, transactor bind.ContractTransactor) (*ERC20Transactor, error) {
	contract, err := bindERC20(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20Transactor{contract: contract}, nil
}

// NewERC20Filterer creates a new log filterer instance of ERC20, bound to a specific deployed contract.
func NewERC20Filterer(address common.Address, filterer bind.ContractFilterer) (*ERC20Filterer, error) {
	contract, err := bindERC20(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ERC20Filterer{contract: contract}, nil
}

// bindERC20 binds a generic wrapper to an already deployed contract.
func bindERC20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20 *ERC20Raw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ERC20.Contract.ERC20Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20 *ERC20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20.Contract.ERC20Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20 *ERC20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20.Contract.ERC20Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20 *ERC20CallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ERC20.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20 *ERC20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20 *ERC20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(tokenOwner address, spender address) constant returns(remaining uint256)
func (_ERC20 *ERC20Caller) Allowance(opts *bind.CallOpts, tokenOwner common.Address, spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20.contract.Call(opts, out, "allowance", tokenOwner, spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(tokenOwner address, spender address) constant returns(remaining uint256)
func (_ERC20 *ERC20Session) Allowance(tokenOwner common.Address, spender common.Address) (*big.Int, error) {
	return _ERC20.Contract.Allowance(&_ERC20.CallOpts, tokenOwner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(tokenOwner address, spender address) constant returns(remaining uint256)
func (_ERC20 *ERC20CallerSession) Allowance(tokenOwner common.Address, spender common.Address) (*big.Int, error) {
	return _ERC20.Contract.Allowance(&_ERC20.CallOpts, tokenOwner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(tokenOwner address) constant returns(balance uint256)
func (_ERC20 *ERC20Caller) BalanceOf(opts *bind.CallOpts, tokenOwner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20.contract.Call(opts, out, "balanceOf", tokenOwner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(tokenOwner address) constant returns(balance uint256)
func (_ERC20 *ERC20Session) BalanceOf(tokenOwner common.Address) (*big.Int, error) {
	return _ERC20.Contract.BalanceOf(&_ERC20.CallOpts, tokenOwner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(tokenOwner address) constant returns(balance uint256)
func (_ERC20 *ERC20CallerSession) BalanceOf(tokenOwner common.Address) (*big.Int, error) {
	return _ERC20.Contract.BalanceOf(&_ERC20.CallOpts, tokenOwner)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ERC20 *ERC20Caller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ERC20 *ERC20Session) TotalSupply() (*big.Int, error) {
	return _ERC20.Contract.TotalSupply(&_ERC20.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ERC20 *ERC20CallerSession) TotalSupply() (*big.Int, error) {
	return _ERC20.Contract.TotalSupply(&_ERC20.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, tokens uint256) returns(success bool)
func (_ERC20 *ERC20Transactor) Approve(opts *bind.TransactOpts, spender common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _ERC20.contract.Transact(opts, "approve", spender, tokens)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, tokens uint256) returns(success bool)
func (_ERC20 *ERC20Session) Approve(spender common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.Approve(&_ERC20.TransactOpts, spender, tokens)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, tokens uint256) returns(success bool)
func (_ERC20 *ERC20TransactorSession) Approve(spender common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.Approve(&_ERC20.TransactOpts, spender, tokens)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, tokens uint256) returns(success bool)
func (_ERC20 *ERC20Transactor) Transfer(opts *bind.TransactOpts, to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _ERC20.contract.Transact(opts, "transfer", to, tokens)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, tokens uint256) returns(success bool)
func (_ERC20 *ERC20Session) Transfer(to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.Transfer(&_ERC20.TransactOpts, to, tokens)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, tokens uint256) returns(success bool)
func (_ERC20 *ERC20TransactorSession) Transfer(to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.Transfer(&_ERC20.TransactOpts, to, tokens)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, tokens uint256) returns(success bool)
func (_ERC20 *ERC20Transactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _ERC20.contract.Transact(opts, "transferFrom", from, to, tokens)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, tokens uint256) returns(success bool)
func (_ERC20 *ERC20Session) TransferFrom(from common.Address, to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.TransferFrom(&_ERC20.TransactOpts, from, to, tokens)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, tokens uint256) returns(success bool)
func (_ERC20 *ERC20TransactorSession) TransferFrom(from common.Address, to common.Address, tokens *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.TransferFrom(&_ERC20.TransactOpts, from, to, tokens)
}

// ERC20ApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the ERC20 contract.
type ERC20ApprovalIterator struct {
	Event *ERC20Approval // Event containing the contract specifics and raw log

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
func (it *ERC20ApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20Approval)
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
		it.Event = new(ERC20Approval)
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
func (it *ERC20ApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20ApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20Approval represents a Approval event raised by the ERC20 contract.
type ERC20Approval struct {
	TokenOwner common.Address
	Spender    common.Address
	Tokens     *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(tokenOwner indexed address, spender indexed address, tokens uint256)
func (_ERC20 *ERC20Filterer) FilterApproval(opts *bind.FilterOpts, tokenOwner []common.Address, spender []common.Address) (*ERC20ApprovalIterator, error) {

	var tokenOwnerRule []interface{}
	for _, tokenOwnerItem := range tokenOwner {
		tokenOwnerRule = append(tokenOwnerRule, tokenOwnerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ERC20.contract.FilterLogs(opts, "Approval", tokenOwnerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &ERC20ApprovalIterator{contract: _ERC20.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(tokenOwner indexed address, spender indexed address, tokens uint256)
func (_ERC20 *ERC20Filterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ERC20Approval, tokenOwner []common.Address, spender []common.Address) (event.Subscription, error) {

	var tokenOwnerRule []interface{}
	for _, tokenOwnerItem := range tokenOwner {
		tokenOwnerRule = append(tokenOwnerRule, tokenOwnerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ERC20.contract.WatchLogs(opts, "Approval", tokenOwnerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20Approval)
				if err := _ERC20.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ERC20TransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the ERC20 contract.
type ERC20TransferIterator struct {
	Event *ERC20Transfer // Event containing the contract specifics and raw log

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
func (it *ERC20TransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20Transfer)
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
		it.Event = new(ERC20Transfer)
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
func (it *ERC20TransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20TransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20Transfer represents a Transfer event raised by the ERC20 contract.
type ERC20Transfer struct {
	From   common.Address
	To     common.Address
	Tokens *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, tokens uint256)
func (_ERC20 *ERC20Filterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*ERC20TransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ERC20.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &ERC20TransferIterator{contract: _ERC20.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, tokens uint256)
func (_ERC20 *ERC20Filterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ERC20Transfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ERC20.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20Transfer)
				if err := _ERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// GatewayABI is the input ABI used to generate the binding from.
const GatewayABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"isChild\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getKLAY\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"contractAddress\",\"type\":\"address\"}],\"name\":\"depositERC20\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"contractAddress\",\"type\":\"address\"}],\"name\":\"getERC20\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"contractAddress\",\"type\":\"address\"}],\"name\":\"withdrawERC20\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"to\",\"type\":\"address\"}],\"name\":\"withdrawKLAY\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"onERC20Received\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_isChild\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"KLAYReceived\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"contractAddress\",\"type\":\"address\"}],\"name\":\"ERC20Received\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"kind\",\"type\":\"uint8\"},{\"indexed\":false,\"name\":\"contractAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"TokenWithdrawn\",\"type\":\"event\"}]"

// GatewayBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const GatewayBinRuntime = `0x6080604052600436106100825763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166311f2d49481146100c65780632980c75b146100ef578063392d661c146101165780634e0dc5571461013c5780635d29dab41461015d578063840f014514610187578063bc04f0af146101ab575b61008a610204565b6040805133815234602082015281517f4fc4e83ddb7c772abac3f1a4755dc8431e5c21e06e7782e77aaa476619348529929181900390910190a1005b3480156100d257600080fd5b506100db61021c565b604080519115158252519081900360200190f35b3480156100fb57600080fd5b50610104610225565b60408051918252519081900360200190f35b34801561012257600080fd5b5061013a600435600160a060020a036024351661022b565b005b34801561014857600080fd5b50610104600160a060020a036004351661034f565b34801561016957600080fd5b5061013a600435600160a060020a036024358116906044351661036a565b34801561019357600080fd5b5061013a600435600160a060020a03602435166104b7565b3480156101b757600080fd5b506101cf600160a060020a0360043516602435610558565b604080517fffffffff000000000000000000000000000000000000000000000000000000009092168252519081900360200190f35b600154610217903463ffffffff6105d316565b600155565b60005460ff1681565b60015490565b604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152306024820152604481018490529051600160a060020a038316916323b872dd9160648083019260209291908290030181600087803b15801561029957600080fd5b505af11580156102ad573d6000803e3d6000fd5b505050506040513d60208110156102c357600080fd5b5050600160a060020a0381166000908152600260205260409020546102ee908363ffffffff6105d316565b600160a060020a03821660008181526002602090815260409182902093909355805133815292830185905282810191909152517fa13cf347fb36122550e414f6fd1a0c2e490cff76331c4dcc20f760891ecca12a9181900360600190a15050565b600160a060020a031660009081526002602052604090205490565b60005460ff1615156103b957600160a060020a03811660009081526002602052604090205461039f908463ffffffff6105ec16565b600160a060020a0382166000908152600260205260409020555b80600160a060020a031663a9059cbb83856040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083600160a060020a0316600160a060020a0316815260200182815260200192505050602060405180830381600087803b15801561043557600080fd5b505af1158015610449573d6000803e3d6000fd5b505050506040513d602081101561045f57600080fd5b505060408051600160a060020a038481168252600160208301528316818301526060810185905290517f591f2d33d85291e32c4067b5a497caf3ddb5b1830eba9909e66006ec3a0051b49181900360800190a1505050565b6001546104ca908363ffffffff6105ec16565b600155604051600160a060020a0382169083156108fc029084906000818181858888f19350505050158015610503573d6000803e3d6000fd5b5060408051600160a060020a0383168152600060208201819052818301526060810184905290517f591f2d33d85291e32c4067b5a497caf3ddb5b1830eba9909e66006ec3a0051b49181900360800190a15050565b600061056382610603565b60408051600160a060020a038516815260208101849052338183015290517fa13cf347fb36122550e414f6fd1a0c2e490cff76331c4dcc20f760891ecca12a9181900360600190a1507fbc04f0af0000000000000000000000000000000000000000000000000000000092915050565b6000828201838110156105e557600080fd5b9392505050565b600080838311156105fc57600080fd5b5050900390565b33600090815260026020526040902054610623908263ffffffff6105d316565b33600090815260026020526040902055505600a165627a7a72305820998566e550ed5fec182f13bebfa06c8643ad5aef3b5dfd9a41b398d80f791d730029`

// GatewayBin is the compiled bytecode used for deploying new contracts.
const GatewayBin = `0x608060405234801561001057600080fd5b506040516020806106a683398101604052516000805491151560ff19909216919091179055610662806100446000396000f3006080604052600436106100825763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166311f2d49481146100c65780632980c75b146100ef578063392d661c146101165780634e0dc5571461013c5780635d29dab41461015d578063840f014514610187578063bc04f0af146101ab575b61008a610204565b6040805133815234602082015281517f4fc4e83ddb7c772abac3f1a4755dc8431e5c21e06e7782e77aaa476619348529929181900390910190a1005b3480156100d257600080fd5b506100db61021c565b604080519115158252519081900360200190f35b3480156100fb57600080fd5b50610104610225565b60408051918252519081900360200190f35b34801561012257600080fd5b5061013a600435600160a060020a036024351661022b565b005b34801561014857600080fd5b50610104600160a060020a036004351661034f565b34801561016957600080fd5b5061013a600435600160a060020a036024358116906044351661036a565b34801561019357600080fd5b5061013a600435600160a060020a03602435166104b7565b3480156101b757600080fd5b506101cf600160a060020a0360043516602435610558565b604080517fffffffff000000000000000000000000000000000000000000000000000000009092168252519081900360200190f35b600154610217903463ffffffff6105d316565b600155565b60005460ff1681565b60015490565b604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152306024820152604481018490529051600160a060020a038316916323b872dd9160648083019260209291908290030181600087803b15801561029957600080fd5b505af11580156102ad573d6000803e3d6000fd5b505050506040513d60208110156102c357600080fd5b5050600160a060020a0381166000908152600260205260409020546102ee908363ffffffff6105d316565b600160a060020a03821660008181526002602090815260409182902093909355805133815292830185905282810191909152517fa13cf347fb36122550e414f6fd1a0c2e490cff76331c4dcc20f760891ecca12a9181900360600190a15050565b600160a060020a031660009081526002602052604090205490565b60005460ff1615156103b957600160a060020a03811660009081526002602052604090205461039f908463ffffffff6105ec16565b600160a060020a0382166000908152600260205260409020555b80600160a060020a031663a9059cbb83856040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083600160a060020a0316600160a060020a0316815260200182815260200192505050602060405180830381600087803b15801561043557600080fd5b505af1158015610449573d6000803e3d6000fd5b505050506040513d602081101561045f57600080fd5b505060408051600160a060020a038481168252600160208301528316818301526060810185905290517f591f2d33d85291e32c4067b5a497caf3ddb5b1830eba9909e66006ec3a0051b49181900360800190a1505050565b6001546104ca908363ffffffff6105ec16565b600155604051600160a060020a0382169083156108fc029084906000818181858888f19350505050158015610503573d6000803e3d6000fd5b5060408051600160a060020a0383168152600060208201819052818301526060810184905290517f591f2d33d85291e32c4067b5a497caf3ddb5b1830eba9909e66006ec3a0051b49181900360800190a15050565b600061056382610603565b60408051600160a060020a038516815260208101849052338183015290517fa13cf347fb36122550e414f6fd1a0c2e490cff76331c4dcc20f760891ecca12a9181900360600190a1507fbc04f0af0000000000000000000000000000000000000000000000000000000092915050565b6000828201838110156105e557600080fd5b9392505050565b600080838311156105fc57600080fd5b5050900390565b33600090815260026020526040902054610623908263ffffffff6105d316565b33600090815260026020526040902055505600a165627a7a72305820998566e550ed5fec182f13bebfa06c8643ad5aef3b5dfd9a41b398d80f791d730029`

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

// GetERC20 is a free data retrieval call binding the contract method 0x4e0dc557.
//
// Solidity: function getERC20(contractAddress address) constant returns(uint256)
func (_Gateway *GatewayCaller) GetERC20(opts *bind.CallOpts, contractAddress common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Gateway.contract.Call(opts, out, "getERC20", contractAddress)
	return *ret0, err
}

// GetERC20 is a free data retrieval call binding the contract method 0x4e0dc557.
//
// Solidity: function getERC20(contractAddress address) constant returns(uint256)
func (_Gateway *GatewaySession) GetERC20(contractAddress common.Address) (*big.Int, error) {
	return _Gateway.Contract.GetERC20(&_Gateway.CallOpts, contractAddress)
}

// GetERC20 is a free data retrieval call binding the contract method 0x4e0dc557.
//
// Solidity: function getERC20(contractAddress address) constant returns(uint256)
func (_Gateway *GatewayCallerSession) GetERC20(contractAddress common.Address) (*big.Int, error) {
	return _Gateway.Contract.GetERC20(&_Gateway.CallOpts, contractAddress)
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

// DepositERC20 is a paid mutator transaction binding the contract method 0x392d661c.
//
// Solidity: function depositERC20(amount uint256, contractAddress address) returns()
func (_Gateway *GatewayTransactor) DepositERC20(opts *bind.TransactOpts, amount *big.Int, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "depositERC20", amount, contractAddress)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0x392d661c.
//
// Solidity: function depositERC20(amount uint256, contractAddress address) returns()
func (_Gateway *GatewaySession) DepositERC20(amount *big.Int, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.DepositERC20(&_Gateway.TransactOpts, amount, contractAddress)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0x392d661c.
//
// Solidity: function depositERC20(amount uint256, contractAddress address) returns()
func (_Gateway *GatewayTransactorSession) DepositERC20(amount *big.Int, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.DepositERC20(&_Gateway.TransactOpts, amount, contractAddress)
}

// OnERC20Received is a paid mutator transaction binding the contract method 0xbc04f0af.
//
// Solidity: function onERC20Received(_from address, amount uint256) returns(bytes4)
func (_Gateway *GatewayTransactor) OnERC20Received(opts *bind.TransactOpts, _from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "onERC20Received", _from, amount)
}

// OnERC20Received is a paid mutator transaction binding the contract method 0xbc04f0af.
//
// Solidity: function onERC20Received(_from address, amount uint256) returns(bytes4)
func (_Gateway *GatewaySession) OnERC20Received(_from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.OnERC20Received(&_Gateway.TransactOpts, _from, amount)
}

// OnERC20Received is a paid mutator transaction binding the contract method 0xbc04f0af.
//
// Solidity: function onERC20Received(_from address, amount uint256) returns(bytes4)
func (_Gateway *GatewayTransactorSession) OnERC20Received(_from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Gateway.Contract.OnERC20Received(&_Gateway.TransactOpts, _from, amount)
}

// WithdrawERC20 is a paid mutator transaction binding the contract method 0x5d29dab4.
//
// Solidity: function withdrawERC20(amount uint256, to address, contractAddress address) returns()
func (_Gateway *GatewayTransactor) WithdrawERC20(opts *bind.TransactOpts, amount *big.Int, to common.Address, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.contract.Transact(opts, "withdrawERC20", amount, to, contractAddress)
}

// WithdrawERC20 is a paid mutator transaction binding the contract method 0x5d29dab4.
//
// Solidity: function withdrawERC20(amount uint256, to address, contractAddress address) returns()
func (_Gateway *GatewaySession) WithdrawERC20(amount *big.Int, to common.Address, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.WithdrawERC20(&_Gateway.TransactOpts, amount, to, contractAddress)
}

// WithdrawERC20 is a paid mutator transaction binding the contract method 0x5d29dab4.
//
// Solidity: function withdrawERC20(amount uint256, to address, contractAddress address) returns()
func (_Gateway *GatewayTransactorSession) WithdrawERC20(amount *big.Int, to common.Address, contractAddress common.Address) (*types.Transaction, error) {
	return _Gateway.Contract.WithdrawERC20(&_Gateway.TransactOpts, amount, to, contractAddress)
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

// GatewayERC20ReceivedIterator is returned from FilterERC20Received and is used to iterate over the raw logs and unpacked data for ERC20Received events raised by the Gateway contract.
type GatewayERC20ReceivedIterator struct {
	Event *GatewayERC20Received // Event containing the contract specifics and raw log

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
func (it *GatewayERC20ReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayERC20Received)
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
		it.Event = new(GatewayERC20Received)
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
func (it *GatewayERC20ReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayERC20ReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayERC20Received represents a ERC20Received event raised by the Gateway contract.
type GatewayERC20Received struct {
	From            common.Address
	Amount          *big.Int
	ContractAddress common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterERC20Received is a free log retrieval operation binding the contract event 0xa13cf347fb36122550e414f6fd1a0c2e490cff76331c4dcc20f760891ecca12a.
//
// Solidity: e ERC20Received(from address, amount uint256, contractAddress address)
func (_Gateway *GatewayFilterer) FilterERC20Received(opts *bind.FilterOpts) (*GatewayERC20ReceivedIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "ERC20Received")
	if err != nil {
		return nil, err
	}
	return &GatewayERC20ReceivedIterator{contract: _Gateway.contract, event: "ERC20Received", logs: logs, sub: sub}, nil
}

// WatchERC20Received is a free log subscription operation binding the contract event 0xa13cf347fb36122550e414f6fd1a0c2e490cff76331c4dcc20f760891ecca12a.
//
// Solidity: e ERC20Received(from address, amount uint256, contractAddress address)
func (_Gateway *GatewayFilterer) WatchERC20Received(opts *bind.WatchOpts, sink chan<- *GatewayERC20Received) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "ERC20Received")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayERC20Received)
				if err := _Gateway.contract.UnpackLog(event, "ERC20Received", log); err != nil {
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

// GatewayKLAYReceivedIterator is returned from FilterKLAYReceived and is used to iterate over the raw logs and unpacked data for KLAYReceived events raised by the Gateway contract.
type GatewayKLAYReceivedIterator struct {
	Event *GatewayKLAYReceived // Event containing the contract specifics and raw log

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
func (it *GatewayKLAYReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GatewayKLAYReceived)
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
		it.Event = new(GatewayKLAYReceived)
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
func (it *GatewayKLAYReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GatewayKLAYReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GatewayKLAYReceived represents a KLAYReceived event raised by the Gateway contract.
type GatewayKLAYReceived struct {
	From   common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterKLAYReceived is a free log retrieval operation binding the contract event 0x4fc4e83ddb7c772abac3f1a4755dc8431e5c21e06e7782e77aaa476619348529.
//
// Solidity: e KLAYReceived(from address, amount uint256)
func (_Gateway *GatewayFilterer) FilterKLAYReceived(opts *bind.FilterOpts) (*GatewayKLAYReceivedIterator, error) {

	logs, sub, err := _Gateway.contract.FilterLogs(opts, "KLAYReceived")
	if err != nil {
		return nil, err
	}
	return &GatewayKLAYReceivedIterator{contract: _Gateway.contract, event: "KLAYReceived", logs: logs, sub: sub}, nil
}

// WatchKLAYReceived is a free log subscription operation binding the contract event 0x4fc4e83ddb7c772abac3f1a4755dc8431e5c21e06e7782e77aaa476619348529.
//
// Solidity: e KLAYReceived(from address, amount uint256)
func (_Gateway *GatewayFilterer) WatchKLAYReceived(opts *bind.WatchOpts, sink chan<- *GatewayKLAYReceived) (event.Subscription, error) {

	logs, sub, err := _Gateway.contract.WatchLogs(opts, "KLAYReceived")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GatewayKLAYReceived)
				if err := _Gateway.contract.UnpackLog(event, "KLAYReceived", log); err != nil {
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
const SafeMathBinRuntime = `0x73000000000000000000000000000000000000000030146080604052600080fd00a165627a7a72305820e434dde608567a8fc8c109afd76c0b1d1399e0b08b8ff031bf130a61d75901430029`

// SafeMathBin is the compiled bytecode used for deploying new contracts.
const SafeMathBin = `0x604c602c600b82828239805160001a60731460008114601c57601e565bfe5b5030600052607381538281f30073000000000000000000000000000000000000000030146080604052600080fd00a165627a7a72305820e434dde608567a8fc8c109afd76c0b1d1399e0b08b8ff031bf130a61d75901430029`

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
