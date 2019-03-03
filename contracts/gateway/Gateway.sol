pragma solidity ^0.4.24;

import "./KRC20Receiver.sol";
import "./KRC20.sol";
import "./SafeMath.sol";

// TODO-Klaytn-Servicechain should know why if IERC721Recevier used, ERC20 token can't call receiver.
contract Gateway is IKRC20Receiver {
    // TODO-Klaytn-Servicechain should add ownership for Gateway.

    bool public isChild;   // Child gateway can withdraw with no limit.

    using SafeMath for uint256;

    struct Balance {
        uint256 klay;
        mapping(address => uint256) erc20;
    }

    Balance balances;

    event KLAYReceived(address from, uint256 amount);
    event ERC20Received(address from, uint256 amount, address contractAddress);

    enum TokenKind {
        KLAY,
        ERC20
    }

    /**
     * Event to log the withdrawal of a token from the Gateway.
     * @param owner Address of the entity that made the withdrawal.
     * @param kind The type of token withdrawn (ERC20/KLAY).
     * @param contractAddress Address of token contract the token belong to.
     * @param value For KLAY/ERC20 this is the amount.
     */
    event TokenWithdrawn(address owner, TokenKind kind, address contractAddress, uint256 value);

    constructor (bool _isChild) public {
        isChild = _isChild;
    }

    // Deposit functions
    function depositKLAY() private {
        balances.klay = balances.klay.add(msg.value);
    }

    function depositERC20(uint256 amount) private {
        balances.erc20[msg.sender] = balances.erc20[msg.sender].add(amount);
    }

    ///////////////////////////////////////////////////////////////////////////////////////
    // Withdrawal functions
    function withdrawERC20(uint256 amount, address to, address contractAddress)
    external
    {
        // TODO-Klaytn-Servicechain should add require to check if msg.sender is same with the owner of Gateway.
        if (isChild == false){
            balances.erc20[contractAddress] = balances.erc20[contractAddress].sub(amount);
        }
        ERC20(contractAddress).transfer(to, amount);
        emit TokenWithdrawn(to, TokenKind.ERC20, contractAddress, amount);
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
    // Requires first to have called `approve` on the specified ERC20 contract
    function depositERC20(uint256 amount, address contractAddress) external {
        ERC20(contractAddress).transferFrom(msg.sender, address(this), amount);
        balances.erc20[contractAddress] = balances.erc20[contractAddress].add(amount);
        emit ERC20Received(msg.sender, amount, contractAddress);
    }

    //////////////////////////////////////////////////////////////////////////////
    // Receiver functions for 1-step deposits to the gateway

    function onERC20Received(address _from, uint256 amount)
    public
    returns (bytes4)
    {
        // TODO-Klaytn-Servicechain should add allowedToken list in this Gateway.
        //require(allowedTokens[msg.sender], "Not a valid token");
        depositERC20(amount);
        emit ERC20Received(_from, amount, msg.sender);
        return ERC20_RECEIVED;
    }

    function () external payable {
        depositKLAY();
        emit KLAYReceived(msg.sender, msg.value);
    }
    //////////////////////////////////////////////////////////////////////////////

    // Returns all the KLAY you own
    function getKLAY() external view returns (uint256) {
        return balances.klay;
    }

    // Returns all the KLAY you own
    function getERC20(address contractAddress) external view returns (uint256) {
        return balances.erc20[contractAddress];
    }
}