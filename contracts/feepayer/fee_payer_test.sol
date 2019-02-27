pragma solidity ^0.4.24;

contract FeePayer {
    function GetFeePayerDirect() public returns (address) {
        assembly {
            if iszero(call(gas, 0x0a, 0, 0, 0, 12, 20)) {
              invalid()
            }
            return(0, 32)
        }
    }

    function feePayer() internal returns (address) {
        assembly {
            if iszero(call(gas, 0x0a, 0, 0, 0, 12, 20)) {
              invalid()
            }
            return(0, 32)
        }
    }

    function GetFeePayer() public returns (address) {
        return feePayer();
    }
}
