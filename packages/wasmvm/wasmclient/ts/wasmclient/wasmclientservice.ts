// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from "./isc"
import * as wasmlib from "wasmlib"


export interface IClientService {
    callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[]): u8[];

    Err(): isc.Error;

    postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[], allowance: wasmlib.ScAssets, keyPair: isc.KeyPair): wasmlib.ScRequestID;

    subscribeEvents(msg: /* chan */ string[], done: /* chan */ bool): isc.Error;

    waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): isc.Error;
}

export class WasmClientService implements IClientService {
    cvt: isc.WasmConvertor;
    waspClient: isc.WaspClient;
    lastError: isc.Error = null;
    eventPort: string;
    nonce: u64;

    public constructor(waspAPI: string, eventPort: string) {
        this.cvt = new isc.WasmConvertor();
        this.waspClient = new isc.WaspClient(waspAPI);
        this.eventPort = eventPort;
        this.nonce = 0;
    }

    public static DefaultWasmClientService(): WasmClientService {
        return new WasmClientService("127.0.0.1:9090", "127.0.0.1:5550");
    }

    public callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[]): u8[] {
        let iscChainID = this.cvt.iscChainID(chainID);
        let iscContract = this.cvt.iscHname(hContract);
        let iscFunction = this.cvt.iscHname(hFunction);
        let res = this.waspClient.callViewByHname(iscChainID, iscContract, iscFunction, args);
        this.lastError = this.waspClient.Err;
        if (this.lastError != null) {
            return [];
        }
        return res;
    }

    public Err(): isc.Error {
        return this.lastError;
    }

    public postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[], allowance: wasmlib.ScAssets, keyPair: isc.KeyPair): wasmlib.ScRequestID {
        let iscChainID = this.cvt.iscChainID(chainID);
        let iscContract = this.cvt.iscHname(hContract);
        let iscFunction = this.cvt.iscHname(hFunction);
        this.nonce++;
        let req = new isc.OffLedgerRequest(iscChainID, iscContract, iscFunction, args, this.nonce);
        let iscAllowance = this.cvt.iscAllowance(allowance);
        req.withAllowance(iscAllowance);
        let signed = req.sign(keyPair);
        this.lastError = this.waspClient.postOffLedgerRequest(iscChainID, signed);
        if (this.lastError != null) {
            return wasmlib.requestIDFromBytes([]);
        }
        return this.cvt.scRequestID(signed.ID());
    }

    public subscribeEvents(msg: /* chan */ string[], done: /* chan */ bool): isc.Error {
        //TODO
        // return subscribe.Subscribe(this.eventPort, msg, done, false, "");
        return null;
    }

    public waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): isc.Error {
        let iscChainID = this.cvt.iscChainID(chainID);
        let iscReqID = this.cvt.iscRequestID(reqID);
        return this.waspClient.waitUntilRequestProcessed(iscChainID, iscReqID, timeout);
    }
}
