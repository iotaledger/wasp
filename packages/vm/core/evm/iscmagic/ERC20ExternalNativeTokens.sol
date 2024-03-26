// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ERC20NativeTokens.sol";

/**
 * @title ERC20ExternalNativeTokens
 * @dev The ERC20 contract for externally registered native tokens (off-chain foundry).
 */
contract ERC20ExternalNativeTokens is ERC20NativeTokens {
    NativeTokenID private _nativeTokenID;

    // TODO: this value is set at contract creation, and may get outdated
    uint256 private _maximumSupply;

    /**
     * @dev Returns the native token ID.
     * @return The native token ID.
     */
    function nativeTokenID()
        public
        view
        override
        returns (NativeTokenID memory)
    {
        return _nativeTokenID;
    }

    /**
     * @dev Returns the total supply of the native tokens.
     * @return The total supply of the native tokens.
     */
    function totalSupply() public view override returns (uint256) {
        return _maximumSupply;
    }
}
