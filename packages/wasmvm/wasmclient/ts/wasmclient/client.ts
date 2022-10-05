// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import { ChainID,error, Hname,RequestID } from "./wasmhost"

export class WaspClient {
    Err : error = null;

    public constructor(waspAPI: string) {

    }

    public callViewByHname(chainID: ChainID, hContract: Hname, hFunction: Hname, args: u8[]): u8[]|null {
        this.Err = null;
        return null;
    }

    public postOffLedgerRequest(chainID: ChainID, signed: u8[]): error {
        this.Err = null;
        return null;
    }

    public waitUntilRequestProcessed(chainID: ChainID, reqID: RequestID, timeout: u32): error {
        return null;
    }
}
