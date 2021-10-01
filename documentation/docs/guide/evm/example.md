---
keywords:
- ISCP
- Smart Contracts
- EVM
- Solidity
description: Solidity smart contract example
image: /img/logo/WASP_logo_dark.png
---
# Solidity Smart Contract Example

Given Solidity smart contracts on ISCP are compatible with Solidity smart contracts on any other network most smart contracts will work directly without modification. To give a rough indication of how a very simple Solidity smart contract looks like see the example below:


```solidity
pragma solidity ^0.8.6;
// No SafeMath needed for Solidity 0.8+

contract Counter { 
   
    uint256 private _count;
        
    function current() public view returns (uint256) {
        return _count;
    }   

    function increment() public {
        _count += 1;
    }   

    function decrement() public {
        _count -= 1;
    }   
}
```

The full documentation for Solidity is [available here](https://docs.soliditylang.org/).
