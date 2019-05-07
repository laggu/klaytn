
// File: contracts/SafeMath.sol from commit 0ddfd0f

pragma solidity ^0.4.24;

/**
 * @title SafeMath
 * @dev Unsigned math operations with safety checks that revert on error
 */
library SafeMath {
    /**
     * @dev Multiplies two unsigned integers, reverts on overflow.
     */
    function mul(uint256 a, uint256 b) internal pure returns (uint256) {
        // Gas optimization: this is cheaper than requiring 'a' not being zero, but the
        // benefit is lost if 'b' is also tested.
        // See: https://github.com/OpenZeppelin/openzeppelin-solidity/pull/522
        if (a == 0) {
            return 0;
        }

        uint256 c = a * b;
        require(c / a == b);

        return c;
    }

    /**
     * @dev Integer division of two unsigned integers truncating the quotient, reverts on division by zero.
     */
    function div(uint256 a, uint256 b) internal pure returns (uint256) {
        // Solidity only automatically asserts when dividing by 0
        require(b > 0);
        uint256 c = a / b;
        // assert(a == b * c + a % b); // There is no case in which this doesn't hold

        return c;
    }

    /**
     * @dev Subtracts two unsigned integers, reverts on overflow (i.e. if subtrahend is greater than minuend).
     */
    function sub(uint256 a, uint256 b) internal pure returns (uint256) {
        require(b <= a);
        uint256 c = a - b;

        return c;
    }

    /**
     * @dev Adds two unsigned integers, reverts on overflow.
     */
    function add(uint256 a, uint256 b) internal pure returns (uint256) {
        uint256 c = a + b;
        require(c >= a);

        return c;
    }

    /**
     * @dev Divides two unsigned integers and returns the remainder (unsigned integer modulo),
     * reverts when dividing by zero.
     */
    function mod(uint256 a, uint256 b) internal pure returns (uint256) {
        require(b != 0);
        return a % b;
    }
}

// File: contracts/MultisigBase.sol from commit cbe6e15

pragma solidity ^0.4.24;

/**
 * @title MultisigBase
 */
