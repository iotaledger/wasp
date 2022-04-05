// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

var (
	allParams = []string{
		testwasmlib.ParamAddress,
		testwasmlib.ParamAgentID,
		testwasmlib.ParamBool,
		testwasmlib.ParamChainID,
		testwasmlib.ParamHash,
		testwasmlib.ParamHname,
		testwasmlib.ParamInt8,
		testwasmlib.ParamInt16,
		testwasmlib.ParamInt32,
		testwasmlib.ParamInt64,
		testwasmlib.ParamNftID,
		testwasmlib.ParamRequestID,
		testwasmlib.ParamTokenID,
		testwasmlib.ParamUint8,
		testwasmlib.ParamUint16,
		testwasmlib.ParamUint32,
		testwasmlib.ParamUint64,
	}
	allLengths    = []int{33, 33, 1, 20, 32, 4, 1, 2, 4, 8, 20, 34, 38, 1, 2, 4, 8}
	invalidValues = map[string][][]byte{
		testwasmlib.ParamAddress: {
			append([]byte{3}, zeroHash...),
			append([]byte{4}, zeroHash...),
			append([]byte{255}, zeroHash...),
		},
		testwasmlib.ParamChainID: {
			append([]byte{0}, zeroHash...),
			append([]byte{1}, zeroHash...),
			append([]byte{3}, zeroHash...),
			append([]byte{4}, zeroHash...),
			append([]byte{255}, zeroHash...),
		},
		testwasmlib.ParamRequestID: {
			append(zeroHash, []byte{128, 0}...),
			append(zeroHash, []byte{127, 1}...),
			append(zeroHash, []byte{0, 1}...),
			append(zeroHash, []byte{255, 255}...),
			append(zeroHash, []byte{4, 4}...),
		},
	}
	zeroHash = make([]byte, 32)
)

func setupTest(t *testing.T) *wasmsolo.SoloContext {
	*wasmsolo.RsWasm = true
	return wasmsolo.NewSoloContext(t, testwasmlib.ScName, testwasmlib.OnLoad)
}

func TestDeploy(t *testing.T) {
	ctx := setupTest(t)
	require.NoError(t, ctx.ContractExists(testwasmlib.ScName))
}

