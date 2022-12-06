// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;
use inccounter::*;
use crate::*;

const HEX: &str = "0123456789abcdef";

static mut LOCAL_STATE_MUST_INCREMENT: bool = false;

pub fn func_init(_ctx: &ScFuncContext, f: &InitContext) {
    if f.params.counter().exists() {
        let counter = f.params.counter().value();
        f.state.counter().set_value(counter);
    }
}

pub fn func_call_increment(ctx: &ScFuncContext, f: &CallIncrementContext) {
    let counter = f.state.counter();
    let value = counter.value();
    counter.set_value(value + 1);
    if value == 0 {
        ScFuncs::call_increment(ctx).func.call();
    }
}

pub fn func_call_increment_recurse5x(ctx: &ScFuncContext, f: &CallIncrementRecurse5xContext) {
    let counter = f.state.counter();
    let value = counter.value();
    counter.set_value(value + 1);
    if value < 5 {
        ScFuncs::call_increment_recurse5x(ctx).func.call();
    }
}

pub fn func_endless_loop(_ctx: &ScFuncContext, _f: &EndlessLoopContext) {
    loop {}
}

pub fn func_increment(_ctx: &ScFuncContext, f: &IncrementContext) {
    let counter = f.state.counter();
    counter.set_value(counter.value() + 1);
}

pub fn func_increment_with_delay(ctx: &ScFuncContext, f: &IncrementWithDelayContext) {
    let delay = f.params.delay().value();
    let inc = ScFuncs::call_increment(ctx);
    inc.func.delay(delay).post();
}

pub fn func_local_state_internal_call(ctx: &ScFuncContext, f: &LocalStateInternalCallContext) {
    unsafe {
        LOCAL_STATE_MUST_INCREMENT = false;
    }
    when_must_increment_state(ctx, &f.state);
    unsafe {
        LOCAL_STATE_MUST_INCREMENT = true;
    }
    when_must_increment_state(ctx, &f.state);
    when_must_increment_state(ctx, &f.state);
    // counter ends up as 2
}

pub fn func_local_state_post(ctx: &ScFuncContext, _f: &LocalStatePostContext) {
    unsafe {
        LOCAL_STATE_MUST_INCREMENT = false;
    }
    // prevent multiple identical posts, need a dummy param to differentiate them
    local_state_post(ctx, 1);
    unsafe {
        LOCAL_STATE_MUST_INCREMENT = true;
    }
    local_state_post(ctx, 2);
    local_state_post(ctx, 3);
    // counter ends up as 0
}

pub fn func_local_state_sandbox_call(ctx: &ScFuncContext, _f: &LocalStateSandboxCallContext) {
    unsafe {
        LOCAL_STATE_MUST_INCREMENT = false;
    }
    ScFuncs::when_must_increment(ctx).func.call();
    unsafe {
        LOCAL_STATE_MUST_INCREMENT = true;
    }
    ScFuncs::when_must_increment(ctx).func.call();
    ScFuncs::when_must_increment(ctx).func.call();
    // counter ends up as 0
}

pub fn func_post_increment(ctx: &ScFuncContext, f: &PostIncrementContext) {
    let counter = f.state.counter();
    let value = counter.value();
    counter.set_value(value + 1);
    if value == 0 {
        ScFuncs::increment(ctx).func.post();
    }
}

pub fn func_repeat_many(ctx: &ScFuncContext, f: &RepeatManyContext) {
    let counter = f.state.counter();
    let value = counter.value();
    counter.set_value(value + 1);
    let state_repeats = f.state.num_repeats();
    let mut repeats = f.params.num_repeats().value();
    if repeats == 0 {
        repeats = state_repeats.value();
        if repeats == 0 {
            return;
        }
    }
    state_repeats.set_value(repeats - 1);
    ScFuncs::repeat_many(ctx).func.post();
}

pub fn func_test_vli_codec(ctx: &ScFuncContext, _f: &TestVliCodecContext) {
    vli_save(ctx, "v-129", -129);
    vli_save(ctx, "v-128", -128);
    vli_save(ctx, "v-127", -127);
    vli_save(ctx, "v-126", -126);
    vli_save(ctx, "v-65", -65);
    vli_save(ctx, "v-64", -64);
    vli_save(ctx, "v-63", -63);
    vli_save(ctx, "v-62", -62);
    vli_save(ctx, "v-2", -2);
    vli_save(ctx, "v-1", -1);
    vli_save(ctx, "v 0", 0);
    vli_save(ctx, "v+1", 1);
    vli_save(ctx, "v+2", 2);
    vli_save(ctx, "v+62", 62);
    vli_save(ctx, "v+63", 63);
    vli_save(ctx, "v+64", 64);
    vli_save(ctx, "v+65", 65);
    vli_save(ctx, "v+126", 126);
    vli_save(ctx, "v+127", 127);
    vli_save(ctx, "v+128", 128);
    vli_save(ctx, "v+129", 129);
}

