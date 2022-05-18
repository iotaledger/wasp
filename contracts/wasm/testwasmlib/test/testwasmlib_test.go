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
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
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
	scChainID := ctx.ChainID()
	chainID := ctx.Chain.ChainID
	require.True(t, scChainID == wasmtypes.ChainIDFromBytes(wasmtypes.ChainIDToBytes(scChainID)))
	require.True(t, scChainID == wasmtypes.ChainIDFromString(wasmtypes.ChainIDToString(scChainID)))
	require.EqualValues(t, scChainID.Bytes(), chainID.Bytes())
	require.EqualValues(t, scChainID.String(), chainID.String())

	// check alias address
	scAliasAddress := scChainID.Address()
	aliasAddress := chainID.AsAddress()
	require.True(t, scAliasAddress == wasmtypes.AddressFromBytes(wasmtypes.AddressToBytes(scAliasAddress)))
	require.True(t, scAliasAddress == wasmtypes.AddressFromString(wasmtypes.AddressToString(scAliasAddress)))
	require.EqualValues(t, scAliasAddress.Bytes(), iscp.BytesFromAddress(aliasAddress))
	// TODO require.EqualValues(t, scAddress.String(), address.String())

	// check ed25519 address
	scEd25519Address := ctx.Originator().ScAgentID().Address()
	ed25519Address := ctx.Chain.OriginatorAddress
	require.True(t, scEd25519Address == wasmtypes.AddressFromBytes(wasmtypes.AddressToBytes(scEd25519Address)))
	require.True(t, scEd25519Address == wasmtypes.AddressFromString(wasmtypes.AddressToString(scEd25519Address)))
	require.EqualValues(t, scEd25519Address.Bytes(), iscp.BytesFromAddress(ed25519Address))
	// TODO require.EqualValues(t, scAddress.String(), address.String())

	// check nft address (currently simply use
	// serialized alias address and overwrite the kind byte)
	nftBytes := scAliasAddress.Bytes()
	nftBytes[0] = wasmtypes.ScAddressNFT
	scNftAddress := wasmtypes.AddressFromBytes(nftBytes)
	nftBytes[0] = byte(iotago.AddressNFT)
	nftAddress, _, _ := iscp.AddressFromBytes(nftBytes)
	require.True(t, scNftAddress == wasmtypes.AddressFromBytes(wasmtypes.AddressToBytes(scNftAddress)))
	require.True(t, scNftAddress == wasmtypes.AddressFromString(wasmtypes.AddressToString(scNftAddress)))
	require.EqualValues(t, scNftAddress.Bytes(), iscp.BytesFromAddress(nftAddress))
	// TODO require.EqualValues(t, scAddress.String(), address.String())

	// check agent id of alias address (hname zero)
	scAgentID := wasmtypes.NewScAgentIDFromAddress(scAliasAddress)
	agentID := iscp.NewAgentID(aliasAddress)
	require.True(t, scAgentID == wasmtypes.AgentIDFromBytes(wasmtypes.AgentIDToBytes(scAgentID)))
	//	require.True(t, scAgentID == wasmtypes.AgentIDFromString(wasmtypes.AgentIDToString(scAgentID)))
	require.EqualValues(t, scAgentID.Bytes(), agentID.Bytes())
	// TODO require.EqualValues(t, scAgentID.String(), agentID.String("atoi"))

	// check agent id of ed25519 address (hname zero)
	scAgentID = wasmtypes.NewScAgentIDFromAddress(scEd25519Address)
	agentID = iscp.NewAgentID(ed25519Address)
	require.True(t, scAgentID == wasmtypes.AgentIDFromBytes(wasmtypes.AgentIDToBytes(scAgentID)))
	//	require.True(t, scAgentID == wasmtypes.AgentIDFromString(wasmtypes.AgentIDToString(scAgentID)))
	require.EqualValues(t, scAgentID.Bytes(), agentID.Bytes())
	// TODO require.EqualValues(t, scAgentID.String(), agentID.String("atoi"))

	// check agent id of NFT address (hname zero)
	scAgentID = wasmtypes.NewScAgentIDFromAddress(scNftAddress)
	agentID = iscp.NewAgentID(nftAddress)
	require.True(t, scAgentID == wasmtypes.AgentIDFromBytes(wasmtypes.AgentIDToBytes(scAgentID)))
	//	require.True(t, scAgentID == wasmtypes.AgentIDFromString(wasmtypes.AgentIDToString(scAgentID)))
	require.EqualValues(t, scAgentID.Bytes(), agentID.Bytes())
	// TODO require.EqualValues(t, scAgentID.String(), agentID.String("atoi"))

	// check agent id of contract (hname non-zero)
	scAgentID = wasmtypes.NewScAgentID(scAliasAddress, testwasmlib.HScName)
	agentID = iscp.NewContractAgentID(chainID, iscp.Hname(testwasmlib.HScName))
	require.True(t, scAgentID == wasmtypes.AgentIDFromBytes(wasmtypes.AgentIDToBytes(scAgentID)))
	//	require.True(t, scAgentID == wasmtypes.AgentIDFromString(wasmtypes.AgentIDToString(scAgentID)))
	require.EqualValues(t, scAgentID.Bytes(), agentID.Bytes())
	// TODO require.EqualValues(t, scAgentID.String(), agentID.String("atoi"))
}