func TestNoParams(t *testing.T) {
	ctx := setupTest(t)

	f := testwasmlib.ScFuncs.ParamTypes(ctx)
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestValidParams(t *testing.T) {
	_ = testValidParams(t)
}

func testValidParams(t *testing.T) *wasmsolo.SoloContext {
	ctx := setupTest(t)

	pt := testwasmlib.ScFuncs.ParamTypes(ctx)
	pt.Params.Address().SetValue(ctx.ChainID().Address())
	pt.Params.AgentID().SetValue(ctx.AccountID())
	pt.Params.Bool().SetValue(true)
	pt.Params.Bytes().SetValue([]byte("these are bytes"))
	pt.Params.ChainID().SetValue(ctx.ChainID())
	pt.Params.Hash().SetValue(wasmtypes.HashFromBytes([]byte("0123456789abcdeffedcba9876543210")))
	pt.Params.Hname().SetValue(testwasmlib.HScName)
	pt.Params.Int8().SetValue(-123)
	pt.Params.Int16().SetValue(-12345)
	pt.Params.Int32().SetValue(-1234567890)
	pt.Params.Int64().SetValue(-1234567890123456789)
	pt.Params.NftID().SetValue(wasmtypes.NftIDFromBytes([]byte("01234567890123456789")))
	pt.Params.RequestID().SetValue(wasmtypes.RequestIDFromBytes([]byte("abcdefghijklmnopqrstuvwxyz123456\x00\x00")))
	pt.Params.String().SetValue("this is a string")
	pt.Params.TokenID().SetValue(wasmtypes.TokenIDFromBytes([]byte("RedGreenBlueYellowCyanBlackWhitePurple")))
	pt.Params.Uint8().SetValue(123)
	pt.Params.Uint16().SetValue(12345)
	pt.Params.Uint32().SetValue(1234567890)
	pt.Params.Uint64().SetValue(1234567890123456789)
	pt.Func.Post()
	require.NoError(t, ctx.Err)
	return ctx
}

func TestValidSizeParams(t *testing.T) {
	ctx := setupTest(t)
	for index, param := range allParams {
		t.Run("ValidSize "+param, func(t *testing.T) {
			paramMismatch := fmt.Sprintf("mismatch: %s%s", strings.ToUpper(param[:1]), param[1:])
			pt := testwasmlib.ScFuncs.ParamTypes(ctx)
			bytes := make([]byte, allLengths[index])
			if param == testwasmlib.ParamChainID {
				bytes[0] = byte(iotago.AddressAlias)
			}
			pt.Params.Param().GetBytes(param).SetValue(bytes)
			pt.Func.Post()
			require.Error(t, ctx.Err)
			require.Contains(t, ctx.Err.Error(), paramMismatch)
		})
	}
}

func TestInvalidSizeParams(t *testing.T) {
	ctx := setupTest(t)
	for index, param := range allParams {
		t.Run("InvalidSize "+param, func(t *testing.T) {
			invalidLength := fmt.Sprintf("invalid %s%s length", strings.ToUpper(param[:1]), param[1:])

			// note that zero lengths are valid and will return a default value

			// no need to check bool/int8/uint8
			if allLengths[index] != 1 {
				pt := testwasmlib.ScFuncs.ParamTypes(ctx)
				pt.Params.Param().GetBytes(param).SetValue(make([]byte, 1))
				pt.Func.Post()
				require.Error(t, ctx.Err)
				require.Contains(t, ctx.Err.Error(), invalidLength)

				pt = testwasmlib.ScFuncs.ParamTypes(ctx)
				pt.Params.Param().GetBytes(param).SetValue(make([]byte, allLengths[index]-1))
				pt.Func.Post()
				require.Error(t, ctx.Err)
				require.Contains(t, ctx.Err.Error(), invalidLength)
			}

			pt := testwasmlib.ScFuncs.ParamTypes(ctx)
			pt.Params.Param().GetBytes(param).SetValue(make([]byte, allLengths[index]+1))
			pt.Func.Post()
			require.Error(t, ctx.Err)
			require.Contains(t, ctx.Err.Error(), invalidLength)
		})
	}
}

func TestInvalidTypeParams(t *testing.T) {
	ctx := setupTest(t)
	for param, values := range invalidValues {
		for index, value := range values {
			t.Run("InvalidType "+param+" "+strconv.Itoa(index), func(t *testing.T) {
				invalidParam := fmt.Sprintf("invalid %s%s", strings.ToUpper(param[:1]), param[1:])
				req := solo.NewCallParams(testwasmlib.ScName, testwasmlib.FuncParamTypes,
					param, value,
				).AddIotas(1).WithMaxAffordableGasBudget()
				_, err := ctx.Chain.PostRequestSync(req, nil)
				require.Error(t, err)
				require.Contains(t, err.Error(), invalidParam)
			})
		}
	}
}

func TestViewBlockRecords(t *testing.T) {
	ctx := testValidParams(t)

	recs := testwasmlib.ScFuncs.BlockRecords(ctx)
	recs.Params.BlockIndex().SetValue(1)
	recs.Func.Call()
	require.NoError(t, ctx.Err)
	count := recs.Results.Count()
	require.True(t, count.Exists())
	require.EqualValues(t, 1, count.Value())

	rec := testwasmlib.ScFuncs.BlockRecord(ctx)
	rec.Params.BlockIndex().SetValue(1)
	rec.Params.RecordIndex().SetValue(0)
	rec.Func.Call()
	require.NoError(t, ctx.Err)
	require.True(t, rec.Results.Record().Exists())
	require.EqualValues(t, 218, len(rec.Results.Record().Value()))
}

func TestStringMapOfStringArrayClear(t *testing.T) {
	ctx := setupTest(t)

	as := testwasmlib.ScFuncs.StringMapOfStringArrayAppend(ctx)
	as.Params.Name().SetValue("bands")
	as.Params.Value().SetValue("Simple Minds")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.StringMapOfStringArrayAppend(ctx)
	as.Params.Name().SetValue("bands")
	as.Params.Value().SetValue("Dire Straits")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.StringMapOfStringArrayAppend(ctx)
	as.Params.Name().SetValue("bands")
	as.Params.Value().SetValue("ELO")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	al := testwasmlib.ScFuncs.StringMapOfStringArrayLength(ctx)
	al.Params.Name().SetValue("bands")
	al.Func.Call()
	require.NoError(t, ctx.Err)
	length := al.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 3, length.Value())

	av := testwasmlib.ScFuncs.StringMapOfStringArrayValue(ctx)
	av.Params.Name().SetValue("bands")
	av.Params.Index().SetValue(0)
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value := av.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "Simple Minds", value.Value())

	av = testwasmlib.ScFuncs.StringMapOfStringArrayValue(ctx)
	av.Params.Name().SetValue("bands")
	av.Params.Index().SetValue(1)
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value = av.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "Dire Straits", value.Value())

	av = testwasmlib.ScFuncs.StringMapOfStringArrayValue(ctx)
	av.Params.Name().SetValue("bands")
	av.Params.Index().SetValue(2)
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value = av.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "ELO", value.Value())

	ac := testwasmlib.ScFuncs.StringMapOfStringArrayClear(ctx)
	ac.Params.Name().SetValue("bands")
	ac.Func.Post()
	require.NoError(t, ctx.Err)

	al = testwasmlib.ScFuncs.StringMapOfStringArrayLength(ctx)
	al.Params.Name().SetValue("bands")
	al.Func.Call()
	require.NoError(t, ctx.Err)
	length = al.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 0, length.Value())

	av = testwasmlib.ScFuncs.StringMapOfStringArrayValue(ctx)
	av.Params.Name().SetValue("bands")
	av.Params.Index().SetValue(0)
	av.Func.Call()
	require.Error(t, ctx.Err)
}

func genTestAddress(ctx *wasmsolo.SoloContext, num int) []wasmtypes.ScAddress {
	addrs := make([]wasmtypes.ScAddress, num)
	for i := 0; i < num; i++ {
		_, addr := ctx.Chain.Env.NewKeyPair()
		addrs[i] = wasmhost.WasmConvertor{}.ScAddress(addr)
	}

	return addrs
}

