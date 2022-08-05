// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

// The interface of the ISC Magic Contract
interface ISC {
    // Get the AgentID of the caller
	function getCaller() external view returns (ISCAgentID memory);

    // Get the ISC request ID
	function getRequestID() external view returns (ISCRequestID memory);

    // Get the AgentID of the sender of the ISC request
	function getSenderAccount() external view returns (ISCAgentID memory);

    // Get the base tokens specified in the allowance
	function getAllowanceBaseTokens() external view returns (uint64);

    // Get the amount of native token IDs specified in the allowance
	function getAllowanceNativeTokensLen() external view returns (uint16);

    // Get the amount of native tokens at position i specified in the allowance
	function getAllowanceNativeToken(uint16 i) external view returns (NativeToken memory);

    // Get the remaining base tokens in the allowance
	function getAllowanceAvailableBaseTokens() external view returns (uint64);

    // Get the amount of native token IDs in the remaining allowance
	function getAllowanceAvailableNativeToken(uint16 i) external view returns (NativeToken memory);

    // Get the remaining native tokens at position i in the allowance
	function getAllowanceAvailableNativeTokensLen() external view returns (uint16);

    // Get the amount of NFTs specified in the allowance
	function getAllowanceNFTsLen() external view returns (uint16);

    // Get the NFT at position i specified in the allowance
	function getAllowanceNFT(uint16 i) external view returns (ISCNFT memory);

    // Get the amount of NFTs in the remaining allowance
	function getAllowanceAvailableNFTsLen() external view returns (uint16);

    // Get the NFT at position i in the remaining allowance
	function getAllowanceAvailableNFT(uint16 i) external view returns (ISCNFT memory);

    // Trigger an ISC event
	function triggerEvent(string memory s) external;

    // Get a random 32-bit value based on the hash of the current ISC state transaction
	function getEntropy() external view returns (bytes32);

    // Send an on-ledger request (or a regular transaction to any L1 address).
    // The specified funds are taken from the ISC request caller's L2 account.
	function send(L1Address memory targetAddress, ISCFungibleTokens memory fungibleTokens, bool adjustMinimumStorageDeposit, ISCSendMetadata memory metadata, ISCSendOptions memory sendOptions) external;

    // Send an on-ledger request as an NFTOutput
	function sendAsNFT(L1Address memory targetAddress, ISCFungibleTokens memory fungibleTokens, bool adjustMinimumStorageDeposit, ISCSendMetadata memory metadata, ISCSendOptions memory sendOptions, NFTID id) external;

    // Register a custom ISC error message
    //
    // Usage example:
    //
    //   ISCError TestError = isc.registerError("TestError");
    //   function revertWithVMError() public view {
    //       revert VMError(TestError);
    //   }
	function registerError(string memory s) external view returns (ISCError);

    // Call the entry point of an ISC contract on the same chain.
    // The specified funds in the allowance are taken from the ISC request caller's L2 account.
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

	// Like revert() but with a custom message that is saved in the ISC receipt
	function logPanic(string memory s) external view;

	// Log a message with INFO level.
    // Use for debugging purposes only -- the message shows up only on the wasp node's / Solo log.
	function logInfo(string memory s) external view;

	// Log a message with DEBUG level.
    // Use for debugging purposes only -- the message shows up only on the wasp node's / Solo log.
	function logDebug(string memory s) external view;
}

// Every ISC chain is initialized with an instance of the Magic contract at address 0x1074
ISC constant isc = ISC(0x0000000000000000000000000000000000001074);

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

// The allowance of an ISC request
struct ISCAllowance {
	uint64 baseTokens;
	NativeToken[] assets;
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

// An ISC error code.
// See [registerError].
type ISCError is uint16;

// When reverting with VMError, the ISC receipt will include the given ISC error code.
// See [registerError].
error VMError(ISCError);
