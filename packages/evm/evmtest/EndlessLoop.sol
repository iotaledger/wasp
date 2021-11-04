// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity ^0.8.0;

contract EndlessLoop {
    function loop() public pure {
        while (true) {}
    }
}
