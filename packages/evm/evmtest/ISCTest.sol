// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.5;

import "@isccontract/ISC.sol";

contract ISCTest {
	ISCError test = isc.registerError("TestError");

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

	event SenderAddressEvent(IotaAddress sender);
	function emitSenderAddress() public {
		IotaAddress memory sender = isc.getSenderAddress();
		emit SenderAddressEvent(sender);
	}

	event AllowanceIotasEvent(uint64 iotas);
	function emitAllowanceIotas() public {
		emit AllowanceIotasEvent(isc.getAllowanceIotas());
	}

	event AllowanceNativeTokenEvent(IotaNativeToken token);
	function emitAllowanceNativeTokens() public {
		uint16 n = isc.getAllowanceNativeTokensLen();
		for (uint16 i = 0; i < n; i++) {
			emit AllowanceNativeTokenEvent(isc.getAllowanceNativeToken(i));
		}
	}

  event AllowanceNFTEvent(ISCNFT token);
  function emitAllowanceNFTs() public {
      uint16 n = isc.getAllowanceNFTsLen();
      for (uint16 i = 0; i < n; i++) {
          emit AllowanceNFTEvent(isc.getNFTData(isc.getAllowanceNFTID(i)));
      }
	}

	event SendEvent();
	function emitSend() public {
		// WIP
		ISCRequestParameters memory params;
		params.fungibleTokens.iotas = 1074;
		params.fungibleTokens.tokens = new IotaNativeToken[](1);
		params.fungibleTokens.tokens[0].amount = 1074;
		params.metadata.entrypoint = ISCHname.wrap(0x1337);
		params.metadata.targetContract = ISCHname.wrap(0xd34db33f);
		params.adjustMinimumDustDeposit = true;

		isc.send(params);
		emit SendEvent();
	}


	function emitRevertVMError() public view {
		revert VMError(test);
  }
  
  function callInccounter() public {
		ISCDict memory params = ISCDict(new ISCDictItem[](1));
        bytes memory int64Encoded42 = hex"2A00000000000000";
		params.items[0] = ISCDictItem("counter", int64Encoded42);
		ISCAllowance memory allowance;
		isc.call(isc.hn("inccounter"), isc.hn("incCounter"), params, allowance);
	}

}
