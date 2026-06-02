// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract Counter {
    uint256 public count;
    address public owner;

    event CountIncremented(uint256 newCount);
    event CountDecremented(uint256 newCount);
    event CountSet(uint256 newCount);

    constructor() {
        owner = msg.sender;
        count = 0;
    }

    function increment() public {
        count += 1;
        emit CountIncremented(count);
    }

    function decrement() public {
        require(count > 0, "Counter cannot be negative");
        count -= 1;
        emit CountDecremented(count);
    }

    function setCount(uint256 newCount) public {
        count = newCount;
        emit CountSet(newCount);
    }

    function getCount() public view returns (uint256) {
        return count;
    }
}