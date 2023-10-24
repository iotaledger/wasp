// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: MIT
pragma solidity >=0.8.5;

/// @title Pseudorandom Number Generator (PRNG) Library
/// @notice This library is used to generate pseudorandom numbers
/// @dev Not recommended for generating cryptographic secure randomness
library PRNG {

    /// @dev Represents the state of the PRNG
    struct PRNGState {
        bytes32 state;
    }

    /// @notice Generate a new pseudorandom hash
    /// @dev Takes the current state, hashes it and returns the new state.
    /// @param self The PRNGState struct to use and alter the state
    /// @return The generated pseudorandom hash
    function generateRandomHash(PRNGState storage self) internal returns (bytes32) {
        require(self.state != bytes32(0), "state must be seeded first");
        self.state = keccak256(abi.encodePacked(self.state));
        return self.state;
    }

    /// @notice Generate a new pseudorandom number
    /// @dev Takes the current state, hashes it and returns the new state.
    /// @param self The PRNGState struct to use and alter the state
    /// @return The generated pseudorandom number
    function generateRandomNumber(PRNGState storage self) internal returns (uint256) {
        return uint256(PRNG.generateRandomHash(self));
    }

    /// @notice Generate a new pseudorandom number in a given range [min, max)
    /// @dev Takes the current state, hashes it and returns the new state. It constrains the returned number to the bounds of min (inclusive) and max (exclusive).
    /// @param self The PRNGState struct to use and alter the state
    /// @return The generated pseudorandom number constrained to the bounds of [min, max)
    function generateRandomNumberInRange(PRNGState storage self, uint256 min, uint256 max) internal returns (uint256) {
        require(max > min, "max should be greater than min");
        uint256 diff = max - min;
        uint256 randomNumber = PRNG.generateRandomNumber(self);
        randomNumber = randomNumber % diff;
        return randomNumber + min;
    }

    /// @notice Seed the PRNG
    /// @dev The seed should not be zero
    /// @param self The PRNGState struct to update the state
    /// @param entropy The seed value (entropy)
    function seed(PRNGState storage self, bytes32 entropy) internal {
        require(entropy != bytes32(0), "seed must not be zero");
        self.state = entropy;
    }
}