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
	IotaAddress iotaAddress;
	ISCHname hname;
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

struct IotaFungibleTokens {
	uint64 iotas;
	IotaNativeToken[] tokens;
}

struct IotaAllowance {
	IotaFungibleTokens assets;
	IotaNFTID[] nfts;
}

struct ISCSendMetadata  {
	ISCHname targetContract;
	ISCHname entrypoint;
	//mapping(string => bytes) params;
	IotaAllowance allowance;
	uint64 gasBudget;
}

struct ISCTimeData {
	uint32 milestoneIndex;
	int64 time;
}

struct ISCExpiration {
	uint32 milestoneIndex;
	uint256 time;
	IotaAddress returnAddress;
}

struct ISCSendOptions {
	ISCTimeData timelock;
	ISCExpiration expiration;
}

struct ISCRequestParameters {
	IotaAddress targetAddress;
	IotaFungibleTokens fungibleTokens;
	bool adjustMinimumDustDeposit;
	ISCSendMetadata metadata;
	ISCSendOptions sendOptions;
}

type ISCError is uint16;

function test3() {

	ISCRequestParameters memory params;
	params.fungibleTokens.iotas = 1074;
	params.fungibleTokens.tokens = new IotaNativeToken[](1);
	params.fungibleTokens.tokens[0].amount = 1074;
	params.metadata.entrypoint = ISCHname.wrap(0x1337);
	params.metadata.targetContract = ISCHname.wrap(0xd34db33f);
	params.adjustMinimumDustDeposit = true;
}

// The interface of the native ISC contract
interface ISC {
	// ----- SandboxBase -----

	function hasParam(string memory key) external view returns (bool);
	function getParam(string memory key) external view returns (bytes memory);

	function getChainID() external view returns (ISCChainID);
	function getChainOwnerID() external view returns (ISCAgentID memory);
	function getTimestampUnixNano() external view returns (int64);

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
    function getAllowanceNFTsLen() external view returns (uint16);
    function getAllowanceNFTID(uint16 i) external view returns (IotaNFTID);
	function triggerEvent(string memory s) external;
	function getEntropy() external view returns (bytes32);
	function send(ISCRequestParameters memory params) external view;

	function registerError(string memory s) external view returns (ISCError);

	// ----- SandboxView -----

	function callView(ISCHname contractHname, ISCHname entryPoint, ISCDict memory params) external view returns (ISCDict memory);
}

ISC constant isc = ISC(0x0000000000000000000000000000000000001074);
error VMError(ISCError);
error TError();
