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

uint8 constant L1AddressTypeEd25519 = 0;
uint8 constant L1AddressTypeAlias = 8;
uint8 constant L1AddressTypeNFT = 16;

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

struct IRC27NFTMetadata {
    string standard;
    string version;
    string mimeType;
    string uri;
    string name;
}

// Information about an on-chain IRC27 NFT
struct IRC27NFT {
    ISCNFT nft;
    IRC27NFTMetadata metadata;
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

uint8 constant ISCAgentIDKindNil = 0;
uint8 constant ISCAgentIDKindAddress = 1;
uint8 constant ISCAgentIDKindContract = 2;
uint8 constant ISCAgentIDKindEthereumAddress = 3;

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

// Parameters for building an on-ledger request
struct ISCSendMetadata  {
    ISCHname targetContract;
    ISCHname entrypoint;
    ISCDict params;
    ISCAssets allowance;
    uint64 gasBudget;
}

// The specifies an amount of funds (tokens) for an ISC call.
struct ISCAssets {
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
    function L1AddressType(L1Address memory addr) internal pure returns (uint8) {
        return uint8(addr.data[0]);
    }

    function newEthereumAgentID(address addr) internal pure returns (ISCAgentID memory) {
        bytes memory addrBytes = abi.encodePacked(addr);
        ISCAgentID memory r;
        r.data = new bytes(1+addrBytes.length);
        r.data[0] = bytes1(ISCAgentIDKindEthereumAddress);
        for (uint i = 0; i < addrBytes.length; i++) {
            r.data[i+1] = addrBytes[i];
        }
        return r;
    }

    function isEthereum(ISCAgentID memory a) internal pure returns (bool) {
        return uint8(a.data[0]) == ISCAgentIDKindEthereumAddress;
    }

    function ethAddress(ISCAgentID memory a) internal pure returns (address) {
        bytes memory b = new bytes(20);
        for (uint i = 0; i < 20; i++) b[i] = a.data[i+1];
        return address(uint160(bytes20(b)));
    }

    function asNFTID(uint256 tokenID) internal pure returns (NFTID) {
        return NFTID.wrap(bytes32(tokenID));
    }

    function isInCollection(ISCNFT memory nft, NFTID collectionId) internal pure returns (bool) {
        if (L1AddressType(nft.issuer) != L1AddressTypeNFT) {
            return false;
        }
        assert(nft.issuer.data.length == 33);
        bytes memory collectionIdBytes = abi.encodePacked(collectionId);
        for (uint i = 0; i < 32; i++) {
            if (collectionIdBytes[i] != nft.issuer.data[i+1]) {
                return false;
            }
        }
        return true;
    }
}
