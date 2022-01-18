// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as sc from "./index";

const hex = "0123456789abcdef";

let localStateMustIncrement: boolean = false;

export function funcInit(ctx: wasmlib.ScFuncContext, f: sc.InitContext): void {
    if (f.params.counter().exists()) {
        let counter = f.params.counter().value();
        f.state.counter().setValue(counter);
    }
}

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

export function funcTestVliCodec(ctx: wasmlib.ScFuncContext, f: sc.TestVliCodecContext): void {
    for (let i: i64 = -1000000; i < 1000000; i++) {
        let enc = new wasmlib.BytesEncoder();
        enc.int64(i);
        let buf = enc.data();
        // let txt = i.toString() + " -";
        // for (let j = 0; j < buf.length; j++) {
        //     let b = buf[j];
        //     txt += " " + hex.charAt((b >> 4) & 0x0f) + hex.charAt(b & 0x0f);
        // }
        let dec = new wasmlib.BytesDecoder(buf);
        let v = dec.int64();
        // txt += " - " + v.toString();
        // ctx.log(txt);
        ctx.require(i == v, "coder value mismatch")
    }

    vliSave(ctx, "v-129", -129);
    vliSave(ctx, "v-128", -128);
    vliSave(ctx, "v-127", -127);
    vliSave(ctx, "v-126", -126);
    vliSave(ctx, "v-65", -65);
    vliSave(ctx, "v-64", -64);
    vliSave(ctx, "v-63", -63);
    vliSave(ctx, "v-62", -62);
    vliSave(ctx, "v-2", -2);
    vliSave(ctx, "v-1", -1);
    vliSave(ctx, "v 0", 0);
    vliSave(ctx, "v+1", 1);
    vliSave(ctx, "v+2", 2);
    vliSave(ctx, "v+62", 62);
    vliSave(ctx, "v+63", 63);
    vliSave(ctx, "v+64", 64);
    vliSave(ctx, "v+65", 65);
    vliSave(ctx, "v+126", 126);
    vliSave(ctx, "v+127", 127);
    vliSave(ctx, "v+128", 128);
    vliSave(ctx, "v+129", 129);
}

export function funcTestVluCodec(ctx: wasmlib.ScFuncContext, f: sc.TestVluCodecContext): void {
    for (let i: u64 = 0; i < 2000000; i++) {
        let enc = new wasmlib.BytesEncoder();
        enc.uint64(i);
        let buf = enc.data();
        // let txt = i.toString() + " -";
        // for (let j = 0; j < buf.length; j++) {
        //     let b = buf[j];
        //     txt += " " + hex.charAt((b >> 4) & 0x0f) + hex.charAt(b & 0x0f);
        // }
        let dec = new wasmlib.BytesDecoder(buf);
        let v = dec.uint64();
        // txt += " - " + v.toString();
        // ctx.log(txt);
        ctx.require(i == v, "coder value mismatch")
    }

    vluSave(ctx, "v 0", 0);
    vluSave(ctx, "v+1", 1);
    vluSave(ctx, "v+2", 2);
    vluSave(ctx, "v+62", 62);
    vluSave(ctx, "v+63", 63);
    vluSave(ctx, "v+64", 64);
    vluSave(ctx, "v+65", 65);
    vluSave(ctx, "v+126", 126);
    vluSave(ctx, "v+127", 127);
    vluSave(ctx, "v+128", 128);
    vluSave(ctx, "v+129", 129);
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

export function viewGetVli(ctx: wasmlib.ScViewContext, f: sc.GetVliContext): void {
    let enc = new wasmlib.BytesEncoder();
    let n = f.params.ni64().value();
    enc = enc.int64(n);
    let buf = enc.data();
    let dec = new wasmlib.BytesDecoder(buf);
    let x = dec.int64();

    let str = n.toString() + " -";
    for (let j = 0; j < buf.length; j++) {
        let b = buf[j];
        str += " " + hex.charAt((b >> 4) & 0x0f) + hex.charAt(b & 0x0f);
    }
    str += " - " + x.toString();

    f.results.ni64().setValue(n);
    f.results.xi64().setValue(x);
    f.results.str().setValue(str);
    f.results.buf().setValue(buf);
}

export function viewGetVlu(ctx: wasmlib.ScViewContext, f: sc.GetVluContext): void {
    let enc = new wasmlib.BytesEncoder();
    let n = f.params.nu64().value();
    enc = enc.uint64(n);
    let buf = enc.data();
    let dec = new wasmlib.BytesDecoder(buf);
    let x = dec.uint64();

    let str = n.toString() + " -";
    for (let j = 0; j < buf.length; j++) {
        let b = buf[j];
        str += " " + hex.charAt((b >> 4) & 0x0f) + hex.charAt(b & 0x0f);
    }
    str += " - " + x.toString();

    f.results.nu64().setValue(n);
    f.results.xu64().setValue(x);
    f.results.str().setValue(str);
    f.results.buf().setValue(buf);
}

//////////////////////////////// util funcs \\\\\\\\\\\\\\\\\\\\\\\\\\\\\

function localStatePost(ctx: wasmlib.ScFuncContext, nr: i64): void {
    //note: we add a dummy parameter here to prevent "duplicate outputs not allowed" error
    let f = sc.ScFuncs.whenMustIncrement(ctx);
    f.params.dummy().setValue(nr);
    f.func.transferIotas(1).post();
}

function vliSave(ctx: wasmlib.ScFuncContext, name: string, value: i64): void {
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

function vluSave(ctx: wasmlib.ScFuncContext, name: string, value: u64): void {
    let encoder = new wasmlib.BytesEncoder();
    encoder.uint64(value);
    let spot = ctx.state().getBytes(wasmlib.Key32.fromString(name));
    spot.setValue(encoder.data());

    let bytes = spot.value();
    let decoder = new wasmlib.BytesDecoder(bytes);
    let retrieved = decoder.uint64();
    if (retrieved != value) {
        ctx.log(name.toString() + " in : " + value.toString());
        ctx.log(name.toString() + " out: " + retrieved.toString());
    }
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
