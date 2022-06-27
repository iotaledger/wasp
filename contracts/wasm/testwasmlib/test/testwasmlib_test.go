// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreblocklog"
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
	allLengths = []int{
		wasmtypes.ScAddressLength,
		wasmtypes.ScAddressLength + 1,
		wasmtypes.ScBoolLength,
		wasmtypes.ScChainIDLength,
		wasmtypes.ScHashLength,
		wasmtypes.ScHnameLength,
		wasmtypes.ScInt8Length,
		wasmtypes.ScInt16Length,
		wasmtypes.ScInt32Length,
		wasmtypes.ScInt64Length,
		wasmtypes.ScNftIDLength,
		wasmtypes.ScRequestIDLength,
		wasmtypes.ScTokenIDLength,
		wasmtypes.ScUint8Length,
		wasmtypes.ScUint16Length,
		wasmtypes.ScUint32Length,
		wasmtypes.ScUint64Length,
	}
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
	zeroHash = make([]byte, wasmtypes.ScHashLength)
)

func setupTest(t *testing.T) *wasmsolo.SoloContext {
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
	pt.Params.Address().SetValue(ctx.CurrentChainID().Address())
	pt.Params.AgentID().SetValue(ctx.AccountID())
	pt.Params.BigInt().SetValue(wasmtypes.BigIntFromString("100000000000000000000"))
	pt.Params.Bool().SetValue(true)
	pt.Params.Bytes().SetValue([]byte("these are bytes"))
	pt.Params.ChainID().SetValue(ctx.CurrentChainID())
	pt.Params.Hash().SetValue(wasmtypes.HashFromBytes([]byte("0123456789abcdeffedcba9876543210")))
	pt.Params.Hname().SetValue(testwasmlib.HScName)
	pt.Params.Int8().SetValue(-123)
	pt.Params.Int16().SetValue(-12345)
	pt.Params.Int32().SetValue(-1234567890)
	pt.Params.Int64().SetValue(-1234567890123456789)
	pt.Params.NftID().SetValue(wasmtypes.NftIDFromBytes([]byte("abcdefghijklmnopqrstuvwxyz123456")))
	pt.Params.RequestID().SetValue(wasmtypes.RequestIDFromBytes([]byte("abcdefghijklmnopqrstuvwxyz123456\x00\x00")))
	pt.Params.String().SetValue("this is a string")
	pt.Params.TokenID().SetValue(wasmtypes.TokenIDFromBytes([]byte("abcdefghijklmnopqrstuvwxyz1234567890AB")))
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
			if param == testwasmlib.ParamAgentID {
				bytes[0] = wasmtypes.ScAgentIDAddress
			}
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
			// note that zero lengths are valid and will return a default value

			// no need to check bool/int8/uint8
			if allLengths[index] != 1 {
				testInvalidSizeParams(t, ctx, param, make([]byte, 1))
				testInvalidSizeParams(t, ctx, param, make([]byte, allLengths[index]-1))
			}
			testInvalidSizeParams(t, ctx, param, make([]byte, allLengths[index]+1))
		})
	}
}

func testInvalidSizeParams(t *testing.T, ctx *wasmsolo.SoloContext, param string, bytes []byte) {
	invalidLength := fmt.Sprintf("invalid %s%s length", strings.ToUpper(param[:1]), param[1:])
	pt := testwasmlib.ScFuncs.ParamTypes(ctx)
	if param == testwasmlib.ParamAgentID {
		bytes[0] = wasmtypes.ScAgentIDAddress
	}
	pt.Params.Param().GetBytes(param).SetValue(bytes)
	pt.Func.Post()
	require.Error(t, ctx.Err)
	require.Contains(t, ctx.Err.Error(), invalidLength)
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
	require.EqualValues(t, 230, len(rec.Results.Record().Value()))
}

