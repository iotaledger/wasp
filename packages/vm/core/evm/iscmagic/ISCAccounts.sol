// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";

/**
 * @title ISCAccounts
 * @notice Functions of the ISC Magic Contract to access the core accounts functionality
 */
interface ISCAccounts {
    
    /**
     * @notice Retrieves the L2 balance of base tokens for a given ISC Agent ID.
     * @param agentID The ISC Agent ID whose L2 base token balance is to be queried.
     * @return The L2 balance of base tokens.
     */
    function getL2BalanceBaseTokens(ISCAgentID memory agentID) external view returns (uint64);

    /**
     * @notice Retrieves the L2 balance of a given coin type for a given ISC Agent ID.
     * @param coinType The type of the coin as a string.
     * @param agentID The ISC Agent ID whose L2 coin balance is to be queried.
     * @return The L2 balance of the given coin type.
     */
    function getL2BalanceCoin(
        string memory coinType,
        ISCAgentID memory agentID
    ) external view returns (uint64);

    /**
     * @notice Retrieves a list of Objects owned by a given ISC Agent ID.
     * @param agentID The ISC Agent ID whose L2 objects are to be queried.
     * @return An array of Object IDs.
     * 
     * @notice This function returns all objects except coins.
     */
    function getL2Objects(ISCAgentID memory agentID) external view
        returns (IotaObjectID[] memory);

    /**
     * @notice Retrieves the amount of Objects owned by a given ISC Agent ID.
     * @param agentID The ISC Agent ID whose L2 object count is to be queried.
     * @return The amount of Objects owned by a given ISC Agent ID.
     */
    function getL2ObjectsCount(ISCAgentID memory agentID) external view
        returns (uint256);

    /**
     * @notice Retrieves a list of Objects of a given collection owned by a given ISC Agent ID.
     * @param agentID The ISC Agent ID whose L2 objects are to be queried.
     * @param collectionId The collection ID whose objects are to be queried.
     * @return An array of Object IDs.
     */
    function getL2ObjectsInCollection(
        ISCAgentID memory agentID,
        IotaObjectID collectionId
    ) external view returns (IotaObjectID[] memory);

    /**
     * @notice Retrieves the amount of Objects of a given collection owned by a given ISC Agent ID.
     * @param agentID The ISC Agent ID whose L2 object count is to be queried.
     * @param collectionId The collection ID whose objects are to be queried.
     * @return The amount of Objects of a given collection owned by a given ISC Agent ID.
     */
    function getL2ObjectsCountInCollection(
        ISCAgentID memory agentID,
        IotaObjectID collectionId
    ) external view returns (uint256);
}

ISCAccounts constant __iscAccounts = ISCAccounts(ISC_MAGIC_ADDRESS);
