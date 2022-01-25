// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package inccounter

import (
	"strconv"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmtypes"
)

const hex = "0123456789abcdef"

var LocalStateMustIncrement = false

func funcInit(ctx wasmlib.ScFuncContext, f *InitContext) {
	if f.Params.Counter().Exists() {
		counter := f.Params.Counter().Value()
		f.State.Counter().SetValue(counter)
	}
}

func funcCallIncrement(ctx wasmlib.ScFuncContext, f *CallIncrementContext) {
	counter := f.State.Counter()
	value := counter.Value()
	counter.SetValue(value + 1)
	if value == 0 {
		ScFuncs.CallIncrement(ctx).Func.Call()
	}
}

func funcCallIncrementRecurse5x(ctx wasmlib.ScFuncContext, f *CallIncrementRecurse5xContext) {
	counter := f.State.Counter()
	value := counter.Value()
	counter.SetValue(value + 1)
	if value < 5 {
		ScFuncs.CallIncrementRecurse5x(ctx).Func.Call()
	}
}

//nolint:unparam
func funcEndlessLoop(ctx wasmlib.ScFuncContext, f *EndlessLoopContext) {
	//nolint:staticcheck
	for {
		// intentional endless loop to see if Wasm VM can be interrupted
	}
}

func funcIncrement(ctx wasmlib.ScFuncContext, f *IncrementContext) {
	counter := f.State.Counter()
	counter.SetValue(counter.Value() + 1)
}

func funcIncrementWithDelay(ctx wasmlib.ScFuncContext, f *IncrementWithDelayContext) {
	delay := f.Params.Delay().Value()
	inc := ScFuncs.CallIncrement(ctx)
	inc.Func.Delay(delay).TransferIotas(1).Post()
}

func funcLocalStateInternalCall(ctx wasmlib.ScFuncContext, f *LocalStateInternalCallContext) {
	LocalStateMustIncrement = false
	whenMustIncrementState(ctx, f.State)
	LocalStateMustIncrement = true
	whenMustIncrementState(ctx, f.State)
	whenMustIncrementState(ctx, f.State)
	// counter ends up as 2
}

//nolint:unparam
func funcLocalStatePost(ctx wasmlib.ScFuncContext, f *LocalStatePostContext) {
	LocalStateMustIncrement = false
	// prevent multiple identical posts, need a dummy param to differentiate them
	localStatePost(ctx, 1)
	LocalStateMustIncrement = true
	localStatePost(ctx, 2)
	localStatePost(ctx, 3)
	// counter ends up as 0
}

//nolint:unparam
func funcLocalStateSandboxCall(ctx wasmlib.ScFuncContext, f *LocalStateSandboxCallContext) {
	LocalStateMustIncrement = false
	ScFuncs.WhenMustIncrement(ctx).Func.Call()
	LocalStateMustIncrement = true
	ScFuncs.WhenMustIncrement(ctx).Func.Call()
	ScFuncs.WhenMustIncrement(ctx).Func.Call()
	// counter ends up as 0
}

func funcPostIncrement(ctx wasmlib.ScFuncContext, f *PostIncrementContext) {
	counter := f.State.Counter()
	value := counter.Value()
	counter.SetValue(value + 1)
	if value == 0 {
		ScFuncs.PostIncrement(ctx).Func.TransferIotas(1).Post()
	}
}

func funcRepeatMany(ctx wasmlib.ScFuncContext, f *RepeatManyContext) {
	counter := f.State.Counter()
	value := counter.Value()
	counter.SetValue(value + 1)
	stateRepeats := f.State.NumRepeats()
	repeats := f.Params.NumRepeats().Value()
	if repeats == 0 {
		repeats = stateRepeats.Value()
		if repeats == 0 {
			return
		}
	}
	stateRepeats.SetValue(repeats - 1)
	ScFuncs.RepeatMany(ctx).Func.TransferIotas(1).Post()
}

//nolint:unparam
func funcTestVliCodec(ctx wasmlib.ScFuncContext, f *TestVliCodecContext) {
	vliSave(ctx, "v-129", -129)
	vliSave(ctx, "v-128", -128)
	vliSave(ctx, "v-127", -127)
	vliSave(ctx, "v-126", -126)
	vliSave(ctx, "v-65", -65)
	vliSave(ctx, "v-64", -64)
	vliSave(ctx, "v-63", -63)
	vliSave(ctx, "v-62", -62)
	vliSave(ctx, "v-2", -2)
	vliSave(ctx, "v-1", -1)
	vliSave(ctx, "v 0", 0)
	vliSave(ctx, "v+1", 1)
	vliSave(ctx, "v+2", 2)
	vliSave(ctx, "v+62", 62)
	vliSave(ctx, "v+63", 63)
	vliSave(ctx, "v+64", 64)
	vliSave(ctx, "v+65", 65)
	vliSave(ctx, "v+126", 126)
	vliSave(ctx, "v+127", 127)
	vliSave(ctx, "v+128", 128)
	vliSave(ctx, "v+129", 129)
}

