// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {panic} from 'wasmlib';
import {IClientService} from './';

export class WasmClientSandbox {
    static hrpForClient = '';

    chainID: wasmlib.ScChainID = new wasmlib.ScChainID();
    Err: isc.Error = null;
    eventReceived: bool = false;
    keyPair: isc.KeyPair | null = null;
    nonce: u64 = 0n;
    ReqID: wasmlib.ScRequestID = new wasmlib.ScRequestID();
    scName: string;
    scHname: wasmlib.ScHname;
    svcClient: IClientService;

    public constructor(svcClient: IClientService, chain: string, scName: string) {
        // local client implementations for sandboxed functions
        wasmlib.sandboxWrappers(clientBech32Decode, clientBech32Encode, clientHashName);

        this.svcClient = svcClient;
        this.scName = scName;
        this.scHname = wasmlib.hnameFromBytes(isc.Codec.hNameBytes(scName));

        // set the network prefix for the current network
        const [hrp, _addr, err] = isc.Codec.bech32Decode(chain);
        if (err != null) {
            this.Err = err;
            return this;
        }
        if (WasmClientSandbox.hrpForClient != hrp && WasmClientSandbox.hrpForClient != '') {
            panic('WasmClient can only connect to one Tangle network per app');
        }
        WasmClientSandbox.hrpForClient = hrp;

        // note that hrpForClient needs to be set
        this.chainID = wasmlib.chainIDFromString(chain);
    }

    public fnCall(req: wasmlib.CallRequest): Uint8Array {
        this.eventReceived = false;

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
        this.eventReceived = false;

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

export function clientBech32Decode(bech32: string): wasmlib.ScAddress {
    const [hrp, addr, err] = isc.Codec.bech32Decode(bech32);
    if (err != null) {
        panic(err);
    }
    if (hrp != WasmClientSandbox.hrpForClient) {
        panic('invalid protocol prefix: ' + hrp);
    }
    return addr;
}

export function clientBech32Encode(addr: wasmlib.ScAddress): string {
    return isc.Codec.bech32Encode(WasmClientSandbox.hrpForClient, addr);
}

export function clientHashName(name: string): wasmlib.ScHname {
    const hName = new wasmlib.ScHname(0);
    hName.id = isc.Codec.hNameBytes(name);
    return hName;
}
