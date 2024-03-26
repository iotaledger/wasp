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
        uint256 numTokens
    ) external;

    function setAllowanceNativeTokens(
        address from,
        address to,
        NativeTokenID memory nativeTokenID,
        uint256 numTokens
    ) external;

    function moveAllowedFunds(
        address from,
        address to,
        ISCAssets memory allowance
    ) external;
}

ISCPrivileged constant __iscPrivileged = ISCPrivileged(ISC_MAGIC_ADDRESS);
