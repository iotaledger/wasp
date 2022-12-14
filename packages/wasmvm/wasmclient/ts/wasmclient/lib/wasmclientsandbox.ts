// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import { panic } from 'wasmlib';
import { IClientService } from './';

export class WasmClientSandbox implements wasmlib.ScHost {
    chID: wasmlib.ScChainID;
    Err: isc.Error = null;
    eventDone: bool = false;
    eventHandlers: wasmlib.IEventHandlers[] = [];
    eventReceived: bool = false;
    keyPair: isc.KeyPair | null = null;
    nonce: u64 = 0n;
    ReqID: wasmlib.ScRequestID = wasmlib.requestIDFromBytes(new Uint8Array(0));
    scName: string;
    scHname: wasmlib.ScHname;
    svcClient: IClientService;

    public constructor(svcClient: IClientService, chainID: wasmlib.ScChainID, scName: string) {
        this.svcClient = svcClient;
        this.chID = chainID;
        this.scName = scName;
        this.scHname = wasmlib.hnameFromBytes(isc.Codec.hNameBytes(scName));
    }

    public exportName(index: i32, name: string) {
        panic('WasmClientContext.ExportName');
    }

    public sandbox(funcNr: i32, args: Uint8Array): Uint8Array {
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
        panic('implement WasmClientContext.Sandbox');
        return new Uint8Array(0);
    }

    public stateDelete(key: Uint8Array) {
        panic('WasmClientContext.StateDelete');
    }

    public stateExists(key: Uint8Array): bool {
        panic('WasmClientContext.StateExists');
        return false;
    }

    public stateGet(key: Uint8Array): Uint8Array {
        panic('WasmClientContext.StateGet');
        return new Uint8Array(0);
    }

    public stateSet(key: Uint8Array, value: Uint8Array) {
        panic('WasmClientContext.StateSet');
    }

    /////////////////////////////////////////////////////////////////

    public fnCall(args: Uint8Array): Uint8Array {
        const req = wasmlib.CallRequest.fromBytes(args);
        if (req.contract != this.scHname) {
            this.Err = 'unknown contract: ' + req.contract.toString();
            return new Uint8Array(0);
        }
        const [res, err] = this.svcClient.callViewByHname(this.chID, req.contract, req.function, req.params);
        this.Err = err;
        if (this.Err != null) {
            return new Uint8Array(0);
        }
        return res;
    }

    public fnPost(args: Uint8Array): Uint8Array {
        if (this.keyPair == null) {
            this.Err = 'missing key pair';
            return new Uint8Array(0);
        }
        const req = wasmlib.PostRequest.fromBytes(args);
        if (!req.chainID.equals(this.chID)) {
            this.Err = 'unknown chain id: ' + req.chainID.toString();
            return new Uint8Array(0);
        }
        if (!req.contract.equals(this.scHname)) {
            this.Err = 'unknown contract:' + req.contract.toString();
            return new Uint8Array(0);
        }
        const scAssets = new wasmlib.ScAssets(req.transfer);
        this.nonce++;
        const [reqId, err] = this.svcClient.postRequest(this.chID, req.contract, req.function, req.params, scAssets, this.keyPair, this.nonce);
        this.ReqID = reqId;
        this.Err = err;
        return new Uint8Array(0);
    }

    public fnUtilsBech32Decode(args: Uint8Array): Uint8Array {
        const bech32 = wasmlib.stringFromBytes(args);
        const addr = isc.Codec.bech32Decode(bech32);
        if (addr == null) {
            this.Err = 'Invalid bech32 address encoding';
            return new Uint8Array(0);
        }
        return addr.toBytes();
    }

    public fnUtilsBech32Encode(args: Uint8Array): Uint8Array {
        const addr = wasmlib.addressFromBytes(args);
        const bech32 = isc.Codec.bech32Encode(addr);
        return wasmlib.stringToBytes(bech32);
    }

    public fnUtilsHashName(args: Uint8Array): Uint8Array {
        const name = wasmlib.stringFromBytes(args);
        return isc.Codec.hNameBytes(name);
    }
}
