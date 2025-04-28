// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

/// @dev Collection of deprecated types and constants used in the ISC system

import "./ISCTypes.sol";

// An L1 stardust IOTA address
struct LegacyL1Address {
    bytes data;
}

// An IOTA stardust native token ID
struct LegacyNativeTokenID {
    bytes data;
}

// An amount of some IOTA stardust native token
struct LegacyNativeToken {
    LegacyNativeTokenID ID;
    uint256 amount;
}

// An IOTA stardust NFT ID
type LegacyNFTID is bytes32;

// The specifies an amount of funds (tokens) for an ISC call.
struct LegacyISCAssets {
    uint64 baseTokens;
    LegacyNativeToken[] nativeTokens;
    LegacyNFTID[] nfts;
}

// Parameters for building an on-ledger request
struct LegacyISCSendMetadata {
    ISCHname targetContract;
    ISCHname entrypoint;
    ISCDict params;
    LegacyISCAssets allowance;
    uint64 gasBudget;
}

// Parameters for building an on-ledger request
struct LegacyISCSendOptions {
    int64 timelock;
    LegacyISCExpiration expiration;
}

// Expiration of an on-ledger request
struct LegacyISCExpiration {
    int64 time;
    LegacyL1Address returnAddress;
}
