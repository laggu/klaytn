pragma solidity ^0.4.24;

import "openzeppelin-solidity/contracts/token/ERC20/ERC20.sol";
import "openzeppelin-solidity/contracts/utils/Address.sol";
import "./ITokenReceiver.sol";


contract ServiceChainToken is ERC20 {
    string public constant name = "ServiceChainToken";
    string public constant symbol = "SCT";
    uint8 public constant decimals = 18;

    address gateway;

    // TODO-Klaytn-Servicechain define proper bytes4 value.
    bytes4 constant TOKEN_RECEIVED = 0xbc04f0af;

    using Address for address;

    // one billion in initial supply
    uint256 public constant INITIAL_SUPPLY = 1000000000 * (10 ** uint256(decimals));

    constructor (address _gateway) public {
        _mint(msg.sender, INITIAL_SUPPLY);
        gateway = _gateway;
    }

    // Additional functions for gateway interaction, influenced from Zeppelin ERC721 Impl.

    function depositToGateway(uint256 _amount, address _to) external {
        safeTransferAndCall(gateway, _amount, _to);
    }

    function safeTransferAndCall(address _gateway, uint256 amount, address _to) public {
        transfer(_gateway, amount);
        require(
            checkAndCallSafeTransfer(msg.sender, _gateway, amount, _to),
            "Sent to a contract which is not an TOKEN receiver"
        );
    }

    function checkAndCallSafeTransfer(address _from, address _gateway, uint256 amount, address _to) internal returns (bool) {
        if (!_gateway.isContract()) {
            return true;
        }

        bytes4 retval = ITokenReceiver(_gateway).onTokenReceived(_from, amount, _to);
        return(retval == TOKEN_RECEIVED);
    }
}
