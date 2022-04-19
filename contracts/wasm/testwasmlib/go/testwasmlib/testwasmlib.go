// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testwasmlib

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreblocklog"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

func funcParamTypes(ctx wasmlib.ScFuncContext, f *ParamTypesContext) {
	if f.Params.Address().Exists() {
		ctx.Require(f.Params.Address().Value() == ctx.AccountID().Address(), "mismatch: Address")
	}
	if f.Params.AgentID().Exists() {
		ctx.Require(f.Params.AgentID().Value() == ctx.AccountID(), "mismatch: AgentID")
	}
	if f.Params.Bool().Exists() {
		ctx.Require(f.Params.Bool().Value(), "mismatch: Bool")
	}
	if f.Params.Bytes().Exists() {
		byteData := []byte("these are bytes")
		ctx.Require(bytes.Equal(f.Params.Bytes().Value(), byteData), "mismatch: Bytes")
	}
	if f.Params.ChainID().Exists() {
		ctx.Require(f.Params.ChainID().Value() == ctx.ChainID(), "mismatch: ChainID")
	}
	if f.Params.Color().Exists() {
		color := wasmtypes.ColorFromBytes([]byte("RedGreenBlueYellowCyanBlackWhite"))
		ctx.Require(f.Params.Color().Value() == color, "mismatch: Color")
	}
	if f.Params.Hash().Exists() {
		hash := wasmtypes.HashFromBytes([]byte("0123456789abcdeffedcba9876543210"))
		ctx.Require(f.Params.Hash().Value() == hash, "mismatch: Hash")
	}
	if f.Params.Hname().Exists() {
		ctx.Require(f.Params.Hname().Value() == ctx.AccountID().Hname(), "mismatch: Hname")
	}
	if f.Params.Int8().Exists() {
		ctx.Require(f.Params.Int8().Value() == -123, "mismatch: Int8")
	}
	if f.Params.Int16().Exists() {
		ctx.Require(f.Params.Int16().Value() == -12345, "mismatch: Int16")
	}
	if f.Params.Int32().Exists() {
		ctx.Require(f.Params.Int32().Value() == -1234567890, "mismatch: Int32")
	}
	if f.Params.Int64().Exists() {
		ctx.Require(f.Params.Int64().Value() == -1234567890123456789, "mismatch: Int64")
	}
	if f.Params.RequestID().Exists() {
		requestID := wasmtypes.RequestIDFromBytes([]byte("abcdefghijklmnopqrstuvwxyz123456\x00\x00"))
		ctx.Require(f.Params.RequestID().Value() == requestID, "mismatch: RequestID")
	}
	if f.Params.String().Exists() {
		ctx.Require(f.Params.String().Value() == "this is a string", "mismatch: String")
	}
	if f.Params.Uint8().Exists() {
		ctx.Require(f.Params.Uint8().Value() == 123, "mismatch: Uint8")
	}
	if f.Params.Uint16().Exists() {
		ctx.Require(f.Params.Uint16().Value() == 12345, "mismatch: Uint16")
	}
	if f.Params.Uint32().Exists() {
		ctx.Require(f.Params.Uint32().Value() == 1234567890, "mismatch: Uint32")
	}
	if f.Params.Uint64().Exists() {
		ctx.Require(f.Params.Uint64().Value() == 1234567890123456789, "mismatch: Uint64")
	}
}

func funcRandom(ctx wasmlib.ScFuncContext, f *RandomContext) {
	f.State.Random().SetValue(ctx.Random(1000))
}

func funcTriggerEvent(ctx wasmlib.ScFuncContext, f *TriggerEventContext) {
	f.Events.Test(f.Params.Address().Value(), f.Params.Name().Value())
}

func viewBlockRecord(ctx wasmlib.ScViewContext, f *BlockRecordContext) {
	records := coreblocklog.ScFuncs.GetRequestReceiptsForBlock(ctx)
	records.Params.BlockIndex().SetValue(f.Params.BlockIndex().Value())
	records.Func.Call()
	recordIndex := f.Params.RecordIndex().Value()
	ctx.Require(recordIndex < records.Results.RequestRecord().Length(), "invalid recordIndex")
	f.Results.Record().SetValue(records.Results.RequestRecord().GetBytes(recordIndex).Value())
}

func viewBlockRecords(ctx wasmlib.ScViewContext, f *BlockRecordsContext) {
	records := coreblocklog.ScFuncs.GetRequestReceiptsForBlock(ctx)
	records.Params.BlockIndex().SetValue(f.Params.BlockIndex().Value())
	records.Func.Call()
	f.Results.Count().SetValue(records.Results.RequestRecord().Length())
}

func viewGetRandom(ctx wasmlib.ScViewContext, f *GetRandomContext) {
	f.Results.Random().SetValue(f.State.Random().Value())
}

func viewIotaBalance(ctx wasmlib.ScViewContext, f *IotaBalanceContext) {
	f.Results.Iotas().SetValue(ctx.Balances().Balance(wasmtypes.IOTA))
}

//////////////////// array of array \\\\\\\\\\\\\\\\\\\\

func funcArrayOfArraysAppend(ctx wasmlib.ScFuncContext, f *ArrayOfArraysAppendContext) {
	index := f.Params.Index().Value()
	valLen := f.Params.Value().Length()

	var sa ArrayOfMutableString
	if f.State.StringArrayOfArrays().Length() <= index {
		sa = f.State.StringArrayOfArrays().AppendStringArray()
	} else {
		sa = f.State.StringArrayOfArrays().GetStringArray(index)
	}

	for i := uint32(0); i < valLen; i++ {
		elt := f.Params.Value().GetString(i).Value()
		sa.AppendString().SetValue(elt)
	}
}

