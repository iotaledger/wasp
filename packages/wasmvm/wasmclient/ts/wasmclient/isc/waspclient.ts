// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from "./index"

//TODO
export class WaspClient {
    Err: isc.Error = null;

    public constructor(waspAPI: string) {

    }

    public callViewByHname(chainID: isc.ChainID, hContract: isc.Hname, hFunction: isc.Hname, args: u8[]): u8[] {
        this.Err = null;
        return [];
    }

    public postOffLedgerRequest(chainID: isc.ChainID, signed: isc.OffLedgerRequest): isc.Error {
        this.Err = null;
        return null;
    }

    public waitUntilRequestProcessed(chainID: isc.ChainID, reqID: isc.RequestID, timeout: u32): isc.Error {
        return null;
    }
}
