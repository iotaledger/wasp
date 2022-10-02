// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as wc from "./index";

export interface IClientService  {
	callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[]): (u8[], error);
	postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[], allowance: wasmlib.ScAssets, keyPair: cryptolib.KeyPair): (wasmlib.ScRequestID, error);
	subscribeEvents(msg: chan string[], done: chan bool): error;
	waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: time.Duration): error;
}

export class WasmClientService {
	cvt        : wasmhost.WasmConvertor;
	waspClient : client.WaspClient;
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
    
    public callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[]): (u8[], error) {
        let iscpChainID = this.cvt.IscpChainID(chainID);
        let iscpContract = this.cvt.IscpHname(hContract);
        let iscpFunction = this.cvt.IscpHname(hFunction);
        let params,  err = dict.FromBytes(args);
        if (err != nil) {
            return nil, err;
        }
        let res,  err = this.waspClient.CallViewByHname(iscpChainID, iscpContract, iscpFunction, params);
        if (err != nil) {
            return nil, err;
        }
        return res.Bytes(), nil;
    }
    
    public postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[], allowance: wasmlib.ScAssets, keyPair: cryptolib.KeyPair): (reqID: wasmlib.ScRequestID, err: error) {
        let iscpChainID = this.cvt.IscpChainID(&chainID);
        let iscpContract = this.cvt.IscpHname(hContract);
        let iscpFunction = this.cvt.IscpHname(hFunction);
        let params,  err = dict.FromBytes(args);
        if (err != nil) {
            return reqID, err;
        }
        this.nonce++;
        let req = iscp.NewOffLedgerRequest(iscpChainID, iscpContract, iscpFunction, params, this.nonce);
        let iscpAllowance = this.cvt.IscpAllowance(allowance);
        req.WithAllowance(iscpAllowance);
        let signed = req.Sign(keyPair);
        err = this.waspClient.PostOffLedgerRequest(iscpChainID, signed);
        if (err == nil) {
            reqID = this.cvt.ScRequestID(signed.ID());
        }
        return reqID, err;
    }
    
    public subscribeEvents(msg: chan []string, done: chan bool): error {
        return subscribe.Subscribe(this.eventPort, msg, done, false, "");
    }
    
    public waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: time.Duration): error {
        let iscpChainID = this.cvt.IscpChainID(chainID);
        let iscpReqID = this.cvt.IscpRequestID(reqID);
        let _,  err = this.waspClient.WaitUntilRequestProcessed(iscpChainID, iscpReqID, timeout);
        return err;
    }
}
