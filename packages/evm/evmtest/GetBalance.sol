// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

import "@iscmagic/ISC.sol";

contract GetBalance {
    event GotAgentID(bytes agentID);
    event GotBaseBalance(uint64 baseBalance);
    event GotNativeTokenBalance(uint256 nativeTokenBalance);
    event GotNFTIDs(uint256 nftBalance);

    function getBalance(bytes memory nativeTokenID) public {
        ISCAgentID memory agentID = ISC.sandbox.getSenderAccount();
        emit GotAgentID(agentID.data);
        
        uint64 baseBalance = ISC.accounts.getL2BalanceBaseTokens(agentID);
        emit GotBaseBalance(baseBalance);

        NativeTokenID memory id = NativeTokenID({ data: nativeTokenID});
        uint256 nativeTokens = ISC.accounts.getL2BalanceNativeTokens(id, agentID);
        emit GotNativeTokenBalance(nativeTokens);

        uint256 nfts = ISC.accounts.getL2NFTAmount(agentID);
        emit GotNFTIDs(nfts);
    }
}