// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from "./isc";
import * as wasmlib from "wasmlib"
import {panic} from "wasmlib"
import * as wc from "./index";

export class WasmClientSandbox implements wasmlib.ScHost {
    chainID: wasmlib.ScChainID;
    Err: isc.Error = null;
    eventDone: bool = false;
    eventHandlers: wc.IEventHandler[] = [];
    keyPair: isc.KeyPair | null = null;
    ReqID: wasmlib.ScRequestID = wasmlib.requestIDFromBytes([]);
    scName: string;
    scHname: wasmlib.ScHname;
    svcClient: wc.IClientService;

    public constructor(svcClient: wc.IClientService, chainID: wasmlib.ScChainID, scName: string) {
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
        let res = this.svcClient.callViewByHname(this.chainID, req.contract, req.function, req.params);
        this.Err = this.svcClient.Err()
        if (this.Err != null) {
            return [];
        }
        return res;
    }

    public fnPost(args: u8[]): u8[] {
        if (this.keyPair == null) {
            this.Err = "missing key pair";
            return [];
        }
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
        this.Err = this.svcClient.Err()
        return [];
    }

    public fnUtilsBech32Decode(args: u8[]): u8[] {
        let bech32 = wasmlib.stringFromBytes(args);
        let addr = isc.Codec.bech32Decode(bech32);
        if (addr == null) {
            this.Err = "Invalid bech32 address encoding";
            return [];
        }
        let cvt = new isc.WasmConvertor();
        return cvt.scAddress(addr).toBytes();
    }

    public fnUtilsBech32Encode(args: u8[]): u8[] {
        let scAddress = wasmlib.addressFromBytes(args);
        let cvt = new isc.WasmConvertor();
        let addr = cvt.iscAddress(scAddress);
        let bech32 = isc.Codec.bech32Encode(addr);
        return wasmlib.stringToBytes(bech32);
    }

    public fnUtilsHashName(args: u8[]): u8[] {
        let name = wasmlib.stringFromBytes(args);
        let iscHname = isc.Codec.hNameEncode(name);
        let cvt = new isc.WasmConvertor();
        let hName = cvt.scHname(iscHname);
        return wasmlib.hnameToBytes(hName);
    }
}
