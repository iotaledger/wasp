// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

/// @dev Collection of types and constants used in the ISC system

/// @dev Address of the Magic contract deployed on every chain
address constant ISC_MAGIC_ADDRESS = 0x1074000000000000000000000000000000000000;

/// Move Address
type IotaAddress is bytes32;

/// Object ID
type IotaObjectID is bytes32;

/// Information about a given Move coin
struct IotaCoinInfo {
    string coinType;
    uint8 decimals;
    string name;
    string symbol;
    string description;
    string iconUrl;
    uint64 totalSupply;
}

/// An ISC hname
type ISCHname is uint32;

/// An ISC chain ID
type ISCChainID is bytes32;

/// An ISC agent ID
struct ISCAgentID {
    bytes data;
}

/// @dev ISC Agent ID kinds
uint8 constant ISCAgentIDKindNil = 0;
/// @dev IOTA Address kind
uint8 constant ISCAgentIDKindAddress = 1;
/// @dev Ethereum Contract kind
uint8 constant ISCAgentIDKindContract = 2;
/// @dev Ethereum Address kind
uint8 constant ISCAgentIDKindEthereumAddress = 3;

/// An ISC request ID
type ISCRequestID is bytes32;

/// A single key-value pair
struct ISCDictItem {
    bytes key;
    bytes value;
}

/// Wrapper for the isc.Dict type, a collection of key-value pairs
struct ISCDict {
    ISCDictItem[] items;
}

/// Struct containing the target contract and entry point hnames
struct ISCTarget {
    ISCHname contractHname;
    ISCHname entryPoint;
}

/// A single message to be sent to an ISC contract
struct ISCMessage {
    ISCTarget target;
    bytes[] params;
}

/// A collection of coins and object IDs
struct ISCAssets {
    CoinBalance[] coins;
    IotaObject[] objects;
}

/// An amount of some IOTA coin
struct CoinBalance {
    string coinType;
    uint64 amount;
}

/// An IOTA object
struct IotaObject {
    IotaObjectID id;
    string objectType;
}

/**
 * @title ISCTypes
 * @notice Collection of utility functions used in the ISC system
 */
library ISCTypes {
    /**
     * @notice Create a new Agent ID from an Ethereum address.
     * @param addr The Ethereum address.
     * @return The new ISCAgentID.
     */
    function newEthereumAgentID(
        address addr
    ) internal pure returns (ISCAgentID memory) {
        bytes memory addrBytes = abi.encodePacked(addr);
        ISCAgentID memory r;
        r.data = new bytes(1 + addrBytes.length);

        // write agentID kind
        r.data[0] = bytes1(ISCAgentIDKindEthereumAddress);

        // write eth addr
        for (uint i = 0; i < addrBytes.length; i++) {
            r.data[i + 1] = addrBytes[i];
        }
        return r;
    }

    /**
     * @notice Creates a new Agent ID from an IOTA Address.
     * @param addr The IOTA address.
     * @return The newly created Agent ID.
     */
    function newL1AgentID(
        IotaAddress addr
    ) internal pure returns (ISCAgentID memory) {
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

    /**
     * @notice Check if an Agent ID is an Ethereum Agent ID.
     * @param a The Agent ID.
     * @return True if the Agent ID is an Ethereum Agent ID.
     */
    function isEthereum(ISCAgentID memory a) internal pure returns (bool) {
        return uint8(a.data[0]) == ISCAgentIDKindEthereumAddress;
    }

    /**
     * @notice Get the Ethereum address from an Agent ID.
     * @param a The Agent ID.
     * @return The Ethereum address.
     */
    function ethAddress(ISCAgentID memory a) internal pure returns (address) {
        require(isEthereum(a));
        bytes memory b = new bytes(20);
        // offset of 1 (kind byte + address)
        for (uint i = 0; i < 20; i++) b[i] = a.data[i + 1];
        return address(uint160(bytes20(b)));
    }

    function asObjectID(uint256 tokenID) internal pure returns (IotaObjectID) {
        return IotaObjectID.wrap(bytes32(tokenID));
    }

    /**
     * @notice Get the amount of a specific coin type from a list of CoinBalance.
     * @param coins The list of CoinBalance.
     * @param coinType The coin type to search for.
     * @return The amount of the specified coin type.
     */
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

    /**
     * @notice Compare two byte arrays for equality.
     * @param a The first byte array.
     * @param b The second byte array.
     * @return True if the byte arrays are equal.
     */
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

    /**
     * @notice Compare two strings for equality.
     * @param a The first string.
     * @param b The second string.
     * @return True if the strings are equal.
     */
    function stringsEqual(
        string memory a,
        string memory b
    ) internal pure returns (bool) {
        return bytesEqual(bytes(a), bytes(b));
    }
}
