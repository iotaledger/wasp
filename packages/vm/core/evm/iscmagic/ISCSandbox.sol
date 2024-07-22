// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";

/**
 * @title ISCSandbox
 * @dev This is the main interface of the ISC Magic Contract.
 */
interface ISCSandbox {
    /**
     * @notice Get the ISC request ID
     * @dev Retrieves the ID of the current ISC request.
     * @return The ISCRequestID of the current request.
     */
    function getRequestID() external view returns (ISCRequestID memory);

    /**
     * @notice Get the AgentID of the sender of the ISC request
     * @dev Retrieves the AgentID of the account that sent the current ISC request.
     * @return The ISCAgentID of the sender.
     */
    function getSenderAccount() external view returns (ISCAgentID memory);

    /**
     * @notice Trigger an ISC event
     * @dev Triggers an event in the ISC system with the given string.
     * @param s The string to include in the event.
     */
    function triggerEvent(string memory s) external;

    /**
     * @notice Get a random 32-bit value based on the hash of the current ISC state transaction
     * @dev Retrieves a random 32-bit value derived from the hash of the current ISC state transaction.
     * @return A random bytes32 value.
     */
    function getEntropy() external view returns (bytes32);

    /**
     * @notice Allow the `target` EVM contract to take some funds from the caller's L2 account
     * @dev Authorizes the specified target address to take the given assets from the caller's account.
     * @param target The address of the target EVM contract.
     * @param allowance The assets to be allowed.
     */
    function allow(address target, ISCAssets memory allowance) external;

    /**
     * @notice Take some funds from the given address, which must have authorized first with `allow`.
     * @dev Takes the specified assets from the given address if they have authorized the caller. If the allowance is empty, all allowed funds are taken.
     * @param addr The address to take funds from.
     * @param allowance The assets to take.
     */
    function takeAllowedFunds(
        address addr,
        ISCAssets memory allowance
    ) external;

    /**
     * @notice Get the amount of funds currently allowed by the given address to the caller
     * @dev Retrieves the amount of assets the specified address has allowed the caller to take.
     * @param addr The address that has allowed funds.
     * @return The allowed ISCAssets.
     */
    function getAllowanceFrom(
        address addr
    ) external view returns (ISCAssets memory);

    /**
     * @notice Get the amount of funds currently allowed by the caller to the given address
     * @dev Retrieves the amount of assets the caller has allowed the specified address to take.
     * @param target The address allowed to take funds.
     * @return The allowed ISCAssets.
     */
    function getAllowanceTo(
        address target
    ) external view returns (ISCAssets memory);

    /**
     * @notice Get the amount of funds currently allowed between the given addresses
     * @dev Retrieves the amount of assets allowed between the specified addresses.
     * @param from The address that has allowed funds.
     * @param to The address allowed to take funds.
     * @return The allowed ISCAssets.
     */
    function getAllowance(
        address from,
        address to
    ) external view returns (ISCAssets memory);

    /**
     * @notice Send an on-ledger request (or a regular transaction to any L1 address).
     * @dev Sends the specified assets from the caller's L2 account to the EVM core contract's account and includes the specified metadata and options.
     * @param targetAddress The L1 address to send the assets to.
     * @param assets The assets to be sent.
     * @param adjustMinimumStorageDeposit Whether to adjust the minimum storage deposit.
     * @param metadata The metadata to include in the request.
     * @param sendOptions The options for the send operation.
     */
    function send(
        L1Address memory targetAddress,
        ISCAssets memory assets,
        bool adjustMinimumStorageDeposit,
        ISCSendMetadata memory metadata,
        ISCSendOptions memory sendOptions
    ) external payable;

    /**
     * @notice Call the entry point of an ISC contract on the same chain.
     * @dev Calls the specified entry point of the ISC contract with the given parameters and allowance.
     * @param contractHname The hash name of the contract.
     * @param entryPoint The entry point to be called.
     * @param params The parameters to pass to the entry point.
     * @param allowance The assets to be allowed for the call.
     * @return The return data from the ISC contract call.
     */
    function call(
        ISCHname contractHname,
        ISCHname entryPoint,
        ISCDict memory params,
        ISCAssets memory allowance
    ) external returns (ISCDict memory);

    /**
     * @notice Call a view entry point of an ISC contract on the same chain.
     * @dev Calls the specified view entry point of the ISC contract with the given parameters.
     * @param contractHname The hash name of the contract.
     * @param entryPoint The view entry point to be called.
     * @param params The parameters to pass to the view entry point.
     * @return The return data from the ISC contract view call.
     */
    function callView(
        ISCHname contractHname,
        ISCHname entryPoint,
        ISCDict memory params
    ) external view returns (ISCDict memory);

    /**
     * @notice Get the ChainID of the underlying ISC chain
     * @dev Retrieves the ChainID of the current ISC chain.
     * @return The ISCChainID of the current chain.
     */
    function getChainID() external view returns (ISCChainID);

