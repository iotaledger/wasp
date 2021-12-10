// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.5;

// ISCP addresses are 33 bytes which is sadly larger than EVM's bytes32 type, so
// it will use two 32-byte slots.
struct ISCPAddress {
	bytes1  typeId;
	bytes32 digest;
}

address constant ISCP_CONTRACT_ADDRESS = 0x0000000000000000000000000000000000001074;
address constant ISCP_YUL_ADDRESS      = 0x0000000000000000000000000000000000001075;

// The standard ISCP contract present in all EVM ISCP chains at ISCP_CONTRACT_ADDRESS
contract ISCP {

	// The ChainID of the underlying ISCP chain
    ISCPAddress chainId;

	function getChainId() public view returns (ISCPAddress memory) {
		return chainId;
	}
}

function iscpTriggerEvent(string memory s) {
	(bool success, ) = ISCP_YUL_ADDRESS.delegatecall(abi.encodeWithSignature("triggerEvent(string)", s));
	assert(success);
}

function iscpEntropy() returns (bytes32) {
	(bool success, bytes memory result) = ISCP_YUL_ADDRESS.delegatecall(abi.encodeWithSignature("entropy()"));
	assert(success);
    return abi.decode(result, (bytes32));
}
