// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testwasmlib

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/coreblocklog"
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
		color := wasmlib.NewScColorFromBytes([]byte("RedGreenBlueYellowCyanBlackWhite"))
		ctx.Require(f.Params.Color().Value() == color, "mismatch: Color")
	}
	if f.Params.Hash().Exists() {
		hash := wasmlib.NewScHashFromBytes([]byte("0123456789abcdeffedcba9876543210"))
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
		requestID := wasmlib.NewScRequestIDFromBytes([]byte("abcdefghijklmnopqrstuvwxyz123456\x00\x00"))
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

func funcArrayClear(ctx wasmlib.ScFuncContext, f *ArrayClearContext) {
	name := f.Params.Name().Value()
	array := f.State.Arrays().GetStringArray(name)
	array.Clear()
}

func funcArrayCreate(ctx wasmlib.ScFuncContext, f *ArrayCreateContext) {
	name := f.Params.Name().Value()
	array := f.State.Arrays().GetStringArray(name)
	array.Clear()
}

func funcArraySet(ctx wasmlib.ScFuncContext, f *ArraySetContext) {
	name := f.Params.Name().Value()
	array := f.State.Arrays().GetStringArray(name)
	index := f.Params.Index().Value()
	value := f.Params.Value().Value()
	array.GetString(index).SetValue(value)
}

func viewArrayLength(ctx wasmlib.ScViewContext, f *ArrayLengthContext) {
	name := f.Params.Name().Value()
	array := f.State.Arrays().GetStringArray(name)
	length := array.Length()
	f.Results.Length().SetValue(length)
}

func viewArrayValue(ctx wasmlib.ScViewContext, f *ArrayValueContext) {
	name := f.Params.Name().Value()
	array := f.State.Arrays().GetStringArray(name)
	index := f.Params.Index().Value()
	value := array.GetString(index).Value()
	f.Results.Value().SetValue(value)
}

func viewIotaBalance(ctx wasmlib.ScViewContext, f *IotaBalanceContext) {
	f.Results.Iotas().SetValue(ctx.Balances().Balance(wasmlib.IOTA))
}

func funcRandom(ctx wasmlib.ScFuncContext, f *RandomContext) {
	f.State.Random().SetValue(ctx.Random(1000))
}

func viewGetRandom(ctx wasmlib.ScViewContext, f *GetRandomContext) {
	f.Results.Random().SetValue(f.State.Random().Value())
}

func funcTriggerEvent(ctx wasmlib.ScFuncContext, f *TriggerEventContext) {
	f.Events.Test(f.Params.Address().Value(), f.Params.Name().Value())
}

func funcMapClear(ctx wasmlib.ScFuncContext, f *MapClearContext) {
	name := f.Params.Name().Value()
	myMap := f.State.Maps().GetStringMap(name)
	myMap.Clear()
}

func funcMapCreate(ctx wasmlib.ScFuncContext, f *MapCreateContext) {
	name := f.Params.Name().Value()
	myMap := f.State.Maps().GetStringMap(name)
	myMap.Clear()
}

func funcMapSet(ctx wasmlib.ScFuncContext, f *MapSetContext) {
	name := f.Params.Name().Value()
	myMap := f.State.Maps().GetStringMap(name)
	key := f.Params.Key().Value()
	value := f.Params.Value().Value()
	myMap.GetString(key).SetValue(value)
}

func viewMapValue(ctx wasmlib.ScViewContext, f *MapValueContext) {
	name := f.Params.Name().Value()
	myMap := f.State.Maps().GetStringMap(name)
	key := f.Params.Key().Value()
	f.Results.Value().SetValue(myMap.GetString(key).Value())
}