func TestTakeAllowance(t *testing.T) {
	ctx := setupTest(t)
	bal := ctx.Balances()

	f := testwasmlib.ScFuncs.TakeAllowance(ctx)
	const iotasToSend = 1 * iscp.Mi
	f.Func.TransferIotas(iotasToSend).Post()
	require.NoError(t, ctx.Err)

	bal.Account += iotasToSend
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
	bal = ctx.Balances()
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
	const iotasToSend = 1 * iscp.Mi
	f.Func.TransferIotas(iotasToSend).Post()
	require.NoError(t, ctx.Err)
	ctx.Balances()

	bal.Chain += ctx.GasFee
	bal.Originator += iotasToSend - ctx.GasFee
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
	bal = ctx.Balances()
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

func TestWasmTypes(t *testing.T) {
	ctx := setupTest(t)

	// check chain id
	scChainID := ctx.CurrentChainID()
	chainID := ctx.Chain.ChainID
	require.True(t, scChainID == wasmtypes.ChainIDFromBytes(wasmtypes.ChainIDToBytes(scChainID)))
	require.True(t, scChainID == wasmtypes.ChainIDFromString(wasmtypes.ChainIDToString(scChainID)))
	require.EqualValues(t, scChainID.Bytes(), chainID.Bytes())
	require.EqualValues(t, scChainID.String(), chainID.String())

	// check alias address
	scAliasAddress := scChainID.Address()
	aliasAddress := chainID.AsAddress()
	checkAddress(t, ctx, scAliasAddress, aliasAddress)

	// check ed25519 address
	scEd25519Address := ctx.Originator().ScAgentID().Address()
	ed25519Address := ctx.Chain.OriginatorAddress
	checkAddress(t, ctx, scEd25519Address, ed25519Address)

	// check nft address (currently simply use
	// serialized alias address and overwrite the kind byte)
	nftBytes := scAliasAddress.Bytes()
	nftBytes[0] = wasmtypes.ScAddressNFT
	scNftAddress := wasmtypes.AddressFromBytes(nftBytes)
	nftBytes[0] = byte(iotago.AddressNFT)
	nftAddress, _, _ := iscp.AddressFromBytes(nftBytes)
	checkAddress(t, ctx, scNftAddress, nftAddress)

	// check agent id of alias address (hname zero)
	scAgentID := wasmtypes.NewScAgentIDFromAddress(scAliasAddress)
	agentID := iscp.NewAgentID(aliasAddress)
	checkAgentID(t, ctx, scAgentID, agentID)

	// check agent id of ed25519 address (hname zero)
	scAgentID = wasmtypes.NewScAgentIDFromAddress(scEd25519Address)
	agentID = iscp.NewAgentID(ed25519Address)
	checkAgentID(t, ctx, scAgentID, agentID)

	// check agent id of NFT address (hname zero)
	scAgentID = wasmtypes.NewScAgentIDFromAddress(scNftAddress)
	agentID = iscp.NewAgentID(nftAddress)
	checkAgentID(t, ctx, scAgentID, agentID)

	// check agent id of contract (hname non-zero)
	scAgentID = wasmtypes.NewScAgentID(scAliasAddress, testwasmlib.HScName)
	agentID = iscp.NewContractAgentID(chainID, iscp.Hname(testwasmlib.HScName))
	checkAgentID(t, ctx, scAgentID, agentID)

	// check nil agent id
	scAgentID = wasmtypes.ScAgentID{}
	agentID = &iscp.NilAgentID{}
	checkAgentID(t, ctx, scAgentID, agentID)

	// eth
	addressEth := "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
	checkerEth := testwasmlib.ScFuncs.CheckEthAddressAndAgentID(ctx)
	checkerEth.Params.EthAddress().SetValue(addressEth)
	checkerEth.Func.Call()
	require.NoError(t, ctx.Err)

	goInt8 := int8(math.MaxInt8)
	require.Equal(t, goInt8, wasmtypes.Int8FromBytes(wasmtypes.Int8ToBytes(goInt8)))
	require.Equal(t, goInt8, wasmtypes.Int8FromString(wasmtypes.Int8ToString(goInt8)))
	goInt8 = math.MinInt8
	require.Equal(t, goInt8, wasmtypes.Int8FromBytes(wasmtypes.Int8ToBytes(goInt8)))
	require.Equal(t, goInt8, wasmtypes.Int8FromString(wasmtypes.Int8ToString(goInt8)))
	goInt8 = 1
	require.Equal(t, goInt8, wasmtypes.Int8FromBytes(wasmtypes.Int8ToBytes(goInt8)))
	require.Equal(t, goInt8, wasmtypes.Int8FromString(wasmtypes.Int8ToString(goInt8)))
	goInt8 = 0
	require.Equal(t, goInt8, wasmtypes.Int8FromBytes(wasmtypes.Int8ToBytes(goInt8)))
	require.Equal(t, goInt8, wasmtypes.Int8FromString(wasmtypes.Int8ToString(goInt8)))
	goInt8 = -1
	require.Equal(t, goInt8, wasmtypes.Int8FromBytes(wasmtypes.Int8ToBytes(goInt8)))
	require.Equal(t, goInt8, wasmtypes.Int8FromString(wasmtypes.Int8ToString(goInt8)))
	goUint8 := uint8(0)
	require.Equal(t, goUint8, wasmtypes.Uint8FromBytes(wasmtypes.Uint8ToBytes(goUint8)))
	require.Equal(t, goUint8, wasmtypes.Uint8FromString(wasmtypes.Uint8ToString(goUint8)))
	goUint8 = math.MaxUint8
	require.Equal(t, goUint8, wasmtypes.Uint8FromBytes(wasmtypes.Uint8ToBytes(goUint8)))
	require.Equal(t, goUint8, wasmtypes.Uint8FromString(wasmtypes.Uint8ToString(goUint8)))

	goInt16 := int16(math.MaxInt16)
	require.Equal(t, goInt16, wasmtypes.Int16FromBytes(wasmtypes.Int16ToBytes(goInt16)))
	require.Equal(t, goInt16, wasmtypes.Int16FromString(wasmtypes.Int16ToString(goInt16)))
	goInt16 = math.MinInt16
	require.Equal(t, goInt16, wasmtypes.Int16FromBytes(wasmtypes.Int16ToBytes(goInt16)))
	require.Equal(t, goInt16, wasmtypes.Int16FromString(wasmtypes.Int16ToString(goInt16)))
	goInt16 = 1
	require.Equal(t, goInt16, wasmtypes.Int16FromBytes(wasmtypes.Int16ToBytes(goInt16)))
	require.Equal(t, goInt16, wasmtypes.Int16FromString(wasmtypes.Int16ToString(goInt16)))
	goInt16 = 0
	require.Equal(t, goInt16, wasmtypes.Int16FromBytes(wasmtypes.Int16ToBytes(goInt16)))
	require.Equal(t, goInt16, wasmtypes.Int16FromString(wasmtypes.Int16ToString(goInt16)))
	goInt16 = -1
	require.Equal(t, goInt16, wasmtypes.Int16FromBytes(wasmtypes.Int16ToBytes(goInt16)))
	require.Equal(t, goInt16, wasmtypes.Int16FromString(wasmtypes.Int16ToString(goInt16)))
	goUint16 := uint16(0)
	require.Equal(t, goUint16, wasmtypes.Uint16FromBytes(wasmtypes.Uint16ToBytes(goUint16)))
	require.Equal(t, goUint16, wasmtypes.Uint16FromString(wasmtypes.Uint16ToString(goUint16)))
	goUint16 = math.MaxUint16
	require.Equal(t, goUint16, wasmtypes.Uint16FromBytes(wasmtypes.Uint16ToBytes(goUint16)))
	require.Equal(t, goUint16, wasmtypes.Uint16FromString(wasmtypes.Uint16ToString(goUint16)))

	goInt32 := int32(math.MaxInt32)
	require.Equal(t, goInt32, wasmtypes.Int32FromBytes(wasmtypes.Int32ToBytes(goInt32)))
	require.Equal(t, goInt32, wasmtypes.Int32FromString(wasmtypes.Int32ToString(goInt32)))
	goInt32 = math.MinInt32
	require.Equal(t, goInt32, wasmtypes.Int32FromBytes(wasmtypes.Int32ToBytes(goInt32)))
	require.Equal(t, goInt32, wasmtypes.Int32FromString(wasmtypes.Int32ToString(goInt32)))
	goInt32 = 1
	require.Equal(t, goInt32, wasmtypes.Int32FromBytes(wasmtypes.Int32ToBytes(goInt32)))
	require.Equal(t, goInt32, wasmtypes.Int32FromString(wasmtypes.Int32ToString(goInt32)))
	goInt32 = 0
	require.Equal(t, goInt32, wasmtypes.Int32FromBytes(wasmtypes.Int32ToBytes(goInt32)))
	require.Equal(t, goInt32, wasmtypes.Int32FromString(wasmtypes.Int32ToString(goInt32)))
	goInt32 = -1
	require.Equal(t, goInt32, wasmtypes.Int32FromBytes(wasmtypes.Int32ToBytes(goInt32)))
	require.Equal(t, goInt32, wasmtypes.Int32FromString(wasmtypes.Int32ToString(goInt32)))
	goUint32 := uint32(0)
	require.Equal(t, goUint32, wasmtypes.Uint32FromBytes(wasmtypes.Uint32ToBytes(goUint32)))
	require.Equal(t, goUint32, wasmtypes.Uint32FromString(wasmtypes.Uint32ToString(goUint32)))
	goUint32 = math.MaxUint32
	require.Equal(t, goUint32, wasmtypes.Uint32FromBytes(wasmtypes.Uint32ToBytes(goUint32)))
	require.Equal(t, goUint32, wasmtypes.Uint32FromString(wasmtypes.Uint32ToString(goUint32)))

	goInt64 := int64(math.MaxInt64)
	require.Equal(t, goInt64, wasmtypes.Int64FromBytes(wasmtypes.Int64ToBytes(goInt64)))
	require.Equal(t, goInt64, wasmtypes.Int64FromString(wasmtypes.Int64ToString(goInt64)))
	goInt64 = math.MinInt64
	require.Equal(t, goInt64, wasmtypes.Int64FromBytes(wasmtypes.Int64ToBytes(goInt64)))
	require.Equal(t, goInt64, wasmtypes.Int64FromString(wasmtypes.Int64ToString(goInt64)))
	goInt64 = 1
	require.Equal(t, goInt64, wasmtypes.Int64FromBytes(wasmtypes.Int64ToBytes(goInt64)))
	require.Equal(t, goInt64, wasmtypes.Int64FromString(wasmtypes.Int64ToString(goInt64)))
	goInt64 = 0
	require.Equal(t, goInt64, wasmtypes.Int64FromBytes(wasmtypes.Int64ToBytes(goInt64)))
	require.Equal(t, goInt64, wasmtypes.Int64FromString(wasmtypes.Int64ToString(goInt64)))
	goInt64 = -1
	require.Equal(t, goInt64, wasmtypes.Int64FromBytes(wasmtypes.Int64ToBytes(goInt64)))
	require.Equal(t, goInt64, wasmtypes.Int64FromString(wasmtypes.Int64ToString(goInt64)))
	goUint64 := uint64(0)
	require.Equal(t, goUint64, wasmtypes.Uint64FromBytes(wasmtypes.Uint64ToBytes(goUint64)))
	require.Equal(t, goUint64, wasmtypes.Uint64FromString(wasmtypes.Uint64ToString(goUint64)))
	goUint64 = math.MaxUint64
	require.Equal(t, goUint64, wasmtypes.Uint64FromBytes(wasmtypes.Uint64ToBytes(goUint64)))
	require.Equal(t, goUint64, wasmtypes.Uint64FromString(wasmtypes.Uint64ToString(goUint64)))

	scBigInt := wasmtypes.NewScBigInt(123213)
	bigInt := big.NewInt(123213)
	checkBigInt(t, ctx, scBigInt, bigInt)

	goBool := true
	require.Equal(t, goBool, wasmtypes.BoolFromBytes(wasmtypes.BoolToBytes(goBool)))
	require.Equal(t, goBool, wasmtypes.BoolFromString(wasmtypes.BoolToString(goBool)))
	goBool = false
	require.Equal(t, goBool, wasmtypes.BoolFromBytes(wasmtypes.BoolToBytes(goBool)))
	require.Equal(t, goBool, wasmtypes.BoolFromString(wasmtypes.BoolToString(goBool)))

	goBytes := []byte{0xc3, 0x77, 0xf3, 0xf1}
	require.Equal(t, goBytes, wasmtypes.BytesFromBytes(wasmtypes.BytesToBytes(goBytes)))
	require.Equal(t, goBytes, wasmtypes.BytesFromString(wasmtypes.BytesToString(goBytes)))

	hashString := "7c106d42ca17fdbfb03f6b45b91effcef2cff61215a3552dbc1ab8fd46817719"
	hash, err := hashing.HashValueFromHex(hashString)
	require.NoError(t, err)
	scHash := wasmtypes.HashFromString(hashString)
	checkHash(t, ctx, scHash, hash)

	scHname := testwasmlib.HScName
	require.Equal(t, scHname, wasmtypes.HnameFromString(wasmtypes.HnameToString(scHname)))
	require.Equal(t, scHname.String(), wasmtypes.HnameToString(scHname))
	require.Equal(t, scHname, wasmtypes.HnameFromBytes(wasmtypes.HnameToBytes(scHname)))
	require.Equal(t, scHname.Bytes(), wasmtypes.HnameToBytes(scHname))

	goString := "this is a go string example"
	require.Equal(t, goString, wasmtypes.StringToString(wasmtypes.StringFromString(goString)))
	require.Equal(t, []byte(goString), wasmtypes.StringToBytes(wasmtypes.StringFromBytes([]byte(goString))))

	tokenID, err := getTokenID(ctx)
	require.NoError(t, err)
	scTokenID := ctx.Cvt.ScTokenID(&tokenID)
	checkTokenID(t, ctx, scTokenID, tokenID)

	nftID, err := getNftID(ctx)
	require.NoError(t, err)
	scNftID := ctx.Cvt.ScNftID(&nftID)
	checkNftID(t, ctx, scNftID, nftID)

	blockNum := uint32(3)
	ctxBlocklog := ctx.SoloContextForCore(t, coreblocklog.ScName, coreblocklog.OnLoad)
	require.NoError(t, ctxBlocklog.Err)
	fblocklog := coreblocklog.ScFuncs.GetRequestIDsForBlock(ctxBlocklog)
	fblocklog.Params.BlockIndex().SetValue(blockNum)
	fblocklog.Func.Call()
	scReq := fblocklog.Results.RequestID().GetRequestID(0).Value()
	req := ctxBlocklog.Chain.GetRequestIDsForBlock(blockNum)[0]
	checkRequestID(t, ctx, scReq, req)
}

func getTokenID(ctx *wasmsolo.SoloContext) (tokenID iotago.NativeTokenID, err error) {
	maxSupply := 100
	fp := ctx.Chain.NewFoundryParams(ctx.Cvt.ToBigInt(maxSupply))
	_, tokenID, err = fp.CreateFoundry()
	if err != nil {
		return iotago.NativeTokenID{}, err
	}
	return tokenID, nil
}

func getNftID(ctx *wasmsolo.SoloContext) (iotago.NFTID, error) {
	agent := ctx.NewSoloAgent()
	addr, ok := iscp.AddressFromAgentID(agent.AgentID())
	if !ok {
		return iotago.NFTID{}, fmt.Errorf("can't get address from AgentID")
	}
	nftInfo, err := ctx.Chain.Env.MintNFTL1(agent.Pair, addr, []byte("test data"))
	if err != nil {
		return iotago.NFTID{}, err
	}
	return nftInfo.NFTID, nil
}

func checkBigInt(t *testing.T, ctx *wasmsolo.SoloContext, scBigInt wasmtypes.ScBigInt, bigInt *big.Int) {
	require.Equal(t, scBigInt, wasmtypes.BigIntFromBytes(wasmtypes.BigIntToBytes(scBigInt)))
	require.Equal(t, bigInt.Bytes(), scBigInt.Bytes())
	require.Equal(t, scBigInt, wasmtypes.BigIntFromString(wasmtypes.BigIntToString(scBigInt)))
	require.Equal(t, bigInt.String(), scBigInt.String())

	bigIntBytes := bigInt.Bytes()
	bigIntString := bigInt.String()
	checker := testwasmlib.ScFuncs.CheckBigInt(ctx)
	checker.Params.ScBigInt().SetValue(scBigInt)
	checker.Params.BigIntBytes().SetValue(bigIntBytes)
	checker.Params.BigIntString().SetValue(bigIntString)
	checker.Func.Call()
	require.NoError(t, ctx.Err, fmt.Sprintf("scBigInt: %s, bigInt: %s", scBigInt.String(), bigInt.String()))
}

//nolint:dupl
func checkAgentID(t *testing.T, ctx *wasmsolo.SoloContext, scAgentID wasmtypes.ScAgentID, agentID iscp.AgentID) {
	agentBytes := agentID.Bytes()
	agentString := agentID.String()

	require.EqualValues(t, scAgentID.Bytes(), agentBytes)
	require.EqualValues(t, scAgentID.String(), agentString)
	require.True(t, scAgentID == wasmtypes.AgentIDFromBytes(wasmtypes.AgentIDToBytes(scAgentID)))
	require.True(t, scAgentID == wasmtypes.AgentIDFromString(wasmtypes.AgentIDToString(scAgentID)))

	checker := testwasmlib.ScFuncs.CheckAgentID(ctx)
	checker.Params.ScAgentID().SetValue(scAgentID)
	checker.Params.AgentBytes().SetValue(agentBytes)
	checker.Params.AgentString().SetValue(agentString)
	checker.Func.Call()
	require.NoError(t, ctx.Err)
}

func checkAddress(t *testing.T, ctx *wasmsolo.SoloContext, scAddress wasmtypes.ScAddress, address iotago.Address) {
	addressBytes := iscp.BytesFromAddress(address)
	addressString := address.Bech32(parameters.L1.Protocol.Bech32HRP)

	require.EqualValues(t, scAddress.Bytes(), addressBytes)
	require.EqualValues(t, scAddress.String(), addressString)
	require.True(t, scAddress == wasmtypes.AddressFromBytes(wasmtypes.AddressToBytes(scAddress)))
	require.True(t, scAddress == wasmtypes.AddressFromString(wasmtypes.AddressToString(scAddress)))

	checker := testwasmlib.ScFuncs.CheckAddress(ctx)
	checker.Params.ScAddress().SetValue(scAddress)
	checker.Params.AddressBytes().SetValue(addressBytes)
	checker.Params.AddressString().SetValue(addressString)
	checker.Func.Call()
	require.NoError(t, ctx.Err)
}

//nolint:dupl
func checkHash(t *testing.T, ctx *wasmsolo.SoloContext, scHash wasmtypes.ScHash, hash hashing.HashValue) {
	hashBytes := hash.Bytes()
	hashString := hash.String()

	require.EqualValues(t, scHash.Bytes(), hashBytes)
	require.EqualValues(t, scHash.String(), hashString)
	require.True(t, scHash == wasmtypes.HashFromBytes(wasmtypes.HashToBytes(scHash)))
	require.True(t, scHash == wasmtypes.HashFromString(wasmtypes.HashToString(scHash)))

	checker := testwasmlib.ScFuncs.CheckHash(ctx)
	checker.Params.ScHash().SetValue(scHash)
	checker.Params.HashBytes().SetValue(hashBytes)
	checker.Params.HashString().SetValue(hashString)
	checker.Func.Call()
	require.NoError(t, ctx.Err)
}

//nolint:dupl
func checkNftID(t *testing.T, ctx *wasmsolo.SoloContext, scNftID wasmtypes.ScNftID, nftID iotago.NFTID) {
	nftIDBytes := nftID[:]
	nftIDString := nftID.String()
	require.Equal(t, scNftID, wasmtypes.NftIDFromString(wasmtypes.NftIDToString(scNftID)))
	require.Equal(t, scNftID, wasmtypes.NftIDFromBytes(wasmtypes.NftIDToBytes(scNftID)))
	require.Equal(t, scNftID.String(), nftID.String())
	require.Equal(t, scNftID.Bytes(), nftID[:])

	checker := testwasmlib.ScFuncs.CheckNftID(ctx)
	checker.Params.ScNftID().SetValue(scNftID)
	checker.Params.NftIDBytes().SetValue(nftIDBytes)
	checker.Params.NftIDString().SetValue(nftIDString)
	checker.Func.Call()
	require.NoError(t, ctx.Err)
}

//nolint:dupl
func checkTokenID(t *testing.T, ctx *wasmsolo.SoloContext, scTokenID wasmtypes.ScTokenID, tokenID iotago.NativeTokenID) {
	tokenIDBytes := tokenID[:]
	tokenIDString := tokenID.String()

	require.Equal(t, scTokenID, wasmtypes.TokenIDFromString(wasmtypes.TokenIDToString(scTokenID)))
	require.Equal(t, scTokenID, wasmtypes.TokenIDFromBytes(wasmtypes.TokenIDToBytes(scTokenID)))
	require.Equal(t, scTokenID.String(), tokenID.String())
	require.Equal(t, scTokenID.Bytes(), tokenID[:])

	checker := testwasmlib.ScFuncs.CheckTokenID(ctx)
	checker.Params.ScTokenID().SetValue(scTokenID)
	checker.Params.TokenIDBytes().SetValue(tokenIDBytes)
	checker.Params.TokenIDString().SetValue(tokenIDString)
	checker.Func.Call()
	require.NoError(t, ctx.Err)
}

func checkRequestID(t *testing.T, ctx *wasmsolo.SoloContext, scRequestID wasmtypes.ScRequestID, requestID iscp.RequestID) {
	requestIDBytes := requestID.Bytes()
	requestIDString := requestID.String()

	require.Equal(t, scRequestID, wasmtypes.RequestIDFromBytes(wasmtypes.RequestIDToBytes(scRequestID)))
	require.Equal(t, scRequestID, wasmtypes.RequestIDFromString(wasmtypes.RequestIDToString(scRequestID)))
	require.Equal(t, scRequestID.Bytes(), requestID.Bytes())
	require.Equal(t, scRequestID.String(), requestID.String())

	checker := testwasmlib.ScFuncs.CheckRequestID(ctx)
	checker.Params.ScRequestID().SetValue(scRequestID)
	checker.Params.RequestIDBytes().SetValue(requestIDBytes)
	checker.Params.RequestIDString().SetValue(requestIDString)
	checker.Func.Call()
	require.NoError(t, ctx.Err)
}
