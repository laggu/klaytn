pragma solidity ^0.4.24;

import "./ITokenReceiver.sol";
import "./IToken.sol";
import "./SafeMath.sol";

// TODO-Klaytn-Servicechain should know why if IERC721Recevier used, TOKEN token can't call receiver.
contract Gateway is ITokenReceiver {
    // TODO-Klaytn-Servicechain should add ownership for Gateway.

    bool public isChild;   // Child gateway can withdraw with no limit.

    using SafeMath for uint256;

    struct Balance {
        uint256 klay;
        mapping(address => uint256) token;  // erc20 -> token
        //mapping(address => mapping(uint256 => bool)) nft;    // erc721 -> nft
    }

    Balance balances;

//    event KLAYReceived(address from, uint256 amount); // TODO-Klaytn It will be removed.
//    event ERC20Received(address from, uint256 amount, address contractAddress); // TODO-Klaytn It will be removed.
    event TokenReceived(TokenKind kind, address from, uint256 amount, address contractAddress);

    enum TokenKind {
        KLAY,
        TOKEN,
        NFT
    }

    /**
     * Event to log the withdrawal of a token from the Gateway.
     * @param owner Address of the entity that made the withdrawal.ga
     * @param kind The type of token withdrawn (TOKEN/KLAY).
     * @param contractAddress Address of token contract the token belong to.
     * @param value For KLAY/TOKEN this is the amount.
     */
    event TokenWithdrawn(address owner, TokenKind kind, address contractAddress, uint256 value);

    constructor (bool _isChild) public payable {
        isChild = _isChild;
    }

    // Deposit functions
    function depositKLAY() private {
        balances.klay = balances.klay.add(msg.value);
    }

    function depositToken(uint256 amount) private {
        balances.token[msg.sender] = balances.token[msg.sender].add(amount);
    }

    ///////////////////////////////////////////////////////////////////////////////////////
    // Withdrawal functions
    function withdrawToken(uint256 amount, address to, address contractAddress)
    external
    {
        // TODO-Klaytn-Servicechain should add require to check if msg.sender is same with the owner of Gateway.
        if (isChild == false){
            balances.token[contractAddress] = balances.token[contractAddress].sub(amount);
        }
        IToken(contractAddress).transfer(to, amount);
        emit TokenWithdrawn(to, TokenKind.TOKEN, contractAddress, amount);
    }

    function withdrawKLAY(uint256 amount, address to)
    external
    {
        // TODO-Klaytn-Servicechain should add require to check if msg.sender is same with the owner of Gateway.
        // TODO-Klaytn-Servicechain for KLAY, we can replace below variable with embedded variable.
        balances.klay = balances.klay.sub(amount);
        to.transfer(amount); // ensure it's not reentrant
        emit TokenWithdrawn(to, TokenKind.KLAY, address(0), amount);
    }

    // Approve and Deposit function for 2-step deposits
    // Requires first to have called `approve` on the specified TOKEN contract
    function depositToken(uint256 amount, address contractAddress) external {
        IToken(contractAddress).transferFrom(msg.sender, address(this), amount);
        balances.token[contractAddress] = balances.token[contractAddress].add(amount);
//        emit ERC20Received(msg.sender, amount, contractAddress);    // TODO-Klaytn It will be removed.
        emit TokenReceived(TokenKind.TOKEN, msg.sender, amount, contractAddress);
    }

    //////////////////////////////////////////////////////////////////////////////
    // Receiver functions for 1-step deposits to the gateway

    function onTokenReceived(address _from, uint256 amount)
    public
    returns (bytes4)
    {
        // TODO-Klaytn-Servicechain should add allowedToken list in this Gateway.
        //require(allowedTokens[msg.sender], "Not a valid token");
        depositToken(amount);
//        emit ERC20Received(_from, amount, msg.sender);              // TODO-Klaytn It will be removed.
        emit TokenReceived(TokenKind.TOKEN, _from, amount, msg.sender);
        return TOKEN_RECEIVED;
    }

    function () external payable {
        depositKLAY();
//        emit KLAYReceived(msg.sender, msg.value);
        emit TokenReceived(TokenKind.KLAY, msg.sender, msg.value, address(0));
    }

    function DepositWithoutEvent() external payable {
        depositKLAY();
    }
    //////////////////////////////////////////////////////////////////////////////

    // Returns all the KLAY you own
    function getKLAY() external view returns (uint256) {
        return balances.klay;
    }

    // Returns all the KLAY you own
    function getToken(address contractAddress) external view returns (uint256) {
        return balances.token[contractAddress];
    }
}
