pragma solidity ^0.4.24;

import "../openzeppelin-solidity/contracts/token/ERC721/ERC721Full.sol";
import "../openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "./INFTReceiver.sol";


contract ServiceChainNFT is ERC721Full("ServiceChainNFT", "SCN"), Ownable {
    mapping(address => bool) private registered;
    address public bridge;

    constructor (address _bridge) public { bridge = _bridge; }

    // Owner mints the NFT to the user.
    function register(address _user, uint256 _tokenId) onlyOwner external {
        _mint(_user, _tokenId);
    }

    // TODO-Klaytn needs to consider how to prevent mint duplicated id in another NFT contract.
    // Bridge mints the NFT to itself to support value transfer.
    function mintToBridge(uint256 _uid) public {
        require(msg.sender == bridge);
        _mint(bridge, _uid);
    }

    // user request value transfer to main / service chain.
    function requestValueTransfer(uint256 _uid, address _to) external {
        safeTransferAndCall(_uid, _to);
    }

    function safeTransferAndCall(uint256 _uid, address _to) public {
        transferFrom(msg.sender, bridge, _uid);
        require(
            checkAndCallSafeTransfer(msg.sender, _uid, _to),
            "Sent to a contract which is not an TOKEN receiver"
        );
    }

    // TODO-Klaytn-Servicechain define proper bytes4 value.
    bytes4 private constant _ERC721_RECEIVED = 0x150b7a02;

    function checkAndCallSafeTransfer(address _from, uint256 _uid, address _to) internal returns (bool) {
        if (!bridge.isContract()) {
            return true;
        }
        bytes4 retval = INFTReceiver(bridge).onNFTReceived(_from, _uid, _to);
        return (retval == _ERC721_RECEIVED);
    }
}
