// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {WasmClientService} from './';

export class WasmClientSandbox {
    Err: isc.Error = null;
    keyPair: isc.KeyPair | null = null;
    nonce: u64 = 0n;
    ReqID: wasmlib.ScRequestID = new wasmlib.ScRequestID();
    scName: string;
    scHname: wasmlib.ScHname;
    svcClient: WasmClientService;

    public constructor(svcClient: WasmClientService, scName: string) {
        this.svcClient = svcClient;
        this.scName = scName;
        this.scHname = wasmlib.hnameFromBytes(isc.Codec.hNameBytes(scName));
    }

    public currentChainID(): wasmlib.ScChainID {
        return this.svcClient.currentChainID();
    }

    public fnCall(req: wasmlib.CallRequest): Uint8Array {
        if (!req.contract.equals(this.scHname)) {
            this.Err = 'unknown contract: ' + req.contract.toString();
            return new Uint8Array(0);
        }

        const [res, err] = this.svcClient.callViewByHname(req.contract, req.function, req.params);
        this.Err = err;
        return res;
    }

    public fnChainID(): wasmlib.ScChainID {
        return this.currentChainID();
    }

    public fnPost(req: wasmlib.PostRequest): Uint8Array {
        if (this.keyPair == null) {
            this.Err = 'missing key pair';
            return new Uint8Array(0);
        }

        if (!req.chainID.equals(this.currentChainID())) {
            this.Err = 'unknown chain id: ' + req.chainID.toString();
            return new Uint8Array(0);
        }

        if (!req.contract.equals(this.scHname)) {
            this.Err = 'unknown contract:' + req.contract.toString();
            return new Uint8Array(0);
        }

        const scAssets = new wasmlib.ScAssets(req.transfer);
        [this.ReqID, this.Err] = this.svcClient.postRequest(req.chainID, req.contract, req.function, req.params, scAssets, this.keyPair);
        return new Uint8Array(0);
    }
}
