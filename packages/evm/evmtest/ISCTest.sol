// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.5;

import "@isccontract/ISC.sol";

contract ISCTest {
    ISCError TestError = isc.registerError("TestError");

    function getChainID() public view returns (ISCChainID) {
        return isc.getChainID();
    }

    function triggerEvent(string memory s) public {
        isc.triggerEvent(s);
    }

    function triggerEventFail(string memory s) public {
        isc.triggerEvent(s);
        revert();
    }

    event EntropyEvent(bytes32 entropy);

    function emitEntropy() public {
        bytes32 e = isc.getEntropy();
        emit EntropyEvent(e);
    }

    event RequestIDEvent(ISCRequestID reqID);

    function emitRequestID() public {
        ISCRequestID memory reqID = isc.getRequestID();
        emit RequestIDEvent(reqID);
    }

    event GetCallerEvent(ISCAgentID agentID);

    function emitGetCaller() public {
        ISCAgentID memory agentID = isc.getCaller();
        emit GetCallerEvent(agentID);
    }

    event SenderAccountEvent(ISCAgentID sender);

    function emitSenderAccount() public {
        ISCAgentID memory sender = isc.getSenderAccount();
        emit SenderAccountEvent(sender);
    }

    function send(L1Address memory receiver) public {
        ISCDict memory params = ISCDict(new ISCDictItem[](1));
        bytes memory int64Encoded42 = hex"2A00000000000000";
        params.items[0] = ISCDictItem("x", int64Encoded42);

        bytes memory emptyID = new bytes(38);
        NativeTokenID memory tokenId;
        tokenId.data = emptyID;

        ISCFungibleTokens memory fungibleTokens;
        fungibleTokens.baseTokens = 1074;

        ISCSendMetadata memory metadata;
        metadata.entrypoint = ISCHname.wrap(0x1337);
        metadata.targetContract = ISCHname.wrap(0xd34db33f);
        metadata.params = params;

        ISCSendOptions memory options;

        isc.send(receiver, fungibleTokens, true, metadata, options);
    }

    function revertWithVMError() public view {
        revert VMError(TestError);
    }

    event AllowanceBaseTokensEvent(uint64 baseTokens);

    function emitAllowanceBaseTokens() public {
        emit AllowanceBaseTokensEvent(isc.getAllowanceBaseTokens());
    }

    event AllowanceNativeTokenEvent(NativeToken token);

    function emitAllowanceNativeTokens() public {
        uint16 n = isc.getAllowanceNativeTokensLen();
        for (uint16 i = 0; i < n; i++) {
            emit AllowanceNativeTokenEvent(isc.getAllowanceNativeToken(i));
        }
    }

    event AllowanceAvailableBaseTokensEvent(uint64 baseTokens);

    function emitAllowanceAvailableBaseTokens() public {
        emit AllowanceAvailableBaseTokensEvent(
            isc.getAllowanceAvailableBaseTokens()
        );
    }

    event AllowanceAvailableNativeTokenEvent(NativeToken token);

    function emitAllowanceAvailableNativeTokens() public {
        uint16 n = isc.getAllowanceAvailableNativeTokensLen();
        for (uint16 i = 0; i < n; i++) {
            emit AllowanceAvailableNativeTokenEvent(
                isc.getAllowanceAvailableNativeToken(i)
            );
        }
    }

    event AllowanceNFTEvent(ISCNFT nft);

    function emitAllowanceNFTs() public {
        uint16 n = isc.getAllowanceNFTsLen();
        for (uint16 i = 0; i < n; i++) {
            emit AllowanceNFTEvent(isc.getAllowanceNFT(i));
        }
    }

    event AllowanceAvailableNFTEvent(ISCNFT nft);

    function emitAllowanceAvailableNFTs() public {
        uint16 n = isc.getAllowanceAvailableNFTsLen();
        for (uint16 i = 0; i < n; i++) {
            emit AllowanceAvailableNFTEvent(isc.getAllowanceAvailableNFT(i));
        }
    }

    function callInccounter() public {
        ISCDict memory params = ISCDict(new ISCDictItem[](1));
        bytes memory int64Encoded42 = hex"2A00000000000000";
        params.items[0] = ISCDictItem("counter", int64Encoded42);
        ISCAllowance memory allowance;
        isc.call(isc.hn("inccounter"), isc.hn("incCounter"), params, allowance);
    }

    function callSendAsNFT(L1Address memory receiver, NFTID id) public {
        ISCFungibleTokens memory fungibleTokens;
        fungibleTokens.baseTokens = 1074;
        fungibleTokens.tokens = new NativeToken[](0);

        ISCSendMetadata memory metadata;
        metadata.entrypoint = ISCHname.wrap(0x1337);
        metadata.targetContract = ISCHname.wrap(0xd34db33f);

        ISCDict memory optParams = ISCDict(new ISCDictItem[](1));
        bytes memory int64Encoded42 = hex"2A00000000000000";
        optParams.items[0] = ISCDictItem("x", int64Encoded42);
        metadata.params = optParams;

        ISCSendOptions memory options;

        isc.sendAsNFT(receiver, fungibleTokens, true, metadata, options, id);
    }
}
