// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISCTypes.sol";

// The ISC magic contract has some extra methods not included in the standard ISC interface:
// (only callable from privileged contracts)
interface ISCPrivileged {
    function moveBetweenAccounts(address sender, address receiver, ISCAllowance memory allowance) external;
    function addToAllowance(address from, address to, ISCAllowance memory allowance) external;
    function moveAllowedFunds(address from, address to, ISCAllowance memory allowance) external;
}

ISCPrivileged constant __iscPrivileged = ISCPrivileged(ISC_MAGIC_ADDRESS);
