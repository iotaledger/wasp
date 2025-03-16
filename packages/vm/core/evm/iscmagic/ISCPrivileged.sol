// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import {ISCAssets, ISC_MAGIC_ADDRESS} from "./ISCTypes.sol";

/**
 * @title ISCPrivileged
 * @dev The ISCPrivileged interface represents a contract that has some extra methods not included in the standard ISC interface.
 * These methods can only be called from privileged contracts.
 */
interface ISCPrivileged {
    /**
     * @dev This function allows privileged contracts to move assets between accounts.
     * @param sender The address of the sender account
     * @param receiver The address of the receiver account
     * @param allowance The assets to be moved from the sender to the receiver
     */
    function moveBetweenAccounts(
        address sender,
        address receiver,
        ISCAssets memory allowance
    ) external;

    /**
     * @dev This function allows privileged contracts to set the allowance of base tokens from one account to another.
     * @param from The address of the account from which tokens are allowed
     * @param to The address of the account to which tokens are allowed
     * @param amount The number of base tokens to be allowed
     */
    function setAllowanceBaseTokens(
        address from,
        address to,
        uint64 amount
    ) external;

    /**
     * @dev This function allows privileged contracts to set the allowance of a
     *      given coin from one account to another.
     * @param from The address of the account from which tokens are allowed
     * @param to The address of the account to which tokens are allowed
     * @param coinType The coin type
     * @param amount The number of base tokens to be allowed
     */
    function setAllowanceCoin(
        address from,
        address to,
        string memory coinType,
        uint64 amount
    ) external;

    /**
     * @dev This function allows privileged contracts to move allowed funds from one account to another.
     * @param from The address of the account from which funds are allowed
     * @param to The address of the account to which funds are allowed
     * @param allowance The assets to be moved from the sender to the receiver
     */
    function moveAllowedFunds(
        address from,
        address to,
        ISCAssets memory allowance
    ) external;
}

ISCPrivileged constant __iscPrivileged = ISCPrivileged(ISC_MAGIC_ADDRESS);
