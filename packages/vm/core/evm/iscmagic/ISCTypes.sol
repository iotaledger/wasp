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

// A L1 address
type IotaAddress is bytes32;

// A Iota ObjectID
type IotaObjectID is bytes32;

// Information about a given L1 coin
struct IotaCoinInfo {
    string coinType;
    uint8 decimals;
    string name;
    string symbol;
    string description;
    string iconUrl;
    uint64 totalSupply;
}

struct IRC27NFTMetadata {
    string standard;
    string version;
    string mimeType;
    string uri;
    string name;
    string description;
}

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
type ISCRequestID is bytes32;

// A single key-value pair
struct ISCDictItem {
    bytes key;
    bytes value;
}

// Wrapper for the isc.Dict type, a collection of key-value pairs
struct ISCDict {
    ISCDictItem[] items;
}

struct ISCTarget {
    ISCHname contractHname;
    ISCHname entryPoint;
}

struct ISCMessage {
    ISCTarget target;
    bytes[] params;
}

// Parameters for building an on-ledger request
struct ISCSendMetadata {
    ISCMessage message;
    ISCAssets allowance;
    uint64 gasBudget;
}

// Parameters for building an on-ledger request
struct ISCSendOptions {
    int64 timelock;
    ISCExpiration expiration;
}

// Expiration of an on-ledger request
struct ISCExpiration {
    int64 time;
    IotaAddress returnAddress;
}


// A collection of coins and object IDs
struct ISCAssets {
    CoinBalance[] coins;
    IotaObjectID[] objects;
}

// An amount of some Iota coin
struct CoinBalance {
    string coinType;
    uint64 amount;
}

library ISCTypes {
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

        // write agentID kind
        r.data[0] = bytes1(ISCAgentIDKindEthereumAddress);

        // write chainID
        for (uint i = 0; i < chainIDBytes.length; i++) {
            r.data[i + 1] = chainIDBytes[i];
        }

        // write eth addr
        for (uint i = 0; i < addrBytes.length; i++) {
            r.data[i + 1 + chainIDBytes.length] = addrBytes[i];
        }
        return r;
    }

    function newL1AgentID(IotaAddress addr) internal pure returns (ISCAgentID memory) {
        bytes memory addrBytes = abi.encodePacked(addr);

        ISCAgentID memory r;
        r.data = new bytes(1 + addrBytes.length);

        // write agentID kind
        r.data[0] = bytes1(ISCAgentIDKindAddress); // isc agentID kind

        // write l1 address
        for (uint i = 0; i < addrBytes.length; i++) {
            r.data[i + 1] = addrBytes[i];
        }
        return r;
    }

    function isEthereum(ISCAgentID memory a) internal pure returns (bool) {
        return uint8(a.data[0]) == ISCAgentIDKindEthereumAddress;
    }

    /**
     * @dev Get the Ethereum address from an ISCAgentID.
     * @param a The ISCAgentID.
     * @return The Ethereum address.
     */
    function ethAddress(ISCAgentID memory a) internal pure returns (address) {
        require(isEthereum(a));
        bytes memory b = new bytes(20);
        // offset of 33 (kind byte + chainID)
        for (uint i = 0; i < 20; i++) b[i] = a.data[i + 33];
        return address(uint160(bytes20(b)));
    }

    /**
     * @dev Get the chain ID from an ISCAgentID.
     * @param a The ISCAgentID.
     * @return The ISCChainID.
     */
    function chainID(ISCAgentID memory a) internal pure returns (ISCChainID) {
        require(isEthereum(a));
        bytes32 out;
        for (uint i = 0; i < 32; i++) {
            // offset of 1 (kind byte)
            out |= bytes32(a.data[1 + i] & 0xFF) >> (i * 8);
        }
        return ISCChainID.wrap(out);
    }

    function asObjectID(uint256 tokenID) internal pure returns (IotaObjectID) {
        return IotaObjectID.wrap(bytes32(tokenID));
    }

    /**
     * @dev Check if an NFT is part of a given collection.
     * @param nft The NFT to check.
     * @param collectionId The collection ID to check against.
     * @return True if the NFT is part of the collection, false otherwise.
     */
    function isInCollection(
        ISCNFT memory,
        IotaObjectID
    ) internal pure returns (bool) {
        assert(false); // TODO
        // return nft.issuer == collectionId;
        return false;
    }

    function getCoinAmount(
        CoinBalance[] memory coins,
        string memory coinType
    ) internal pure returns (uint64) {
        for (uint i = 0; i < coins.length; i++) {
            if (stringsEqual(coins[i].coinType, coinType)) {
                return coins[i].amount;
            }
        }
        return 0;
    }

    function bytesEqual(
        bytes memory a,
        bytes memory b
    ) internal pure returns (bool) {
        if (a.length != b.length) {
            return false;
        }
        for (uint i = 0; i < a.length; i++) {
            if (a[i] != b[i]) {
                return false;
            }
        }
        return true;
    }

    function stringsEqual(
        string memory a,
        string memory b
    ) internal pure returns (bool) {
        return bytesEqual(bytes(a), bytes(b));
    }
}
