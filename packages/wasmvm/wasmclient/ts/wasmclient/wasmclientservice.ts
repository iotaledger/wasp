// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as client from "./client";
import * as wasmhost from "./wasmhost";
import * as wasmlib from "wasmlib"
import * as wc from "./index";
import * as cryptolib from "./cryptolib";
import { error } from "./wasmhost"


export interface IClientService  {
	callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[]): u8[]|null;
	postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[], allowance: wasmlib.ScAssets, keyPair: cryptolib.KeyPair): wasmlib.ScRequestID|null;
	subscribeEvents(msg: /* chan */ string[], done: /* chan */ bool): error;
	waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): error;
}

export class WasmClientService {
	cvt        : wasmhost.WasmConvertor;
	waspClient : client.WaspClient;
    Err        : error = null;
	eventPort  : string;
	nonce      : u64;

    public constructor(waspAPI: string, eventPort: string)  {
        this.cvt = new wasmhost.WasmConvertor();
        this.waspClient= new client.WaspClient(waspAPI);
        this.eventPort= eventPort;
        this.nonce = 0;
    }
    
    static DefaultWasmClientService(): WasmClientService {
        return new WasmClientService("127.0.0.1:9090", "127.0.0.1:5550");
    }
    
    public callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[]): u8[]|null {
        let iscChainID = this.cvt.iscChainID(chainID);
        let iscContract = this.cvt.iscHname(hContract);
        let iscFunction = this.cvt.iscHname(hFunction);
        let res = this.waspClient.callViewByHname(iscChainID, iscContract, iscFunction, args);
        this.Err = this.waspClient.Err;
        if (this.Err != null) {
            return null;
        }
        return res;
    }
    
    public postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[], allowance: wasmlib.ScAssets, keyPair: cryptolib.KeyPair):  wasmlib.ScRequestID|null {
        let iscChainID = this.cvt.iscChainID(chainID);
        let iscContract = this.cvt.iscHname(hContract);
        let iscFunction = this.cvt.iscHname(hFunction);
        this.nonce++;
        let req = isc.NewOffLedgerRequest(iscChainID, iscContract, iscFunction, args, this.nonce);
        let iscAllowance = this.cvt.iscAllowance(allowance);
        req.WithAllowance(iscAllowance);
        let signed = req.Sign(keyPair);
        this.Err = this.waspClient.postOffLedgerRequest(iscChainID, signed);
        if (this.Err != null) {
            return null;
        }
        return this.cvt.scRequestID(signed.ID());
    }
    
    public subscribeEvents(msg: /* chan */ string[], done: /* chan */ bool): error {
        return subscribe.Subscribe(this.eventPort, msg, done, false, "");
    }
    
    public waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): error {
        let iscChainID = this.cvt.iscChainID(chainID);
        let iscReqID = this.cvt.iscRequestID(reqID);
        return this.waspClient.waitUntilRequestProcessed(iscChainID, iscReqID, timeout);
    }
}
