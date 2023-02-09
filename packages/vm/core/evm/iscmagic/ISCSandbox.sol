// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISCTypes.sol";

// The main interface of the ISC Magic Contract
interface ISCSandbox {
    // Get the ISC request ID
    function getRequestID() external returns (ISCRequestID memory);

    // Get the AgentID of the sender of the ISC request
    function getSenderAccount() external returns (ISCAgentID memory);

    // Trigger an ISC event
    function triggerEvent(string memory s) external;

    // Get a random 32-bit value based on the hash of the current ISC state transaction
    function getEntropy() external returns (bytes32);

    // Allow the `target` EVM contract to take some funds from the caller's L2 account
    function allow(address target, ISCAssets memory allowance) external;

    // Take some funds from the given address, which must have authorized first with `allow`.
    // If `allowance` is empty, all allowed funds are taken.
    function takeAllowedFunds(address addr, ISCAssets memory allowance)
        external;

    // Get the amount of funds currently allowed by the given address to the caller
    function getAllowanceFrom(address addr)
        external
        view
        returns (ISCAssets memory);

    // Get the amount of funds currently allowed by the caller to the given address
    function getAllowanceTo(address target)
        external
        view
        returns (ISCAssets memory);

    // Get the amount of funds currently allowed between the given addresses
    function getAllowance(address from, address to)
        external
        view
        returns (ISCAssets memory);

    // Send an on-ledger request (or a regular transaction to any L1 address).
    // The specified `assets` are transferred from the caller's
    // L2 account to the `evm` core contract's account.
    // The sent request will have the `evm` core contract as sender. It will
    // include the transferred `assets`.
    // The specified `allowance` must not be greater than `assets`.
    function send(
        L1Address memory targetAddress,
        ISCAssets memory assets,
        bool adjustMinimumStorageDeposit,
        ISCSendMetadata memory metadata,
        ISCSendOptions memory sendOptions
    ) external;

    // Call the entry point of an ISC contract on the same chain.
    // The specified funds in the allowance are taken from the caller's L2 account.
    function call(
        ISCHname contractHname,
        ISCHname entryPoint,
        ISCDict memory params,
        ISCAssets memory allowance
    ) external returns (ISCDict memory);

    // Call a view entry point of an ISC contract on the same chain.
    function callView(
        ISCHname contractHname,
        ISCHname entryPoint,
        ISCDict memory params
    ) external view returns (ISCDict memory);

    // Get the ChainID of the underlying ISC chain
    function getChainID() external view returns (ISCChainID);

    // Get the ISC chain's owner
    function getChainOwnerID() external view returns (ISCAgentID memory);

    // Get the timestamp of the ISC block (seconds since UNIX epoch)
    function getTimestampUnixSeconds() external view returns (int64);

    // Get the properties of the ISC base token
    function getBaseTokenProperties()
        external
        view
        returns (ISCTokenProperties memory);

    // Get the ID of a L2-controlled native token, given its foundry serial number
    function getNativeTokenID(uint32 foundrySN)
        external
        view
        returns (NativeTokenID memory);

    // Get the token scheme of a L2-controlled native token, given its foundry serial number
    function getNativeTokenScheme(uint32 foundrySN)
        external
        view
        returns (NativeTokenScheme memory);

    // Get information about an on-chain NFT
    function getNFTData(NFTID id) external view returns (ISCNFT memory);

    // Get information about an on-chain IRC27 NFT
    function getIRC27NFTData(NFTID id) external view returns (IRC27NFT memory);

    // Get the address of an ERC20NativeTokens contract for the given foundry serial number
    function erc20NativeTokensAddress(uint32 foundrySN)
        external
        view
        returns (address);

    // Get the address of an ERC721NFTCollection contract for the given collection ID
    function erc721NFTCollectionAddress(NFTID collectionID)
        external
        view
        returns (address);

    // Extract the foundry serial number from an ERC20NativeTokens contract's address
    function erc20NativeTokensFoundrySerialNumber(address addr)
        external
        view
        returns (uint32);
}

ISCSandbox constant __iscSandbox = ISCSandbox(ISC_MAGIC_ADDRESS);
