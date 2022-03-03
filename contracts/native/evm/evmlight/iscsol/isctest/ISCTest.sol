// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.5;

import "@isccontract/ISC.sol";

contract ISCTest {
    function getChainID() public view returns (ISCChainID) {
		return isc.getChainID();
    }

	function triggerEvent(string memory s) public {
		isc.triggerEvent(s);
	}

	event EntropyEvent(bytes32 entropy);

	function emitEntropy() public {
		bytes32 e = isc.getEntropy();
		emit EntropyEvent(e);
	}
}
