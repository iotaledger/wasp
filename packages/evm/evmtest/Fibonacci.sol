// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity ^0.8.0;

contract Fibonacci {

    function fib(uint32 n) public view returns(uint32) {
        if (n <= 1) {
            return n;
        }
        return this.fib(n-1) + this.fib(n-2);
    }
}