    /**
     * @notice Get the ISC chain's owner
     * @dev Retrieves the AgentID of the owner of the current ISC chain.
     * @return The ISCAgentID of the chain owner.
     */
    function getChainOwnerID() external view returns (ISCAgentID memory);

    /**
     * @notice Get the timestamp of the ISC block (seconds since UNIX epoch)
     * @dev Retrieves the timestamp of the current ISC block in seconds since the UNIX epoch.
     * @return The timestamp of the current block.
     */
    function getTimestampUnixSeconds() external view returns (int64);

    /**
     * @notice Get the properties of the ISC base token
     * @dev Retrieves the properties of the base token used in the ISC system.
     * @return The ISCTokenProperties of the base token.
     */
    function getBaseTokenProperties()
        external
        view
        returns (ISCTokenProperties memory);

    /**
     * @notice Get the ID of a L2-controlled native token, given its foundry serial number
     * @dev Retrieves the NativeTokenID of a native token based on its foundry serial number.
     * @param foundrySN The serial number of the foundry.
     * @return The NativeTokenID of the specified native token.
     */
    function getNativeTokenID(
        uint32 foundrySN
    ) external view returns (NativeTokenID memory);

    /**
     * @notice Get the token scheme of a L2-controlled native token, given its foundry serial number
     * @dev Retrieves the NativeTokenScheme of a native token based on its foundry serial number.
     * @param foundrySN The serial number of the foundry.
     * @return The NativeTokenScheme of the specified native token.
     */
    function getNativeTokenScheme(
        uint32 foundrySN
    ) external view returns (NativeTokenScheme memory);

    /**
     * @notice Get information about an on-chain NFT
     * @dev Retrieves the details of an NFT based on its ID.
     * @param id The ID of the NFT.
     * @return The ISCNFT data of the specified NFT.
     */
    function getNFTData(NFTID id) external view returns (ISCNFT memory);

    /**
     * @notice Get information about an on-chain IRC27 NFT
     * @dev Retrieves the details of an IRC27 NFT based on its ID.
     * @param id The ID of the IRC27 NFT.
     * @return The IRC27NFT data of the specified NFT.
     */
    // Note: the metadata.uri field is encoded as a data URL with:
    // base64(jsonEncode({
    //   "name": NFT.name,
    //   "description": NFT.description,
    //   "image": NFT.URI
    // }))
    function getIRC27NFTData(NFTID id) external view returns (IRC27NFT memory);

    /**
     * @notice Get the URI of an on-chain IRC27 NFT
     * @dev Retrieves the URI of an IRC27 NFT based on its ID.
     * @param id The ID of the IRC27 NFT.
     * @return The URI of the specified IRC27 NFT.
     */
    // returns a JSON file encoded with the following format:
    // base64(jsonEncode({
    //   "name": NFT.name,
    //   "description": NFT.description,
    //   "image": NFT.URI
    // }))
    function getIRC27TokenURI(NFTID id) external view returns (string memory);

    /**
     * @notice Get the address of an ERC20NativeTokens contract for the given foundry serial number
     * @dev Retrieves the address of an ERC20NativeTokens contract based on the foundry serial number.
     * @param foundrySN The serial number of the foundry.
     * @return The address of the specified ERC20NativeTokens contract.
     */
    function erc20NativeTokensAddress(
        uint32 foundrySN
    ) external view returns (address);

    /**
     * @notice Get the address of an ERC721NFTCollection contract for the given collection ID
     * @dev Retrieves the address of an ERC721NFTCollection contract based on the collection ID.
     * @param collectionID The ID of the NFT collection.
     * @return The address of the specified ERC721NFTCollection contract.
     */
    function erc721NFTCollectionAddress(
        NFTID collectionID
    ) external view returns (address);

    /**
     * @notice Extract the foundry serial number from an ERC20NativeTokens contract's address
     * @dev Retrieves the foundry serial number from the address of an ERC20NativeTokens contract.
     * @param addr The address of the ERC20NativeTokens contract.
     * @return The foundry serial number.
     */
    function erc20NativeTokensFoundrySerialNumber(
        address addr
    ) external view returns (uint32);

    /**
     * @notice Creates an ERC20NativeTokens contract instance and register it with the foundry as a native token. Only the foundry owner can call this function.
     * @dev Registers a new ERC20NativeTokens contract with the specified foundry and token details. Only callable by the foundry owner.
     * @param foundrySN The serial number of the foundry.
     * @param name The name of the new token.
     * @param symbol The symbol of the new token.
     * @param decimals The decimals of the new token.
     * @param allowance The assets to be allowed for the registration.
     */
    function registerERC20NativeToken(
        uint32 foundrySN,
        string memory name,
        string memory symbol,
        uint8 decimals,
        ISCAssets memory allowance
    ) external;
}

ISCSandbox constant __iscSandbox = ISCSandbox(ISC_MAGIC_ADDRESS);
