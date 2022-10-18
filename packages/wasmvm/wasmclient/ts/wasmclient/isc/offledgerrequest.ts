// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from "./index";
import * as wasmlib from "wasmlib"

export class OffLedgerSignatureScheme {
    keyPair: isc.KeyPair;
    signature: u8[] = [];

    public constructor(keyPair: isc.KeyPair) {
        this.keyPair = keyPair;
    }
}

export class OffLedgerRequest {
    chainID: wasmlib.ScChainID;
    contract: wasmlib.ScHname;
    entryPoint: wasmlib.ScHname;
    params: u8[];
    signatureScheme: isc.OffLedgerSignatureScheme = new isc.OffLedgerSignatureScheme(new isc.KeyPair([]));
    nonce: u64;
    allowance: wasmlib.ScAssets = new wasmlib.ScAssets([]);
    gasBudget: u64 = 0;

    public constructor(chainID: wasmlib.ScChainID, contract: wasmlib.ScHname, entryPoint: wasmlib.ScHname, params: u8[], nonce: u64) {
        this.chainID = chainID;
        this.contract = contract;
        this.entryPoint = entryPoint;
        this.params = params;
        this.nonce = nonce;
    }

    public bytes(): u8[] {
        return this.essence().concat(this.signatureScheme.signature);
    }

    public essence(): u8[] {
        let data: u8[] = [1]; // requestKindTagOffLedgerISC
        data = data.concat(this.chainID.toBytes());
        data = data.concat(this.contract.toBytes());
        data = data.concat(this.entryPoint.toBytes());
        data = data.concat(this.params);
        data = data.concat(wasmlib.uint64ToBytes(this.nonce));
        data = data.concat(wasmlib.uint64ToBytes(this.gasBudget));
        const pubKey = wasmlib.bytesFromUint8Array(this.signatureScheme.keyPair.publicKey);
        data = data.concat([pubKey.length as u8]);
        data = data.concat(pubKey);
        //TODO convert to bytes according to isc.Allowance?
        data = data.concat(this.allowance.toBytes());
        return data;
    }

    public ID(): wasmlib.ScRequestID {
        //TODO
        return wasmlib.requestIDFromBytes([]);
    }

    public sign(keyPair: isc.KeyPair): OffLedgerRequest {
        const req = new OffLedgerRequest(this.chainID, this.contract, this.entryPoint, this.params, this.nonce);
        req.signatureScheme = new isc.OffLedgerSignatureScheme(keyPair);
        req.signatureScheme.signature = keyPair.sign(this.essence());
        return req;
    }

    public withAllowance(allowance: wasmlib.ScAssets): void {
        this.allowance = allowance
    }
}