//nolint:unparam
func funcTestVluCodec(ctx wasmlib.ScFuncContext, f *TestVluCodecContext) {
	vluSave(ctx, "v 0", 0)
	vluSave(ctx, "v+1", 1)
	vluSave(ctx, "v+2", 2)
	vluSave(ctx, "v+62", 62)
	vluSave(ctx, "v+63", 63)
	vluSave(ctx, "v+64", 64)
	vluSave(ctx, "v+65", 65)
	vluSave(ctx, "v+126", 126)
	vluSave(ctx, "v+127", 127)
	vluSave(ctx, "v+128", 128)
	vluSave(ctx, "v+129", 129)
}

func funcWhenMustIncrement(ctx wasmlib.ScFuncContext, f *WhenMustIncrementContext) {
	whenMustIncrementState(ctx, f.State)
}

// note that getCounter mirrors the state of the 'counter' state variable
// which means that if the state variable was not present it also will not be present in the result
func viewGetCounter(ctx wasmlib.ScViewContext, f *GetCounterContext) {
	counter := f.State.Counter()
	if counter.Exists() {
		f.Results.Counter().SetValue(counter.Value())
	}
}

//nolint:dupl
func viewGetVli(ctx wasmlib.ScViewContext, f *GetVliContext) {
	enc := wasmcodec.NewWasmEncoder()
	n := f.Params.Ni64().Value()
	buf := enc.VliEncode(n).Buf()
	dec := wasmcodec.NewWasmDecoder(buf)
	x := wasmtypes.DecodeInt64(dec)

	str := strconv.FormatInt(n, 10) + " -"
	for j := 0; j < len(buf); j++ {
		b := buf[j]
		str += " " + string(append([]byte(nil), hex[(b>>4)&0x0f], hex[b&0x0f]))
	}
	str += " - " + strconv.FormatInt(x, 10)

	f.Results.Ni64().SetValue(n)
	f.Results.Xi64().SetValue(x)
	f.Results.Str().SetValue(str)
	f.Results.Buf().SetValue(buf)
}

//nolint:dupl
func viewGetVlu(ctx wasmlib.ScViewContext, f *GetVluContext) {
	enc := wasmcodec.NewWasmEncoder()
	n := f.Params.Nu64().Value()
	buf := enc.VluEncode(n).Buf()
	dec := wasmcodec.NewWasmDecoder(buf)
	x := wasmtypes.DecodeUint64(dec)

	str := strconv.FormatUint(n, 10) + " -"
	for j := 0; j < len(buf); j++ {
		b := buf[j]
		str += " " + string(append([]byte(nil), hex[(b>>4)&0x0f], hex[b&0x0f]))
	}
	str += " - " + strconv.FormatUint(x, 10)

	f.Results.Nu64().SetValue(n)
	f.Results.Xu64().SetValue(x)
	f.Results.Str().SetValue(str)
	f.Results.Buf().SetValue(buf)
}

//////////////////////////////// util funcs \\\\\\\\\\\\\\\\\\\\\\\\\\\\\

func localStatePost(ctx wasmlib.ScFuncContext, nr int64) {
	// note: we add a dummy parameter here to prevent "duplicate outputs not allowed" error
	f := ScFuncs.WhenMustIncrement(ctx)
	f.Params.Dummy().SetValue(nr)
	f.Func.TransferIotas(1).Post()
}

func vliSave(ctx wasmlib.ScFuncContext, name string, value int64) {
	enc := wasmcodec.NewWasmEncoder()
	state := wasmlib.ScState{}
	state.Set([]byte(name), enc.VliEncode(value).Buf())

	buf := state.Get([]byte(name))
	dec := wasmcodec.NewWasmDecoder(buf)
	retrieved := wasmtypes.DecodeInt64(dec)
	if retrieved != value {
		ctx.Log(name + " in : " + ctx.Utility().String(value))
		ctx.Log(name + " out: " + ctx.Utility().String(retrieved))
	}
}

func vluSave(ctx wasmlib.ScFuncContext, name string, value uint64) {
	enc := wasmcodec.NewWasmEncoder()
	state := wasmlib.ScState{}
	state.Set([]byte(name), enc.VluEncode(value).Buf())

	buf := state.Get([]byte(name))
	dec := wasmcodec.NewWasmDecoder(buf)
	retrieved := wasmtypes.DecodeUint64(dec)
	if retrieved != value {
		ctx.Log(name + " in : " + ctx.Utility().String(int64(value)))
		ctx.Log(name + " out: " + ctx.Utility().String(int64(retrieved)))
	}
}

func whenMustIncrementState(ctx wasmlib.ScFuncContext, state MutableIncCounterState) {
	ctx.Log("when_must_increment called")
	if !LocalStateMustIncrement {
		return
	}
	counter := state.Counter()
	counter.SetValue(counter.Value() + 1)
}
