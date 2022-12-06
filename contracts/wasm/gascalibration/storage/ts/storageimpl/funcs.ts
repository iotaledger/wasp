// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib";
import * as wasmtypes from "wasmlib/wasmtypes";
import * as sc from "../storage/index";

export function funcF(_ctx: wasmlib.ScFuncContext, f: sc.FContext): void {
    let v = f.state.v();
    const n = f.params.n().value();
    for (let i: u32 = 0; i < n; i++) {
        v.appendUint32().setValue(i);
    }
}
