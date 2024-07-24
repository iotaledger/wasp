// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";

/**
 * @title ISCAccounts
 * @dev Functions of the ISC Magic Contract to access the core accounts functionality
 */
interface ISCAccounts {
    /**
     * @dev This function retrieves the balance of L2 base tokens for a given account.
     * @param agentID The ID of the agent (account) whose balance is to be retrieved
     * @return The L2 base tokens balance of the specified account
     */
    function getL2BalanceBaseTokens(ISCAgentID memory agentID) external view returns (uint64);

    /**
     * @dev This function retrieves the balance of L2 native tokens for a given account.
     * @param id The ID of the native token
     * @param agentID The ID of the agent (account) whose balance is to be retrieved
     * @return The L2 native tokens balance of the specified account
     */
    function getL2BalanceNativeTokens(NativeTokenID memory id, ISCAgentID memory agentID) external view returns (uint256);

    /**
     * @dev This function retrieves the IDs of NFTs owned by a given account.
     * @param agentID The ID of the agent (account) whose NFTs are to be retrieved
     * @return An array of NFTIDs representing the NFTs owned by the specified account
     */
    function getL2NFTs(ISCAgentID memory agentID) external view returns (NFTID[] memory);

    /**
     * @dev This function retrieves the number of NFTs owned by a given account.
     * @param agentID The ID of the agent (account) whose NFT amount is to be retrieved
     * @return The amount of L2 NFTs owned by the specified account
     */
    function getL2NFTAmount(ISCAgentID memory agentID) external view returns (uint256);

    /**
     * @dev This function retrieves the NFTs of a specific collection owned by a given account.
     * @param agentID The ID of the agent (account) whose NFTs are to be retrieved
     * @param collectionId The ID of the NFT collection
     * @return An array of NFTIDs representing the NFTs in the specified collection owned by the account
     */
    function getL2NFTsInCollection(ISCAgentID memory agentID, NFTID collectionId) external view returns (NFTID[] memory);

    /**
     * @dev This function retrieves the number of NFTs in a specific collection owned by a given account.
     * @param agentID The ID of the agent (account) whose NFT amount is to be retrieved
     * @param collectionId The ID of the NFT collection
     * @return The amount of L2 NFTs in the specified collection owned by the account
     */
    function getL2NFTAmountInCollection(ISCAgentID memory agentID, NFTID collectionId) external view returns (uint256);

    /**
     * @dev This function allows the creation of a new foundry with a specified token scheme and asset allowance.
     * @param tokenScheme The token scheme for the new foundry
     * @param allowance The assets to be allowed for the foundry creation
     * @return The serial number of the newly created foundry
     */
    function foundryCreateNew(NativeTokenScheme memory tokenScheme, ISCAssets memory allowance) external returns(uint32);

    /**
     * @dev This function allows the creation of a new native token foundry along with its IRC30 metadata and ERC20 token registration.
     * @param tokenName The name of the new token
     * @param tokenSymbol The symbol of the new token
     * @param tokenDecimals The number of decimals for the new token
     * @param tokenScheme The token scheme for the new foundry
     * @param allowance The assets to be allowed for the foundry creation
     * @return The serial number of the newly created foundry
     */
    function createNativeTokenFoundry(string memory tokenName, string memory tokenSymbol, uint8 tokenDecimals, NativeTokenScheme memory tokenScheme, ISCAssets memory allowance) external returns(uint32);

    /**
     * @dev This function allows the owner of a foundry to mint new native tokens.
     * @param foundrySN The serial number of the foundry
     * @param amount The amount of tokens to mint
     * @param allowance The assets to be allowed for the minting process
     */
    function mintNativeTokens(uint32 foundrySN, uint256 amount, ISCAssets memory allowance) external;
}

ISCAccounts constant __iscAccounts = ISCAccounts(ISC_MAGIC_ADDRESS);
