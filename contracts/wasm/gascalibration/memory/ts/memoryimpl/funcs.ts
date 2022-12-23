// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib";
import * as wasmtypes from "wasmlib/wasmtypes";
import * as sc from "../memory/index";

export function funcF(ctx: wasmlib.ScFuncContext, f: sc.FContext): void {
    const n = f.params.n().value();
    const store: u32[] = [];
    for (var i: u32 = 0; i < n; i++) {
        store.push(i);
    }
}
