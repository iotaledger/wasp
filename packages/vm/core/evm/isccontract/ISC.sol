// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

type ISCChainID is bytes20;

struct IotaAddress {
	bytes data;
}

type ISCHname is uint32;

struct ISCAgentID {
	IotaAddress iotaAddress;
	ISCHname hname;
}

type IotaNFTID is bytes20;

struct ISCNFT {
	IotaNFTID ID;
	IotaAddress issuer;
	bytes metadata;
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

	function triggerEvent(string memory s) external;
	function getEntropy() external returns (bytes32);
}

ISC constant isc = ISC(0x0000000000000000000000000000000000001074);