func TestAddressMapOfAddressArrayClear(t *testing.T) {
	ctx := setupTest(t)
	mapNames, mapVals := genTestAddress(ctx, 2), genTestAddress(ctx, 3)

	as := testwasmlib.ScFuncs.AddressMapOfAddressArrayAppend(ctx)
	as.Params.NameAddr().SetValue(mapNames[0])
	as.Params.ValueAddr().SetValue(mapVals[0])
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.AddressMapOfAddressArrayAppend(ctx)
	as.Params.NameAddr().SetValue(mapNames[0])
	as.Params.ValueAddr().SetValue(mapVals[1])
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.AddressMapOfAddressArrayAppend(ctx)
	as.Params.NameAddr().SetValue(mapNames[1])
	as.Params.ValueAddr().SetValue(mapVals[2])
	as.Func.Post()
	require.NoError(t, ctx.Err)

	al := testwasmlib.ScFuncs.AddressMapOfAddressArrayLength(ctx)
	al.Params.NameAddr().SetValue(mapNames[0])
	al.Func.Call()
	require.NoError(t, ctx.Err)
	length := al.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 2, length.Value())

	al = testwasmlib.ScFuncs.AddressMapOfAddressArrayLength(ctx)
	al.Params.NameAddr().SetValue(mapNames[1])
	al.Func.Call()
	require.NoError(t, ctx.Err)
	length = al.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 1, length.Value())

	av := testwasmlib.ScFuncs.AddressMapOfAddressArrayValue(ctx)
	av.Params.NameAddr().SetValue(mapNames[0])
	av.Params.Index().SetValue(0)
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value := av.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[0], value.Value())

	av = testwasmlib.ScFuncs.AddressMapOfAddressArrayValue(ctx)
	av.Params.NameAddr().SetValue(mapNames[0])
	av.Params.Index().SetValue(1)
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value = av.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[1], value.Value())

	av = testwasmlib.ScFuncs.AddressMapOfAddressArrayValue(ctx)
	av.Params.NameAddr().SetValue(mapNames[1])
	av.Params.Index().SetValue(0)
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value = av.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[2], value.Value())

	ac := testwasmlib.ScFuncs.AddressMapOfAddressArrayClear(ctx)
	ac.Params.NameAddr().SetValue(mapNames[0])
	ac.Func.Post()
	require.NoError(t, ctx.Err)

	al = testwasmlib.ScFuncs.AddressMapOfAddressArrayLength(ctx)
	al.Params.NameAddr().SetValue(mapNames[0])
	al.Func.Call()
	require.NoError(t, ctx.Err)
	length = al.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 0, length.Value())

	av = testwasmlib.ScFuncs.AddressMapOfAddressArrayValue(ctx)
	av.Params.NameAddr().SetValue(mapNames[0])
	av.Params.Index().SetValue(0)
	av.Func.Call()
	require.Error(t, ctx.Err)
}

func TestStringMapOfStringArraySet(t *testing.T) {
	ctx := setupTest(t)

	ap := testwasmlib.ScFuncs.StringMapOfStringArrayAppend(ctx)
	ap.Params.Name().SetValue("bands")
	ap.Params.Value().SetValue("Simple Minds")
	ap.Func.Post()
	require.NoError(t, ctx.Err)

	ap = testwasmlib.ScFuncs.StringMapOfStringArrayAppend(ctx)
	ap.Params.Name().SetValue("bands")
	ap.Params.Value().SetValue("Dire Straits")
	ap.Func.Post()
	require.NoError(t, ctx.Err)

	al := testwasmlib.ScFuncs.StringMapOfStringArrayLength(ctx)
	al.Params.Name().SetValue("bands")
	al.Func.Call()
	require.NoError(t, ctx.Err)
	length := al.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 2, length.Value())

	av := testwasmlib.ScFuncs.StringMapOfStringArrayValue(ctx)
	av.Params.Name().SetValue("bands")
	av.Params.Index().SetValue(0)
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value := av.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "Simple Minds", value.Value())

	av = testwasmlib.ScFuncs.StringMapOfStringArrayValue(ctx)
	av.Params.Name().SetValue("bands")
	av.Params.Index().SetValue(1)
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value = av.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "Dire Straits", value.Value())

	as := testwasmlib.ScFuncs.StringMapOfStringArraySet(ctx)
	as.Params.Name().SetValue("bands")
	as.Params.Index().SetValue(0)
	as.Params.Value().SetValue("Collage")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	al = testwasmlib.ScFuncs.StringMapOfStringArrayLength(ctx)
	al.Params.Name().SetValue("bands")
	al.Func.Call()
	require.NoError(t, ctx.Err)
	length = al.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 2, length.Value())

	av = testwasmlib.ScFuncs.StringMapOfStringArrayValue(ctx)
	av.Params.Name().SetValue("bands")
	av.Params.Index().SetValue(0)
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value = av.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "Collage", value.Value())

	av = testwasmlib.ScFuncs.StringMapOfStringArrayValue(ctx)
	av.Params.Name().SetValue("bands")
	av.Params.Index().SetValue(1)
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value = av.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "Dire Straits", value.Value())
}