pub fn func_test_vlu_codec(ctx: &ScFuncContext, _f: &TestVluCodecContext) {
    vlu_save(ctx, "v 0", 0);
    vlu_save(ctx, "v+1", 1);
    vlu_save(ctx, "v+2", 2);
    vlu_save(ctx, "v+62", 62);
    vlu_save(ctx, "v+63", 63);
    vlu_save(ctx, "v+64", 64);
    vlu_save(ctx, "v+65", 65);
    vlu_save(ctx, "v+126", 126);
    vlu_save(ctx, "v+127", 127);
    vlu_save(ctx, "v+128", 128);
    vlu_save(ctx, "v+129", 129);
}

pub fn func_when_must_increment(ctx: &ScFuncContext, f: &WhenMustIncrementContext) {
    when_must_increment_state(ctx, &f.state);
}

// note that get_counter mirrors the state of the 'counter' state variable
// which means that if the state variable was not present it also will not be present in the result
pub fn view_get_counter(_ctx: &ScViewContext, f: &GetCounterContext) {
    let counter = f.state.counter();
    if counter.exists() {
        f.results.counter().set_value(counter.value());
    }
}

pub fn view_get_vli(_ctx: &ScViewContext, f: &GetVliContext) {
    let mut enc = WasmEncoder::new();
    let n = f.params.ni64().value();
    int64_encode(&mut enc,n);
    let buf = enc.buf();
    let mut dec = WasmDecoder::new(&buf);
    let x = int64_decode(&mut dec);

    let mut str = n.to_string() + " -";
    for b in &buf {
        let h1 = ((b >> 4) & 0x0f) as usize;
        let h2 = (b & 0x0f) as usize;
        str += &(" ".to_string() + &HEX[h1..h1+1] + &HEX[h2..h2+1]);
    }
    str += &(" - ".to_string() + &x.to_string());

    f.results.ni64().set_value(n);
    f.results.xi64().set_value(x);
    f.results.str().set_value(&str);
    f.results.buf().set_value(&buf);
}

pub fn view_get_vlu(_ctx: &ScViewContext, f: &GetVluContext) {
    let mut enc = WasmEncoder::new();
    let n = f.params.nu64().value();
    uint64_encode(&mut enc, n);
    let buf = enc.buf();
    let mut dec = WasmDecoder::new(&buf);
    let x = uint64_decode(&mut dec);

    let mut str = n.to_string() + " -";
    for b in &buf {
        let h1 = ((b >> 4) & 0x0f) as usize;
        let h2 = (b & 0x0f) as usize;
        str += &(" ".to_string() + &HEX[h1..h1+1] + &HEX[h2..h2+1]);
    }
    str += &(" - ".to_string() + &x.to_string());

    f.results.nu64().set_value(n);
    f.results.xu64().set_value(x);
    f.results.str().set_value(&str);
    f.results.buf().set_value(&buf);
}

//////////////////////////////// util funcs \\\\\\\\\\\\\\\\\\\\\\\\\\\\\

fn local_state_post(ctx: &ScFuncContext, nr: i64) {
    //note: we add a dummy parameter here to prevent "duplicate outputs not allowed" error
    let f = ScFuncs::when_must_increment(ctx);
    f.params.dummy().set_value(nr);
    f.func.post();
}

fn vli_save(ctx: &ScFuncContext, name: &str, value: i64) {
    let mut enc = WasmEncoder::new();
    let state = ctx.raw_state();
    let key = string_to_bytes(name);
    state.set(&key, &enc.vli_encode(value).buf());

    let buf = state.get(&key);
    let mut dec = WasmDecoder::new(&buf);
    let val = dec.vli_decode(64);
    if val != value {
        ctx.log(&(name.to_string() + " in : " + &value.to_string()));
        ctx.log(&(name.to_string() + " out: " + &val.to_string()));
    }
}

fn vlu_save(ctx: &ScFuncContext, name: &str, value: u64) {
    let mut enc = WasmEncoder::new();
    let state = ctx.raw_state();
    let key = string_to_bytes(name);
    state.set(&key, &enc.vlu_encode(value).buf());

    let buf = state.get(&key);
    let mut dec = WasmDecoder::new(&buf);
    let val = dec.vlu_decode(64);
    if val != value {
        ctx.log(&(name.to_string() + " in : " + &value.to_string()));
        ctx.log(&(name.to_string() + " out: " + &val.to_string()));
    }
}

fn when_must_increment_state(ctx: &ScFuncContext, state: &MutableIncCounterState) {
    ctx.log("when_must_increment called");
    unsafe {
        if !LOCAL_STATE_MUST_INCREMENT {
            return;
        }
    }
    let counter = state.counter();
    counter.set_value(counter.value() + 1);
}
