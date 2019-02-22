// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"strings"

	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
)

// AddressBookABI is the input ABI used to generate the binding from.
const AddressBookABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_adminList\",\"type\":\"address[]\"},{\"name\":\"_cnNodeIdList\",\"type\":\"address[]\"},{\"name\":\"_cnStakingContractList\",\"type\":\"address[]\"},{\"name\":\"_cnRewardAddressList\",\"type\":\"address[]\"},{\"name\":\"_pocContract\",\"type\":\"address\"},{\"name\":\"_kirContract\",\"type\":\"address\"}],\"name\":\"init\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getAllAddressInfo\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"},{\"name\":\"\",\"type\":\"address[]\"},{\"name\":\"\",\"type\":\"address[]\"},{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"completeInitialization\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"testAddCn\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"initTest\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// AddressBookBinRuntime is the compiled bytecode used for adding genesis block without deploying code.
const AddressBookBinRuntime = `0x60806040526004361061006c5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630f97899c8114610071578063160370b8146101895780635827b25014610293578063eaee0716146102bc578063f4c7bc34146102d1575b600080fd5b34801561007d57600080fd5b506040805160206004803580820135838102808601850190965280855261018795369593946024949385019291829185019084908082843750506040805187358901803560208181028481018201909552818452989b9a998901989297509082019550935083925085019084908082843750506040805187358901803560208181028481018201909552818452989b9a998901989297509082019550935083925085019084908082843750506040805187358901803560208181028481018201909552818452989b9a99890198929750908201955093508392508501908490808284375094975050600160a060020a0385358116965060209095013590941693506102e692505050565b005b34801561019557600080fd5b5061019e610385565b60408051600160a060020a0380851660608301528316608082015260a080825287519082015286519091829160208084019284019160c08501918b8101910280838360005b838110156101fb5781810151838201526020016101e3565b50505050905001848103835288818151815260200191508051906020019060200280838360005b8381101561023a578181015183820152602001610222565b50505050905001848103825287818151815260200191508051906020019060200280838360005b83811015610279578181015183820152602001610261565b505050509050019850505050505050505060405180910390f35b34801561029f57600080fd5b506102a86104e6565b604080519115158252519081900360200190f35b3480156102c857600080fd5b506102a8610515565b3480156102dd57600080fd5b5061018761063b565b33156102f157600080fd5b6001600281905586516103099190602089019061090e565b50845161031d90600390602088019061090e565b50835161033190600490602087019061090e565b50825161034590600590602086019061090e565b5060068054600160a060020a0393841673ffffffffffffffffffffffffffffffffffffffff19918216179091556007805492909316911617905550505050565b600080546060918291829190819060ff1615156001146103a457600080fd5b60065460075460038054604080516020808402820181019092528281529294600494600594600160a060020a0392831694919092169287919083018282801561041657602002820191906000526020600020905b8154600160a060020a031681526001909101906020018083116103f8575b505050505094508380548060200260200160405190810160405280929190818152602001828054801561047257602002820191906000526020600020905b8154600160a060020a03168152600190910190602001808311610454575b50505050509350828054806020026020016040519081016040528092919081815260200182805480156104ce57602002820191906000526020600020905b8154600160a060020a031681526001909101906020018083116104b0575b50505050509250945094509450945094509091929394565b6000805460ff16156104f757600080fd5b600154151561050557600080fd5b6000805460ff1916600117905590565b60008054819060ff16151560011461052c57600080fd5b5060005b600a811015610637576003805460018181019092557fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b01805473ffffffffffffffffffffffffffffffffffffffff1990811673a22499738b961e56fb833d9368241a6a789e77c417909155600480548084019091557f8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65c6b64bfe7fe36bd19b018054821673c63ba6b8a9f2d33ee26a7fd0a59d155131f2701a1790556005805480840182556000919091527f036b6384b5eca791c62761152d0c79bb0604c104a5fb6f4eb0703f3154bb3db001805490911673193efb18f18f93e32e418e383714279bdade4a8117905501610530565b5090565b60016002819055805480820182557fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf601805473ffffffffffffffffffffffffffffffffffffffff199081167318254160af9c10f43db77566c294db7d8182caca179091556003805480840182557fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b9081018054841673060107a87178fcd01822d616e17eb5b26dd279ad179055815480850183558101805484167382ad1aea897ba3e72ee3632c08fe7596a9ff9c3f17905581548085018355810180548416730994ebd481f77768a8bf20d35d8f405ce21d2176179055815480850190925501805482167378d2ad8c09dce2eba9c25ef8201607ead8637b8e1790556004805480840182557f8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65c6b64bfe7fe36bd19b9081018054841673ac5e047d39692be8c81d0724543d5de721d0dd541790558154808501835581018054841673382ef85439dc874a0f55ab4d9801a5056e371b371790558154808501835581018054841673b821a659c21cb39745144931e71d0e9d09c8647f1790558154808501909255018054821673c1094c666657937ab7ed23040207a8ee6878135017905560058054808401825560008290527f036b6384b5eca791c62761152d0c79bb0604c104a5fb6f4eb0703f3154bb3db090810180548416739b0bf94f5ab62c73e454dfc55adb2d2fa6cd3af517905581548085018355810180548416734d83f1795ecdd684e94f1c5893ae6904ebeaeb9417905581548085018355810180548416739b58fe24f7a7cb9d102e21b3376bd80eefdc320b1790558154938401909155919091018054821673e60bf7b625e54e9f67767fad0e564f6aec29765217905560068054821673142441cb0896d4cd2ecdf3328d8d841a07f4b04f1790556007805490911673a36c921743d63361258fdbd107b906be0ad87940179055565b828054828255906000526020600020908101928215610970579160200282015b82811115610970578251825473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0390911617825560209092019160019091019061092e565b50610637926109a79250905b8082111561063757805473ffffffffffffffffffffffffffffffffffffffff1916815560010161097c565b905600a165627a7a72305820c1ba68771b503d0510928439c7f56625af3727ac2b3070b8cd42191f13b110ea0029`

// AddressBookBin is the compiled bytecode used for deploying new contracts.
const AddressBookBin = `0x60806040526000805460ff1916905534801561001a57600080fd5b506109d68061002a6000396000f30060806040526004361061006c5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630f97899c8114610071578063160370b8146101895780635827b25014610293578063eaee0716146102bc578063f4c7bc34146102d1575b600080fd5b34801561007d57600080fd5b506040805160206004803580820135838102808601850190965280855261018795369593946024949385019291829185019084908082843750506040805187358901803560208181028481018201909552818452989b9a998901989297509082019550935083925085019084908082843750506040805187358901803560208181028481018201909552818452989b9a998901989297509082019550935083925085019084908082843750506040805187358901803560208181028481018201909552818452989b9a99890198929750908201955093508392508501908490808284375094975050600160a060020a0385358116965060209095013590941693506102e692505050565b005b34801561019557600080fd5b5061019e610385565b60408051600160a060020a0380851660608301528316608082015260a080825287519082015286519091829160208084019284019160c08501918b8101910280838360005b838110156101fb5781810151838201526020016101e3565b50505050905001848103835288818151815260200191508051906020019060200280838360005b8381101561023a578181015183820152602001610222565b50505050905001848103825287818151815260200191508051906020019060200280838360005b83811015610279578181015183820152602001610261565b505050509050019850505050505050505060405180910390f35b34801561029f57600080fd5b506102a86104e6565b604080519115158252519081900360200190f35b3480156102c857600080fd5b506102a8610515565b3480156102dd57600080fd5b5061018761063b565b33156102f157600080fd5b6001600281905586516103099190602089019061090e565b50845161031d90600390602088019061090e565b50835161033190600490602087019061090e565b50825161034590600590602086019061090e565b5060068054600160a060020a0393841673ffffffffffffffffffffffffffffffffffffffff19918216179091556007805492909316911617905550505050565b600080546060918291829190819060ff1615156001146103a457600080fd5b60065460075460038054604080516020808402820181019092528281529294600494600594600160a060020a0392831694919092169287919083018282801561041657602002820191906000526020600020905b8154600160a060020a031681526001909101906020018083116103f8575b505050505094508380548060200260200160405190810160405280929190818152602001828054801561047257602002820191906000526020600020905b8154600160a060020a03168152600190910190602001808311610454575b50505050509350828054806020026020016040519081016040528092919081815260200182805480156104ce57602002820191906000526020600020905b8154600160a060020a031681526001909101906020018083116104b0575b50505050509250945094509450945094509091929394565b6000805460ff16156104f757600080fd5b600154151561050557600080fd5b6000805460ff1916600117905590565b60008054819060ff16151560011461052c57600080fd5b5060005b600a811015610637576003805460018181019092557fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b01805473ffffffffffffffffffffffffffffffffffffffff1990811673a22499738b961e56fb833d9368241a6a789e77c417909155600480548084019091557f8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65c6b64bfe7fe36bd19b018054821673c63ba6b8a9f2d33ee26a7fd0a59d155131f2701a1790556005805480840182556000919091527f036b6384b5eca791c62761152d0c79bb0604c104a5fb6f4eb0703f3154bb3db001805490911673193efb18f18f93e32e418e383714279bdade4a8117905501610530565b5090565b60016002819055805480820182557fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf601805473ffffffffffffffffffffffffffffffffffffffff199081167318254160af9c10f43db77566c294db7d8182caca179091556003805480840182557fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b9081018054841673060107a87178fcd01822d616e17eb5b26dd279ad179055815480850183558101805484167382ad1aea897ba3e72ee3632c08fe7596a9ff9c3f17905581548085018355810180548416730994ebd481f77768a8bf20d35d8f405ce21d2176179055815480850190925501805482167378d2ad8c09dce2eba9c25ef8201607ead8637b8e1790556004805480840182557f8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65c6b64bfe7fe36bd19b9081018054841673ac5e047d39692be8c81d0724543d5de721d0dd541790558154808501835581018054841673382ef85439dc874a0f55ab4d9801a5056e371b371790558154808501835581018054841673b821a659c21cb39745144931e71d0e9d09c8647f1790558154808501909255018054821673c1094c666657937ab7ed23040207a8ee6878135017905560058054808401825560008290527f036b6384b5eca791c62761152d0c79bb0604c104a5fb6f4eb0703f3154bb3db090810180548416739b0bf94f5ab62c73e454dfc55adb2d2fa6cd3af517905581548085018355810180548416734d83f1795ecdd684e94f1c5893ae6904ebeaeb9417905581548085018355810180548416739b58fe24f7a7cb9d102e21b3376bd80eefdc320b1790558154938401909155919091018054821673e60bf7b625e54e9f67767fad0e564f6aec29765217905560068054821673142441cb0896d4cd2ecdf3328d8d841a07f4b04f1790556007805490911673a36c921743d63361258fdbd107b906be0ad87940179055565b828054828255906000526020600020908101928215610970579160200282015b82811115610970578251825473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0390911617825560209092019160019091019061092e565b50610637926109a79250905b8082111561063757805473ffffffffffffffffffffffffffffffffffffffff1916815560010161097c565b905600a165627a7a72305820c1ba68771b503d0510928439c7f56625af3727ac2b3070b8cd42191f13b110ea0029`

// DeployAddressBook deploys a new klaytn contract, binding an instance of AddressBook to it.
func DeployAddressBook(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *AddressBook, error) {
	parsed, err := abi.JSON(strings.NewReader(AddressBookABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(AddressBookBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AddressBook{AddressBookCaller: AddressBookCaller{contract: contract}, AddressBookTransactor: AddressBookTransactor{contract: contract}, AddressBookFilterer: AddressBookFilterer{contract: contract}}, nil
}

// AddressBook is an auto generated Go binding around a klaytn contract.
type AddressBook struct {
	AddressBookCaller     // Read-only binding to the contract
	AddressBookTransactor // Write-only binding to the contract
	AddressBookFilterer   // Log filterer for contract events
}

// AddressBookCaller is an auto generated read-only Go binding around a klaytn contract.
type AddressBookCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddressBookTransactor is an auto generated write-only Go binding around a klaytn contract.
type AddressBookTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddressBookFilterer is an auto generated log filtering Go binding around a klaytn contract events.
type AddressBookFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AddressBookSession is an auto generated Go binding around a klaytn contract,
// with pre-set call and transact options.
type AddressBookSession struct {
	Contract     *AddressBook      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AddressBookCallerSession is an auto generated read-only Go binding around a klaytn contract,
// with pre-set call options.
type AddressBookCallerSession struct {
	Contract *AddressBookCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// AddressBookTransactorSession is an auto generated write-only Go binding around a klaytn contract,
// with pre-set transact options.
type AddressBookTransactorSession struct {
	Contract     *AddressBookTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// AddressBookRaw is an auto generated low-level Go binding around a klaytn contract.
type AddressBookRaw struct {
	Contract *AddressBook // Generic contract binding to access the raw methods on
}

// AddressBookCallerRaw is an auto generated low-level read-only Go binding around a klaytn contract.
type AddressBookCallerRaw struct {
	Contract *AddressBookCaller // Generic read-only contract binding to access the raw methods on
}

// AddressBookTransactorRaw is an auto generated low-level write-only Go binding around a klaytn contract.
type AddressBookTransactorRaw struct {
	Contract *AddressBookTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAddressBook creates a new instance of AddressBook, bound to a specific deployed contract.
func NewAddressBook(address common.Address, backend bind.ContractBackend) (*AddressBook, error) {
	contract, err := bindAddressBook(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AddressBook{AddressBookCaller: AddressBookCaller{contract: contract}, AddressBookTransactor: AddressBookTransactor{contract: contract}, AddressBookFilterer: AddressBookFilterer{contract: contract}}, nil
}

// NewAddressBookCaller creates a new read-only instance of AddressBook, bound to a specific deployed contract.
func NewAddressBookCaller(address common.Address, caller bind.ContractCaller) (*AddressBookCaller, error) {
	contract, err := bindAddressBook(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AddressBookCaller{contract: contract}, nil
}

// NewAddressBookTransactor creates a new write-only instance of AddressBook, bound to a specific deployed contract.
func NewAddressBookTransactor(address common.Address, transactor bind.ContractTransactor) (*AddressBookTransactor, error) {
	contract, err := bindAddressBook(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AddressBookTransactor{contract: contract}, nil
}

// NewAddressBookFilterer creates a new log filterer instance of AddressBook, bound to a specific deployed contract.
func NewAddressBookFilterer(address common.Address, filterer bind.ContractFilterer) (*AddressBookFilterer, error) {
	contract, err := bindAddressBook(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AddressBookFilterer{contract: contract}, nil
}

// bindAddressBook binds a generic wrapper to an already deployed contract.
func bindAddressBook(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AddressBookABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AddressBook *AddressBookRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AddressBook.Contract.AddressBookCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AddressBook *AddressBookRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AddressBook.Contract.AddressBookTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AddressBook *AddressBookRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AddressBook.Contract.AddressBookTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AddressBook *AddressBookCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AddressBook.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AddressBook *AddressBookTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AddressBook.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AddressBook *AddressBookTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AddressBook.Contract.contract.Transact(opts, method, params...)
}

// GetAllAddressInfo is a free data retrieval call binding the contract method 0x160370b8.
//
// Solidity: function getAllAddressInfo() constant returns(address[], address[], address[], address, address)
func (_AddressBook *AddressBookCaller) GetAllAddressInfo(opts *bind.CallOpts) ([]common.Address, []common.Address, []common.Address, common.Address, common.Address, error) {
	var (
		ret0 = new([]common.Address)
		ret1 = new([]common.Address)
		ret2 = new([]common.Address)
		ret3 = new(common.Address)
		ret4 = new(common.Address)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
		ret4,
	}
	err := _AddressBook.contract.Call(opts, out, "getAllAddressInfo")
	return *ret0, *ret1, *ret2, *ret3, *ret4, err
}

// GetAllAddressInfo is a free data retrieval call binding the contract method 0x160370b8.
//
// Solidity: function getAllAddressInfo() constant returns(address[], address[], address[], address, address)
func (_AddressBook *AddressBookSession) GetAllAddressInfo() ([]common.Address, []common.Address, []common.Address, common.Address, common.Address, error) {
	return _AddressBook.Contract.GetAllAddressInfo(&_AddressBook.CallOpts)
}

// GetAllAddressInfo is a free data retrieval call binding the contract method 0x160370b8.
//
// Solidity: function getAllAddressInfo() constant returns(address[], address[], address[], address, address)
func (_AddressBook *AddressBookCallerSession) GetAllAddressInfo() ([]common.Address, []common.Address, []common.Address, common.Address, common.Address, error) {
	return _AddressBook.Contract.GetAllAddressInfo(&_AddressBook.CallOpts)
}

// CompleteInitialization is a paid mutator transaction binding the contract method 0x5827b250.
//
// Solidity: function completeInitialization() returns(bool)
func (_AddressBook *AddressBookTransactor) CompleteInitialization(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AddressBook.contract.Transact(opts, "completeInitialization")
}

// CompleteInitialization is a paid mutator transaction binding the contract method 0x5827b250.
//
// Solidity: function completeInitialization() returns(bool)
func (_AddressBook *AddressBookSession) CompleteInitialization() (*types.Transaction, error) {
	return _AddressBook.Contract.CompleteInitialization(&_AddressBook.TransactOpts)
}

// CompleteInitialization is a paid mutator transaction binding the contract method 0x5827b250.
//
// Solidity: function completeInitialization() returns(bool)
func (_AddressBook *AddressBookTransactorSession) CompleteInitialization() (*types.Transaction, error) {
	return _AddressBook.Contract.CompleteInitialization(&_AddressBook.TransactOpts)
}

// Init is a paid mutator transaction binding the contract method 0x0f97899c.
//
// Solidity: function init(_adminList address[], _cnNodeIdList address[], _cnStakingContractList address[], _cnRewardAddressList address[], _pocContract address, _kirContract address) returns()
func (_AddressBook *AddressBookTransactor) Init(opts *bind.TransactOpts, _adminList []common.Address, _cnNodeIdList []common.Address, _cnStakingContractList []common.Address, _cnRewardAddressList []common.Address, _pocContract common.Address, _kirContract common.Address) (*types.Transaction, error) {
	return _AddressBook.contract.Transact(opts, "init", _adminList, _cnNodeIdList, _cnStakingContractList, _cnRewardAddressList, _pocContract, _kirContract)
}

// Init is a paid mutator transaction binding the contract method 0x0f97899c.
//
// Solidity: function init(_adminList address[], _cnNodeIdList address[], _cnStakingContractList address[], _cnRewardAddressList address[], _pocContract address, _kirContract address) returns()
func (_AddressBook *AddressBookSession) Init(_adminList []common.Address, _cnNodeIdList []common.Address, _cnStakingContractList []common.Address, _cnRewardAddressList []common.Address, _pocContract common.Address, _kirContract common.Address) (*types.Transaction, error) {
	return _AddressBook.Contract.Init(&_AddressBook.TransactOpts, _adminList, _cnNodeIdList, _cnStakingContractList, _cnRewardAddressList, _pocContract, _kirContract)
}

// Init is a paid mutator transaction binding the contract method 0x0f97899c.
//
// Solidity: function init(_adminList address[], _cnNodeIdList address[], _cnStakingContractList address[], _cnRewardAddressList address[], _pocContract address, _kirContract address) returns()
func (_AddressBook *AddressBookTransactorSession) Init(_adminList []common.Address, _cnNodeIdList []common.Address, _cnStakingContractList []common.Address, _cnRewardAddressList []common.Address, _pocContract common.Address, _kirContract common.Address) (*types.Transaction, error) {
	return _AddressBook.Contract.Init(&_AddressBook.TransactOpts, _adminList, _cnNodeIdList, _cnStakingContractList, _cnRewardAddressList, _pocContract, _kirContract)
}

// InitTest is a paid mutator transaction binding the contract method 0xf4c7bc34.
//
// Solidity: function initTest() returns()
func (_AddressBook *AddressBookTransactor) InitTest(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AddressBook.contract.Transact(opts, "initTest")
}

// InitTest is a paid mutator transaction binding the contract method 0xf4c7bc34.
//
// Solidity: function initTest() returns()
func (_AddressBook *AddressBookSession) InitTest() (*types.Transaction, error) {
	return _AddressBook.Contract.InitTest(&_AddressBook.TransactOpts)
}

// InitTest is a paid mutator transaction binding the contract method 0xf4c7bc34.
//
// Solidity: function initTest() returns()
func (_AddressBook *AddressBookTransactorSession) InitTest() (*types.Transaction, error) {
	return _AddressBook.Contract.InitTest(&_AddressBook.TransactOpts)
}

// TestAddCn is a paid mutator transaction binding the contract method 0xeaee0716.
//
// Solidity: function testAddCn() returns(bool)
func (_AddressBook *AddressBookTransactor) TestAddCn(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AddressBook.contract.Transact(opts, "testAddCn")
}

// TestAddCn is a paid mutator transaction binding the contract method 0xeaee0716.
//
// Solidity: function testAddCn() returns(bool)
func (_AddressBook *AddressBookSession) TestAddCn() (*types.Transaction, error) {
	return _AddressBook.Contract.TestAddCn(&_AddressBook.TransactOpts)
}

// TestAddCn is a paid mutator transaction binding the contract method 0xeaee0716.
//
// Solidity: function testAddCn() returns(bool)
func (_AddressBook *AddressBookTransactorSession) TestAddCn() (*types.Transaction, error) {
	return _AddressBook.Contract.TestAddCn(&_AddressBook.TransactOpts)
}
