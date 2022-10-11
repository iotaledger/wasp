// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as isc from "./index";

export class OffLedgerSignatureScheme {
    keyPair: isc.KeyPair;
    signature: u8[] = [];

    public constructor(keyPair: isc.KeyPair) {
        this.keyPair = keyPair;
    }
}

export class OffLedgerRequest {
    chainID: isc.ChainID;
    contract: isc.Hname;
    entryPoint: isc.Hname;
    params: isc.Dict;
    signatureScheme: isc.OffLedgerSignatureScheme = new isc.OffLedgerSignatureScheme(new isc.KeyPair([]));
    nonce: u64;
    allowance: isc.Allowance = [ 1 ]; // empty allowance
    gasBudget: u64 = 0;

    public constructor(chainID: isc.ChainID, contract: isc.Hname, entryPoint: isc.Hname, params: u8[], nonce: u64) {
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
        let data: u8[] = [ 1 ]; // requestKindTagOffLedgerISC
        data = data.concat(wasmlib.bytesFromUint8Array(this.chainID));
        data = data.concat(wasmlib.uint32ToBytes(this.contract));
        data = data.concat(wasmlib.uint32ToBytes(this.entryPoint));
        data = data.concat(this.params);
        data = data.concat(wasmlib.uint64ToBytes(this.nonce));
        data = data.concat(wasmlib.uint64ToBytes(this.gasBudget));
        const pubKey = wasmlib.bytesFromUint8Array(this.signatureScheme.keyPair.publicKey);
        data = data.concat([pubKey.length as u8]);
        data = data.concat(pubKey);
        data = data.concat(this.allowance);
        return data;
    }

    public ID(): isc.RequestID {
        //TODO
        return wasmlib.bytesToUint8Array([]);
    }

    public sign(keyPair: isc.KeyPair): OffLedgerRequest {
        const req =  new OffLedgerRequest(this.chainID, this.contract, this.entryPoint, this.params, this.nonce);
        req.signatureScheme = new isc.OffLedgerSignatureScheme(keyPair);
        req.signatureScheme.signature = keyPair.sign(this.essence());
        return req;
    }

    public withAllowance(allowance: isc.Allowance): void {
        //TODO
        this.allowance = allowance;
    }
}
