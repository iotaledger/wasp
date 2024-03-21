// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ERC20NativeTokens.sol";

// The ERC20 contract for ISC L2 native tokens (off-chain foundry).
contract ERC20ExternalNativeTokens is ERC20NativeTokens {
    NativeTokenID private _nativeTokenID;

    // TODO: this value is set at contract creation, and may get outdated
    uint256 private _maximumSupply;

    function nativeTokenID()
        public
        view
        override
        returns (NativeTokenID memory)
    {
        return _nativeTokenID;
    }

    function totalSupply() public view override returns (uint256) {
        return _maximumSupply;
    }
}
