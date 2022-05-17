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
	// TODO big.Int
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
	if f.Params.NftID().Exists() {
		nftID := wasmtypes.NftIDFromBytes([]byte("abcdefghijklmnopqrstuvwxyz123456"))
		ctx.Require(f.Params.NftID().Value() == nftID, "mismatch: NftID")
	}
	if f.Params.RequestID().Exists() {
		requestID := wasmtypes.RequestIDFromBytes([]byte("abcdefghijklmnopqrstuvwxyz123456\x00\x00"))
		ctx.Require(f.Params.RequestID().Value() == requestID, "mismatch: RequestID")
	}
	if f.Params.String().Exists() {
		ctx.Require(f.Params.String().Value() == "this is a string", "mismatch: String")
	}
	if f.Params.TokenID().Exists() {
		tokenID := wasmtypes.TokenIDFromBytes([]byte("abcdefghijklmnopqrstuvwxyz1234567890\x00\x00"))
		ctx.Require(f.Params.TokenID().Value() == tokenID, "mismatch: TokenID")
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

//nolint:unparam
func funcTakeAllowance(ctx wasmlib.ScFuncContext, f *TakeAllowanceContext) {
	ctx.TransferAllowed(ctx.AccountID(), wasmlib.NewScTransferFromBalances(ctx.Allowance()), false)
	ctx.Log(ctx.Utility().String(int64(ctx.Balances().Iotas())))
}

func funcTakeBalance(ctx wasmlib.ScFuncContext, f *TakeBalanceContext) {
	f.Results.Iotas().SetValue(ctx.Balances().Iotas())
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
	f.Results.Iotas().SetValue(ctx.Balances().Iotas())
}

//////////////////// array of StringArray \\\\\\\\\\\\\\\\\\\\

func funcArrayOfStringArrayAppend(ctx wasmlib.ScFuncContext, f *ArrayOfStringArrayAppendContext) {
	index := f.Params.Index().Value()
	valLen := f.Params.Value().Length()

	var sa ArrayOfMutableString
	if f.State.ArrayOfStringArray().Length() <= index {
		sa = f.State.ArrayOfStringArray().AppendStringArray()
	} else {
		sa = f.State.ArrayOfStringArray().GetStringArray(index)
	}

	for i := uint32(0); i < valLen; i++ {
		elt := f.Params.Value().GetString(i).Value()
		sa.AppendString().SetValue(elt)
	}
}

func funcArrayOfStringArrayClear(ctx wasmlib.ScFuncContext, f *ArrayOfStringArrayClearContext) {
	length := f.State.ArrayOfStringArray().Length()
	for i := uint32(0); i < length; i++ {
		array := f.State.ArrayOfStringArray().GetStringArray(i)
		array.Clear()
	}
	f.State.ArrayOfStringArray().Clear()
}

func funcArrayOfStringArraySet(ctx wasmlib.ScFuncContext, f *ArrayOfStringArraySetContext) {
	index0 := f.Params.Index0().Value()
	index1 := f.Params.Index1().Value()
	array := f.State.ArrayOfStringArray().GetStringArray(index0)
	value := f.Params.Value().Value()
	array.GetString(index1).SetValue(value)
}

func viewArrayOfStringArrayLength(ctx wasmlib.ScViewContext, f *ArrayOfStringArrayLengthContext) {
	length := f.State.ArrayOfStringArray().Length()
	f.Results.Length().SetValue(length)
}

func viewArrayOfStringArrayValue(ctx wasmlib.ScViewContext, f *ArrayOfStringArrayValueContext) {
	index0 := f.Params.Index0().Value()
	index1 := f.Params.Index1().Value()

	elt := f.State.ArrayOfStringArray().GetStringArray(index0).GetString(index1).Value()
	f.Results.Value().SetValue(elt)
}

//////////////////// array of StringMap \\\\\\\\\\\\\\\\\\\\

func funcArrayOfStringMapClear(ctx wasmlib.ScFuncContext, f *ArrayOfStringMapClearContext) {
	length := f.State.ArrayOfStringArray().Length()
	for i := uint32(0); i < length; i++ {
		mmap := f.State.ArrayOfStringMap().GetStringMap(i)
		mmap.Clear()
	}
	f.State.ArrayOfStringMap().Clear()
}

func funcArrayOfStringMapSet(ctx wasmlib.ScFuncContext, f *ArrayOfStringMapSetContext) {
	index := f.Params.Index().Value()
	value := f.Params.Value().Value()
	key := f.Params.Key().Value()
	if f.State.ArrayOfStringMap().Length() <= index {
		mmap := f.State.ArrayOfStringMap().AppendStringMap()
		mmap.GetString(key).SetValue(value)
		return
	}
	mmap := f.State.ArrayOfStringMap().GetStringMap(index)
	mmap.GetString(key).SetValue(value)
}

func viewArrayOfStringMapValue(ctx wasmlib.ScViewContext, f *ArrayOfStringMapValueContext) {
	index := f.Params.Index().Value()
	key := f.Params.Key().Value()
	mmap := f.State.ArrayOfStringMap().GetStringMap(index)
	f.Results.Value().SetValue(mmap.GetString(key).Value())
}

//////////////////// StringMap of StringArray \\\\\\\\\\\\\\\\\\\\

func funcStringMapOfStringArrayAppend(ctx wasmlib.ScFuncContext, f *StringMapOfStringArrayAppendContext) {
	name := f.Params.Name().Value()
	array := f.State.StringMapOfStringArray().GetStringArray(name)
	value := f.Params.Value().Value()
	array.AppendString().SetValue(value)
}

func funcStringMapOfStringArrayClear(ctx wasmlib.ScFuncContext, f *StringMapOfStringArrayClearContext) {
	name := f.Params.Name().Value()
	array := f.State.StringMapOfStringArray().GetStringArray(name)
	array.Clear()
}

func funcStringMapOfStringArraySet(ctx wasmlib.ScFuncContext, f *StringMapOfStringArraySetContext) {
	name := f.Params.Name().Value()
	array := f.State.StringMapOfStringArray().GetStringArray(name)
	index := f.Params.Index().Value()
	value := f.Params.Value().Value()
	array.GetString(index).SetValue(value)
}

func viewStringMapOfStringArrayLength(ctx wasmlib.ScViewContext, f *StringMapOfStringArrayLengthContext) {
	name := f.Params.Name().Value()
	array := f.State.StringMapOfStringArray().GetStringArray(name)
	length := array.Length()
	f.Results.Length().SetValue(length)
}

func viewStringMapOfStringArrayValue(ctx wasmlib.ScViewContext, f *StringMapOfStringArrayValueContext) {
	name := f.Params.Name().Value()
	array := f.State.StringMapOfStringArray().GetStringArray(name)
	index := f.Params.Index().Value()
	value := array.GetString(index).Value()
	f.Results.Value().SetValue(value)
}

//////////////////// StringMap of StringMap \\\\\\\\\\\\\\\\\\\\

func funcStringMapOfStringMapClear(ctx wasmlib.ScFuncContext, f *StringMapOfStringMapClearContext) {
	name := f.Params.Name().Value()
	mmap := f.State.StringMapOfStringMap().GetStringMap(name)
	mmap.Clear()
}

func funcStringMapOfStringMapSet(ctx wasmlib.ScFuncContext, f *StringMapOfStringMapSetContext) {
	name := f.Params.Name().Value()
	mmap := f.State.StringMapOfStringMap().GetStringMap(name)
	key := f.Params.Key().Value()
	value := f.Params.Value().Value()
	mmap.GetString(key).SetValue(value)
}

func viewStringMapOfStringMapValue(ctx wasmlib.ScViewContext, f *StringMapOfStringMapValueContext) {
	name := f.Params.Name().Value()
	mmap := f.State.StringMapOfStringMap().GetStringMap(name)
	key := f.Params.Key().Value()
	f.Results.Value().SetValue(mmap.GetString(key).Value())
}

//////////////////// array of AddressArray \\\\\\\\\\\\\\\\\\\\

func funcArrayOfAddressArrayAppend(ctx wasmlib.ScFuncContext, f *ArrayOfAddressArrayAppendContext) {
	index := f.Params.Index().Value()
	valLen := f.Params.ValueAddr().Length()

	var sa ArrayOfMutableAddress
	if f.State.ArrayOfStringArray().Length() <= index {
		sa = f.State.ArrayOfAddressArray().AppendAddressArray()
	} else {
		sa = f.State.ArrayOfAddressArray().GetAddressArray(index)
	}

	for i := uint32(0); i < valLen; i++ {
		elt := f.Params.ValueAddr().GetAddress(i).Value()
		sa.AppendAddress().SetValue(elt)
	}
}

func funcArrayOfAddressArrayClear(ctx wasmlib.ScFuncContext, f *ArrayOfAddressArrayClearContext) {
	length := f.State.ArrayOfAddressArray().Length()
	for i := uint32(0); i < length; i++ {
		array := f.State.ArrayOfAddressArray().GetAddressArray(i)
		array.Clear()
	}
	f.State.ArrayOfAddressArray().Clear()
}

func funcArrayOfAddressArraySet(ctx wasmlib.ScFuncContext, f *ArrayOfAddressArraySetContext) {
	index0 := f.Params.Index0().Value()
	index1 := f.Params.Index1().Value()
	array := f.State.ArrayOfAddressArray().GetAddressArray(index0)
	value := f.Params.ValueAddr().Value()
	array.GetAddress(index1).SetValue(value)
}

func viewArrayOfAddressArrayLength(ctx wasmlib.ScViewContext, f *ArrayOfAddressArrayLengthContext) {
	length := f.State.ArrayOfAddressArray().Length()
	f.Results.Length().SetValue(length)
}

func viewArrayOfAddressArrayValue(ctx wasmlib.ScViewContext, f *ArrayOfAddressArrayValueContext) {
	index0 := f.Params.Index0().Value()
	index1 := f.Params.Index1().Value()

	elt := f.State.ArrayOfAddressArray().GetAddressArray(index0).GetAddress(index1).Value()
	f.Results.ValueAddr().SetValue(elt)
}

//////////////////// array of AddressMap \\\\\\\\\\\\\\\\\\\\

func funcArrayOfAddressMapClear(ctx wasmlib.ScFuncContext, f *ArrayOfAddressMapClearContext) {
	length := f.State.ArrayOfAddressArray().Length()
	for i := uint32(0); i < length; i++ {
		mmap := f.State.ArrayOfAddressMap().GetAddressMap(i)
		mmap.Clear()
	}
	f.State.ArrayOfAddressMap().Clear()
}

func funcArrayOfAddressMapSet(ctx wasmlib.ScFuncContext, f *ArrayOfAddressMapSetContext) {
	index := f.Params.Index().Value()
	value := f.Params.ValueAddr().Value()
	key := f.Params.KeyAddr().Value()
	if f.State.ArrayOfAddressMap().Length() <= index {
		mmap := f.State.ArrayOfAddressMap().AppendAddressMap()
		mmap.GetAddress(key).SetValue(value)
		return
	}
	mmap := f.State.ArrayOfAddressMap().GetAddressMap(index)
	mmap.GetAddress(key).SetValue(value)
}

func viewArrayOfAddressMapValue(ctx wasmlib.ScViewContext, f *ArrayOfAddressMapValueContext) {
	index := f.Params.Index().Value()
	key := f.Params.KeyAddr().Value()
	mmap := f.State.ArrayOfAddressMap().GetAddressMap(index)
	f.Results.ValueAddr().SetValue(mmap.GetAddress(key).Value())
}

//////////////////// AddressMap of AddressArray \\\\\\\\\\\\\\\\\\\\

func funcAddressMapOfAddressArrayAppend(ctx wasmlib.ScFuncContext, f *AddressMapOfAddressArrayAppendContext) {
	addr := f.Params.NameAddr().Value()
	array := f.State.AddressMapOfAddressArray().GetAddressArray(addr)
	value := f.Params.ValueAddr().Value()
	array.AppendAddress().SetValue(value)
}

func funcAddressMapOfAddressArrayClear(ctx wasmlib.ScFuncContext, f *AddressMapOfAddressArrayClearContext) {
	addr := f.Params.NameAddr().Value()
	array := f.State.AddressMapOfAddressArray().GetAddressArray(addr)
	array.Clear()
}

func funcAddressMapOfAddressArraySet(ctx wasmlib.ScFuncContext, f *AddressMapOfAddressArraySetContext) {
	addr := f.Params.NameAddr().Value()
	array := f.State.AddressMapOfAddressArray().GetAddressArray(addr)
	index := f.Params.Index().Value()
	value := f.Params.ValueAddr().Value()
	array.GetAddress(index).SetValue(value)
}

func viewAddressMapOfAddressArrayLength(ctx wasmlib.ScViewContext, f *AddressMapOfAddressArrayLengthContext) {
	addr := f.Params.NameAddr().Value()
	array := f.State.AddressMapOfAddressArray().GetAddressArray(addr)
	length := array.Length()
	f.Results.Length().SetValue(length)
}

func viewAddressMapOfAddressArrayValue(ctx wasmlib.ScViewContext, f *AddressMapOfAddressArrayValueContext) {
	addr := f.Params.NameAddr().Value()
	array := f.State.AddressMapOfAddressArray().GetAddressArray(addr)
	index := f.Params.Index().Value()
	value := array.GetAddress(index).Value()
	f.Results.ValueAddr().SetValue(value)
}

//////////////////// AddressMap of AddressMap \\\\\\\\\\\\\\\\\\\\

func funcAddressMapOfAddressMapClear(ctx wasmlib.ScFuncContext, f *AddressMapOfAddressMapClearContext) {
	name := f.Params.NameAddr().Value()
	myMap := f.State.AddressMapOfAddressMap().GetAddressMap(name)
	myMap.Clear()
}

func funcAddressMapOfAddressMapSet(ctx wasmlib.ScFuncContext, f *AddressMapOfAddressMapSetContext) {
	name := f.Params.NameAddr().Value()
	myMap := f.State.AddressMapOfAddressMap().GetAddressMap(name)
	key := f.Params.KeyAddr().Value()
	value := f.Params.ValueAddr().Value()
	myMap.GetAddress(key).SetValue(value)
}

func viewAddressMapOfAddressMapValue(ctx wasmlib.ScViewContext, f *AddressMapOfAddressMapValueContext) {
	name := f.Params.NameAddr().Value()
	myMap := f.State.AddressMapOfAddressMap().GetAddressMap(name)
	key := f.Params.KeyAddr().Value()
	f.Results.ValueAddr().SetValue(myMap.GetAddress(key).Value())
}

func viewBigIntAdd(ctx wasmlib.ScViewContext, f *BigIntAddContext) {
	lhs := f.Params.Lhs().Value()
	rhs := f.Params.Rhs().Value()
	res := lhs.Add(rhs)
	f.Results.Res().SetValue(res)
}

func viewBigIntDiv(ctx wasmlib.ScViewContext, f *BigIntDivContext) {
	lhs := f.Params.Lhs().Value()
	rhs := f.Params.Rhs().Value()
	res := lhs.Div(rhs)
	f.Results.Res().SetValue(res)
}

func viewBigIntMod(ctx wasmlib.ScViewContext, f *BigIntModContext) {
	lhs := f.Params.Lhs().Value()
	rhs := f.Params.Rhs().Value()
	res := lhs.Modulo(rhs)
	f.Results.Res().SetValue(res)
}

func viewBigIntMul(ctx wasmlib.ScViewContext, f *BigIntMulContext) {
	lhs := f.Params.Lhs().Value()
	rhs := f.Params.Rhs().Value()
	res := lhs.Mul(rhs)
	f.Results.Res().SetValue(res)
}

func viewBigIntSub(ctx wasmlib.ScViewContext, f *BigIntSubContext) {
	lhs := f.Params.Lhs().Value()
	rhs := f.Params.Rhs().Value()
	res := lhs.Sub(rhs)
	f.Results.Res().SetValue(res)
}

func viewBigIntShl(ctx wasmlib.ScViewContext, f *BigIntShlContext) {
	lhs := f.Params.Lhs().Value()
	shift := f.Params.Shift().Value()
	res := lhs.Shl(shift)
	f.Results.Res().SetValue(res)
}

func viewBigIntShr(ctx wasmlib.ScViewContext, f *BigIntShrContext) {
	lhs := f.Params.Lhs().Value()
	shift := f.Params.Shift().Value()
	res := lhs.Shr(shift)
	f.Results.Res().SetValue(res)
}
