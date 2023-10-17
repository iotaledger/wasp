// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: MIT
pragma solidity >=0.8.5;

import "./prng.sol";

contract PRNGTest {
    using PRNG for PRNG.PRNGState;

    PRNG.PRNGState internal prngState;

    event RandomNumberGenerated(uint256 randomNumber);
    event RandomHashGenerated(bytes32 randomHash);

    constructor() {
        PRNG.seed(prngState, keccak256("test"));
    }

    function generateRandomHash() public returns (bytes32) {
        bytes32 hash = prngState.generateRandomHash();
        emit RandomHashGenerated(hash);
        return hash;
    }

    function generateRandomNumber() public returns (uint256) {
        uint256 randomNumber = prngState.generateRandomNumber();
        emit RandomNumberGenerated(randomNumber);
        return randomNumber;
    }

    function generateRandomNumberInRange(uint256 min, uint256 max) public returns (uint256) {
        uint256 randomNumberInRange = prngState.generateRandomNumberInRange(min, max);
        emit RandomNumberGenerated(randomNumberInRange);
        return randomNumberInRange;
    }

    function getPRNGState() public view returns (bytes32) {
        return prngState.state;
    }
}

