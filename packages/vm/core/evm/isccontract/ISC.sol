// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

// Assuming that ChainID is AliasID which is 20 bytes
type ISCChainID is bytes20;

// The interface of the native ISC contract
interface ISC {
	function triggerEvent(string memory s) external;
	function getEntropy() external returns (bytes32);
	function getChainID() external view returns (ISCChainID);
}

ISC constant isc = ISC(0x0000000000000000000000000000000000001074);
