// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity ^0.8.0;

contract GasTestExecutionTime {
    
    function f(uint32 n) public pure returns(uint32){
       uint32 x = 0;
       uint32 y = 0;

       for (uint32 i = 0;i < n;i++) {
           x += 1;
           y += 3 * (x % 10);
       }
       return y;
    }
}

