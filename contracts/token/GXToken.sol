pragma solidity ^0.4.24;

import "./ITokenReceiver.sol";
import "./SafeMath.sol";

contract GXToken {

  using SafeMath for uint256;

  // Transfer Gateway contract address
  address gateway;

  string public name = "GXToken";
  string public symbol = "GX";
  uint8 public decimals = 18;

  uint256 private totalSupply;
  mapping (address => uint256) private balances;

  uint256 public constant INITIAL_SUPPLY = 1000000000;

  event Transfer(address from, address to, uint256 amount);

  // TODO-Klaytn Undeclared identifier Err while compiling using abigen
  bytes4 constant TOKEN_RECEIVED = 0xbc04f0af;

  constructor (address _gateway) public {
    gateway = _gateway;
    totalSupply = INITIAL_SUPPLY * (10 ** uint256(decimals));
    balances[_gateway] = totalSupply;
  }

  function transfer(address to, uint256 value) public returns (bool) {
     _transfer(msg.sender,to,value);
     return true;
  }

  function balanceOfMine() public view returns (uint256) {
     return balances[msg.sender];
  }

  function _transfer(address from, address to, uint256 value) internal {
    require(to != address(0));

    balances[from] = balances[from].sub(value);
    balances[to] = balances[to].add(value);
    emit Transfer(from, to, value);
  }

  function depositToGateway(uint256 amount) external {
    safeTransferAndCall(gateway, amount);
  }

  // Called by the gateway contract to mint tokens that have been deposited to the Mainnet gateway.
  function mintToGateway(uint256 _amount) public {
    require(msg.sender == gateway);
    totalSupply = totalSupply.add(_amount);
    balances[gateway] = balances[gateway].add(_amount);
  }

  function safeTransferAndCall(address _to, uint256 amount) public {
    transfer(_to, amount);
    require(checkAndCallSafeTransfer(msg.sender, _to, amount));
  }

  function checkAndCallSafeTransfer(address _from, address _to, uint256 amount) internal returns (bool) {
    bytes4 retval = ITokenReceiver(_to).onTokenReceived(_from, amount);
    return(retval == TOKEN_RECEIVED);
  }
}