func TestAddressMapOfAddressArraySet(t *testing.T) {
	ctx := setupTest(t)
	mapNames, mapVals := genTestAddress(ctx, 2), genTestAddress(ctx, 4)

	aap := testwasmlib.ScFuncs.AddressMapOfAddressArrayAppend(ctx)
	aap.Params.NameAddr().SetValue(mapNames[0])
	aap.Params.ValueAddr().SetValue(mapVals[0])
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aap = testwasmlib.ScFuncs.AddressMapOfAddressArrayAppend(ctx)
	aap.Params.NameAddr().SetValue(mapNames[0])
	aap.Params.ValueAddr().SetValue(mapVals[1])
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aap = testwasmlib.ScFuncs.AddressMapOfAddressArrayAppend(ctx)
	aap.Params.NameAddr().SetValue(mapNames[1])
	aap.Params.ValueAddr().SetValue(mapVals[2])
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aal := testwasmlib.ScFuncs.AddressMapOfAddressArrayLength(ctx)
	aal.Params.NameAddr().SetValue(mapNames[0])
	aal.Func.Call()
	require.NoError(t, ctx.Err)
	length := aal.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 2, length.Value())

	aav := testwasmlib.ScFuncs.AddressMapOfAddressArrayValue(ctx)
	aav.Params.NameAddr().SetValue(mapNames[0])
	aav.Params.Index().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	value := aav.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[0], value.Value())

	aas := testwasmlib.ScFuncs.AddressMapOfAddressArraySet(ctx)
	aas.Params.NameAddr().SetValue(mapNames[0])
	aas.Params.Index().SetValue(0)
	aas.Params.ValueAddr().SetValue(mapVals[3])
	aas.Func.Post()
	require.NoError(t, ctx.Err)

	aal = testwasmlib.ScFuncs.AddressMapOfAddressArrayLength(ctx)
	aal.Params.NameAddr().SetValue(mapNames[0])
	aal.Func.Call()
	require.NoError(t, ctx.Err)
	length = aal.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 2, length.Value())

	aav = testwasmlib.ScFuncs.AddressMapOfAddressArrayValue(ctx)
	aav.Params.NameAddr().SetValue(mapNames[0])
	aav.Params.Index().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	value = aav.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[3], value.Value())

	aav = testwasmlib.ScFuncs.AddressMapOfAddressArrayValue(ctx)
	aav.Params.NameAddr().SetValue(mapNames[0])
	aav.Params.Index().SetValue(1)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	value = aav.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[1], value.Value())

	aav = testwasmlib.ScFuncs.AddressMapOfAddressArrayValue(ctx)
	aav.Params.NameAddr().SetValue(mapNames[1])
	aav.Params.Index().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	value = aav.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[2], value.Value())
}

func TestInvalidIndexInGetStringMapOfStringArrayElt(t *testing.T) {
	ctx := setupTest(t)

	as := testwasmlib.ScFuncs.StringMapOfStringArrayAppend(ctx)
	as.Params.Name().SetValue("bands")
	as.Params.Value().SetValue("Simple Minds")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.StringMapOfStringArrayAppend(ctx)
	as.Params.Name().SetValue("bands")
	as.Params.Value().SetValue("Dire Straits")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.StringMapOfStringArrayAppend(ctx)
	as.Params.Name().SetValue("bands")
	as.Params.Value().SetValue("ELO")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	al := testwasmlib.ScFuncs.StringMapOfStringArrayLength(ctx)
	al.Params.Name().SetValue("bands")
	al.Func.Call()
	require.NoError(t, ctx.Err)
	length := al.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 3, length.Value())

	av := testwasmlib.ScFuncs.StringMapOfStringArrayValue(ctx)
	av.Params.Name().SetValue("bands")
	av.Params.Index().SetValue(100)
	av.Func.Call()
	require.Contains(t, ctx.Err.Error(), "invalid index")
}

func TestArrayOfStringArrayAppend(t *testing.T) {
	ctx := setupTest(t)

	aap := testwasmlib.ScFuncs.ArrayOfStringArrayAppend(ctx)
	aap.Params.Index().SetValue(0)
	aap.Params.Value().AppendString().SetValue("support")
	aap.Params.Value().AppendString().SetValue("freedom")
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aap = testwasmlib.ScFuncs.ArrayOfStringArrayAppend(ctx)
	aap.Params.Index().SetValue(1)
	aap.Params.Value().AppendString().SetValue("hail")
	aap.Params.Value().AppendString().SetValue("life")
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aav := testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(0)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "support", aav.Results.Value().Value())

	aav = testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(0)
	aav.Params.Index1().SetValue(1)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "freedom", aav.Results.Value().Value())

	aav = testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "hail", aav.Results.Value().Value())

	aav = testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(1)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "life", aav.Results.Value().Value())
}

