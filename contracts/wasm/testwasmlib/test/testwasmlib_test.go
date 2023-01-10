// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/erc721/go/erc721"
	"github.com/iotaledger/wasp/contracts/wasm/erc721/go/erc721impl"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlibimpl"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreblocklog"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
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
	return wasmsolo.NewSoloContext(t, testwasmlib.ScName, testwasmlibimpl.OnDispatch)
}

func TestDeploy(t *testing.T) {
	ctx := setupTest(t)
	require.NoError(t, ctx.Err)
	require.NoError(t, ctx.ContractExists(testwasmlib.ScName))
}

func TestDeployErc721Too(t *testing.T) {
	//NOTE: when trying to crash a node using WasmTime we need to run some Wasm code
	//*wasmsolo.RsWasm = true
	ctx := setupTest(t)
	require.NoError(t, ctx.Err)
	require.NoError(t, ctx.ContractExists(testwasmlib.ScName))

	init := erc721.ScFuncs.Init(nil)
	init.Params.Name().SetValue("Name")
	init.Params.Symbol().SetValue("Symbol")

	ctxErc721 := wasmsolo.NewSoloContextForChain(t, ctx.Chain, nil, erc721.ScName, erc721impl.OnDispatch, init.Func)
	require.NoError(t, ctxErc721.Err)
	require.NoError(t, ctxErc721.ContractExists(erc721.ScName))

	mint := erc721.ScFuncs.Mint(ctxErc721)
	tokenID := wasmtypes.HashFromString("0xd4735e3a265e16eee03f59718b9b5d03019c07d8b6c51f90da3a666eec13ab37")
	mint.Params.TokenID().SetValue(tokenID)
	mint.Params.TokenURI().SetValue("information about the token")
	mint.Func.Post()
	require.NoError(t, ctxErc721.Err)

	oo := erc721.ScFuncs.OwnerOf(ctxErc721)
	oo.Params.TokenID().SetValue(tokenID)
	oo.Func.Call()
	require.NoError(t, ctxErc721.Err)
	require.EqualValues(t, oo.Results.Owner().Value(), ctxErc721.ChainOwnerID())

	// NOTE: this post() can bring a node down when reactivating the commented out line
	// in WasmTimeVM.RunScFunction() because it triggers a Rust panic() in the WasmTime code.
	// We need to find out how to catch such an error from within Go.
	// Note that this post() triggers a call() to an Erc721 view, which then triggers the
	// Rust panic() after returning from the call.
	f := testwasmlib.ScFuncs.VerifyErc721(ctx)
	f.Params.TokenHash().SetValue(tokenID)
	f.Func.Post()
	require.NoError(t, ctx.Err)
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
				).AddBaseTokens(1).WithMaxAffordableGasBudget()
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
	const tokensToSend = 1 * isc.Million
	f.Func.TransferBaseTokens(tokensToSend).Post()
	require.NoError(t, ctx.Err)

	bal.Account += tokensToSend
	bal.Chain += ctx.GasFee
	bal.Originator -= ctx.GasFee
	bal.VerifyBalances(t)

	g := testwasmlib.ScFuncs.TakeBalance(ctx)
	g.Func.Post()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, bal.Account, g.Results.Tokens().Value())

	bal.Chain += ctx.GasFee
	bal.Originator += ctx.StorageDeposit - ctx.GasFee
	bal.VerifyBalances(t)

	v := testwasmlib.ScFuncs.TokenBalance(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	bal = ctx.Balances()
	require.True(t, v.Results.Tokens().Exists())
	require.EqualValues(t, bal.Account, v.Results.Tokens().Value())

	bal.VerifyBalances(t)
}

