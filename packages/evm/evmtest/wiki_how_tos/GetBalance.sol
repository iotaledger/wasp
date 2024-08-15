// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

import "@iscmagic/ISC.sol";

contract GetBalance {
    event GotAgentID(bytes agentID);
    event GotBaseBalance(uint64 baseBalance);
    event GotCoinBalance(uint64 nativeTokenBalance);
    event GotNFTIDs(uint256 nftBalance);

    function getBalanceBaseTokens() public {
        ISCAgentID memory agentID = ISC.sandbox.getSenderAccount();
        uint64 baseBalance = ISC.accounts.getL2BalanceBaseTokens(agentID);
        emit GotBaseBalance(baseBalance);
    }

    function getBalanceCoin(string memory coinType) public {
        ISCAgentID memory agentID = ISC.sandbox.getSenderAccount();
        emit GotCoinBalance(ISC.accounts.getL2BalanceCoin(
            coinType,
            agentID
        ));
    }

    function getBalanceNFTs() public {
        ISCAgentID memory agentID = ISC.sandbox.getSenderAccount();
        uint256 nfts = ISC.accounts.getL2ObjectsCount(agentID);
        emit GotNFTIDs(nfts);
    }

    function getAgentID() public {
        ISCAgentID memory agentID = ISC.sandbox.getSenderAccount();
        emit GotAgentID(agentID.data);
    }
}
