// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

/**
 * @dev Collection of types and constants used in the ISC system
 */

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
    string description;
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
struct ISCSendMetadata {
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
    /**
     * @dev Get the type of an L1 address.
     * @param addr The L1 address.
     * @return The type of the L1 address.
     *
     * For more details about the types of L1 addresses, please refer to the IOTA Tangle Improvement Proposal (TIP) 18:
     * https://wiki.iota.org/tips/tips/TIP-0018/#address-unlock-condition
     */
    function L1AddressType(
        L1Address memory addr
    ) internal pure returns (uint8) {
        return uint8(addr.data[0]);
    }

    /**
     * @dev Create a new Ethereum AgentID.
     * @param addr The Ethereum address.
     * @param iscChainID The ISC chain ID.
     * @return The new ISCAgentID.
     */
    function newEthereumAgentID(
        address addr,
        ISCChainID iscChainID
    ) internal pure returns (ISCAgentID memory) {
        bytes memory chainIDBytes = abi.encodePacked(iscChainID);
        bytes memory addrBytes = abi.encodePacked(addr);
        ISCAgentID memory r;
        r.data = new bytes(1 + addrBytes.length + chainIDBytes.length);
        r.data[0] = bytes1(ISCAgentIDKindEthereumAddress);

        //write chainID
        for (uint i = 0; i < chainIDBytes.length; i++) {
            r.data[i + 1] = chainIDBytes[i];
        }

        //write eth addr
        for (uint i = 0; i < addrBytes.length; i++) {
            r.data[i + 1 + chainIDBytes.length] = addrBytes[i];
        }
        return r;
    }

    /**
     * @dev Create a new L1 AgentID.
     * @param l1Addr The L1 address.
     * @return The new ISCAgentID.
     */
    function newL1AgentID(
        bytes memory l1Addr
    ) internal pure returns (ISCAgentID memory) {
        if (l1Addr.length != 32) {
            revert("bad address length");
        }
        ISCAgentID memory r;
        r.data = new bytes(2 + 32); // 2 for the kind + The hash size of BLAKE2b-256 in bytes.
        r.data[0] = bytes1(ISCAgentIDKindAddress); // isc agentID kind
        r.data[1] = bytes1(0); // iota go AddressEd25519 AddressType = 0

        //write l1 address
        for (uint i = 0; i < l1Addr.length; i++) {
            r.data[i + 2] = l1Addr[i];
        }
        return r;
    }

    /**
     * @dev Check if an ISCAgentID is of Ethereum type.
     * @param a The ISCAgentID to check.
     * @return True if the ISCAgentID is of Ethereum type, false otherwise.
     */
    function isEthereum(ISCAgentID memory a) internal pure returns (bool) {
        return uint8(a.data[0]) == ISCAgentIDKindEthereumAddress;
    }

    /**
     * @dev Get the Ethereum address from an ISCAgentID.
     * @param a The ISCAgentID.
     * @return The Ethereum address.
     */
    function ethAddress(ISCAgentID memory a) internal pure returns (address) {
        bytes memory b = new bytes(20);
        //offset of 33 (kind byte + chainID)
        for (uint i = 0; i < 20; i++) b[i] = a.data[i + 33];
        return address(uint160(bytes20(b)));
    }

    /**
     * @dev Get the chain ID from an ISCAgentID.
     * @param a The ISCAgentID.
     * @return The ISCChainID.
     */
    function chainID(ISCAgentID memory a) internal pure returns (ISCChainID) {
        bytes32 out;
        for (uint i = 0; i < 32; i++) {
            //offset of 1 (kind byte)
            out |= bytes32(a.data[1 + i] & 0xFF) >> (i * 8);
        }
        return ISCChainID.wrap(out);
    }

    /**
     * @notice Convert a token ID to an NFTID.
     * @param tokenID The token ID.
     * @return The NFTID.
     */
    function asNFTID(uint256 tokenID) internal pure returns (NFTID) {
        return NFTID.wrap(bytes32(tokenID));
    }

    /**
     * @dev Check if an NFT is part of a given collection.
     * @param nft The NFT to check.
     * @param collectionId The collection ID to check against.
     * @return True if the NFT is part of the collection, false otherwise.
     */
    function isInCollection(
        ISCNFT memory nft,
        NFTID collectionId
    ) internal pure returns (bool) {
        if (L1AddressType(nft.issuer) != L1AddressTypeNFT) {
            return false;
        }
        assert(nft.issuer.data.length == 33);
        bytes memory collectionIdBytes = abi.encodePacked(collectionId);
        for (uint i = 0; i < 32; i++) {
            if (collectionIdBytes[i] != nft.issuer.data[i + 1]) {
                return false;
            }
        }
        return true;
    }
}