func TestTakeNoAllowance(t *testing.T) {
	ctx := setupTest(t)
	bal := ctx.Balances()

	// FuncParamTypes without params does nothing to SC balance
	// because it does not take the allowance
	f := testwasmlib.ScFuncs.ParamTypes(ctx)
	const tokensToSend = 1 * isc.Million
	f.Func.TransferBaseTokens(tokensToSend).Post()
	require.NoError(t, ctx.Err)
	ctx.Balances()

	bal.Chain += ctx.GasFee
	bal.Originator += tokensToSend - ctx.GasFee
	bal.VerifyBalances(t)

	g := testwasmlib.ScFuncs.TakeBalance(ctx)
	g.Func.Post()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, bal.Account, g.Results.Tokens().Value())

	bal.Chain += ctx.GasFee
	bal.Originator += ctx.StorageDeposit - ctx.GasFee
	bal.VerifyBalances(t)

	v := testwasmlib.ScFuncs.TokenBalance(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	bal = ctx.Balances()
	require.True(t, v.Results.Tokens().Exists())
	require.EqualValues(t, bal.Account, v.Results.Tokens().Value())

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
	nftAddress, _, _ := isc.AddressFromBytes(nftBytes)
	checkAddress(t, ctx, scNftAddress, nftAddress)

	// check agent id of alias address (hname zero)
	scAgentID := wasmtypes.NewScAgentIDFromAddress(scAliasAddress)
	agentID := isc.NewAgentID(aliasAddress)
	checkAgentID(t, ctx, scAgentID, agentID)

	// check agent id of ed25519 address (hname zero)
	scAgentID = wasmtypes.NewScAgentIDFromAddress(scEd25519Address)
	agentID = isc.NewAgentID(ed25519Address)
	checkAgentID(t, ctx, scAgentID, agentID)

	// check agent id of NFT address (hname zero)
	scAgentID = wasmtypes.NewScAgentIDFromAddress(scNftAddress)
	agentID = isc.NewAgentID(nftAddress)
	checkAgentID(t, ctx, scAgentID, agentID)

	// check agent id of contract (hname non-zero)
	scAgentID = wasmtypes.NewScAgentID(scAliasAddress, testwasmlib.HScName)
	agentID = isc.NewContractAgentID(chainID, isc.Hname(testwasmlib.HScName))
	checkAgentID(t, ctx, scAgentID, agentID)

	// check nil agent id
	scAgentID = wasmtypes.ScAgentID{}
	agentID = &isc.NilAgentID{}
	checkAgentID(t, ctx, scAgentID, agentID)

	// eth
	addressEth := "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
	checkerEth := testwasmlib.ScFuncs.CheckEthAddressAndAgentID(ctx)
	checkerEth.Params.EthAddress().SetValue(addressEth)
	checkerEth.Func.Call()
	require.NoError(t, ctx.Err)

	// check int types and uint types
	checkerIntAndUint := testwasmlib.ScFuncs.CheckIntAndUint(ctx)
	checkerIntAndUint.Func.Call()
	require.NoError(t, ctx.Err)

	scBigInt := wasmtypes.NewScBigInt(123213)
	bigInt := big.NewInt(123213)
	checkBigInt(t, ctx, scBigInt, bigInt)

	checkerBool := testwasmlib.ScFuncs.CheckBool(ctx)
	checkerBool.Func.Call()
	require.NoError(t, ctx.Err)

	checkerBytes := testwasmlib.ScFuncs.CheckBytes(ctx)
	length := 100
	byteData := make([]byte, length)
	for i := 0; i < length; i++ {
		byteData[i] = byte(rand.Intn(256))
	}
	checkerBytes.Params.Bytes().SetValue(byteData)
	checkerBytes.Func.Call()
	require.NoError(t, ctx.Err)

	hashString := hashing.HashData([]byte("foobar")).String()
	hash, err := hashing.HashValueFromHex(hashString)
	require.NoError(t, err)
	scHash := wasmtypes.HashFromString(hashString)
	checkHash(t, ctx, scHash, hash)

	scHname := testwasmlib.HScName
	checkerHname := testwasmlib.ScFuncs.CheckHname(ctx)
	checkerHname.Params.ScHname().SetValue(scHname)
	checkerHname.Params.HnameBytes().SetValue(scHname.Bytes())
	checkerHname.Params.HnameString().SetValue(scHname.String())
	checkerHname.Func.Call()
	require.NoError(t, ctx.Err)

	checkerString := testwasmlib.ScFuncs.CheckString(ctx)
	stringData := "this is a go string example"
	checkerString.Params.String().SetValue(stringData)
	checkerString.Func.Call()
	require.NoError(t, ctx.Err)

	nativeTokenID, err := getTokenID(ctx)
	require.NoError(t, err)
	scTokenID := ctx.Cvt.ScTokenID(nativeTokenID)
	checkTokenID(t, ctx, scTokenID, nativeTokenID)

	nftID, err := getNftID(ctx)
	require.NoError(t, err)
	scNftID := ctx.Cvt.ScNftID(&nftID)
	checkNftID(t, ctx, scNftID, nftID)

	blockNum := uint32(3)
	ctxBlocklog := ctx.SoloContextForCore(t, coreblocklog.ScName, coreblocklog.OnDispatch)
	require.NoError(t, ctxBlocklog.Err)
	fblocklog := coreblocklog.ScFuncs.GetRequestIDsForBlock(ctxBlocklog)
	fblocklog.Params.BlockIndex().SetValue(blockNum)
	fblocklog.Func.Call()
	scReq := fblocklog.Results.RequestID().GetRequestID(0).Value()
	req := ctxBlocklog.Chain.GetRequestIDsForBlock(blockNum)[0]
	checkRequestID(t, ctx, scReq, req)
}

