pragma solidity ^0.4.24;

import "../externals/openzeppelin-solidity/contracts/token/ERC20/ERC20.sol";
import "../externals/openzeppelin-solidity/contracts/utils/Address.sol";
import "./ITokenReceiver.sol";


contract ServiceChainToken is ERC20 {
    string public constant name = "ServiceChainToken";
    string public constant symbol = "SCT";
    uint8 public constant decimals = 18;

    address bridge;

    // TODO-Klaytn-Servicechain define proper bytes4 value.
    bytes4 constant TOKEN_RECEIVED = 0xbc04f0af;

    using Address for address;

    // one billion in initial supply
    uint256 public constant INITIAL_SUPPLY = 1000000000 * (10 ** uint256(decimals));

    constructor (address _bridge) public {
        _mint(msg.sender, INITIAL_SUPPLY);
        bridge = _bridge;
    }

    // Additional functions for gateway interaction, influenced from Zeppelin ERC721 Impl.
    function requestValueTransfer(uint256 _amount, address _to) external {
        safeTransferAndCall(bridge, _amount, _to);
    }

    function safeTransferAndCall(address _bridge, uint256 _amount, address _to) public {
        transfer(_bridge, _amount);
        require(
            checkAndCallSafeTransfer(msg.sender, _bridge, _amount, _to),
            "Sent to a contract which is not an TOKEN receiver"
        );
    }

    function checkAndCallSafeTransfer(address _from,
        address _bridge,
        uint256 _amount,
        address _to)
    internal
    returns (bool) {
        if (!_bridge.isContract()) {
            return true;
        }

        bytes4 retval = ITokenReceiver(_bridge).onTokenReceived(_from, _amount, _to);
        return(retval == TOKEN_RECEIVED);
    }
}
