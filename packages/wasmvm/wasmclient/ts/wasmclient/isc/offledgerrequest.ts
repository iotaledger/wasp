// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

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
    signatureScheme: isc.OffLedgerSignatureScheme | null = null; // null if unsigned
    nonce: u64;
    allowance: isc.Allowance | null = null;
    gasBudget: u64 = 0;

    public constructor(chainID: isc.ChainID, contract: isc.Hname, entryPoint: isc.Hname, params: u8[], nonce: u64) {
        this.chainID = chainID;
        this.contract = contract;
        this.entryPoint = entryPoint;
        this.params = params;
        this.nonce = nonce;
    }

    public essence(): u8[] {
        //TODO
        return [];
    }

    public ID(): isc.RequestID {
        //TODO
        return [];
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
