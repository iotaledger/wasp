// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {WasmClientService} from './';

export class WasmClientSandbox {
    chainID: wasmlib.ScChainID = new wasmlib.ScChainID();
    Err: isc.Error = null;
    keyPair: isc.KeyPair | null = null;
    nonce: u64 = 0n;
    ReqID: wasmlib.ScRequestID = new wasmlib.ScRequestID();
    scName: string;
    scHname: wasmlib.ScHname;
    svcClient: WasmClientService;

    public constructor(svcClient: WasmClientService, chainID: string, scName: string) {
        this.Err = isc.setSandboxWrappers(chainID);
        this.svcClient = svcClient;
        this.scName = scName;
        this.scHname = wasmlib.hnameFromBytes(isc.Codec.hNameBytes(scName));
        if (this.Err == null) {
            // only do this when setSandboxWrappers() was successful
            this.chainID = wasmlib.chainIDFromString(chainID);
        }
    }

    public fnCall(req: wasmlib.CallRequest): Uint8Array {
        if (!req.contract.equals(this.scHname)) {
            this.Err = 'unknown contract: ' + req.contract.toString();
            return new Uint8Array(0);
        }

        const [res, err] = this.svcClient.callViewByHname(this.chainID, req.contract, req.function, req.params);
        this.Err = err;
        return res;
    }

    public fnChainID(): wasmlib.ScChainID {
        return this.chainID;
    }

    public fnPost(req: wasmlib.PostRequest): Uint8Array {
        if (this.keyPair == null) {
            this.Err = 'missing key pair';
            return new Uint8Array(0);
        }

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
        [this.ReqID, this.Err] = this.svcClient.postRequest(req.chainID, req.contract, req.function, req.params, scAssets, this.keyPair, this.nonce);
        return new Uint8Array(0);
    }
}
