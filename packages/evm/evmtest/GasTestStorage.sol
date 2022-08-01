// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity ^0.8.0;

contract GasTestStorage {
    uint32[] store;

    function f(uint32 n) public {
        for (uint32 i = 0;i < n;i++) {
            store.push(i);
        }
    }
}

