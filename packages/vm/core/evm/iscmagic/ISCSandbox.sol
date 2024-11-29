// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";

/**
 * @title ISCSandbox
 * @dev This is the main interface of the ISC Magic Contract.
 */
interface ISCSandbox {
    // Get the ISC request ID
    function getRequestID() external view returns (ISCRequestID);

    // Get the AgentID of the sender of the ISC request
    function getSenderAccount() external view returns (ISCAgentID memory);

    // Trigger an ISC event
    function triggerEvent(string memory s) external;

    // Get a random 32-bit value based on the hash of the current ISC state transaction
    function getEntropy() external view returns (bytes32);

    // Allow the `target` EVM contract to take some funds from the caller's L2 account
    function allow(address target, ISCAssets memory allowance) external;

    // Take some funds from the given address, which must have authorized first with `allow`.
    // If `allowance` is empty, all allowed funds are taken.
    function takeAllowedFunds(address addr, ISCAssets memory allowance) external;

    // Get the amount of funds currently allowed by the given address to the caller
    function getAllowanceFrom(address addr) external view returns (ISCAssets memory);

    // Get the amount of funds currently allowed by the caller to the given address
    function getAllowanceTo(address target) external view returns (ISCAssets memory);

    // Get the amount of funds currently allowed between the given addresses
    function getAllowance(address from, address to) external view
        returns (ISCAssets memory);

    // Send an on-ledger request (or a regular transaction to any L1 address).
    // The specified `assets` are transferred from the caller's
    // L2 account to the `evm` core contract's account.
    // The sent request will have the `evm` core contract as sender. It will
    // include the transferred `assets`.
    // The specified `allowance` must not be greater than `assets`.
    function send(
        IotaAddress targetAddress,
        ISCAssets memory assets,
        ISCSendMetadata memory metadata,
        ISCSendOptions memory sendOptions
    ) external payable;

    // Call the entry point of an ISC contract on the same chain.
    function call(
        ISCMessage memory message,
        ISCAssets memory allowance
    ) external returns (bytes[] memory);

    // Call a view entry point of an ISC contract on the same chain.
    // The called entry point will have the `evm` core contract as caller.
    function callView(ISCMessage memory message) external view returns (bytes[] memory);

    // Get the ChainID of the underlying ISC chain
    function getChainID() external view returns (ISCChainID);

    // Get the ISC chain's owner
    function getChainOwnerID() external view returns (ISCAgentID memory);

    // Get the timestamp of the ISC block (seconds since UNIX epoch)
    function getTimestampUnixSeconds() external view returns (int64);

    // Get the properties of the L1 base token
    function getBaseTokenInfo() external view returns (IotaCoinInfo memory);

    // Get the properties of a L2-controlled coin
    function getCoinInfo(string memory coinType) external view returns (IotaCoinInfo memory);

    // Get information about an on-chain object
    function getObjectBCS(IotaObjectID id) external view returns (bytes memory);

    // TODO
    // // Get information about an on-chain IRC27 NFT
    // // NOTE: metadata does not include attributes, use `getIRC27TokenURI` to get those attributes off-chain in JSON form
    // function getIRC27NFTData(IotaObjectID id) external view returns (IRC27NFT memory);

    // Get information about an on-chain IRC27 NFT
    // returns a JSON file encoded with the following format:
    // base64(jsonEncode({
    //   "name": NFT.name,
    //   "description": NFT.description,
    //   "image": NFT.URI
    // }))
    function getIRC27TokenURI(IotaObjectID id) external view returns (string memory);

    // Get the address of an ERC20Coin contract
    function ERC20CoinAddress(string memory coinType) external view returns (address);

    // Get the address of an ERC721NFTCollection contract for the given collection ID
    function erc721NFTCollectionAddress(IotaObjectID collectionID) external view
        returns (address);
}

ISCSandbox constant __iscSandbox = ISCSandbox(ISC_MAGIC_ADDRESS);
