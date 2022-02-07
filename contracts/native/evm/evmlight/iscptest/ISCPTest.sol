// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.5;

import "@iscpcontract/ISCP.sol";

ISCP constant iscp = ISCP(ISCP_CONTRACT_ADDRESS);

contract ISCPTest {
    function getChainID() public view returns (ISCPChainID) {
		return iscp.getChainID();
    }

	function triggerEvent(string memory s) public {
		iscpTriggerEvent(s);
	}

	event EntropyEvent(bytes32 entropy);

	function emitEntropy() public {
		bytes32 e = iscpEntropy();
		emit EntropyEvent(e);
	}
}
