// SPDX-License-Identifier: MIT

pragma solidity >=0.8.5;

import "@iscmagic/ISC.sol";

contract Entropy {
    event EntropyEvent(bytes32 entropy);

    // this will emit a "random" value taken from the ISC entropy value
    function emitEntropy() public {
        bytes32 e = ISC.sandbox.getEntropy();
        emit EntropyEvent(e);
    }
}