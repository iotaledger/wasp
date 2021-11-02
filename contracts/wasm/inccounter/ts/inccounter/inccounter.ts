// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as sc from "./index";

let localStateMustIncrement: boolean = false;

export function funcCallIncrement(ctx: wasmlib.ScFuncContext, f: sc.CallIncrementContext): void {
    let counter = f.state.counter();
    let value = counter.value();
    counter.setValue(value + 1);
    if (value == 0) {
        sc.ScFuncs.callIncrement(ctx).func.call();
    }
}

export function funcCallIncrementRecurse5x(ctx: wasmlib.ScFuncContext, f: sc.CallIncrementRecurse5xContext): void {
    let counter = f.state.counter();
    let value = counter.value();
    counter.setValue(value + 1);
    if (value < 5) {
        sc.ScFuncs.callIncrementRecurse5x(ctx).func.call();
    }
}

export function funcEndlessLoop(ctx: wasmlib.ScFuncContext, f: sc.EndlessLoopContext): void {
    for (; ;) {
    }
}

export function funcIncrement(ctx: wasmlib.ScFuncContext, f: sc.IncrementContext): void {
    let counter = f.state.counter();
    counter.setValue(counter.value() + 1);
}

export function funcIncrementWithDelay(ctx: wasmlib.ScFuncContext, f: sc.IncrementWithDelayContext): void {
    let delay = f.params.delay().value();
    let inc = sc.ScFuncs.callIncrement(ctx);
    inc.func.delay(delay).transferIotas(1).post();
}

export function funcInit(ctx: wasmlib.ScFuncContext, f: sc.InitContext): void {
    if (f.params.counter().exists()) {
        let counter = f.params.counter().value();
        f.state.counter().setValue(counter);
    }
}

export function funcLocalStateInternalCall(ctx: wasmlib.ScFuncContext, f: sc.LocalStateInternalCallContext): void {
    localStateMustIncrement = false;
    whenMustIncrementState(ctx, f.state);
    localStateMustIncrement = true;
    whenMustIncrementState(ctx, f.state);
    whenMustIncrementState(ctx, f.state);
    // counter ends up as 2
}

export function funcLocalStatePost(ctx: wasmlib.ScFuncContext, f: sc.LocalStatePostContext): void {
    localStateMustIncrement = false;
    // prevent multiple identical posts, need a dummy param to differentiate them
    localStatePost(ctx, 1);
    localStateMustIncrement = true;
    localStatePost(ctx, 2);
    localStatePost(ctx, 3);
    // counter ends up as 0
}

export function funcLocalStateSandboxCall(ctx: wasmlib.ScFuncContext, f: sc.LocalStateSandboxCallContext): void {
    localStateMustIncrement = false;
    sc.ScFuncs.whenMustIncrement(ctx).func.call();
    localStateMustIncrement = true;
    sc.ScFuncs.whenMustIncrement(ctx).func.call();
    sc.ScFuncs.whenMustIncrement(ctx).func.call();
    // counter ends up as 0
}

export function funcPostIncrement(ctx: wasmlib.ScFuncContext, f: sc.PostIncrementContext): void {
    let counter = f.state.counter();
    let value = counter.value();
    counter.setValue(value + 1);
    if (value == 0) {
        sc.ScFuncs.increment(ctx).func.transferIotas(1).post();
    }
}

export function funcRepeatMany(ctx: wasmlib.ScFuncContext, f: sc.RepeatManyContext): void {
    let counter = f.state.counter();
    let value = counter.value();
    counter.setValue(value + 1);
    let stateRepeats = f.state.numRepeats();
    let repeats = f.params.numRepeats().value();
    if (repeats == 0) {
        repeats = stateRepeats.value();
        if (repeats == 0) {
            return;
        }
    }
    stateRepeats.setValue(repeats - 1);
    sc.ScFuncs.repeatMany(ctx).func.transferIotas(1).post();
}

let hex = "0123456789abcdef";

export function funcTestLeb128(ctx: wasmlib.ScFuncContext, f: sc.TestLeb128Context): void {
    for (let i: i64 = -1000000; i < 1000000; i++) {
        let d = new wasmlib.BytesEncoder();
        d.int64(i);
        // let txt = i.toString() + " -";
        // for (let j = 0; j < d.buf.length; j++) {
        //     let b = d.buf[j];
        //     txt += " " + hex[(b >> 4) & 0x0f] + hex[d.buf[j] & 0x0f];
        // }
        let e = new wasmlib.BytesDecoder(d.buf);
        let v = e.int64();
        // txt += " - " + v.toString();
        // ctx.log(txt);
        ctx.require(i == v, "coder value mismatch")
    }

    leb128Save(ctx, "v-1", -1);
    leb128Save(ctx, "v-2", -2);
    leb128Save(ctx, "v-126", -126);
    leb128Save(ctx, "v-127", -127);
    leb128Save(ctx, "v-128", -128);
    leb128Save(ctx, "v-129", -129);
    leb128Save(ctx, "v0", 0);
    leb128Save(ctx, "v+1", 1);
    leb128Save(ctx, "v+2", 2);
    leb128Save(ctx, "v+126", 126);
    leb128Save(ctx, "v+127", 127);
    leb128Save(ctx, "v+128", 128);
    leb128Save(ctx, "v+129", 129);
}

export function funcWhenMustIncrement(ctx: wasmlib.ScFuncContext, f: sc.WhenMustIncrementContext): void {
    whenMustIncrementState(ctx, f.state);
}

// note that getCounter mirrors the state of the 'counter' state variable
// which means that if the state variable was not present it also will not be present in the result
export function viewGetCounter(ctx: wasmlib.ScViewContext, f: sc.GetCounterContext): void {
    let counter = f.state.counter();
    if (counter.exists()) {
        f.results.counter().setValue(counter.value());
    }
}

function leb128Save(ctx: wasmlib.ScFuncContext, name: string, value: i64): void {
    let encoder = new wasmlib.BytesEncoder();
    encoder.int64(value);
    let spot = ctx.state().getBytes(wasmlib.Key32.fromString(name));
    spot.setValue(encoder.data());

    let bytes = spot.value();
    let decoder = new wasmlib.BytesDecoder(bytes);
    let retrieved = decoder.int64();
    if (retrieved != value) {
        ctx.log(name.toString() + " in : " + value.toString());
        ctx.log(name.toString() + " out: " + retrieved.toString());
    }
}

function localStatePost(ctx: wasmlib.ScFuncContext, nr: i64): void {
    //note: we add a dummy parameter here to prevent "duplicate outputs not allowed" error
    let f = sc.ScFuncs.whenMustIncrement(ctx);
    f.params.dummy().setValue(nr);
    f.func.transferIotas(1).post();
}

function whenMustIncrementState(ctx: wasmlib.ScFuncContext, state: sc.MutableIncCounterState): void {
    ctx.log("whenMustIncrement called");
    if (!localStateMustIncrement) {
        return;
    }
    let counter = state.counter();
    counter.setValue(counter.value() + 1);
    ctx.log("whenMustIncrement incremented");
}
