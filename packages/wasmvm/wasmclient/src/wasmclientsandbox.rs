// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;
use crate::*;

pub struct WasmClientSandbox {
}

impl ScHost for WasmClientSandbox {
	fn export_name(&self, index: i32, name: string) {
		panic("WasmClientContext.ExportName");
	}

	fn sandbox(&self, funcNr: i32, args: &[u8]) -> Vec<u8> {
		self.Err = nil;
		match funcNr {
		FnCall=> return self.fnCall(args),
		FnPost=> return self.fnPost(args),
		FnUtilsBech32Decode=> return self.fnUtilsBech32Decode(args),
		FnUtilsBech32Encode=> return self.fnUtilsBech32Encode(args),
		FnUtilsHashName=> return self.fnUtilsHashName(args),
			_ => panic("implement WasmClientContext.Sandbox"),
		}
	}

	fn state_delete(&self, _key: &[u8]) {
		panic("WasmClientContext.StateDelete");
	}

	fn state_exists(&self, _key: &[u8]) -> bool {
		panic("WasmClientContext.StateExists");
	}

	fn state_get(&self, _key: &[u8]) -> Vec<u8> {
		panic("WasmClientContext.StateGet");
	}

	fn state_set(&self, _key: &[u8], _value: &[u8]) {
		panic("WasmClientContext.StateSet");
	}
}

impl WasmClientSandbox {
	pub fn fnCall(&self, args: &[u8]) -> Vec<u8> {
		let req = wasmrequests.NewCallRequestFromBytes(args);
		if req.Contract != self.scHname {
			self.Err = errors.Errorf("unknown contract: %self", req.Contract.String());
			return nil;
		}
		let res,  err = self.svcClient.CallViewByHname(self.chainID, req.Contract, req.Function, req.Params);
		if err != nil {
			self.Err = err;
			return nil;
		}
		return res;
	}

	pub fn fnPost(&self, args: &[u8]) -> Vec<u8> {
		let req = wasmrequests.NewPostRequestFromBytes(args);
		if req.ChainID != self.chainID {
			self.Err = errors.Errorf("unknown chain id: %self", req.ChainID.String());
			return nil;
		}
		if req.Contract != self.scHname {
			self.Err = errors.Errorf("unknown contract: %self", req.Contract.String());
			return nil;
		}
		let scAssets = wasmlib.NewScAssets(req.Transfer);
		self.ReqID, self.Err = self.svcClient.PostRequest(self.chainID, req.Contract, req.Function, req.Params, scAssets, self.keyPair);
		return nil;
	}

	pub fn fnUtilsBech32Decode(&self, args: &[u8]) -> Vec<u8> {
		let hrp,  addr,  err = iotago.ParseBech32(string(args));
		if err != nil {
			self.Err = err;
			return nil;
		}
		if hrp != parameters.L1.Protocol.Bech32HRP {
			self.Err = errors.Errorf("Invalid protocol prefix: %self", string(hrp));
			return nil;
		}
		var cvt wasmhost.WasmConvertor;
		return cvt.ScAddress(addr).Bytes();
	}

	pub fn fnUtilsBech32Encode(&self, args: &[u8]) -> Vec<u8> {
		var cvt wasmhost.WasmConvertor;
		let scAddress = AddressFromBytes(args);
		let addr = cvt.IscpAddress(&scAddress);
		return &[u8](addr.Bech32(parameters.L1.Protocol.Bech32HRP));
	}

	pub fn fnUtilsHashName(&self, args: &[u8]) -> Vec<u8> {
		var utils iscp.Utils;
		return codec.EncodeHname(utils.Hashing().Hname(string(args)));
	}
}
