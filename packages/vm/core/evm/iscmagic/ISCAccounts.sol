// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";

/**
 * @title ISCAccounts
 * @dev Functions of the ISC Magic Contract to access the core accounts functionality
 */
interface ISCAccounts {
    // Get the L2 base tokens balance of an account
    function getL2BalanceBaseTokens(ISCAgentID memory agentID) external view returns (uint64);

    // Get the L2 coin balance of an account
    function getL2BalanceCoin(
        string memory coinType,
        ISCAgentID memory agentID
    ) external view returns (uint64);

    // Get the list of objects owned by an account on L2
    function getL2Objects(ISCAgentID memory agentID) external view
        returns (SuiObjectID[] memory);

    // Get the amount of objects owned by an account on L2
    function getL2ObjectsCount(ISCAgentID memory agentID) external view
        returns (uint256);

    // Get the objects of a given collection owned by an account on L2
    function getL2ObjectsInCollection(
        ISCAgentID memory agentID,
        SuiObjectID collectionId
    ) external view returns (SuiObjectID[] memory);

    // Get the amount of objects of a given collection owned by an account on L2
    function getL2ObjectsCountInCollection(
        ISCAgentID memory agentID,
        SuiObjectID collectionId
    ) external view returns (uint256);
}

ISCAccounts constant __iscAccounts = ISCAccounts(ISC_MAGIC_ADDRESS);
