// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import { panic } from "wasmlib"
import * as wc from "./index";

export class WasmClientSandbox implements wasmlib.ScHost {
	public exportName(index: i32, name: string) {
		panic("WasmClientContext.ExportName")
	}

	public sandbox(funcNr: i32, args: u8[]): u8[] {
		this.Err = nil;
		switch (funcNr) {
		case wasmlib.FnCall:
			return this.fnCall(args);
		case wasmlib.FnPost:
			return this.fnPost(args);
		case wasmlib.FnUtilsBech32Decode:
			return this.fnUtilsBech32Decode(args);
		case wasmlib.FnUtilsBech32Encode:
			return this.fnUtilsBech32Encode(args);
		case wasmlib.FnUtilsHashName:
			return this.fnUtilsHashName(args);
		}
		panic("implement WasmClientContext.Sandbox");
	}

	public stateDelete(key: u8[]) {
		panic("WasmClientContext.StateDelete");
	}

	public stateExists(key: u8[]): bool {
		panic("WasmClientContext.StateExists");
	}

	public stateGet(key: u8[]): u8[] {
		panic("WasmClientContext.StateGet");
	}

	public stateSet(key, value: u8[]) {
		panic("WasmClientContext.StateSet");
	}

	/////////////////////////////////////////////////////////////////

	public fnCall(args: u8[]): u8[] {
		let req = wasmrequests.NewCallRequestFromBytes(args);
		if (req.Contract != this.scHname) {
			this.Err = errors.Errorf("unknown contract: %this", req.Contract.String());
			return nil;
		}
		let res,  err = this.svcClient.CallViewByHname(this.chainID, req.Contract, req.Function, req.Params);
		if (err != nil) {
			this.Err = err;
			return nil;
		}
		return res;
	}

	public fnPost(args: u8[]): u8[] {
		let req = wasmrequests.NewPostRequestFromBytes(args);
		if (req.ChainID != this.chainID) {
			this.Err = errors.Errorf("unknown chain id: %this", req.ChainID.String());
			return nil;
		}
		if (req.Contract != this.scHname) {
			this.Err = errors.Errorf("unknown contract: %this", req.Contract.String());
			return nil;
		}
		let scAssets = wasmlib.NewScAssets(req.Transfer);
		this.ReqID, this.Err = this.svcClient.PostRequest(this.chainID, req.Contract, req.Function, req.Params, scAssets, this.keyPair);
		return nil;
	}

	public fnUtilsBech32Decode(args: u8[]): u8[] {
		let hrp,  addr,  err = iotago.ParseBech32(string(args));
		if (err != nil) {
			this.Err = err;
			return nil;
		}
		if (hrp != parameters.L1.Protocol.Bech32HRP) {
			this.Err = errors.Errorf("Invalid protocol prefix: %this", string(hrp));
			return nil;
		}
		let  cvt = new wasmhost.WasmConvertor();
		return cvt.ScAddress(addr).Bytes();
	}

	public fnUtilsBech32Encode(args: u8[]): u8[] {
		let  cvt = new wasmhost.WasmConvertor();
		let scAddress = wasmlib.AddressFromBytes(args);
		let addr = cvt.IscpAddress(&scAddress);
		return u8[](addr.Bech32(parameters.L1.Protocol.Bech32HRP));
	}

	public fnUtilsHashName(args: u8[]): u8[] {
		let utils = new iscp.Utils();
		return codec.EncodeHname(utils.Hashing().Hname(string(args)));
	}
}