func TestArrayOfStringArrayClear(t *testing.T) {
	ctx := setupTest(t)

	aap := testwasmlib.ScFuncs.ArrayOfStringArrayAppend(ctx)
	aap.Params.Index().SetValue(0)
	aap.Params.Value().AppendString().SetValue("support")
	aap.Params.Value().AppendString().SetValue("freedom")
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aap = testwasmlib.ScFuncs.ArrayOfStringArrayAppend(ctx)
	aap.Params.Index().SetValue(1)
	aap.Params.Value().AppendString().SetValue("hail")
	aap.Params.Value().AppendString().SetValue("life")
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aav := testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(0)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "support", aav.Results.Value().Value())

	aav = testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(0)
	aav.Params.Index1().SetValue(1)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "freedom", aav.Results.Value().Value())

	aav = testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "hail", aav.Results.Value().Value())

	aav = testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(1)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "life", aav.Results.Value().Value())

	ac := testwasmlib.ScFuncs.ArrayOfStringArrayClear(ctx)
	ac.Func.Post()
	require.NoError(t, ctx.Err)

	aav = testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.Error(t, ctx.Err)
}

func TestArrayOfAddressArrayClear(t *testing.T) {
	ctx := setupTest(t)
	mapVals := genTestAddress(ctx, 4)

	aap := testwasmlib.ScFuncs.ArrayOfAddressArrayAppend(ctx)
	aap.Params.Index().SetValue(0)
	aap.Params.ValueAddr().AppendAddress().SetValue(mapVals[0])
	aap.Params.ValueAddr().AppendAddress().SetValue(mapVals[1])
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aap = testwasmlib.ScFuncs.ArrayOfAddressArrayAppend(ctx)
	aap.Params.Index().SetValue(1)
	aap.Params.ValueAddr().AppendAddress().SetValue(mapVals[2])
	aap.Params.ValueAddr().AppendAddress().SetValue(mapVals[3])
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aav := testwasmlib.ScFuncs.ArrayOfAddressArrayValue(ctx)
	aav.Params.Index0().SetValue(0)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, mapVals[0], aav.Results.ValueAddr().Value())

	aav = testwasmlib.ScFuncs.ArrayOfAddressArrayValue(ctx)
	aav.Params.Index0().SetValue(0)
	aav.Params.Index1().SetValue(1)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, mapVals[1], aav.Results.ValueAddr().Value())

	aav = testwasmlib.ScFuncs.ArrayOfAddressArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, mapVals[2], aav.Results.ValueAddr().Value())

	aav = testwasmlib.ScFuncs.ArrayOfAddressArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(1)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, mapVals[3], aav.Results.ValueAddr().Value())

	ac := testwasmlib.ScFuncs.ArrayOfAddressArrayClear(ctx)
	ac.Func.Post()
	require.NoError(t, ctx.Err)

	aav = testwasmlib.ScFuncs.ArrayOfAddressArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.Error(t, ctx.Err)
}

func TestArrayOfStringArraySet(t *testing.T) {
	ctx := setupTest(t)

	aap := testwasmlib.ScFuncs.ArrayOfStringArrayAppend(ctx)
	aap.Params.Index().SetValue(0)
	aap.Params.Value().AppendString().SetValue("support")
	aap.Params.Value().AppendString().SetValue("freedom")
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aap = testwasmlib.ScFuncs.ArrayOfStringArrayAppend(ctx)
	aap.Params.Index().SetValue(1)
	aap.Params.Value().AppendString().SetValue("hail")
	aap.Params.Value().AppendString().SetValue("life")
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aav := testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(0)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "support", aav.Results.Value().Value())

	aav = testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(0)
	aav.Params.Index1().SetValue(1)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "freedom", aav.Results.Value().Value())

	aav = testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "hail", aav.Results.Value().Value())

	aav = testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(1)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "life", aav.Results.Value().Value())

	aas := testwasmlib.ScFuncs.ArrayOfStringArraySet(ctx)
	aas.Params.Index0().SetValue(1)
	aas.Params.Index1().SetValue(1)
	aas.Params.Value().SetValue("moon")
	aas.Func.Post()
	require.NoError(t, ctx.Err)

	aav = testwasmlib.ScFuncs.ArrayOfStringArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(1)
	aav.Func.Call()
	require.EqualValues(t, "moon", aav.Results.Value().Value())
}

func TestArrayOfAddressArraySet(t *testing.T) {
	ctx := setupTest(t)
	mapVals := genTestAddress(ctx, 4)

	aap := testwasmlib.ScFuncs.ArrayOfAddressArrayAppend(ctx)
	aap.Params.Index().SetValue(0)
	aap.Params.ValueAddr().AppendAddress().SetValue(mapVals[0])
	aap.Params.ValueAddr().AppendAddress().SetValue(mapVals[1])
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aap = testwasmlib.ScFuncs.ArrayOfAddressArrayAppend(ctx)
	aap.Params.Index().SetValue(1)
	aap.Params.ValueAddr().AppendAddress().SetValue(mapVals[2])
	aap.Func.Post()
	require.NoError(t, ctx.Err)

	aav := testwasmlib.ScFuncs.ArrayOfAddressArrayValue(ctx)
	aav.Params.Index0().SetValue(0)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, mapVals[0], aav.Results.ValueAddr().Value())

	aav = testwasmlib.ScFuncs.ArrayOfAddressArrayValue(ctx)
	aav.Params.Index0().SetValue(0)
	aav.Params.Index1().SetValue(1)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, mapVals[1], aav.Results.ValueAddr().Value())

	aav = testwasmlib.ScFuncs.ArrayOfAddressArrayValue(ctx)
	aav.Params.Index0().SetValue(1)
	aav.Params.Index1().SetValue(0)
	aav.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, mapVals[2], aav.Results.ValueAddr().Value())

	aas := testwasmlib.ScFuncs.ArrayOfAddressArraySet(ctx)
	aas.Params.Index0().SetValue(0)
	aas.Params.Index1().SetValue(1)
	aas.Params.ValueAddr().SetValue(mapVals[3])
	aas.Func.Post()
	require.NoError(t, ctx.Err)

	aav = testwasmlib.ScFuncs.ArrayOfAddressArrayValue(ctx)
	aav.Params.Index0().SetValue(0)
	aav.Params.Index1().SetValue(1)
	aav.Func.Call()
	require.EqualValues(t, mapVals[3], aav.Results.ValueAddr().Value())
}

