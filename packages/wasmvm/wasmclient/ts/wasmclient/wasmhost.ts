// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"

export type Address = u8[];
export type ChainID = u8[];
export type error = string|null;
export type Hname = u32;
export type RequestID = u8[];

export class WasmConvertor {
    public constructor() {
    }

    public iscAddress(addr: wasmlib.ScAddress): Address {
        return [];
    }

    public iscAllowance(addr: wasmlib.ScAssets): Allowance {

    }

    public iscChainID(chainID: wasmlib.ScChainID): ChainID {
        return [];
    }

    public iscHname(hName: wasmlib.ScHname): Hname {
        return wasmlib.uint32FromBytes(wasmlib.hnameToBytes(hName))
    }

    public iscRequestID(chainID: wasmlib.ScRequestID): RequestID {
        return [];
    }

    public scAddress(addr: Address): wasmlib.ScAddress {
        return wasmlib.addressFromBytes(addr)
    }

    public scChainID(chainID: ChainID): wasmlib.ScChainID {
        return wasmlib.chainIDFromBytes(chainID)
    }

    public scHname(hName: Hname): wasmlib.ScHname {
        return wasmlib.hnameFromBytes(wasmlib.uint32ToBytes(hName));
    }

    public scRequestID(requestID: RequestID): wasmlib.ScRequestID {
        return wasmlib.requestIDFromBytes(requestID)
    }
}