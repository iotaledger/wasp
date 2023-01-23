// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {panic} from 'wasmlib';
import {IClientService} from './';

export class WasmClientSandbox implements wasmlib.ScHost {
    chainID: wasmlib.ScChainID = new wasmlib.ScChainID();
    Err: isc.Error = null;
    hrp = '';
    keyPair: isc.KeyPair | null = null;
    nonce: u64 = 0n;
    ReqID: wasmlib.ScRequestID = new wasmlib.ScRequestID();
    scName: string;
    scHname: wasmlib.ScHname;
    svcClient: IClientService;

    public constructor(svcClient: IClientService, chain: string, scName: string) {
        this.svcClient = svcClient;
        this.scName = scName;
        this.scHname = wasmlib.hnameFromBytes(isc.Codec.hNameBytes(scName));
        const [hrp, _addr, err] = isc.Codec.bech32Decode(chain);
        if (err != null) {
            this.Err = err;
            return this;
        }
        this.hrp = hrp;

        // note that chainIDFromString needs host to be connected
        wasmlib.connectHost(this);
        this.chainID = wasmlib.chainIDFromString(chain);
    }

    public exportName(index: i32, name: string) {
        panic('WasmClientContext.ExportName');
    }

    public sandbox(funcNr: i32, args: Uint8Array): Uint8Array {
        this.Err = null;
        switch (funcNr) {
            case wasmlib.FnCall:
                return this.fnCall(args);
            case wasmlib.FnChainID:
                return this.chainID.toBytes();
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
        if (!req.contract.equals(this.scHname)) {
            this.Err = 'unknown contract: ' + req.contract.toString();
            return new Uint8Array(0);
        }
        const [res, err] = this.svcClient.callViewByHname(this.chainID, req.contract, req.function, req.params);
        this.Err = err;
        return res;
    }

    public fnPost(args: Uint8Array): Uint8Array {
        if (this.keyPair == null) {
            this.Err = 'missing key pair';
            return new Uint8Array(0);
        }
        const req = wasmlib.PostRequest.fromBytes(args);
        if (!req.chainID.equals(this.chainID)) {
            this.Err = 'unknown chain id: ' + req.chainID.toString();
            return new Uint8Array(0);
        }
        if (!req.contract.equals(this.scHname)) {
            this.Err = 'unknown contract:' + req.contract.toString();
            return new Uint8Array(0);
        }
        const scAssets = new wasmlib.ScAssets(req.transfer);
        this.nonce++;
        const [reqId, err] = this.svcClient.postRequest(this.chainID, req.contract, req.function, req.params, scAssets, this.keyPair, this.nonce);
        this.ReqID = reqId;
        this.Err = err;
        return new Uint8Array(0);
    }

    public fnUtilsBech32Decode(args: Uint8Array): Uint8Array {
        const bech32 = wasmlib.stringFromBytes(args);
        const [hrp, addr, err] = isc.Codec.bech32Decode(bech32);
        if (err != null) {
            this.Err = err;
            return new Uint8Array(0);
        }
        if (hrp != this.hrp) {
            this.Err = 'invalid protocol prefix: ' + hrp;
            return new Uint8Array(0);
        }
        return addr.toBytes();
    }

    public fnUtilsBech32Encode(args: Uint8Array): Uint8Array {
        const addr = wasmlib.addressFromBytes(args);
        const bech32 = isc.Codec.bech32Encode(this.hrp, addr);
        return wasmlib.stringToBytes(bech32);
    }

    public fnUtilsHashName(args: Uint8Array): Uint8Array {
        const name = wasmlib.stringFromBytes(args);
        return isc.Codec.hNameBytes(name);
    }
}
