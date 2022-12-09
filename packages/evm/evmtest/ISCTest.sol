// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.5;

import "@iscmagic/ISC.sol";

contract ISCTest {
    uint64 public constant TokensForGas = 500;

    function getChainID() public view returns (ISCChainID) {
        return ISC.sandbox.getChainID();
    }

    function triggerEvent(string memory s) public {
        ISC.sandbox.triggerEvent(s);
    }

    function triggerEventFail(string memory s) public {
        ISC.sandbox.triggerEvent(s);
        revert();
    }

    event EntropyEvent(bytes32 entropy);

    function emitEntropy() public {
        bytes32 e = ISC.sandbox.getEntropy();
        emit EntropyEvent(e);
    }

    event RequestIDEvent(ISCRequestID reqID);

    function emitRequestID() public {
        ISCRequestID memory reqID = ISC.sandbox.getRequestID();
        emit RequestIDEvent(reqID);
    }

    event SenderAccountEvent(ISCAgentID sender);

    function emitSenderAccount() public {
        ISCAgentID memory sender = ISC.sandbox.getSenderAccount();
        emit SenderAccountEvent(sender);
    }

    function sendBaseTokens(L1Address memory receiver, uint64 baseTokens)
        public
    {
        ISCAllowance memory allowance;
        if (baseTokens == 0) {
            allowance = ISC.sandbox.getAllowanceFrom(msg.sender);
        } else {
            allowance.baseTokens = baseTokens;
        }

        ISC.sandbox.takeAllowedFunds(msg.sender, allowance);

        ISCFungibleTokens memory fungibleTokens;
        require(allowance.baseTokens > TokensForGas);
        fungibleTokens.baseTokens = allowance.baseTokens - TokensForGas;

        ISCSendMetadata memory metadata;
        ISCSendOptions memory options;
        ISC.sandbox.send(receiver, fungibleTokens, true, metadata, options);
    }

    function sendAsNFT(L1Address memory receiver, NFTID id, uint64 storageDeposit) public {
        ISCAllowance memory allowance;
        allowance.baseTokens = storageDeposit;
        allowance.nfts = new NFTID[](1);
        allowance.nfts[0] = id;

        ISC.sandbox.takeAllowedFunds(msg.sender, allowance);

        ISCFungibleTokens memory fungibleTokens;
        ISCSendMetadata memory metadata;
        ISCSendOptions memory options;
        ISC.sandbox.sendAsNFT(receiver, fungibleTokens, id, true, metadata, options);
    }

    function callInccounter() public {
        ISCDict memory params = ISCDict(new ISCDictItem[](1));
        bytes memory int64Encoded42 = hex"2A00000000000000";
        params.items[0] = ISCDictItem("counter", int64Encoded42);
        ISCAllowance memory allowance;
        ISC.sandbox.call(ISC.util.hn("inccounter"), ISC.util.hn("incCounter"), params, allowance);
    }

    function makeISCPanic() public {
        // will produce a panic in ISC
        ISCDict memory params;
        ISCAllowance memory allowance;
        ISC.sandbox.call(
            ISC.util.hn("governance"),
            ISC.util.hn("claimChainOwnership"),
            params,
            allowance
        );
    }

    function moveToAccount(
        ISCAgentID memory targetAgentID,
        ISCAllowance memory allowance
    ) public {
        // moves funds owned by the current contract to the targetAgentID
        ISCDict memory params = ISCDict(new ISCDictItem[](2));
        params.items[0] = ISCDictItem("a", targetAgentID.data);
        bytes memory forceOpenAccount = "\xFF";
        params.items[1] = ISCDictItem("c", forceOpenAccount);
        ISC.sandbox.call(
            ISC.util.hn("accounts"),
            ISC.util.hn("transferAllowanceTo"),
            params,
            allowance
        );
    }

    function sendTo(address payable to, uint256 amount) public payable {
        to.transfer(amount);
    }

    function testRevertReason() public pure {
        revert("foobar");
    }

    function testStackOverflow() public view {
        bytes memory args = bytes.concat(
            hex"0000000000000000000000000000000000000000" // From address
            hex"01" // Optional field ToAddr exists
            , bytes20(uint160(address(this))), // Put our own address as ToAddr
            hex"0000000000000000" // Gas limit
            hex"00" // Optional field value does not exist
            hex"04000000" // Data length
            hex"b3ee6942" // Function to call: sha3.keccak_256(b'testStackOverflow()').hexdigest()[0:8]
        );
        ISCDict memory params = ISCDict(new ISCDictItem[](1));
        params.items[0] = ISCDictItem("c", args);

        ISC.sandbox.callView(
            ISC.util.hn("evm"),
            ISC.util.hn("callContract"),
            params
        );
    }

    function testStaticCall() public {
        bool success;
        bytes memory result;

        (success, result) = address(ISC.sandbox).call(abi.encodeWithSignature("triggerEvent(string)", "non-static"));
        require(success, "call should succeed");

        (success, result) = address(ISC.sandbox).staticcall(abi.encodeWithSignature("getChainID()"));
        require(success, "staticcall to view should succeed");

        (success, result) = address(ISC.sandbox).staticcall(abi.encodeWithSignature("triggerEvent(string)", "static"));
        require(!success, "staticcall to non-view should fail");
    }

    function testSelfDestruct(address payable beneficiary) public {
        selfdestruct(beneficiary);
    }
}
