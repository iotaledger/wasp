// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: MIT
pragma solidity >=0.8.5;

/// @title Pseudorandom Number Generator (PRNG) Library
/// @notice This library is used to generate pseudorandom numbers
/// @dev Not recommended for generating cryptographic secure randomness
library PRNG {
    /// @notice Parameters of the PRNG (specifically, a Linear Congruential Generator)
    /// @dev These are constants required for generating the random number sequence
    uint256 constant LCG_MULTIPLIER = 1103515245; // Multiplier in LCG
    uint256 constant LCG_INCREMENT = 12345; // Increment in LCG
    uint256 constant LCG_MODULUS = 2**31; // Modulus in LCG

    /// @dev Represents the state of the PRNG
    struct PRNGState {
        uint256 state;
    }

    /// @notice Generate a new pseudorandom number
    /// @dev Uses the formula of Linear Congruential Generator (LCG): new_state = (LCG_MULTIPLIER*state + LCG_INCREMENT) mod LCG_MODULUS
    /// @param self The PRNGState struct to use and alter the state
    /// @return The generated pseudorandom number
    function generateRandomNumber(PRNGState storage self) public returns (uint256) {
        require(self.state != 0, "Generator has not been initialized, state is zero.");
        self.state = (LCG_MULTIPLIER * self.state + LCG_INCREMENT) % LCG_MODULUS;
        return self.state;
    }

    /// @notice Seed the PRNG
    /// @dev The seed should not be zero
    /// @param self The PRNGState struct to update the state
    /// @param entropy The seed value (entropy)
    function seed(PRNGState storage self, bytes32 entropy) internal {
        require(entropy != bytes32(0), "Entered entropy should not be zero");
        self.state = uint256(entropy);
    }
}

