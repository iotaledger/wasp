// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "../wasmlib"
import * as sc from "./index";

export function funcDivide(ctx: wasmlib.ScFuncContext, f: sc.DivideContext): void {
}

export function funcInit(ctx: wasmlib.ScFuncContext, f: sc.InitContext): void {
    if (f.params.owner().exists()) {
        f.state.owner().setValue(f.params.owner().value());
        return;
    }
    f.state.owner().setValue(ctx.contractCreator());
}

export function funcMember(ctx: wasmlib.ScFuncContext, f: sc.MemberContext): void {
}

export function funcSetOwner(ctx: wasmlib.ScFuncContext, f: sc.SetOwnerContext): void {
    f.state.owner().setValue(f.params.owner().value());
}

export function viewGetFactor(ctx: wasmlib.ScViewContext, f: sc.GetFactorContext): void {
}

export function viewGetOwner(ctx: wasmlib.ScViewContext, f: sc.GetOwnerContext): void {
    f.results.owner().setValue(f.state.owner().value());
}