func funcArrayOfArraysClear(ctx wasmlib.ScFuncContext, f *ArrayOfArraysClearContext) {
	length := f.State.StringArrayOfArrays().Length()
	for i := uint32(0); i < length; i++ {
		array := f.State.StringArrayOfArrays().GetStringArray(i)
		array.Clear()
	}
	f.State.StringArrayOfArrays().Clear()
}

func funcArrayOfArraysSet(ctx wasmlib.ScFuncContext, f *ArrayOfArraysSetContext) {
	index0 := f.Params.Index0().Value()
	index1 := f.Params.Index1().Value()
	array := f.State.StringArrayOfArrays().GetStringArray(index0)
	value := f.Params.Value().Value()
	array.GetString(index1).SetValue(value)
}

func viewArrayOfArraysLength(ctx wasmlib.ScViewContext, f *ArrayOfArraysLengthContext) {
	length := f.State.StringArrayOfArrays().Length()
	f.Results.Length().SetValue(length)
}

func viewArrayOfArraysValue(ctx wasmlib.ScViewContext, f *ArrayOfArraysValueContext) {
	index0 := f.Params.Index0().Value()
	index1 := f.Params.Index1().Value()

	elt := f.State.StringArrayOfArrays().GetStringArray(index0).GetString(index1).Value()
	f.Results.Value().SetValue(elt)
}

//////////////////// array of map \\\\\\\\\\\\\\\\\\\\

func funcArrayOfMapsClear(ctx wasmlib.ScFuncContext, f *ArrayOfMapsClearContext) {
	length := f.State.StringArrayOfArrays().Length()
	for i := uint32(0); i < length; i++ {
		mmap := f.State.StringArrayOfMaps().GetStringMap(i)
		mmap.Clear()
	}
	f.State.StringArrayOfMaps().Clear()
}

func funcArrayOfMapsSet(ctx wasmlib.ScFuncContext, f *ArrayOfMapsSetContext) {
	index := f.Params.Index().Value()
	value := f.Params.Value().Value()
	key := f.Params.Key().Value()
	if f.State.StringArrayOfMaps().Length() <= index {
		mmap := f.State.StringArrayOfMaps().AppendStringMap()
		mmap.GetString(key).SetValue(value)
		return
	}
	mmap := f.State.StringArrayOfMaps().GetStringMap(index)
	mmap.GetString(key).SetValue(value)
}

func viewArrayOfMapsValue(ctx wasmlib.ScViewContext, f *ArrayOfMapsValueContext) {
	index := f.Params.Index().Value()
	key := f.Params.Key().Value()
	mmap := f.State.StringArrayOfMaps().GetStringMap(index)
	f.Results.Value().SetValue(mmap.GetString(key).Value())
}

//////////////////// map of array \\\\\\\\\\\\\\\\\\\\

func funcMapOfArraysAppend(ctx wasmlib.ScFuncContext, f *MapOfArraysAppendContext) {
	name := f.Params.Name().Value()
	array := f.State.StringMapOfArrays().GetStringArray(name)
	value := f.Params.Value().Value()
	array.AppendString().SetValue(value)
}

func funcMapOfArraysClear(ctx wasmlib.ScFuncContext, f *MapOfArraysClearContext) {
	name := f.Params.Name().Value()
	array := f.State.StringMapOfArrays().GetStringArray(name)
	array.Clear()
}

func funcMapOfArraysSet(ctx wasmlib.ScFuncContext, f *MapOfArraysSetContext) {
	name := f.Params.Name().Value()
	array := f.State.StringMapOfArrays().GetStringArray(name)
	index := f.Params.Index().Value()
	value := f.Params.Value().Value()
	array.GetString(index).SetValue(value)
}

func viewMapOfArraysLength(ctx wasmlib.ScViewContext, f *MapOfArraysLengthContext) {
	name := f.Params.Name().Value()
	array := f.State.StringMapOfArrays().GetStringArray(name)
	length := array.Length()
	f.Results.Length().SetValue(length)
}

func viewMapOfArraysValue(ctx wasmlib.ScViewContext, f *MapOfArraysValueContext) {
	name := f.Params.Name().Value()
	array := f.State.StringMapOfArrays().GetStringArray(name)
	index := f.Params.Index().Value()
	value := array.GetString(index).Value()
	f.Results.Value().SetValue(value)
}

//////////////////// map of map \\\\\\\\\\\\\\\\\\\\

func funcMapOfMapsClear(ctx wasmlib.ScFuncContext, f *MapOfMapsClearContext) {
	name := f.Params.Name().Value()
	mmap := f.State.StringMapOfMaps().GetStringMap(name)
	mmap.Clear()
}

func funcMapOfMapsSet(ctx wasmlib.ScFuncContext, f *MapOfMapsSetContext) {
	name := f.Params.Name().Value()
	mmap := f.State.StringMapOfMaps().GetStringMap(name)
	key := f.Params.Key().Value()
	value := f.Params.Value().Value()
	mmap.GetString(key).SetValue(value)
}

func viewMapOfMapsValue(ctx wasmlib.ScViewContext, f *MapOfMapsValueContext) {
	name := f.Params.Name().Value()
	mmap := f.State.StringMapOfMaps().GetStringMap(name)
	key := f.Params.Key().Value()
	f.Results.Value().SetValue(mmap.GetString(key).Value())
}