contract MultisigBase {
    using SafeMath for uint256;
    /*
     *  Events
     */
    event DeployMultisigContract(address[] adminList, uint requirement);
    event AddAdmin (address indexed admin);
    event DeleteAdmin(address indexed admin);
    event UpdateRequirement(uint requirement);
    event ClearRequest(uint timestamp);
    event SubmitRequest(uint indexed index, address indexed from, address to, uint value, bytes data, uint timestamp);
    event ConfirmRequest(uint indexed index, address indexed from, uint timestamp);
    event RevokeConfirmation(uint indexed index, address indexed from, uint timestamp);
    event CancelRequest(uint indexed index, address indexed from, uint timestamp);
    event ExecuteRequest(uint indexed index, address indexed from, uint timestamp);
    event ExecuteRequestFailure(uint indexed index, address indexed from, uint timestamp);


    /*
     *  Constants
     */
    uint constant public MAX_ADMIN = 50;


    /*
     *  Storage
     */
    address[] adminList;
    uint public requirement;
    uint public requestCount;
    uint public lastClearedIndex;
    mapping (address => bool) isAdmin;

    mapping(uint => Request) requestMap;
    struct Request {
        address to;
        uint value;
        bytes data;
        address requestProposer;

        uint confirmationCount;
        mapping(address => bool) confirmations;

        bool pending;
        bool executed;
        bool canceled;
    }


    /*
     *  Modifiers
     */
    modifier onlyMultisigTx() {
        require(msg.sender == address(this));
        _;
    }

    modifier onlyAdmin(address _admin) {
        require(isAdmin[_admin]);
        _;
    }

    modifier adminDoesNotExist(address _admin) {
        require(!isAdmin[_admin]);
        _;
    }

    modifier notNull(address _address) {
        require(_address != 0);
        _;
    }

    modifier requestExists(uint _index) {
        require(requestMap[_index].to != 0);
        _;
    }

    modifier confirmed(uint _index, address _admin) {
        require(requestMap[_index].confirmations[_admin]);
        _;
    }

    modifier isPending(uint _index) {
        require(requestMap[_index].pending);
        _;
    }

    modifier notConfirmed(uint _index, address _admin) {
        require(!requestMap[_index].confirmations[_admin]);
        _;
    }

    modifier notPending(uint _index) {
        require(!requestMap[_index].pending);
        _;
    }

    modifier notExecuted(uint _index) {
        require(!requestMap[_index].executed);
        _;
    }

    modifier notCanceled(uint _index) {
        require(!requestMap[_index].canceled);
        _;
    }

    modifier validRequirement(uint _adminCount, uint _requirement) {
        require(_adminCount <= MAX_ADMIN
            && _requirement <= _adminCount
            && _requirement != 0
            && _adminCount != 0);
        _;
    }


    /*
     *  Constructor
     */
    constructor(address[] _adminList, uint _requirement) public validRequirement(_adminList.length, _requirement) {
        for (uint i = 0; i < _adminList.length; i++) {
            require(!isAdmin[_adminList[i]] && _adminList[i] != 0);
            isAdmin[_adminList[i]] = true;
        }
        adminList = _adminList;
        requirement = _requirement;
        emit DeployMultisigContract(adminList, requirement);
    }


    /*
     *  Multisig functions
     */
    function addAdmin(address _admin) public 
    onlyMultisigTx
    adminDoesNotExist(_admin)
    notNull(_admin)
    validRequirement(adminList.length.add(1), requirement) {
        isAdmin[_admin] = true;
        adminList.push(_admin);
        clearRequest();
        emit AddAdmin(_admin);
    }

    function deleteAdmin(address _admin) public
    onlyMultisigTx
    onlyAdmin(_admin)
    validRequirement(adminList.length.sub(1), requirement) {
        isAdmin[_admin] = false;
        
        for (uint i=0; i < adminList.length - 1; i++)
            if (adminList[i] == _admin) {
                adminList[i] = adminList[adminList.length - 1];
                break;
            }
        adminList.length -= 1;
        clearRequest();
        emit DeleteAdmin(_admin);
    }

    function updateRequirement(uint _requirement) public 
    onlyMultisigTx
    validRequirement(adminList.length, _requirement) {
        requirement = _requirement;
        clearRequest();
        emit UpdateRequirement(_requirement);
    }

    function clearRequest() public onlyMultisigTx {
        for (uint i = lastClearedIndex; i < requestCount; i++){
            if (!requestMap[i].executed && !requestMap[i].canceled) {
                requestMap[i].canceled = true;
            }
        }
        lastClearedIndex = requestCount;
        emit ClearRequest(now);
    }

    /*
     *  Public functions
     */
    function submitRequest(address _to, uint _value, bytes _data) public 
    notNull(_to)
    onlyAdmin(msg.sender) {
        uint index = requestCount;
        requestMap[index] = Request({
            to : _to,
            value : _value,
            data : _data,
            requestProposer : msg.sender,
            confirmationCount : 0,
            pending : true,
            executed : false,
            canceled : false
        });
        emit SubmitRequest(index, msg.sender, _to, _value, _data, now);
        requestCount = requestCount.add(1);
        confirmRequest(index);
    }

    function confirmRequest(uint _index) public
    onlyAdmin(msg.sender)
    isPending(_index)
    notConfirmed(_index, msg.sender)
    notExecuted(_index)
    notCanceled(_index) {
        requestMap[_index].confirmations[msg.sender] = true;
        requestMap[_index].confirmationCount = requestMap[_index].confirmationCount.add(1);
        emit ConfirmRequest(_index, msg.sender, now);

        if (requestMap[_index].confirmationCount >= requirement) {
            requestMap[_index].pending = false;
        }

        if (requestMap[_index].pending == false) {
            executeRequest(_index);
        }
    }

    function revokeConfirmation(uint _index) public 
    onlyAdmin(msg.sender)
    confirmed(_index, msg.sender)
    notExecuted(_index)
    notCanceled(_index) {
        requestMap[_index].confirmations[msg.sender] = false;
        requestMap[_index].confirmationCount = requestMap[_index].confirmationCount.sub(1);
        emit RevokeConfirmation(_index, msg.sender, now);

        if (requestMap[_index].confirmationCount < requirement) {
            requestMap[_index].pending = true;
        }

        if (requestMap[_index].requestProposer == msg.sender) {
            requestMap[_index].canceled = true;
            emit CancelRequest(_index, msg.sender, now);
        }
    }

    function executeRequest(uint _index) public
    onlyAdmin(msg.sender)
    notPending(_index)
    notExecuted(_index)
    notCanceled(_index) {
        Request storage request = requestMap[_index];
        request.executed = true;
        if (external_call(request.to, request.value, request.data.length, request.data))
            emit ExecuteRequest(_index, msg.sender, now);
        else {
            emit ExecuteRequestFailure(_index, msg.sender, now);
            request.executed = false;
        }
    }

    function external_call(address destination, uint value, uint dataLength, bytes data) private returns (bool) {
        bool result;
        assembly {
            let x := mload(0x40)    // "Allocate" memory for output (0x40 is where "free memory" pointer is stored by convention)
            let d := add(data, 32)  // First 32 bytes are the padded length of data, so exclude that
            result := call(
                sub(gas, 34710),    // 34710 is the value that solidity is currently emitting
                                    // It includes callGas (700) + callVeryLow (3, to pay for SUB) + callValueTransferGas (9000) +
                                    // callNewAccountGas (25000, in case the destination address does not exist and needs creating)
                destination,
                value,
                d,
                dataLength,         // Size of the input (in bytes) - this is what fixes the padding problem
                x,
                0                   // Output is ignored, therefore the output size is zero
            )
        }
        return result;
    }


    /*
     * Getter functions
     */
    /// @dev return current adminList
    function getAdminInfo() public view returns(address[]) {
        return (adminList);
    }

    /// @return Returns corresponding request index list.
    function getRequestIndexes(uint _from, uint _to, bool _pending, bool _executed, bool _canceled) public view returns(uint[]) {
        if (_to == 0 || _to >= requestCount) {
            _to = requestCount;
        }
        uint cnt = 0;
        uint i;
        for (i = _from; i < _to; i++) {
            if ( requestMap[i].to != 0
            && _pending == requestMap[i].pending
            && _executed == requestMap[i].executed
            && _canceled == requestMap[i].canceled
            ) {
                cnt += 1;
            }  
        }
        uint[] memory requestIndexes = new uint[](cnt);
        cnt = 0;
        for (i = _from; i < _to; i++) {
            if ( requestMap[i].to != 0
            && _pending == requestMap[i].pending
            && _executed == requestMap[i].executed
            && _canceled == requestMap[i].canceled
            ) {
                requestIndexes[cnt] = i;
                cnt += 1;
            }
        }
        return requestIndexes;
    }

    function getRequestInfo(uint _index) public view returns(
        address To,
        uint Value,
        bytes Data,
        uint ConfirmationCount,
        bool Pending,
        bool Executed,
        bool Canceled) {
        return(
            requestMap[_index].to,
            requestMap[_index].value,
            requestMap[_index].data,
            requestMap[_index].confirmationCount,
            requestMap[_index].pending,
            requestMap[_index].executed,
            requestMap[_index].canceled
        );
    }
}