func getTokenID(ctx *wasmsolo.SoloContext) (nativeTokenID iotago.NativeTokenID, err error) {
	maxSupply := 100
	fp := ctx.Chain.NewFoundryParams(ctx.Cvt.ToBigInt(maxSupply))
	_, nativeTokenID, err = fp.CreateFoundry()
	if err != nil {
		return iotago.NativeTokenID{}, err
	}
	return nativeTokenID, nil
}

func getNftID(ctx *wasmsolo.SoloContext) (iotago.NFTID, error) {
	agent := ctx.NewSoloAgent()
	addr, ok := isc.AddressFromAgentID(agent.AgentID())
	if !ok {
		return iotago.NFTID{}, fmt.Errorf("can't get address from AgentID")
	}
	_, nftInfo, err := ctx.Chain.Env.MintNFTL1(agent.Pair, addr, []byte("test data"))
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
func checkAgentID(t *testing.T, ctx *wasmsolo.SoloContext, scAgentID wasmtypes.ScAgentID, agentID isc.AgentID) {
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
	addressBytes := isc.BytesFromAddress(address)
	addressString := address.Bech32(parameters.L1().Protocol.Bech32HRP)

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
func checkTokenID(t *testing.T, ctx *wasmsolo.SoloContext, scTokenID wasmtypes.ScTokenID, nativeTokenID iotago.NativeTokenID) {
	nativeTokenIDBytes := nativeTokenID[:]
	nativeTokenIDString := nativeTokenID.String()

	require.Equal(t, scTokenID, wasmtypes.TokenIDFromString(wasmtypes.TokenIDToString(scTokenID)))
	require.Equal(t, scTokenID, wasmtypes.TokenIDFromBytes(wasmtypes.TokenIDToBytes(scTokenID)))
	require.Equal(t, scTokenID.String(), nativeTokenID.String())
	require.Equal(t, scTokenID.Bytes(), nativeTokenID[:])

	checker := testwasmlib.ScFuncs.CheckTokenID(ctx)
	checker.Params.ScTokenID().SetValue(scTokenID)
	checker.Params.TokenIDBytes().SetValue(nativeTokenIDBytes)
	checker.Params.TokenIDString().SetValue(nativeTokenIDString)
	checker.Func.Call()
	require.NoError(t, ctx.Err)
}

func checkRequestID(t *testing.T, ctx *wasmsolo.SoloContext, scRequestID wasmtypes.ScRequestID, requestID isc.RequestID) {
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