func TestStringMapOfStringMapClear(t *testing.T) {
	// test reproduces a problem that needs fixing
	t.SkipNow()

	ctx := setupTest(t)

	as := testwasmlib.ScFuncs.StringMapOfStringMapSet(ctx)
	as.Params.Name().SetValue("albums")
	as.Params.Key().SetValue("Simple Minds")
	as.Params.Value().SetValue("New Gold Dream")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.StringMapOfStringMapSet(ctx)
	as.Params.Name().SetValue("albums")
	as.Params.Key().SetValue("Dire Straits")
	as.Params.Value().SetValue("Calling Elvis")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.StringMapOfStringMapSet(ctx)
	as.Params.Name().SetValue("albums")
	as.Params.Key().SetValue("ELO")
	as.Params.Value().SetValue("Mr. Blue Sky")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	av := testwasmlib.ScFuncs.StringMapOfStringMapValue(ctx)
	av.Params.Name().SetValue("albums")
	av.Params.Key().SetValue("Dire Straits")
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value := av.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "Calling Elvis", value.Value())

	ac := testwasmlib.ScFuncs.StringMapOfStringMapClear(ctx)
	ac.Params.Name().SetValue("albums")
	ac.Func.Post()
	require.NoError(t, ctx.Err)

	av = testwasmlib.ScFuncs.StringMapOfStringMapValue(ctx)
	av.Params.Name().SetValue("albums")
	av.Params.Key().SetValue("Dire Straits")
	av.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "", value.Value())
}

func TestAddressMapOfAddressMapClear(t *testing.T) {
	// test reproduces a problem that needs fixing
	t.SkipNow()

	ctx := setupTest(t)

	mapNames, mapKeys, mapVals := genTestAddress(ctx, 2), genTestAddress(ctx, 4), genTestAddress(ctx, 4)

	as := testwasmlib.ScFuncs.AddressMapOfAddressMapSet(ctx)
	as.Params.NameAddr().SetValue(mapNames[0])
	as.Params.KeyAddr().SetValue(mapKeys[0])
	as.Params.ValueAddr().SetValue(mapVals[0])
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.AddressMapOfAddressMapSet(ctx)
	as.Params.NameAddr().SetValue(mapNames[0])
	as.Params.KeyAddr().SetValue(mapKeys[1])
	as.Params.ValueAddr().SetValue(mapVals[1])
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.AddressMapOfAddressMapSet(ctx)
	as.Params.NameAddr().SetValue(mapNames[1])
	as.Params.KeyAddr().SetValue(mapKeys[2])
	as.Params.ValueAddr().SetValue(mapVals[2])
	as.Func.Post()
	require.NoError(t, ctx.Err)

	av := testwasmlib.ScFuncs.AddressMapOfAddressMapValue(ctx)
	av.Params.NameAddr().SetValue(mapNames[0])
	av.Params.KeyAddr().SetValue(mapKeys[0])
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value := av.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[0], value.Value())

	ac := testwasmlib.ScFuncs.AddressMapOfAddressMapClear(ctx)
	ac.Params.NameAddr().SetValue(mapNames[0])
	ac.Func.Post()
	require.NoError(t, ctx.Err)

	av = testwasmlib.ScFuncs.AddressMapOfAddressMapValue(ctx)
	av.Params.NameAddr().SetValue(mapNames[0])
	av.Params.KeyAddr().SetValue(mapKeys[0])
	av.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, "", value.Value())
}