// File: contracts/InitContract.sol from commit cbe6e15

pragma solidity ^0.4.24;


/**
 * @title InitContract
 */
contract InitContract is MultisigBase {
    /*
     *  Events
     */
    event Initialize(address[] adminList, uint requirement);
    event RegisterBranchContract(uint index, string branchName, address contractAddress);
    event UnregisterBranchContract(uint index, string branchName, address contractAddress);
    event UpdateBranchContract(uint index, string branchName, address prevAddress, address curAddress);
    event RegisterLeafContract(address branchAddress, uint addressType, address[] leafAddress);
    event UnregisterLeafContract(address branchAddress, uint addressType, address leafAddress);
    event UpdateLeafContract(address branchAddress, uint addressType, address prevLeafAddress, address curLeafAddress, address extraLeafAddress);
    event CompleteInitialization(string branchName, address contractAddress);


    /*
     *  Storage
     */
    mapping(address => bool) isActiveBranch;
    mapping(string => bool) isBranchName;
    mapping(uint => BranchContract) branchContractMap;
    struct BranchContract {
        string branchName;
        address contractAddress;
    }
    uint public branchContractCount;
    bool public isInitialized;


    /*
     *  Modifiers
     */
    modifier afterInit() {
        require(isInitialized == true);
        _;
    }

    modifier activeBranch(address contractAddress) {
        require(isActiveBranch[contractAddress]);
        _;
    }

    modifier notActiveBranch(address contractAddress) {
        require(!isActiveBranch[contractAddress]);
        _;
    }

    modifier branchNameExists(string branchName) {
        require(isBranchName[branchName]);
        _;
    }

    modifier branchNameDoesNotExist(string branchName) {
        require(!isBranchName[branchName]);
        _;
    }

    modifier branchNameCheck(string branchName, address branchAddress) {
        require(keccak256(abi.encodePacked(branchName)) == keccak256(abi.encodePacked(branchContract(branchAddress).THIS_CONTRACT_NAME())));
        _;
    }


    /*
     * Constructor
     */
    constructor(address[] dummyArray, uint dummyUint) public MultisigBase(dummyArray, dummyUint) {}

    function initialize(address[] _adminList, uint _requirement, string _branchName, address _contractAddress) public 
    validRequirement(_adminList.length, _requirement)
    branchNameCheck(_branchName, _contractAddress)
    {
        require(msg.sender == 0x83fdd31030e6cc6a527fb8923801ef843d8488be);
        require(isInitialized == false); 
        require(bytes(_branchName).length > 0);

        // adminList validation
        for (uint i = 0; i < _adminList.length; i++) {
            require(!isAdmin[_adminList[i]] && _adminList[i] != 0);
            isAdmin[_adminList[i]] = true;
        }
        adminList = _adminList;
        requirement = _requirement;

        uint index = branchContractCount;
        branchContractMap[index] = BranchContract({
            branchName : _branchName,
            contractAddress : _contractAddress
        });
        isBranchName[_branchName] = true;
        isActiveBranch[_contractAddress] = true;
        branchContractCount += 1;

        isInitialized = true;
        emit RegisterBranchContract(index, _branchName, _contractAddress);
        emit Initialize(adminList, requirement);
    }


    function registerBranchContract(string _branchName, address _contractAddress) public
    notNull(_contractAddress)
    notActiveBranch(_contractAddress)
    branchNameDoesNotExist(_branchName)
    branchNameCheck(_branchName, _contractAddress)
    onlyMultisigTx
    afterInit {
        require(bytes(_branchName).length > 0);
        uint index = branchContractCount;
        branchContractMap[index] = BranchContract({
            branchName : _branchName,
            contractAddress : _contractAddress
        });
        isBranchName[_branchName] = true;
        isActiveBranch[_contractAddress] = true;
        require(branchContractCount < 2**256 -1); //to prevent overflow
        branchContractCount += 1;
    emit RegisterBranchContract(index, _branchName, _contractAddress);
    }

    function unregisterBranchContract(uint _index, string _branchName, address _contractAddress) public
    notNull(_contractAddress) 
    activeBranch(_contractAddress)
    onlyMultisigTx 
    afterInit {
        require(branchContractMap[_index].contractAddress == _contractAddress);
        require(keccak256(abi.encodePacked(branchContractMap[_index].branchName)) == keccak256(abi.encodePacked(_branchName)));
        isActiveBranch[_contractAddress] = false;
        emit UnregisterBranchContract(_index, branchContractMap[_index].branchName, _contractAddress);
    }

    function updateBranchContract(uint _index, string _branchName, address _prevAddress, address _newAddress) public
    notNull(_newAddress)
    activeBranch(_prevAddress)
    onlyMultisigTx
    afterInit {
        require(branchContractMap[_index].contractAddress == _prevAddress);
        require(keccak256(abi.encodePacked(branchContractMap[_index].branchName)) == keccak256(abi.encodePacked(_branchName)));
        branchContractMap[_index].contractAddress = _newAddress;
        emit UpdateBranchContract(_index, _branchName, _prevAddress, _newAddress);
    }

    
    /*
     * External functions
     */
    function registerLeafContract(address _branchAddress, uint _addressType, address[] _leafAddress) external activeBranch(msg.sender) {
        emit RegisterLeafContract(_branchAddress, _addressType, _leafAddress);
    }

    function unregisterLeafContract(address _branchAddress, uint _addressType, address _leafAddress) external activeBranch(msg.sender) {
        emit UnregisterLeafContract(_branchAddress, _addressType, _leafAddress);
    }

    function updateLeafContract(address _branchAddress, uint _addressType, address _prevLeafAddress, address _curLeafAddress, address _extraLeafAddress) external activeBranch(msg.sender) {
        emit UpdateLeafContract(_branchAddress, _addressType, _prevLeafAddress, _curLeafAddress, _extraLeafAddress);
    }

    function completeInitialization(string _branchName) external branchNameExists(_branchName) activeBranch(msg.sender) {
        emit CompleteInitialization(_branchName, msg.sender);
    }


    /*
     * Getter functions
     */
    function getAllBranchAddress() public view returns(address[], bool[]){
        address [] memory branchAddressList = new address[](branchContractCount);
        bool [] memory isActiveList =  new bool[](branchContractCount);

        for(uint i = 0; i < branchContractCount; i++){
            branchAddressList[i] = branchContractMap[i].contractAddress;
            isActiveList[i] = isActiveBranch[branchContractMap[i].contractAddress];
        }
        return (branchAddressList, isActiveList);
    }

    function getBranchInfo(uint _index) public view returns(string, bool, address){
        return(branchContractMap[_index].branchName, isActiveBranch[branchContractMap[_index].contractAddress], branchContractMap[_index].contractAddress);
    }

    function getAllAddress() public view returns(uint8[], address[]) {
        uint totalArrayLength;
        uint[] memory leafArrayLength = new uint[](branchContractCount);
        uint i;
        uint j;
        for(i = 0; i < branchContractCount; i++){
            if(isActiveBranch[branchContractMap[i].contractAddress]) {
                leafArrayLength[i] = branchContract(branchContractMap[i].contractAddress).getAllAddressCount();
                totalArrayLength += leafArrayLength[i];
            }
        }
        uint8[] memory allTypeList = new uint8[](totalArrayLength);
        address[] memory allAddressList = new address[](totalArrayLength);
        for(i = 0; i < branchContractCount; i++){
            if(leafArrayLength[i] != 0) {
                uint8[] memory leafTypeList = new uint8[](leafArrayLength[i]);
                address[] memory leafAddressList = new address[](leafArrayLength[i]);
                (leafTypeList, leafAddressList) = branchContract(branchContractMap[i].contractAddress).getAllAddress();
                for(j = 0; j < leafArrayLength[i]; j++){
                    allTypeList[i+j] = leafTypeList[j];
                    allAddressList[i+j] = leafAddressList[j];                    
                }
            }
        }
        return (allTypeList, allAddressList);
    }
}

interface branchContract {
    function getAllAddress() external view returns(uint8[], address[]);
    function getAllAddressCount() external view returns(uint);
    function getAllAddressInNiceForm() external view returns(address[], address[], address[], address, address);
    function THIS_CONTRACT_NAME() external view returns(string);
}
