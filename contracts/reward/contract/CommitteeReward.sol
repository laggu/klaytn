pragma solidity ^0.4.24;

contract CommitteeReward {

    uint public totalAmount;
    mapping(address => uint256) public balanceOf;

    function CommitteeReward() public {
    }

    function () payable public {
        uint amount = msg.value;
        balanceOf[msg.sender] += amount;
        totalAmount += amount;
    }

    function reward(address receiver) payable public {
        uint amount = msg.value;
        balanceOf[receiver] += amount;
        totalAmount += amount;
    }

    function safeWithdrawal() public {
        uint amount = balanceOf[msg.sender];
        balanceOf[msg.sender] = 0;
        if (amount > 0) {
             if (msg.sender.send(amount)) {

             } else {
                balanceOf[msg.sender] = amount;
             }
        }
    }
}
