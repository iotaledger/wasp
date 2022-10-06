// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from "./index";

export class OffLedgerSignatureScheme {
    //TODO
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

    public ID(): isc.RequestID {
        //TODO
        return [];
    }

    public sign(keyPair: isc.KeyPair): OffLedgerRequest {
        //TODO
        return new OffLedgerRequest(this.chainID, this.contract, this.entryPoint, this.params, this.nonce);
    }

    public withAllowance(allowance: isc.Allowance): void {
        //TODO
        this.allowance = allowance;
    }
}
