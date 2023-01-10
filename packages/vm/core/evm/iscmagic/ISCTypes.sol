// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

// Every ISC chain is initialized with an instance of the Magic contract at this address
address constant ISC_MAGIC_ADDRESS = 0x1074000000000000000000000000000000000000;

// The ERC20 contract for base tokens is at this address:
address constant ISC_ERC20BASETOKENS_ADDRESS = 0x1074010000000000000000000000000000000000;

// The ERC721 contract for NFTs is at this address:
address constant ISC_ERC721_ADDRESS = 0x1074030000000000000000000000000000000000;

// An L1 IOTA address
struct L1Address {
    bytes data;
}

// An IOTA native token ID
struct NativeTokenID {
    bytes data;
}

// An amount of some IOTA native token
struct NativeToken {
    NativeTokenID ID;
    uint256 amount;
}

// The scheme used to create an IOTA native token (corresponds to iotago.SimpleTokenScheme)
struct NativeTokenScheme {
    uint256 mintedTokens;
    uint256 meltedTokens;
    uint256 maximumSupply;
}


// An IOTA NFT ID
type NFTID is bytes32;

// Information about an on-chain NFT
struct ISCNFT {
    NFTID ID;
    L1Address issuer;
    bytes metadata;
    ISCAgentID owner;
}

// An ISC transaction ID
type ISCTransactionID is bytes32;

// An ISC hname
type ISCHname is uint32;

// An ISC chain ID
type ISCChainID is bytes32;

// An ISC AgentID
struct ISCAgentID {
    bytes data;
}

// An ISC request ID
struct ISCRequestID {
    bytes data;
}

// A single key-value pair
struct ISCDictItem {
    bytes key;
    bytes value;
}

// Wrapper for the isc.Dict type, a collection of key-value pairs
struct ISCDict {
    ISCDictItem[] items;
}

// A collection of fungible tokens (base tokens + native tokens)
struct ISCFungibleTokens {
    uint64 baseTokens;
    NativeToken[] nativeTokens;
}

// Parameters for building an on-ledger request
struct ISCSendMetadata  {
    ISCHname targetContract;
    ISCHname entrypoint;
    ISCDict params;
    ISCAllowance allowance;
    uint64 gasBudget;
}

// The allowance of an ISC call.
// The specified tokens, assets and NFTs are transferred from the caller's L2 account to
// the callee's L2 account.
struct ISCAllowance {
    uint64 baseTokens;
    NativeToken[] nativeTokens;
    NFTID[] nfts;
}

// Parameters for building an on-ledger request
struct ISCSendOptions {
    int64 timelock;
    ISCExpiration expiration;
}

// Expiration of an on-ledger request
struct ISCExpiration {
    int64 time;
    L1Address returnAddress;
}

// Properties of an ISC base/native token
struct ISCTokenProperties {
    string name;
    string tickerSymbol;
    uint8 decimals;
    uint256 totalSupply;
}

library ISCTypes {
    uint8 constant AgentIDKindEthereumAddress = 3;

    function newEthereumAgentID(address addr) internal pure returns (ISCAgentID memory) {
        bytes memory addrBytes = abi.encodePacked(addr);
        ISCAgentID memory r;
        r.data = new bytes(1+addrBytes.length);
        r.data[0] = bytes1(AgentIDKindEthereumAddress);
        for (uint i = 0; i < addrBytes.length; i++) {
            r.data[i+1] = addrBytes[i];
        }
        return r;
    }

    function isEthereum(ISCAgentID memory a) internal pure returns (bool) {
        return uint8(a.data[0]) == AgentIDKindEthereumAddress;
    }

    function ethAddress(ISCAgentID memory a) internal pure returns (address) {
        bytes memory b = new bytes(20);
        for (uint i = 0; i < 20; i++) b[i] = a.data[i+1];
        return address(uint160(bytes20(b)));
    }

    function asNFTID(uint256 tokenID) internal pure returns (NFTID) {
        return NFTID.wrap(bytes32(tokenID));
    }
}
