// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

struct IotaAddress {
	bytes data;
}

struct IotaNativeTokenID {
	bytes data;
}

struct IotaNativeToken {
	IotaNativeTokenID ID;
	uint256 amount;
}

type IotaNFTID is bytes20;

type IotaTransactionID is bytes32;

type ISCHname is uint32;

type ISCChainID is bytes20;

struct ISCAgentID {
	bytes data;
}

struct ISCNFT {
	IotaNFTID ID;
	IotaAddress issuer;
	bytes metadata;
}

struct ISCRequestID {
	IotaTransactionID transactionID;
	uint16 transactionOutputIndex;
}

struct ISCDictItem {
	bytes key;
	bytes value;
}

struct ISCDict {
	ISCDictItem[] items;
}

struct ISCFungibleTokens {
	uint64 iotas;
	IotaNativeToken[] tokens;
}

struct ISCSendMetadata  {
	ISCHname targetContract;
	ISCHname entrypoint;
	ISCDict params;
	ISCAllowance allowance;
	uint64 gasBudget;
}

struct ISCTimeData {
	uint32 milestoneIndex;
	int64 time;
}

struct ISCExpiration {
	uint32 milestoneIndex;
	int64 time;
	IotaAddress returnAddress;
}

struct ISCSendOptions {
	ISCTimeData timelock;
	ISCExpiration expiration;
}

struct ISCRequestParameters {
	IotaAddress targetAddress;
	ISCFungibleTokens fungibleTokens;
	bool adjustMinimumDustDeposit;
	ISCSendMetadata metadata;
	ISCSendOptions sendOptions;
}

type ISCError is uint16;

struct ISCAllowance {
	uint64 iotas;
	IotaNativeToken[] assets;
	IotaNFTID[] nfts;
}

// The interface of the native ISC contract
interface ISC {
	// ----- misc -----
	function hn(string memory s) external view returns (ISCHname);

	// ----- SandboxBase -----

	function hasParam(string memory key) external view returns (bool);
	function getParam(string memory key) external view returns (bytes memory);

	function getChainID() external view returns (ISCChainID);
	function getChainOwnerID() external view returns (ISCAgentID memory);
	function getTimestampUnixSeconds() external view returns (int64);

	// these show up only on the wasp node log
	function logInfo(string memory s) external view;
	function logDebug(string memory s) external view;

	// like revert() but with a custom message that is saved in the ISC receipt
	function logPanic(string memory s) external view;

	function getNFTData(IotaNFTID id) external view returns (ISCNFT memory);

	// ----- Sandbox -----

	function getCaller() external view returns (ISCAgentID memory);
	function getRequestID() external view returns (ISCRequestID memory);
	function getSenderAccount() external view returns (ISCAgentID memory);
	function getSenderAddress() external view returns (IotaAddress memory);
	function getAllowanceIotas() external view returns (uint64);
	function getAllowanceNativeTokensLen() external view returns (uint16);
	function getAllowanceNativeToken(uint16 i) external view returns (IotaNativeToken memory);
	function triggerEvent(string memory s) external;
	function getEntropy() external view returns (bytes32);
	function send(IotaAddress memory targetAddress, ISCFungibleTokens memory fungibleTokens, bool adjustMinimumDustDeposit, ISCSendMetadata memory metadata, ISCSendOptions memory sendOptions) external;
        function sendAsNFT(IotaAddress memory targetAddress, ISCFungibleTokens memory fungibleTokens, bool adjustMinimumDustDeposit, ISCSendMetadata memory metadata, ISCSendOptions memory sendOptions, IotaNFTID id) external;
	function registerError(string memory s) external view returns (ISCError);
	function call(ISCHname contractHname, ISCHname entryPoint, ISCDict memory params, ISCAllowance memory allowance) external returns (ISCDict memory);
	function getAllowanceAvailableIotas() external view returns (uint64);
	function getAllowanceAvailableNativeToken(uint16 i) external view returns (IotaNativeToken memory);
	function getAllowanceAvailableNativeTokensLen() external view returns (uint16);
	function getAllowanceNFTsLen() external view returns (uint16);
	function getAllowanceNFT(uint16 i) external view returns (ISCNFT memory);
	function getAllowanceAvailableNFTsLen() external view returns (uint16);
	function getAllowanceAvailableNFT(uint16 i) external view returns (ISCNFT memory);

	// ----- SandboxView -----

	function callView(ISCHname contractHname, ISCHname entryPoint, ISCDict memory params) external view returns (ISCDict memory);
}

ISC constant isc = ISC(0x0000000000000000000000000000000000001074);

error VMError(ISCError);