func TestAddressMapOfAddressMapSet(t *testing.T) {
	ctx := setupTest(t)

	mapNames, mapKeys, mapVals := genTestAddress(ctx, 2), genTestAddress(ctx, 4), genTestAddress(ctx, 4)

	as := testwasmlib.ScFuncs.AddressMapOfAddressMapSet(ctx)
	as.Params.NameAddr().SetValue(mapNames[0])
	as.Params.KeyAddr().SetValue(mapKeys[0])
	as.Params.ValueAddr().SetValue(mapVals[0])
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.AddressMapOfAddressMapSet(ctx)
	as.Params.NameAddr().SetValue(mapNames[0])
	as.Params.KeyAddr().SetValue(mapKeys[1])
	as.Params.ValueAddr().SetValue(mapVals[1])
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.AddressMapOfAddressMapSet(ctx)
	as.Params.NameAddr().SetValue(mapNames[1])
	as.Params.KeyAddr().SetValue(mapKeys[2])
	as.Params.ValueAddr().SetValue(mapVals[2])
	as.Func.Post()
	require.NoError(t, ctx.Err)

	av := testwasmlib.ScFuncs.AddressMapOfAddressMapValue(ctx)
	av.Params.NameAddr().SetValue(mapNames[0])
	av.Params.KeyAddr().SetValue(mapKeys[0])
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value := av.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[0], value.Value())

	av = testwasmlib.ScFuncs.AddressMapOfAddressMapValue(ctx)
	av.Params.NameAddr().SetValue(mapNames[0])
	av.Params.KeyAddr().SetValue(mapKeys[1])
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value = av.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[1], value.Value())

	av = testwasmlib.ScFuncs.AddressMapOfAddressMapValue(ctx)
	av.Params.NameAddr().SetValue(mapNames[1])
	av.Params.KeyAddr().SetValue(mapKeys[2])
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value = av.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[2], value.Value())

	as = testwasmlib.ScFuncs.AddressMapOfAddressMapSet(ctx)
	as.Params.NameAddr().SetValue(mapNames[1])
	as.Params.KeyAddr().SetValue(mapKeys[2])
	as.Params.ValueAddr().SetValue(mapVals[3])
	as.Func.Post()
	require.NoError(t, ctx.Err)

	av = testwasmlib.ScFuncs.AddressMapOfAddressMapValue(ctx)
	av.Params.NameAddr().SetValue(mapNames[1])
	av.Params.KeyAddr().SetValue(mapKeys[2])
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value = av.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[3], value.Value())

	as = testwasmlib.ScFuncs.AddressMapOfAddressMapSet(ctx)
	as.Params.NameAddr().SetValue(mapNames[0])
	as.Params.KeyAddr().SetValue(mapKeys[1])
	as.Params.ValueAddr().SetValue(mapVals[3])
	as.Func.Post()
	require.NoError(t, ctx.Err)

	av = testwasmlib.ScFuncs.AddressMapOfAddressMapValue(ctx)
	av.Params.NameAddr().SetValue(mapNames[0])
	av.Params.KeyAddr().SetValue(mapKeys[1])
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value = av.Results.ValueAddr()
	require.True(t, value.Exists())
	require.EqualValues(t, mapVals[3], value.Value())
}

func TestStringMapOfStringMapSet(t *testing.T) {
	ctx := setupTest(t)

	as := testwasmlib.ScFuncs.StringMapOfStringMapSet(ctx)
	as.Params.Name().SetValue("albums")
	as.Params.Key().SetValue("Simple Minds")
	as.Params.Value().SetValue("New Gold Dream")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.StringMapOfStringMapSet(ctx)
	as.Params.Name().SetValue("albums")
	as.Params.Key().SetValue("Dire Straits")
	as.Params.Value().SetValue("Calling Elvis")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.StringMapOfStringMapSet(ctx)
	as.Params.Name().SetValue("albums")
	as.Params.Key().SetValue("ELO")
	as.Params.Value().SetValue("Mr. Blue Sky")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	av := testwasmlib.ScFuncs.StringMapOfStringMapValue(ctx)
	av.Params.Name().SetValue("albums")
	av.Params.Key().SetValue("Dire Straits")
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value := av.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "Calling Elvis", value.Value())

	av = testwasmlib.ScFuncs.StringMapOfStringMapValue(ctx)
	av.Params.Name().SetValue("albums")
	av.Params.Key().SetValue("Simple Minds")
	av.Func.Call()
	require.NoError(t, ctx.Err)
	require.True(t, value.Exists())
	require.EqualValues(t, "New Gold Dream", av.Results.Value().Value())

	av = testwasmlib.ScFuncs.StringMapOfStringMapValue(ctx)
	av.Params.Name().SetValue("albums")
	av.Params.Key().SetValue("ELO")
	av.Func.Call()
	require.NoError(t, ctx.Err)
	require.True(t, value.Exists())
	require.EqualValues(t, "Mr. Blue Sky", av.Results.Value().Value())

	as = testwasmlib.ScFuncs.StringMapOfStringMapSet(ctx)
	as.Params.Name().SetValue("albums")
	as.Params.Key().SetValue("Simple Minds")
	as.Params.Value().SetValue("Life in a Day")
	as.Func.Post()
	require.NoError(t, ctx.Err)

	av = testwasmlib.ScFuncs.StringMapOfStringMapValue(ctx)
	av.Params.Name().SetValue("albums")
	av.Params.Key().SetValue("Simple Minds")
	av.Func.Call()
	require.NoError(t, ctx.Err)
	require.True(t, value.Exists())
	require.EqualValues(t, "Life in a Day", av.Results.Value().Value())
}

