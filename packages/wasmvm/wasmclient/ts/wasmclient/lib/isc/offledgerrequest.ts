// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Blake2b} from '@iota/crypto.js';
import * as wasmlib from 'wasmlib';
import {KeyPair} from './keypair';

export class OffLedgerSignature {
    publicKey: Uint8Array;
    signature: Uint8Array;

    public constructor(publicKey: Uint8Array) {
        this.publicKey = publicKey;
        this.signature = new Uint8Array(0);
    }
}

export class OffLedgerRequest {
    chainID: wasmlib.ScChainID;
    contract: wasmlib.ScHname;
    entryPoint: wasmlib.ScHname;
    params: Uint8Array;
    signature: OffLedgerSignature = new OffLedgerSignature(new KeyPair(new Uint8Array(0)).publicKey);
    nonce: u64;
    allowance: wasmlib.ScAssets = new wasmlib.ScAssets(new Uint8Array(0));
    gasBudget: u64 = 2n ** 64n - 1n;

    public constructor(chainID: wasmlib.ScChainID, contract: wasmlib.ScHname, entryPoint: wasmlib.ScHname, params: Uint8Array, nonce: u64) {
        this.chainID = chainID;
        this.contract = contract;
        this.entryPoint = entryPoint;
        this.params = params;
        this.nonce = nonce;
    }

    public bytes(): Uint8Array {
        let data = this.essence()
        const publicKey = this.signature.publicKey;
        data = wasmlib.concat(data, wasmlib.uint8ToBytes(publicKey.length as u8))
        data = wasmlib.concat(data, publicKey)
        const signature = this.signature.signature;
        data = wasmlib.concat(data, wasmlib.uint16ToBytes(signature.length as u16))
        return wasmlib.concat(data, signature);
    }

    public essence(): Uint8Array {
        const oneByte = new Uint8Array(1);
        oneByte[0] = 1; // requestKindTagOffLedgerISC
        let data = wasmlib.concat(oneByte, this.chainID.toBytes());
        data = wasmlib.concat(data, this.contract.toBytes());
        data = wasmlib.concat(data, this.entryPoint.toBytes());
        data = wasmlib.concat(data, this.params);
        data = wasmlib.concat(data, wasmlib.uint64ToBytes(this.nonce));
        data = wasmlib.concat(data, wasmlib.uint64ToBytes(this.gasBudget));
        return wasmlib.concat(data, this.allowance.toBytes());
    }

    public ID(): wasmlib.ScRequestID {
        // req id is hash of req bytes with output index zero
        const hash = Blake2b.sum256(this.bytes());
        const reqId = new wasmlib.ScRequestID();
        reqId.id.set(hash, 0);
        return reqId;
    }

    public sign(keyPair: KeyPair): OffLedgerRequest {
        const req = new OffLedgerRequest(this.chainID, this.contract, this.entryPoint, this.params, this.nonce);
        req.signature = new OffLedgerSignature(keyPair.publicKey);
        req.signature.signature = keyPair.sign(req.essence());
        return req;
    }

    public withAllowance(allowance: wasmlib.ScAssets): void {
        this.allowance = allowance;
    }
}
