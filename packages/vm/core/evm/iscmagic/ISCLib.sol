// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.11;

import "@iscmagic/ISC.sol";

uint8 constant AgentIDKindEthereumAddress = 3;

library ISCLib {
    // Get the L2 balance of an account
    function getL2BalanceBaseTokens(ISCAgentID memory agentID) internal view returns (uint64) {
        ISCDict memory params;
        params.items = new ISCDictItem[](1);
        params.items[0].key = bytes("a");
        params.items[0].value = agentID.data;

        ISCDict memory r = isc.callView(isc.hn("accounts"), isc.hn("balanceBaseToken"), params);

        for (uint i = 0; i < r.items.length; i++) {
            if (r.items[i].key.length == 1 && r.items[i].key[0] == 'B') {
                return decodeUint64(r.items[i].value);
            }
        }
        revert("something went wrong");
    }

    function bytesToUint(bytes memory b) internal pure returns (uint256){
        uint256 number;
        for(uint i = 0; i < b.length; i++){
            number = number + uint(uint8(b[i])) * (2**(8*i));
        }
        return number;
    }


    // Decode 4 bytes little-endian to uint64
    function decodeUint64(bytes memory b) internal pure returns (uint64) {
        require(b.length == 8, "decodeUint64: expected 8 bytes");
        return uint64(bytesToUint(b));
    }

    function newEthereumAgentID(address addr) internal pure returns (ISCAgentID memory) {
        bytes memory addrBytes = abi.encodePacked(addr);
        ISCAgentID memory r;
        r.data = new bytes(1+addrBytes.length);
        r.data[0] = bytes1(AgentIDKindEthereumAddress);
        for (uint i = 0; i < addrBytes.length; i++) {
            r.data[i+1] = addrBytes[i];
        }
        return r;
    }
}
