// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

// The interface of the ISC Magic Contract
interface ISC {
    // Get the ISC request ID
    function getRequestID() external view returns (ISCRequestID memory);

    // Get the AgentID of the sender of the ISC request
    function getSenderAccount() external view returns (ISCAgentID memory);

    // Trigger an ISC event
    function triggerEvent(string memory s) external;

    // Get a random 32-bit value based on the hash of the current ISC state transaction
    function getEntropy() external view returns (bytes32);

    // Allow the `target` EVM contract to take some funds from the caller's L2 account
    function allow(address target, ISCAllowance memory allowance) external;

    // Take some funds from the given address, which must have authorized first with `allow`.
    // If `allowance` is empty, all allowed funds are taken.
    function takeAllowedFunds(address addr, ISCAllowance memory allowance) external;

    // Get the amount of funds currently allowed by the given address to the caller
    function getAllowanceFrom(address addr) external view returns (ISCAllowance memory);

    // Get the amount of funds currently allowed by the caller to the given address
    function getAllowanceTo(address target) external view returns (ISCAllowance memory);

    // Get the amount of funds currently allowed between the given addresses
    function getAllowance(address from, address to) external view returns (ISCAllowance memory);

    // Send an on-ledger request (or a regular transaction to any L1 address).
    // The specified `fungibleTokens` are transferred from the caller's
    // L2 account to the `evm` core contract's account.
    // The sent request will have the `evm` core contract as sender. It will
    // include the transferred `fungibleTokens`.
    // The specified `allowance` must not be greater than `fungibleTokens`.
    function send(L1Address memory targetAddress, ISCFungibleTokens memory fungibleTokens, bool adjustMinimumStorageDeposit, ISCSendMetadata memory metadata, ISCSendOptions memory sendOptions) external;

    // Send an on-ledger request as an NFTOutput
    // The specified `fungibleTokens` and NFT `id` are transferred from the caller's
    // L2 account to the `evm` core contract's account.
    // The sent request will have the `evm` core contract as sender. It will
    // include the transferred assets.
    // The specified `allowance` must not be greater than `fungibleTokens`.
    function sendAsNFT(L1Address memory targetAddress, ISCFungibleTokens memory fungibleTokens, NFTID id, bool adjustMinimumStorageDeposit, ISCSendMetadata memory metadata, ISCSendOptions memory sendOptions) external;

    // Call the entry point of an ISC contract on the same chain.
    // The specified funds in the allowance are taken from the caller's L2 account.
    function call(ISCHname contractHname, ISCHname entryPoint, ISCDict memory params, ISCAllowance memory allowance) external returns (ISCDict memory);

    // Call a view entry point of an ISC contract on the same chain.
    function callView(ISCHname contractHname, ISCHname entryPoint, ISCDict memory params) external view returns (ISCDict memory);

    // Get information about an on-chain NFT
    function getNFTData(NFTID id) external view returns (ISCNFT memory);

    // Get the ChainID of the underlying ISC chain
    function getChainID() external view returns (ISCChainID);

    // Get the ISC chain's owner
    function getChainOwnerID() external view returns (ISCAgentID memory);

    // Get the timestamp of the ISC block (seconds since UNIX epoch)
    function getTimestampUnixSeconds() external view returns (int64);

    // Get an ISC contract's hname given its instance name
    function hn(string memory s) external view returns (ISCHname);

    // Get the properties of the ISC base token
    function getBaseTokenProperties() external view returns (ISCTokenProperties memory);

    // Print something to the console (will only work when debugging contracts with Solo)
    function print(string memory s) external pure;
}

// Every ISC chain is initialized with an instance of the Magic contract at this address
ISC constant isc = ISC(0x1074000000000000000000000000000000000000);

// An L1 IOTA address
struct L1Address {
    bytes data;
}

// An IOTA native token ID
struct NativeTokenID {
    bytes data;
}

// An amount of some IOTA native token
struct NativeToken {
    NativeTokenID ID;
    uint256 amount;
}

// An IOTA NFT ID
type NFTID is bytes32;

// Information about an on-chain NFT
struct ISCNFT {
    NFTID ID;
    L1Address issuer;
    bytes metadata;
}

// An ISC transaction ID
type ISCTransactionID is bytes32;

// An ISC hname
type ISCHname is uint32;

// An ISC chain ID
type ISCChainID is bytes32;

// An ISC AgentID
struct ISCAgentID {
    bytes data;
}

// An ISC request ID
struct ISCRequestID {
    ISCTransactionID transactionID;
    uint16 transactionOutputIndex;
}

// A single key-value pair
struct ISCDictItem {
    bytes key;
    bytes value;
}

// Wrapper for the isc.Dict type, a collection of key-value pairs
struct ISCDict {
    ISCDictItem[] items;
}

// A collection of fungible tokens (base tokens + native tokens)
struct ISCFungibleTokens {
    uint64 baseTokens;
    NativeToken[] tokens;
}

// Parameters for building an on-ledger request
struct ISCSendMetadata  {
    ISCHname targetContract;
    ISCHname entrypoint;
    ISCDict params;
    ISCAllowance allowance;
    uint64 gasBudget;
}

// The allowance of an ISC call.
// The specified tokens, assets and NFTs are transferred from the caller's L2 account to
// the callee's L2 account.
struct ISCAllowance {
    uint64 baseTokens;
    NativeToken[] tokens;
    NFTID[] nfts;
}

// Parameters for building an on-ledger request
struct ISCSendOptions {
    int64 timelock;
    ISCExpiration expiration;
}

// Expiration of an on-ledger request
struct ISCExpiration {
    int64 time;
    L1Address returnAddress;
}

// Properties of an ISC base/native token
struct ISCTokenProperties {
    string name;
    string tickerSymbol;
    uint8 decimals;
    uint256 totalSupply;
}
