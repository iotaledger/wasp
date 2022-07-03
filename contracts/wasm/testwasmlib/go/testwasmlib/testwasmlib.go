// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testwasmlib

import (
	"bytes"
	"math"

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
	if f.Params.BigInt().Exists() {
		bigIntData := wasmtypes.BigIntFromString("100000000000000000000")
		ctx.Require(f.Params.BigInt().Value().Cmp(bigIntData) == 0, "mismatch: BigInt")
	}
	if f.Params.Bool().Exists() {
		ctx.Require(f.Params.Bool().Value(), "mismatch: Bool")
	}
	if f.Params.Bytes().Exists() {
		byteData := []byte("these are bytes")
		ctx.Require(bytes.Equal(f.Params.Bytes().Value(), byteData), "mismatch: Bytes")
	}
	if f.Params.ChainID().Exists() {
		ctx.Require(f.Params.ChainID().Value() == ctx.CurrentChainID(), "mismatch: ChainID")
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
		tokenID := wasmtypes.TokenIDFromBytes([]byte("abcdefghijklmnopqrstuvwxyz1234567890AB"))
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

func viewBigIntDivMod(ctx wasmlib.ScViewContext, f *BigIntDivModContext) {
	lhs := f.Params.Lhs().Value()
	rhs := f.Params.Rhs().Value()
	quo, remainder := lhs.DivMod(rhs)
	f.Results.Quo().SetValue(quo)
	f.Results.Remainder().SetValue(remainder)
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

func viewCheckAgentID(ctx wasmlib.ScViewContext, f *CheckAgentIDContext) {
	scAgentID := f.Params.ScAgentID().Value()
	agentBytes := f.Params.AgentBytes().Value()
	agentString := f.Params.AgentString().Value()
	ctx.Require(scAgentID == wasmtypes.AgentIDFromBytes(wasmtypes.AgentIDToBytes(scAgentID)), "bytes conversion failed")
	ctx.Require(scAgentID == wasmtypes.AgentIDFromString(wasmtypes.AgentIDToString(scAgentID)), "string conversion failed")
	ctx.Require(string(scAgentID.Bytes()) == string(agentBytes), "bytes mismatch")
	ctx.Require(scAgentID.String() == agentString, "string mismatch")
}

func viewCheckAddress(ctx wasmlib.ScViewContext, f *CheckAddressContext) {
	address := f.Params.ScAddress().Value()
	addressBytes := f.Params.AddressBytes().Value()
	addressString := f.Params.AddressString().Value()
	ctx.Require(address == wasmtypes.AddressFromBytes(wasmtypes.AddressToBytes(address)), "bytes conversion failed")
	ctx.Require(address == wasmtypes.AddressFromString(wasmtypes.AddressToString(address)), "string conversion failed")
	ctx.Require(string(address.Bytes()) == string(addressBytes), "bytes mismatch")
	ctx.Require(address.String() == addressString, "string mismatch")
}

func viewCheckEthAddressAndAgentID(ctx wasmlib.ScViewContext, f *CheckEthAddressAndAgentIDContext) {
	addressTxt := f.Params.EthAddress().Value()
	scAddressEth := wasmtypes.AddressFromString(addressTxt)
	ctx.Require(scAddressEth == wasmtypes.AddressFromBytes(wasmtypes.AddressToBytes(scAddressEth)), "eth address bytes conversion failed")
	ctx.Require(scAddressEth == wasmtypes.AddressFromString(wasmtypes.AddressToString(scAddressEth)), "eth address string conversion failed")
	scAgentIDEth := wasmtypes.NewScAgentIDFromAddress(scAddressEth)
	ctx.Require(scAgentIDEth == wasmtypes.AgentIDFromBytes(wasmtypes.AgentIDToBytes(scAgentIDEth)), "eth agentID bytes conversion failed")
	ctx.Require(scAgentIDEth == wasmtypes.AgentIDFromString(wasmtypes.AgentIDToString(scAgentIDEth)), "eth agentID string conversion failed")
}

func viewCheckHash(ctx wasmlib.ScViewContext, f *CheckHashContext) {
	scHash := f.Params.ScHash().Value()
	hashBytes := f.Params.HashBytes().Value()
	hashString := f.Params.HashString().Value()
	ctx.Require(scHash == wasmtypes.HashFromBytes(wasmtypes.HashToBytes(scHash)), "bytes conversion failed")
	ctx.Require(scHash == wasmtypes.HashFromString(wasmtypes.HashToString(scHash)), "string conversion failed")
	ctx.Require(string(scHash.Bytes()) == string(hashBytes), "bytes mismatch")
	ctx.Require(scHash.String() == hashString, "string mismatch")
}

func viewCheckNftID(ctx wasmlib.ScViewContext, f *CheckNftIDContext) {
	scNftID := f.Params.ScNftID().Value()
	nftIDBytes := f.Params.NftIDBytes().Value()
	nftIDString := f.Params.NftIDString().Value()

	ctx.Require(scNftID == wasmtypes.NftIDFromString(wasmtypes.NftIDToString(scNftID)), "bytes conversion failed")
	ctx.Require(scNftID == wasmtypes.NftIDFromBytes(wasmtypes.NftIDToBytes(scNftID)), "string conversion failed")
	ctx.Require(string(scNftID.Bytes()) == string(nftIDBytes), "bytes mismatch")
	ctx.Require(scNftID.String() == nftIDString, "string mismatch")
}

func viewCheckRequestID(ctx wasmlib.ScViewContext, f *CheckRequestIDContext) {
	scRequestID := f.Params.ScRequestID().Value()
	requestIDBytes := f.Params.RequestIDBytes().Value()
	requestIDString := f.Params.RequestIDString().Value()

	ctx.Require(scRequestID == wasmtypes.RequestIDFromString(wasmtypes.RequestIDToString(scRequestID)), "bytes conversion failed")
	ctx.Require(scRequestID == wasmtypes.RequestIDFromBytes(wasmtypes.RequestIDToBytes(scRequestID)), "string conversion failed")
	ctx.Require(string(scRequestID.Bytes()) == string(requestIDBytes), "bytes mismatch")
	ctx.Require(scRequestID.String() == requestIDString, "string mismatch")
}

func viewCheckTokenID(ctx wasmlib.ScViewContext, f *CheckTokenIDContext) {
	scTokenID := f.Params.ScTokenID().Value()
	tokenIDBytes := f.Params.TokenIDBytes().Value()
	tokenIDString := f.Params.TokenIDString().Value()

	ctx.Require(scTokenID == wasmtypes.TokenIDFromString(wasmtypes.TokenIDToString(scTokenID)), "bytes conversion failed")
	ctx.Require(scTokenID == wasmtypes.TokenIDFromBytes(wasmtypes.TokenIDToBytes(scTokenID)), "string conversion failed")
	ctx.Require(string(scTokenID.Bytes()) == string(tokenIDBytes), "bytes mismatch")
	ctx.Require(scTokenID.String() == tokenIDString, "string mismatch")
}

func viewCheckBigInt(ctx wasmlib.ScViewContext, f *CheckBigIntContext) {
	scBigInt := f.Params.ScBigInt().Value()
	bigIntBytes := f.Params.BigIntBytes().Value()
	bigIntString := f.Params.BigIntString().Value()

	ctx.Require(scBigInt.Cmp(wasmtypes.BigIntFromString(wasmtypes.BigIntToString(scBigInt))) == 0, "bytes conversion failed")
	ctx.Require(scBigInt.Cmp(wasmtypes.BigIntFromBytes(wasmtypes.BigIntToBytes(scBigInt))) == 0, "string conversion failed")
	ctx.Require(string(scBigInt.Bytes()) == string(bigIntBytes), "bytes mismatch")
	ctx.Require(scBigInt.String() == bigIntString, "string mismatch")
}

//nolint:funlen
func viewCheckIntAndUint(ctx wasmlib.ScViewContext, f *CheckIntAndUintContext) {
	goInt8 := int8(math.MaxInt8)
	ctx.Require(goInt8 == wasmtypes.Int8FromBytes(wasmtypes.Int8ToBytes(goInt8)), "bytes conversion failed")
	ctx.Require(goInt8 == wasmtypes.Int8FromString(wasmtypes.Int8ToString(goInt8)), "string conversion failed")
	goInt8 = math.MinInt8
	ctx.Require(goInt8 == wasmtypes.Int8FromBytes(wasmtypes.Int8ToBytes(goInt8)), "bytes conversion failed")
	ctx.Require(goInt8 == wasmtypes.Int8FromString(wasmtypes.Int8ToString(goInt8)), "string conversion failed")
	goInt8 = 1
	ctx.Require(goInt8 == wasmtypes.Int8FromBytes(wasmtypes.Int8ToBytes(goInt8)), "bytes conversion failed")
	ctx.Require(goInt8 == wasmtypes.Int8FromString(wasmtypes.Int8ToString(goInt8)), "string conversion failed")
	goInt8 = 0
	ctx.Require(goInt8 == wasmtypes.Int8FromBytes(wasmtypes.Int8ToBytes(goInt8)), "bytes conversion failed")
	ctx.Require(goInt8 == wasmtypes.Int8FromString(wasmtypes.Int8ToString(goInt8)), "string conversion failed")
	goInt8 = -1
	ctx.Require(goInt8 == wasmtypes.Int8FromBytes(wasmtypes.Int8ToBytes(goInt8)), "bytes conversion failed")
	ctx.Require(goInt8 == wasmtypes.Int8FromString(wasmtypes.Int8ToString(goInt8)), "string conversion failed")
	goUint8 := uint8(0)
	ctx.Require(goUint8 == wasmtypes.Uint8FromBytes(wasmtypes.Uint8ToBytes(goUint8)), "bytes conversion failed")
	ctx.Require(goUint8 == wasmtypes.Uint8FromString(wasmtypes.Uint8ToString(goUint8)), "string conversion failed")
	goUint8 = math.MaxUint8
	ctx.Require(goUint8 == wasmtypes.Uint8FromBytes(wasmtypes.Uint8ToBytes(goUint8)), "bytes conversion failed")
	ctx.Require(goUint8 == wasmtypes.Uint8FromString(wasmtypes.Uint8ToString(goUint8)), "string conversion failed")

	goInt16 := int16(math.MaxInt16)
	ctx.Require(goInt16 == wasmtypes.Int16FromBytes(wasmtypes.Int16ToBytes(goInt16)), "bytes conversion failed")
	ctx.Require(goInt16 == wasmtypes.Int16FromString(wasmtypes.Int16ToString(goInt16)), "string conversion failed")
	goInt16 = math.MinInt16
	ctx.Require(goInt16 == wasmtypes.Int16FromBytes(wasmtypes.Int16ToBytes(goInt16)), "bytes conversion failed")
	ctx.Require(goInt16 == wasmtypes.Int16FromString(wasmtypes.Int16ToString(goInt16)), "string conversion failed")
	goInt16 = 1
	ctx.Require(goInt16 == wasmtypes.Int16FromBytes(wasmtypes.Int16ToBytes(goInt16)), "bytes conversion failed")
	ctx.Require(goInt16 == wasmtypes.Int16FromString(wasmtypes.Int16ToString(goInt16)), "string conversion failed")
	goInt16 = 0
	ctx.Require(goInt16 == wasmtypes.Int16FromBytes(wasmtypes.Int16ToBytes(goInt16)), "bytes conversion failed")
	ctx.Require(goInt16 == wasmtypes.Int16FromString(wasmtypes.Int16ToString(goInt16)), "string conversion failed")
	goInt16 = -1
	ctx.Require(goInt16 == wasmtypes.Int16FromBytes(wasmtypes.Int16ToBytes(goInt16)), "bytes conversion failed")
	ctx.Require(goInt16 == wasmtypes.Int16FromString(wasmtypes.Int16ToString(goInt16)), "string conversion failed")
	goUint16 := uint16(0)
	ctx.Require(goUint16 == wasmtypes.Uint16FromBytes(wasmtypes.Uint16ToBytes(goUint16)), "bytes conversion failed")
	ctx.Require(goUint16 == wasmtypes.Uint16FromString(wasmtypes.Uint16ToString(goUint16)), "string conversion failed")
	goUint16 = math.MaxUint16
	ctx.Require(goUint16 == wasmtypes.Uint16FromBytes(wasmtypes.Uint16ToBytes(goUint16)), "bytes conversion failed")
	ctx.Require(goUint16 == wasmtypes.Uint16FromString(wasmtypes.Uint16ToString(goUint16)), "string conversion failed")

	goInt32 := int32(math.MaxInt32)
	ctx.Require(goInt32 == wasmtypes.Int32FromBytes(wasmtypes.Int32ToBytes(goInt32)), "bytes conversion failed")
	ctx.Require(goInt32 == wasmtypes.Int32FromString(wasmtypes.Int32ToString(goInt32)), "string conversion failed")
	goInt32 = math.MinInt32
	ctx.Require(goInt32 == wasmtypes.Int32FromBytes(wasmtypes.Int32ToBytes(goInt32)), "bytes conversion failed")
	ctx.Require(goInt32 == wasmtypes.Int32FromString(wasmtypes.Int32ToString(goInt32)), "string conversion failed")
	goInt32 = 1
	ctx.Require(goInt32 == wasmtypes.Int32FromBytes(wasmtypes.Int32ToBytes(goInt32)), "bytes conversion failed")
	ctx.Require(goInt32 == wasmtypes.Int32FromString(wasmtypes.Int32ToString(goInt32)), "string conversion failed")
	goInt32 = 0
	ctx.Require(goInt32 == wasmtypes.Int32FromBytes(wasmtypes.Int32ToBytes(goInt32)), "bytes conversion failed")
	ctx.Require(goInt32 == wasmtypes.Int32FromString(wasmtypes.Int32ToString(goInt32)), "string conversion failed")
	goInt32 = -1
	ctx.Require(goInt32 == wasmtypes.Int32FromBytes(wasmtypes.Int32ToBytes(goInt32)), "bytes conversion failed")
	ctx.Require(goInt32 == wasmtypes.Int32FromString(wasmtypes.Int32ToString(goInt32)), "string conversion failed")
	goUint32 := uint32(0)
	ctx.Require(goUint32 == wasmtypes.Uint32FromBytes(wasmtypes.Uint32ToBytes(goUint32)), "bytes conversion failed")
	ctx.Require(goUint32 == wasmtypes.Uint32FromString(wasmtypes.Uint32ToString(goUint32)), "string conversion failed")
	goUint32 = math.MaxUint32
	ctx.Require(goUint32 == wasmtypes.Uint32FromBytes(wasmtypes.Uint32ToBytes(goUint32)), "bytes conversion failed")
	ctx.Require(goUint32 == wasmtypes.Uint32FromString(wasmtypes.Uint32ToString(goUint32)), "string conversion failed")

	goInt64 := int64(math.MaxInt64)
	ctx.Require(goInt64 == wasmtypes.Int64FromBytes(wasmtypes.Int64ToBytes(goInt64)), "bytes conversion failed")
	ctx.Require(goInt64 == wasmtypes.Int64FromString(wasmtypes.Int64ToString(goInt64)), "string conversion failed")
	goInt64 = math.MinInt64
	ctx.Require(goInt64 == wasmtypes.Int64FromBytes(wasmtypes.Int64ToBytes(goInt64)), "bytes conversion failed")
	ctx.Require(goInt64 == wasmtypes.Int64FromString(wasmtypes.Int64ToString(goInt64)), "string conversion failed")
	goInt64 = 1
	ctx.Require(goInt64 == wasmtypes.Int64FromBytes(wasmtypes.Int64ToBytes(goInt64)), "bytes conversion failed")
	ctx.Require(goInt64 == wasmtypes.Int64FromString(wasmtypes.Int64ToString(goInt64)), "string conversion failed")
	goInt64 = 0
	ctx.Require(goInt64 == wasmtypes.Int64FromBytes(wasmtypes.Int64ToBytes(goInt64)), "bytes conversion failed")
	ctx.Require(goInt64 == wasmtypes.Int64FromString(wasmtypes.Int64ToString(goInt64)), "string conversion failed")
	goInt64 = -1
	ctx.Require(goInt64 == wasmtypes.Int64FromBytes(wasmtypes.Int64ToBytes(goInt64)), "bytes conversion failed")
	ctx.Require(goInt64 == wasmtypes.Int64FromString(wasmtypes.Int64ToString(goInt64)), "string conversion failed")
	goUint64 := uint64(0)
	ctx.Require(goUint64 == wasmtypes.Uint64FromBytes(wasmtypes.Uint64ToBytes(goUint64)), "bytes conversion failed")
	ctx.Require(goUint64 == wasmtypes.Uint64FromString(wasmtypes.Uint64ToString(goUint64)), "string conversion failed")
	goUint64 = math.MaxUint64
	ctx.Require(goUint64 == wasmtypes.Uint64FromBytes(wasmtypes.Uint64ToBytes(goUint64)), "bytes conversion failed")
	ctx.Require(goUint64 == wasmtypes.Uint64FromString(wasmtypes.Uint64ToString(goUint64)), "string conversion failed")
}

func viewCheckBool(ctx wasmlib.ScViewContext, f *CheckBoolContext) {
	boolData := true
	ctx.Require(boolData == wasmtypes.BoolFromBytes(wasmtypes.BoolToBytes(boolData)), "bytes conversion failed")
	ctx.Require(boolData == wasmtypes.BoolFromString(wasmtypes.BoolToString(boolData)), "string conversion failed")
	boolData = false
	ctx.Require(boolData == wasmtypes.BoolFromBytes(wasmtypes.BoolToBytes(boolData)), "bytes conversion failed")
	ctx.Require(boolData == wasmtypes.BoolFromString(wasmtypes.BoolToString(boolData)), "string conversion failed")
}

func viewCheckBytes(ctx wasmlib.ScViewContext, f *CheckBytesContext) {
	byteData := f.Params.Bytes().Value()
	ctx.Require(bytes.Equal(byteData, wasmtypes.BytesFromBytes(wasmtypes.BytesToBytes(byteData))), "bytes conversion failed")
	ctx.Require(bytes.Equal(byteData, wasmtypes.BytesFromString(wasmtypes.BytesToString(byteData))), "string conversion failed")
}

func viewCheckHname(ctx wasmlib.ScViewContext, f *CheckHnameContext) {
	scHname := f.Params.ScHname().Value()
	hnameBytes := f.Params.HnameBytes().Value()
	hnameString := f.Params.HnameString().Value()
	ctx.Require(scHname == wasmtypes.HnameFromBytes(wasmtypes.HnameToBytes(scHname)), "bytes conversion failed")
	ctx.Require(scHname == wasmtypes.HnameFromString(wasmtypes.HnameToString(scHname)), "string conversion failed")
	ctx.Require(bytes.Equal(hnameBytes, wasmtypes.HnameToBytes(scHname)), "bytes conversion failed")
	ctx.Require(hnameString == wasmtypes.HnameToString(scHname), "string conversion failed")
}

func viewCheckString(ctx wasmlib.ScViewContext, f *CheckStringContext) {
	stringData := f.Params.String().Value()
	ctx.Require(stringData == wasmtypes.StringFromBytes(wasmtypes.StringToBytes(stringData)), "bytes conversion failed")
	ctx.Require(stringData == wasmtypes.StringToString(wasmtypes.StringFromString(stringData)), "string conversion failed")
}
