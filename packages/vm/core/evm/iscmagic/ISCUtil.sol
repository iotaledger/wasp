// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISCTypes.sol";

// Functions of the ISC Magic Contract not directly related to the ISC sandbox
interface ISCUtil {
    // Get an ISC contract's hname given its instance name
    function hn(string memory s) external pure returns (ISCHname);

    // Print something to the console (will only work when debugging contracts with Solo)
    function print(string memory s) external pure;
}

ISCUtil constant __iscUtil = ISCUtil(ISC_MAGIC_ADDRESS);