func TestArrayOfStringMapClear(t *testing.T) {
	ctx := setupTest(t)

	ams := testwasmlib.ScFuncs.ArrayOfStringMapSet(ctx)
	ams.Params.Index().SetValue(0)
	ams.Params.Key().SetValue("Simple Minds")
	ams.Params.Value().SetValue("New Gold Dream")
	ams.Func.Post()
	require.NoError(t, ctx.Err)

	ams = testwasmlib.ScFuncs.ArrayOfStringMapSet(ctx)
	ams.Params.Index().SetValue(0)
	ams.Params.Key().SetValue("Dire Straits")
	ams.Params.Value().SetValue("Calling Elvis")
	ams.Func.Post()
	require.NoError(t, ctx.Err)

	ams = testwasmlib.ScFuncs.ArrayOfStringMapSet(ctx)
	ams.Params.Index().SetValue(1)
	ams.Params.Key().SetValue("ELO")
	ams.Params.Value().SetValue("Mr. Blue Sky")
	ams.Func.Post()
	require.NoError(t, ctx.Err)

	amv := testwasmlib.ScFuncs.ArrayOfStringMapValue(ctx)
	amv.Params.Index().SetValue(0)
	amv.Params.Key().SetValue("Simple Minds")
	amv.Func.Call()
	require.NoError(t, ctx.Err)
	value := amv.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "New Gold Dream", value.Value())

	amv = testwasmlib.ScFuncs.ArrayOfStringMapValue(ctx)
	amv.Params.Index().SetValue(0)
	amv.Params.Key().SetValue("Dire Straits")
	amv.Func.Call()
	require.NoError(t, ctx.Err)
	value = amv.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "Calling Elvis", value.Value())

	amv = testwasmlib.ScFuncs.ArrayOfStringMapValue(ctx)
	amv.Params.Index().SetValue(1)
	amv.Params.Key().SetValue("ELO")
	amv.Func.Call()
	require.NoError(t, ctx.Err)
	value = amv.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "Mr. Blue Sky", value.Value())

	amc := testwasmlib.ScFuncs.ArrayOfStringMapClear(ctx)
	amc.Func.Post()
	require.NoError(t, ctx.Err)

	amv = testwasmlib.ScFuncs.ArrayOfStringMapValue(ctx)
	amv.Params.Index().SetValue(1)
	amv.Params.Key().SetValue("ELO")
	amv.Func.Call()
	require.Error(t, ctx.Err)

	amv = testwasmlib.ScFuncs.ArrayOfStringMapValue(ctx)
	amv.Params.Index().SetValue(0)
	amv.Params.Key().SetValue("Simple Minds")
	amv.Func.Call()
	require.Error(t, ctx.Err)
}

func TestTakeAllowance(t *testing.T) {
	ctx := setupTest(t)
	bal := ctx.Balances()

	f := testwasmlib.ScFuncs.TakeAllowance(ctx)
	f.Func.TransferIotas(1234).Post()
	require.NoError(t, ctx.Err)

	bal.Account += 1234
	bal.Chain += ctx.GasFee
	bal.Originator -= ctx.GasFee
	bal.VerifyBalances(t)

	g := testwasmlib.ScFuncs.TakeBalance(ctx)
	g.Func.Post()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, bal.Account, g.Results.Iotas().Value())

	bal.Chain += ctx.GasFee
	bal.Originator += ctx.Dust - ctx.GasFee
	bal.VerifyBalances(t)

	v := testwasmlib.ScFuncs.IotaBalance(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	ctx.Balances()
	require.True(t, v.Results.Iotas().Exists())
	require.EqualValues(t, bal.Account, v.Results.Iotas().Value())

	bal.VerifyBalances(t)
}

func TestTakeNoAllowance(t *testing.T) {
	ctx := setupTest(t)
	bal := ctx.Balances()

	// FuncParamTypes without params does nothing to SC balance
	// because it does not take the allowance
	f := testwasmlib.ScFuncs.ParamTypes(ctx)
	f.Func.TransferIotas(1234).Post()
	require.NoError(t, ctx.Err)
	ctx.Balances()

	bal.Chain += ctx.GasFee
	bal.Originator += 1234 - ctx.GasFee
	bal.VerifyBalances(t)

	g := testwasmlib.ScFuncs.TakeBalance(ctx)
	g.Func.Post()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, bal.Account, g.Results.Iotas().Value())

	bal.Chain += ctx.GasFee
	bal.Originator += ctx.Dust - ctx.GasFee
	bal.VerifyBalances(t)

	v := testwasmlib.ScFuncs.IotaBalance(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	ctx.Balances()
	require.True(t, v.Results.Iotas().Exists())
	require.EqualValues(t, bal.Account, v.Results.Iotas().Value())

	bal.VerifyBalances(t)
}

func TestRandom(t *testing.T) {
	ctx := setupTest(t)

	f := testwasmlib.ScFuncs.Random(ctx)
	f.Func.Post()
	require.NoError(t, ctx.Err)

	v := testwasmlib.ScFuncs.GetRandom(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	random := v.Results.Random().Value()
	require.True(t, random < 1000)
	fmt.Printf("Random value: %d\n", random)
}

func TestMultiRandom(t *testing.T) {
	ctx := setupTest(t)

	numbers := make([]uint64, 0)
	for i := 0; i < 10; i++ {
		f := testwasmlib.ScFuncs.Random(ctx)
		f.Func.Post()
		require.NoError(t, ctx.Err)

		v := testwasmlib.ScFuncs.GetRandom(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		random := v.Results.Random().Value()
		require.True(t, random < 1000)
		numbers = append(numbers, random)
	}

	for _, number := range numbers {
		fmt.Printf("Random value: %d\n", number)
	}
}
