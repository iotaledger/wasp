// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity ^0.8.0;

contract Storage {
    uint32 n;

    constructor(uint32 _n) {
        n = _n;
    }

    function store(uint32 _n) public {
        n = _n;
    }

    function retrieve() public view returns (uint32) {
        return n;
    }
}

