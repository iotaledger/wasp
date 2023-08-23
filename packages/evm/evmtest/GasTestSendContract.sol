// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity ^0.8.0;

contract GasTestSendContract {
    address private _owner;

    /**
     * @dev Initializes the contract setting the deployer as the initial owner.
     */
    constructor() payable {
        _owner = msg.sender;
    }

    fallback() external payable {}

    receive() external payable {}

    /**
     * @dev Returns the address of the current owner.
     */
    function owner() public view virtual returns (address) {
        return _owner;
    }

    function deposit() public payable {}

    function withdraw(uint256 _value, address payable to) public payable returns(bool) {
        uint256 amount = address(this).balance;
        
        require(amount >= _value);
        require(msg.sender == owner());
        (bool success, ) = to.call{value: _value}("");
        return success;
    }
}
    