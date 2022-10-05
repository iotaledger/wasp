// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmhost from "./wasmhost";
import * as wasmlib from "wasmlib"
import { panic } from "wasmlib"
import * as wc from "./index";
import * as cryptolib from "./cryptolib";
import { error } from "./wasmhost"

export class WasmClientSandbox implements wasmlib.ScHost {
	chainID       : wasmlib.ScChainID;
	Err           : error = null;
	eventDone     : bool = false;
	eventHandlers : wc.IEventHandler[] = [];
	keyPair       : cryptolib.KeyPair|null = null;
	ReqID         : wasmlib.ScRequestID = wasmlib.requestIDFromBytes([]);
	scName        : string;
	scHname       : wasmlib.ScHname;
	svcClient     : wc.IClientService;

	public constructor(svcClient: wc.IClientService, chainID: wasmlib.ScChainID, scName: string) {
		super();
		this.svcClient = svcClient;
		this.scName = scName;
		this.scHname = wasmlib.ScHname.fromName(scName);
		this.chainID = chainID;
	}

	public exportName(index: i32, name: string) {
		panic("WasmClientContext.ExportName")
	}

	public sandbox(funcNr: i32, args: u8[]): u8[] {
		this.Err = null;
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
		return [];
	}

	public stateDelete(key: u8[]) {
		panic("WasmClientContext.StateDelete");
	}

	public stateExists(key: u8[]): bool {
		panic("WasmClientContext.StateExists");
		return false;
	}

	public stateGet(key: u8[]): u8[] {
		panic("WasmClientContext.StateGet");
		return [];
	}

	public stateSet(key: u8[], value: u8[]) {
		panic("WasmClientContext.StateSet");
	}

	/////////////////////////////////////////////////////////////////

	public fnCall(args: u8[]): u8[] {
		let req = wasmlib.CallRequest.fromBytes(args);
		if (req.contract != this.scHname) {
			this.Err = "unknown contract: " + req.contract.toString();
			return [];
		}
		let res,  err = this.svcClient.callViewByHname(this.chainID, req.contract, req.function, req.params);
		if (err != null) {
			this.Err = err;
			return [];
		}
		return res;
	}

	public fnPost(args: u8[]): u8[] {
		let req = wasmlib.PostRequest.fromBytes(args);
		if (req.chainID != this.chainID) {
			this.Err = "unknown chain id: " + req.chainID.toString();
			return [];
		}
		if (req.contract != this.scHname) {
			this.Err = "unknown contract:" + req.contract.toString();
			return [];
		}
		let scAssets = new wasmlib.ScAssets(req.transfer);
		this.ReqID = this.svcClient.postRequest(this.chainID, req.contract, req.function, req.params, scAssets, this.keyPair);
		this.Err = this.svcClient.Err;
		return [];
	}

	public fnUtilsBech32Decode(args: u8[]): u8[] {
		let hrp,  addr,  err = iotago.ParseBech32(string(args));
		if (err != null) {
			this.Err = err;
			return null;
		}
		if (hrp != parameters.L1.Protocol.Bech32HRP) {
			this.Err = errors.Errorf("Invalid protocol prefix: %this", string(hrp));
			return null;
		}
		let  cvt = new wasmhost.WasmConvertor();
		return cvt.scAddress(addr).Bytes();
	}

	public fnUtilsBech32Encode(args: u8[]): u8[] {
		let  cvt = new wasmhost.WasmConvertor();
		let scAddress = wasmlib.addressFromBytes(args);
		let addr = cvt.iscAddress(scAddress);
		return u8[](addr.Bech32(parameters.L1.Protocol.Bech32HRP));
	}

	public fnUtilsHashName(args: u8[]): u8[] {
		let utils = new isc.Utils();
		return codec.EncodeHname(utils.Hashing().Hname(string(args)));
	}
}
