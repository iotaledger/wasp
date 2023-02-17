// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISCTypes.sol";

// Functions of the ISC Magic Contract to access the core accounts functionality
interface ISCAccounts {
    // Get the L2 base tokens balance of an account
    function getL2BalanceBaseTokens(ISCAgentID memory agentID) external view returns (uint64);

    // Get the L2 native tokens balance of an account
    function getL2BalanceNativeTokens(NativeTokenID memory id, ISCAgentID memory agentID) external view returns (uint256);

    // Get the L2 NFTs owned by an account
    function getL2NFTs(ISCAgentID memory agentID) external view returns (NFTID[] memory);

    // Get the amount of L2 NFTs owned by an account
    function getL2NFTAmount(ISCAgentID memory agentID) external view returns (uint256);

    // Get the L2 NFTs of a given collection owned by an account
    function getL2NFTsInCollection(ISCAgentID memory agentID, NFTID collectionId) external view returns (NFTID[] memory);

    // Get the amount of L2 NFTs of a given collection owned by an account
    function getL2NFTAmountInCollection(ISCAgentID memory agentID, NFTID collectionId) external view returns (uint256);
}

ISCAccounts constant __iscAccounts = ISCAccounts(ISC_MAGIC_ADDRESS);
