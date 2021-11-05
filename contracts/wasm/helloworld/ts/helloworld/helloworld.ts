// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as sc from "./index";

export function funcHelloWorld(ctx: wasmlib.ScFuncContext, f: sc.HelloWorldContext): void {
    ctx.log("Hello, world!");
}

export function viewGetHelloWorld(ctx: wasmlib.ScViewContext, f: sc.GetHelloWorldContext): void {
    f.results.helloWorld().setValue("Hello, world!");
}
