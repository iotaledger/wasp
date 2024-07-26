// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

import "@iscmagic/ISC.sol";

contract GetBalance {
    event GotAgentID(bytes agentID);
    event GotBaseBalance(uint64 baseBalance);
    event GotNativeTokenBalance(uint256 nativeTokenBalance);
    event GotNFTIDs(uint256 nftBalance);

    function getBalanceBaseTokens() public {
        ISCAgentID memory agentID = ISC.sandbox.getSenderAccount();
        uint64 baseBalance = ISC.accounts.getL2BalanceBaseTokens(agentID);
        emit GotBaseBalance(baseBalance);
    }

    function getBalanceNativeTokens(bytes memory nativeTokenID) public {
        ISCAgentID memory agentID = ISC.sandbox.getSenderAccount();
        NativeTokenID memory id = NativeTokenID({data: nativeTokenID});
        uint256 nativeTokens = ISC.accounts.getL2BalanceNativeTokens(
            id,
            agentID
        );
        emit GotNativeTokenBalance(nativeTokens);
    }

    function getBalanceNFTs() public {
        ISCAgentID memory agentID = ISC.sandbox.getSenderAccount();
        uint256 nfts = ISC.accounts.getL2NFTAmount(agentID);
        emit GotNFTIDs(nfts);
    }

    function getAgentID() public {
        ISCAgentID memory agentID = ISC.sandbox.getSenderAccount();
        emit GotAgentID(agentID.data);
    }
}
