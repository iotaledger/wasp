// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "./ISCTypes.sol";
import "./ISCTypesLegacy.sol";

/**
 * @title ISCSandbox
 * @notice This is the main interface of the ISC Magic Contract.
 */
interface ISCSandbox {
    /**
     * @notice Retrieves the unique identifier of the current ISC request.
     * @return The ISCRequestID representing the unique identifier of the request.
     */
    function getRequestID() external view returns (ISCRequestID);

    /**
     * @notice Retrieves the sender Agent ID.
     * @return The Agent ID of the sender.
     */
    function getSenderAccount() external view returns (ISCAgentID memory);

    /**
     * @notice Triggers an event with the given string parameter.
     * @param s The string parameter that will be used to trigger the event.
     */
    function triggerEvent(string memory s) external;

    /**
     * @notice Retrieves a 32-byte entropy value.
     * @return A 32-byte value representing the entropy.
     */
    function getEntropy() external view returns (bytes32);

    /**
     * @notice Grants permission to the specified address (`target`) to use the provided allowance of ISC assets.
     * @param target The address to which the allowance is granted.
     * @param allowance The ISCAssets structure specifying the assets and amounts being allowed.
     */
    function allow(address target, ISCAssets memory allowance) external;

    /**
     * @notice Takes funds from the specified address, provided the address has authorized the transfer using `allow`.
     * @dev If the `allowance` parameter is empty, all funds that have been authorized will be transferred.
     * @param addr The address from which the funds will be taken.
     * @param allowance The specific amount and types of assets to transfer. If empty, all authorized funds are taken.
     */
    function takeAllowedFunds(
        address addr,
        ISCAssets memory allowance
    ) external;

    /**
     * @notice Retrieves the allowance of ISC assets for a specific address.
     * @param addr The address for which the allowance is being queried.
     * @return The allowance of ISC assets for the specified address.
     */
    function getAllowanceFrom(
        address addr
    ) external view returns (ISCAssets memory);

    /**
     * @notice Retrieves the allowance of ISC assets granted to a specific target address.
     * @param target The address for which the allowance is being queried.
     * @return The allowance granted to the target address.
     */
    function getAllowanceTo(
        address target
    ) external view returns (ISCAssets memory);

    // Get the amount of funds currently allowed between the given addresses
    /**
     * @notice Retrieves the allowance of ISC assets that `from` has approved for `to`.
     * @param from The address of the account that has approved the allowance.
     * @param to The address of the account that is allowed to spend the assets.
     * @return The approved allowance.
     */
    function getAllowance(
        address from,
        address to
    ) external view returns (ISCAssets memory);

    /**
     * @notice Sends assets to any L1 address.
     * @param targetAddress The target L1 address to which the assets are sent.
     * @param assets The assets to be transferred.
     */
    function transferToL1(
        IotaAddress targetAddress,
        ISCAssets memory assets
    ) external payable;

    /**
     * @notice DEPRECATED: This function is kept only to support tracing legacy transactions.
     * @dev Sends the specified assets from the caller's L2 account to a L1 address and includes
     *      the specified metadata and options. This can also be used to create on-ledger requests to the chain itself.
     * @param targetAddress The L1 address to send the assets to.
     * @param assets The assets to be sent.
     * @param adjustMinimumStorageDeposit Whether to adjust the minimum storage deposit.
     * @param metadata The metadata to include in the request.
     * @param sendOptions The options for the send operation.
     */
    function send(
        LegacyL1Address memory targetAddress,
        LegacyISCAssets memory assets,
        bool adjustMinimumStorageDeposit,
        LegacyISCSendMetadata memory metadata,
        LegacyISCSendOptions memory sendOptions
    ) external payable;

    /**
     * @notice Calls the entry point of an Core contract on the same chain.
     * @param message The details of the message to be sent to the Core contract.
     * @param allowance The assets allowed to be used for this call.
     * @return An array of bytes containing the results returned by the Core contract.
     */
    function call(
        ISCMessage memory message,
        ISCAssets memory allowance
    ) external returns (bytes[] memory);

    // Call a view entry point of an ISC contract on the same chain.
    // The called entry point will have the `evm` core contract as caller.
    /**
     * @notice Executes a view call on the specified Core contract.
     * @param message The details of the view call.
     * @return An array of bytes representing the result of the view call.
     */
    function callView(
        ISCMessage memory message
    ) external view returns (bytes[] memory);

    /**
     * @notice Retrieves the ISC Chain ID (Object ID) associated with the current environment.
     * @return The ISCChainID representing the chain ID of the current environment.
     */
    function getChainID() external view returns (ISCChainID);

    /**
     * @notice Retrieves the Agent ID of the ISC chain admin.
     * @return The ISC chain admin.
     */
    function getChainAdmin() external view returns (ISCAgentID memory);

    /**
     * @notice Retrieves the timestamp of the ISC block.
     * @dev The timestamp is returned as the number of seconds since the UNIX epoch.
     * @return The ISC block timestamp.
     */
    function getTimestampUnixSeconds() external view returns (int64);

    /**
     * @notice Retrieves the metadata of the L1 base token.
     * @return Metadata of the base token.
     */
    function getBaseTokenInfo() external view returns (IotaCoinInfo memory);

    /**
     * @notice Retrieves the metadata of a L1 coin.
     * @param coinType The type of the coin as a string.
     * @return Metadata of the specified coin.
     */
    function getCoinInfo(
        string memory coinType
    ) external view returns (IotaCoinInfo memory);

    /**
     * @notice Retrieves the address of the ERC20 contract linked to the specified coin type.
     * @param coinType The type of the coin as a string.
     * @return The address of the ERC20 contract corresponding to the specified coin type.
     */
    function ERC20CoinAddress(
        string memory coinType
    ) external view returns (address);

    /**
     * @notice Returns the token balance of the specified address.
     * @param account The address to query the balance of.
     * @return The amount of tokens owned by the account.
     */
    function balanceOf(address account) external view returns (uint256);

    /**
     * @notice Returns the token symbol.
     * @return The token symbol.
     */
    function symbol() external view returns (string memory);

    /**
     * @notice Returns the number of decimals used to get its user representation.
     * @return The number of decimals.
     */
    function decimals() external view returns (uint8);

    /**
     * @notice Returns true if this contract implements the interface defined by interfaceId.
     * @param interfaceId The interface identifier, as specified in ERC-165.
     * @return true if the contract implements interfaceId and interfaceId is not 0xffffffff, false otherwise.
     */
    function supportsInterface(bytes4 interfaceId) external view returns (bool);
}

ISCSandbox constant __iscSandbox = ISCSandbox(ISC_MAGIC_ADDRESS);
