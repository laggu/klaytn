pragma solidity ^0.4.24;

import "../openzeppelin-solidity/contracts/math/SafeMath.sol";
import "../openzeppelin-solidity/contracts/token/ERC20/IERC20.sol";
import "../openzeppelin-solidity/contracts/token/ERC721/IERC721.sol";
import "../openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "../servicechain_nft/INFTReceiver.sol";
import "../servicechain_token/ITokenReceiver.sol";

contract Bridge is ITokenReceiver, INFTReceiver, Ownable {
    bool public onServiceChain;

    using SafeMath for uint256;

    struct Balance {
        uint256 klay;
        mapping(address => uint256) token;
        mapping(address => mapping(uint256 => bool)) nft;
    }

    Balance balances;

    uint64 public requestNonce;
    uint64 public handleNonce;

    enum TokenKind {
        KLAY,
        TOKEN,
        NFT
    }


    constructor (bool _onServiceChain) public payable {
        onServiceChain = _onServiceChain;
        updateKLAY();
    }

    /**
     * Event to log the withdrawal of a token from the Bridge.
     * @param kind The type of token withdrawn (KLAY/TOKEN/NFT).
     * @param from is the requester of the request value transfer event.
     * @param contractAddress Address of token contract the token belong to.
     * @param amount is the amount for KLAY/TOKEN and the NFT ID for NFT.
     * @param requestNonce is the order number of the request value transfer.
     */
    event RequestValueTransfer(TokenKind kind,
        address from,
        uint256 amount,
        address contractAddress,
        address to,
        uint64 requestNonce);

    /**
     * Event to log the withdrawal of a token from the Bridge.
     * @param owner Address of the entity that made the withdrawal.ga
     * @param kind The type of token withdrawn (KLAY/TOKEN/NFT).
     * @param contractAddress Address of token contract the token belong to.
     * @param value For KLAY/TOKEN this is the amount.
     * @param handleNonce is the order number of the handle value transfer.
     */
    event HandleValueTransfer(address owner,
        TokenKind kind,
        address contractAddress,
        uint256 value,
        uint64 handleNonce);

    // Internal Deposit functions update the balance in this contract.
    function updateKLAY() private {
        balances.klay = balances.klay.add(msg.value);
    }

    function updateToken(uint256 _amount) private {
        balances.token[msg.sender] = balances.token[msg.sender].add(_amount);
    }

    function updateNFT(uint256 _uid) private {
        balances.nft[msg.sender][_uid] = true;
    }

    // HandleValue(Token/KLAY/NFT)Transfer sends the value by the request.
    function HandleTokenTransfer(uint256 _amount, address _to, address _contractAddress, uint64 _handleNonce)
    onlyOwner
    external
    {
        require(handleNonce == _handleNonce, "mismatched handle nonce");

        if (onServiceChain == false){
            balances.token[_contractAddress] = balances.token[_contractAddress].sub(_amount);
        }
        IERC20(_contractAddress).transfer(_to, _amount);
        emit HandleValueTransfer(_to, TokenKind.TOKEN, _contractAddress, _amount, handleNonce);
        handleNonce++;
    }

    function HandleKLAYTransfer(uint256 _amount, address _to, uint64 _handleNonce)
    onlyOwner
    external
    {
        require(handleNonce == _handleNonce, "mismatched handle nonce");

        // TODO-Klaytn-Servicechain for KLAY, we can replace below variable with embedded variable.
        balances.klay = balances.klay.sub(_amount);
        _to.transfer(_amount); // ensure it's not reentrant
        emit HandleValueTransfer(_to, TokenKind.KLAY, address(0), _amount, handleNonce);
        handleNonce++;
    }

    function HandleNFTTransfer(uint256 _uid, address _contractAddress, address _to, uint64 _handleNonce)
    onlyOwner
    external
    {
        require(handleNonce == _handleNonce, "mismatched handle nonce");
        require(balances.nft[_contractAddress][_uid], "Does not own token");

        IERC721(_contractAddress).safeTransferFrom(address(this), _to, _uid);
        delete balances.nft[_contractAddress][_uid];
        emit HandleValueTransfer(_to, TokenKind.NFT, _contractAddress, _uid, handleNonce);
        handleNonce++;
    }

    // Approve and Deposit function for 2-step deposits
    // Requires first to have called `approve` on the specified TOKEN contract
    // TODO-Klaytn need to consider whether this method is necessary or not.
    function RequestTokenTransfer(uint256 _amount, address _contractAddress, address _to)
    onlyOwner
    external {
        IERC20(_contractAddress).transferFrom(msg.sender, address(this), _amount);
        balances.token[_contractAddress] = balances.token[_contractAddress].add(_amount);
        emit RequestValueTransfer(TokenKind.TOKEN, msg.sender, _amount, _contractAddress, _to, requestNonce);
        requestNonce++;
    }

    //////////////////////////////////////////////////////////////////////////////
    // Receiver functions of Token for 1-step deposits to the Bridge
    bytes4 constant TOKEN_RECEIVED = 0xbc04f0af;

    function onTokenReceived(address _from, uint256 _amount, address _to)
    public
    returns (bytes4)
    {
        // TODO-Klaytn-Servicechain should add allowedToken list in this Bridge.
        //require(allowedTokens[msg.sender], "Not a valid token");
        updateToken(_amount);
        emit RequestValueTransfer(TokenKind.TOKEN, _from, _amount, msg.sender, _to, requestNonce);
        requestNonce++;
        return TOKEN_RECEIVED;
    }

    // Receiver function of NFT for 1-step deposits to the Bridge
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
        updateNFT(tokenId);
        emit RequestValueTransfer(TokenKind.NFT, from, tokenId, msg.sender, to, requestNonce);
        requestNonce++;
        return ERC721_RECEIVED;
    }

    // () requests transfer KLAY to msg.sender address on relative chain.
    function () external payable {
        updateKLAY();
        emit RequestValueTransfer(TokenKind.KLAY, msg.sender, msg.value, address(0), msg.sender, requestNonce);
        requestNonce++;
    }

    // DepositKLAY requests transfer KLAY to _to on relative chain.
    function RequestKLAYTransfer(address _to) external payable {
        updateKLAY();
        emit RequestValueTransfer(TokenKind.KLAY, msg.sender, msg.value, address(0), _to, requestNonce);
        requestNonce++;
    }

    // DepositWithoutEvent send KLAY to this contract without event for increasing the withdrawal limit.
    function ChargeWithoutEvent() external payable {
        updateKLAY();
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
