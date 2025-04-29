// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.5;

import "@iscmagic/ISC.sol";

contract ISCTest {
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
        emit RequestIDEvent(ISC.sandbox.getRequestID());
    }

    event DummyEvent(string s);

    function emitDummyEvent() public {
        emit DummyEvent("foobar");
    }


    event SenderAccountEvent(ISCAgentID sender);

    function emitSenderAccount() public {
        ISCAgentID memory sender = ISC.sandbox.getSenderAccount();
        emit SenderAccountEvent(sender);
    }

    function sendBaseTokens(IotaAddress receiver, uint64 baseTokens)
        public payable
    {
        ISCAssets memory allowance;
        if (baseTokens == 0) {
            allowance = ISC.sandbox.getAllowanceFrom(msg.sender);
        } else {
            allowance.coins = new CoinBalance[](1);
            allowance.coins[0].coinType = ISC.sandbox.getBaseTokenInfo().coinType;
            allowance.coins[0].amount = uint64(baseTokens);
        }

        ISC.sandbox.takeAllowedFunds(msg.sender, allowance);

        ISCAssets memory assets;
        assets.coins = new CoinBalance[](1);
        assets.coins[0].coinType = allowance.coins[0].coinType;
        assets.coins[0].amount = allowance.coins[0].amount;

        ISC.sandbox.transferToL1(receiver, assets);
    }

    function callInccounter() public {
        bytes[] memory params = new bytes[](1);
        params[0] = hex"012A00000000000000"; // optional int64(42) BCS-encoded
        ISC.sandbox.call(
            ISCMessage({
                target: ISCTarget({
                    contractHname: ISC.util.hn("inccounter"),
                    entryPoint: ISC.util.hn("incCounter")
                }),
                params: params
            }),
            ISCAssets({
                coins: new CoinBalance[](0),
                objects: new IotaObject[](0)
            })
        );
    }

    function makeISCPanic() public {
        // will produce a panic in ISC
        ISC.sandbox.call(
            ISCMessage({
                target: ISCTarget({
                    contractHname: ISC.util.hn("governance"),
                    entryPoint: ISC.util.hn("claimChainAdmin")
                }),
                params: new bytes[](0)
            }),
            ISCAssets({
                coins: new CoinBalance[](0),
                objects: new IotaObject[](0)
            })
        );
    }

    function moveToAccount(
        ISCAgentID memory targetAgentID,
        ISCAssets memory allowance
    ) public {
        // moves funds owned by the current contract to the targetAgentID
        bytes[] memory params = new bytes[](1);
        params[0] = targetAgentID.data;
        ISC.sandbox.call(
            ISCMessage({
                target: ISCTarget({
                    contractHname: ISC.util.hn("accounts"),
                    entryPoint: ISC.util.hn("transferAllowanceTo")
                }),
                params: params
            }),
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
        bytes[] memory params = new bytes[](1);
        params[0] = bytes.concat(
            hex"0000000000000000000000000000000000000000" // From address
            hex"01" // Optional field ToAddr exists
            , bytes20(uint160(address(this))), // Put our own address as ToAddr
            hex"00" // Gas limit
            hex"00" // Optional field value does not exist
            hex"04" // Data length
            hex"b3ee6942" // Function to call: sha3.keccak_256(b'testStackOverflow()').hexdigest()[0:8]
        );
        ISC.sandbox.callView(
            ISCMessage({
                target: ISCTarget({
                    contractHname: ISC.util.hn("evm"),
                    entryPoint: ISC.util.hn("callContract")
                }),
                params: params
            })
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

    event TestSelfDestruct6780ContractCreated(address);

    function testSelfDestruct6780() public{
        // deploy a new contract instance
        SelfDestruct6780 c = new SelfDestruct6780();
        emit TestSelfDestruct6780ContractCreated(address(c));
        // call selfdestruct in the same tx
        c.testSelfDestruct(payable(msg.sender));
    }

    event LoopEvent();

    function loopWithGasLeft() public {
        while (gasleft() >= 10000) {
            emit LoopEvent();
        }
    }

    function testCallViewCaller() public view returns (bytes memory) {
        // test that the caller is set to this contract's address
        bytes[] memory params = new bytes[](1);
        params[0] = hex"00"; // Optional field agentID
        bytes[] memory r = ISC.sandbox.callView(
            ISCMessage({
                target: ISCTarget({
                    contractHname: ISC.util.hn("accounts"),
                    entryPoint: ISC.util.hn("balance")
                }),
                params: params
            })
        );
        return r[0];
    }

    error CustomError(uint8);

    function revertWithCustomError() public pure {
        revert CustomError(42);
    }

    event SomeEvent();

    function emitEventAndRevert() public {
        emit SomeEvent();
        revert();
    }
}

contract SelfDestruct6780{
    function testSelfDestruct(address payable beneficiary) public {
        selfdestruct(beneficiary);
    }
}
