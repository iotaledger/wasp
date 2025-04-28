// SPDX-License-Identifier: MIT
pragma solidity 0.8.20;

// remober to run `git clone https://github.com/OpenZeppelin/openzeppelin-contracts.git` under lib/ first
import {ERC20} from "openzeppelin-contracts/contracts/token/ERC20/ERC20.sol";

contract TestTokenERC20 is ERC20 {
    constructor() ERC20("TestToken", "TTK") {
        _mint(msg.sender, 1_000_000); // Mint tokens to deployer
    }
}