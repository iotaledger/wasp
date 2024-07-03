// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";

/**
 * @title ISCUtil
 * @dev Functions of the ISC Magic Contract not directly related to the ISC sandbox
 */
interface ISCUtil {
    /**
     * @notice Get an ISC contract's hname given its instance name
     * @dev Converts a string instance name to its corresponding ISC hname.
     * @param s The instance name of the ISC contract.
     * @return The ISCHname corresponding to the given instance name.
     */
    function hn(string memory s) external pure returns (ISCHname);

    /**
     * @notice Print something to the console (will only work when debugging contracts with Solo)
     * @dev Prints the given string to the console for debugging purposes.
     * @param s The string to print to the console.
     */
    function print(string memory s) external pure;
}

ISCUtil constant __iscUtil = ISCUtil(ISC_MAGIC_ADDRESS);
