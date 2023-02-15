// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ERC20NativeTokens.sol";

// The ERC20 contract for ISC L2 native tokens (off-chain foundry).
contract ERC20ExternalNativeTokens is ERC20NativeTokens {
    NativeTokenID _nativeTokenID;

    // TODO: this value is set at contract creation, and may get outdated
    uint256 _maximumSupply;

    function nativeTokenID() public override view returns (NativeTokenID memory) {
        return _nativeTokenID;
    }

    function totalSupply() public override view returns (uint256) {
        return _maximumSupply;
    }
}
