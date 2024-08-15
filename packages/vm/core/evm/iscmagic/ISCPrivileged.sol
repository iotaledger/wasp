// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";

/**
 * @title ISCPrivileged
 * @dev The ISCPrivileged interface represents a contract that has some extra methods not included in the standard ISC interface.
 * These methods can only be called from privileged contracts.
 */
interface ISCPrivileged {
    function moveBetweenAccounts(
        address sender,
        address receiver,
        ISCAssets memory allowance
    ) external;

    function setAllowanceBaseTokens(
        address from,
        address to,
        uint64 amount
    ) external;

    function setAllowanceCoin(
        address from,
        address to,
        string memory coinType,
        uint64 amount
    ) external;

    function moveAllowedFunds(
        address from,
        address to,
        ISCAssets memory allowance
    ) external;
}

ISCPrivileged constant __iscPrivileged = ISCPrivileged(ISC_MAGIC_ADDRESS);
