// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from "./index";
import * as wasmlib from "wasmlib"

export class WasmConvertor {
    public constructor() {
    }

    public iscAddress(addr: wasmlib.ScAddress): isc.Address {
        return wasmlib.bytesToUint8Array(wasmlib.addressToBytes(addr));
    }

    public iscAllowance(addr: wasmlib.ScAssets): isc.Allowance {
        // TODO
        return [];
    }

    public iscChainID(chainID: wasmlib.ScChainID): isc.ChainID {
        return wasmlib.bytesToUint8Array(wasmlib.chainIDToBytes(chainID));
    }

    public iscHname(hName: wasmlib.ScHname): isc.Hname {
        return wasmlib.uint32FromBytes(wasmlib.hnameToBytes(hName))
    }

    public iscRequestID(chainID: wasmlib.ScRequestID): isc.RequestID {
        return wasmlib.bytesToUint8Array(wasmlib.requestIDToBytes(chainID));
    }

    public scAddress(addr: isc.Address): wasmlib.ScAddress {
        return wasmlib.addressFromBytes(wasmlib.bytesFromUint8Array(addr));
    }

    public scChainID(chainID: isc.ChainID): wasmlib.ScChainID {
        return wasmlib.chainIDFromBytes(wasmlib.bytesFromUint8Array(chainID))
    }

    public scHname(hName: isc.Hname): wasmlib.ScHname {
        return wasmlib.hnameFromBytes(wasmlib.uint32ToBytes(hName));
    }

    public scRequestID(requestID: isc.RequestID): wasmlib.ScRequestID {
        return wasmlib.requestIDFromBytes(wasmlib.bytesFromUint8Array(requestID))
    }
}