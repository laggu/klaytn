pragma solidity ^0.4.24;

import "openzeppelin-solidity/contracts/math/SafeMath.sol";
import "openzeppelin-solidity/contracts/token/ERC20/IERC20.sol";
import "openzeppelin-solidity/contracts/token/ERC721/IERC721.sol";
import "openzeppelin-solidity/contracts/ownership/Ownable.sol";

contract Gateway is Ownable {
    bool public isChild;   // Child gateway can withdraw with no limit.

    using SafeMath for uint256;

    struct Balance {
        uint256 klay;
        mapping(address => uint256) token;  // erc20 -> token
        mapping(address => mapping(uint256 => bool)) nft;    // erc721 -> nft
    }

    Balance balances;

    event TokenReceived(TokenKind kind, address from, uint256 amount, address contractAddress, address to);

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

    // Internal Deposit functions
    function depositKLAY() private {
        balances.klay = balances.klay.add(msg.value);
    }

    function depositToken(uint256 amount) private {
        balances.token[msg.sender] = balances.token[msg.sender].add(amount);
    }

    function depositNFT(uint256 uid) private {
        balances.nft[msg.sender][uid] = true;
    }

    // Withdrawal functions
    function withdrawToken(uint256 amount, address to, address contractAddress)
    onlyOwner
    external
    {
        if (isChild == false){
            balances.token[contractAddress] = balances.token[contractAddress].sub(amount);
        }
        IERC20(contractAddress).transfer(to, amount);
        emit TokenWithdrawn(to, TokenKind.TOKEN, contractAddress, amount);
    }

    function withdrawKLAY(uint256 amount, address to)
    onlyOwner
    external
    {
        // TODO-Klaytn-Servicechain for KLAY, we can replace below variable with embedded variable.
        balances.klay = balances.klay.sub(amount);
        to.transfer(amount); // ensure it's not reentrant
        emit TokenWithdrawn(to, TokenKind.KLAY, address(0), amount);
    }

    function withdrawERC721(uint256 uid, address contractAddress, address to)
    onlyOwner
    external
    {
        require(balances.nft[contractAddress][uid], "Does not own token");
        IERC721(contractAddress).safeTransferFrom(address(this), to, uid);
        delete balances.nft[contractAddress][uid];
        emit TokenWithdrawn(to, TokenKind.NFT, contractAddress, uid);
    }

    // Approve and Deposit function for 2-step deposits
    // Requires first to have called `approve` on the specified TOKEN contract
    // TODO-Klaytn need to consider whether this method is necessary or not.
    function depositToken(uint256 amount, address contractAddress, address to)
    onlyOwner
    external {
        IERC20(contractAddress).transferFrom(msg.sender, address(this), amount);
        balances.token[contractAddress] = balances.token[contractAddress].add(amount);
        emit TokenReceived(TokenKind.TOKEN, msg.sender, amount, contractAddress, to);
    }

    //////////////////////////////////////////////////////////////////////////////
    // Receiver functions of Token for 1-step deposits to the gateway
    bytes4 constant TOKEN_RECEIVED = 0xbc04f0af;

    function onTokenReceived(address _from, uint256 amount, address _to)
    public
    returns (bytes4)
    {
        // TODO-Klaytn-Servicechain should add allowedToken list in this Gateway.
        //require(allowedTokens[msg.sender], "Not a valid token");
        depositToken(amount);
        emit TokenReceived(TokenKind.TOKEN, _from, amount, msg.sender, _to);
        return TOKEN_RECEIVED;
    }

    // Receiver function of NFT for 1-step deposits to the gateway
    bytes4 private constant ERC721_RECEIVED = 0x150b7a02;

    function onNFTReceived(
        address from,
        uint256 tokenId,
        address to
    )
    public
    returns(bytes4)
    {
        //require(allowedTokens[msg.sender], "Not a valid token");
        depositNFT(tokenId);
        emit TokenReceived(TokenKind.NFT, from, tokenId, msg.sender, to);
        return ERC721_RECEIVED;
    }

    // () requests transfer KLAY to msg.sender address on relative chain.
    function () external payable {
        depositKLAY();
        emit TokenReceived(TokenKind.KLAY, msg.sender, msg.value, address(0), msg.sender);
    }

    // DepositKLAY requests transfer KLAY to _to on relative chain.
    function DepositKLAY(address _to) external payable {
        depositKLAY();
        emit TokenReceived(TokenKind.KLAY, msg.sender, msg.value, address(0), _to);
    }

    // DepositWithoutEvent send KLAY to this contract without event for increasing the withdrawal limit.
    function DepositWithoutEvent() external payable {
        depositKLAY();
    }
    //////////////////////////////////////////////////////////////////////////////

    // Returns KLAY withdrawal limit
    function getKLAY() external view returns (uint256) {
        return balances.klay;
    }

    // Returns given Token withdrawal limit
    function getToken(address contractAddress) external view returns (uint256) {
        return balances.token[contractAddress];
    }

    // Returns ERC721 token by uid
    function getNFT(address owner, uint256 uid, address contractAddress) external view returns (bool) {
        return balances.nft[contractAddress][uid];
    }
}
