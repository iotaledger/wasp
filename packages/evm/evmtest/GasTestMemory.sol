// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity ^0.8.0;

contract GasTestMemory {
    function f(uint32 n) public pure {
        uint32[] memory store = new uint32[](n);
        for (uint32 i = 0;i < n;i++) {
            store[i] = i;
        }
    }
}
