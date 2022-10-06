// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from "./index";

// TODO
export class Codec {
    public static bech32Decode(bech32: string): isc.Address | null {
        return null;
    }

    public static bech32Encode(addr: isc.Address): string {
        return "";
    }

    public static hNameEncode(name: string): isc.Hname {
        return 0;
    }
}