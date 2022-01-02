package test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlibclient"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmclient"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/vm/wasmsolo"
	"github.com/stretchr/testify/require"
)

var (
	allParams = []string{
		testwasmlib.ParamAddress,
		testwasmlib.ParamAgentID,
		testwasmlib.ParamBool,
		testwasmlib.ParamChainID,
		testwasmlib.ParamColor,
		testwasmlib.ParamHash,
		testwasmlib.ParamHname,
		testwasmlib.ParamInt8,
		testwasmlib.ParamInt16,
		testwasmlib.ParamInt32,
		testwasmlib.ParamInt64,
		testwasmlib.ParamRequestID,
		testwasmlib.ParamUint8,
		testwasmlib.ParamUint16,
		testwasmlib.ParamUint32,
		testwasmlib.ParamUint64,
	}
	allLengths    = []int{33, 37, 1, 33, 32, 32, 4, 1, 2, 4, 8, 34, 1, 2, 4, 8}
	invalidValues = map[wasmlib.Key][][]byte{
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
	return wasmsolo.NewSoloContext(t, testwasmlib.ScName, testwasmlib.OnLoad)
}

func TestDeploy(t *testing.T) {
	ctx := setupTest(t)
	require.NoError(t, ctx.ContractExists(testwasmlib.ScName))
}

func TestNoParams(t *testing.T) {
	ctx := setupTest(t)

	f := testwasmlib.ScFuncs.ParamTypes(ctx)
	f.Func.TransferIotas(1).Post()
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
	pt.Params.Color().SetValue(wasmlib.NewScColorFromBytes([]byte("RedGreenBlueYellowCyanBlackWhite")))
	pt.Params.Hash().SetValue(wasmlib.NewScHashFromBytes([]byte("0123456789abcdeffedcba9876543210")))
	pt.Params.Hname().SetValue(testwasmlib.HScName)
	pt.Params.Int8().SetValue(-123)
	pt.Params.Int16().SetValue(-12345)
	pt.Params.Int32().SetValue(-1234567890)
	pt.Params.Int64().SetValue(-1234567890123456789)
	pt.Params.RequestID().SetValue(wasmlib.NewScRequestIDFromBytes([]byte("abcdefghijklmnopqrstuvwxyz123456\x00\x00")))
	pt.Params.String().SetValue("this is a string")
	pt.Params.Uint8().SetValue(123)
	pt.Params.Uint16().SetValue(12345)
	pt.Params.Uint32().SetValue(1234567890)
	pt.Params.Uint64().SetValue(1234567890123456789)
	pt.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)
	return ctx
}

func TestValidSizeParams(t *testing.T) {
	ctx := setupTest(t)
	for index, param := range allParams {
		t.Run("ValidSize "+param, func(t *testing.T) {
			pt := testwasmlib.ScFuncs.ParamTypes(ctx)
			bytes := make([]byte, allLengths[index])
			if param == testwasmlib.ParamChainID {
				bytes[0] = byte(ledgerstate.AliasAddressType)
			}
			pt.Params.Param().GetBytes(param).SetValue(bytes)
			pt.Func.TransferIotas(1).Post()
			require.Error(t, ctx.Err)
			require.Contains(t, ctx.Err.Error(), "mismatch: ")
		})
	}
}

func TestInvalidSizeParams(t *testing.T) {
	ctx := setupTest(t)
	for index, param := range allParams {
		t.Run("InvalidSize "+param, func(t *testing.T) {
			pt := testwasmlib.ScFuncs.ParamTypes(ctx)
			pt.Params.Param().GetBytes(param).SetValue(make([]byte, 0))
			pt.Func.TransferIotas(1).Post()
			require.Error(t, ctx.Err)
			require.True(t, strings.HasSuffix(ctx.Err.Error(), "invalid type size"))

			pt = testwasmlib.ScFuncs.ParamTypes(ctx)
			pt.Params.Param().GetBytes(param).SetValue(make([]byte, allLengths[index]-1))
			pt.Func.TransferIotas(1).Post()
			require.Error(t, ctx.Err)
			require.True(t, strings.HasSuffix(ctx.Err.Error(), "invalid type size"))

			pt = testwasmlib.ScFuncs.ParamTypes(ctx)
			pt.Params.Param().GetBytes(param).SetValue(make([]byte, allLengths[index]+1))
			pt.Func.TransferIotas(1).Post()
			require.Error(t, ctx.Err)
			require.Contains(t, ctx.Err.Error(), "invalid type size")
		})
	}
}

func TestInvalidTypeParams(t *testing.T) {
	ctx := setupTest(t)
	for param, values := range invalidValues {
		for index, value := range values {
			t.Run("InvalidType "+string(param)+" "+strconv.Itoa(index), func(t *testing.T) {
				req := solo.NewCallParams(testwasmlib.ScName, testwasmlib.FuncParamTypes,
					string(param), value,
				).WithIotas(1)
				_, err := ctx.Chain.PostRequestSync(req, nil)
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid ")
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
	require.EqualValues(t, 339, len(rec.Results.Record().Value()))
}

func TestClearArray(t *testing.T) {
	ctx := setupTest(t)

	as := testwasmlib.ScFuncs.ArraySet(ctx)
	as.Params.Name().SetValue("bands")
	as.Params.Index().SetValue(0)
	as.Params.Value().SetValue("Simple Minds")
	as.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.ArraySet(ctx)
	as.Params.Name().SetValue("bands")
	as.Params.Index().SetValue(1)
	as.Params.Value().SetValue("Dire Straits")
	as.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	as = testwasmlib.ScFuncs.ArraySet(ctx)
	as.Params.Name().SetValue("bands")
	as.Params.Index().SetValue(2)
	as.Params.Value().SetValue("ELO")
	as.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	al := testwasmlib.ScFuncs.ArrayLength(ctx)
	al.Params.Name().SetValue("bands")
	al.Func.Call()
	require.NoError(t, ctx.Err)
	length := al.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 3, length.Value())

	av := testwasmlib.ScFuncs.ArrayValue(ctx)
	av.Params.Name().SetValue("bands")
	av.Params.Index().SetValue(1)
	av.Func.Call()
	require.NoError(t, ctx.Err)
	value := av.Results.Value()
	require.True(t, value.Exists())
	require.EqualValues(t, "Dire Straits", value.Value())

	ac := testwasmlib.ScFuncs.ArrayClear(ctx)
	ac.Params.Name().SetValue("bands")
	ac.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	al = testwasmlib.ScFuncs.ArrayLength(ctx)
	al.Params.Name().SetValue("bands")
	al.Func.Call()
	require.NoError(t, ctx.Err)
	length = al.Results.Length()
	require.True(t, length.Exists())
	require.EqualValues(t, 0, length.Value())

	av = testwasmlib.ScFuncs.ArrayValue(ctx)
	av.Params.Name().SetValue("bands")
	av.Params.Index().SetValue(0)
	av.Func.Call()
	require.Error(t, ctx.Err)
}

func TestViewBalance(t *testing.T) {
	ctx := setupTest(t)

	v := testwasmlib.ScFuncs.IotaBalance(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	require.True(t, v.Results.Iotas().Exists())
	require.EqualValues(t, 0, v.Results.Iotas().Value())
}

func TestViewBalanceWithTokens(t *testing.T) {
	ctx := setupTest(t)

	// FuncParamTypes without params does nothing
	f := testwasmlib.ScFuncs.ParamTypes(ctx)
	f.Func.TransferIotas(42).Post()
	require.NoError(t, ctx.Err)

	v := testwasmlib.ScFuncs.IotaBalance(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	require.True(t, v.Results.Iotas().Exists())
	require.EqualValues(t, 42, v.Results.Iotas().Value())
}

func TestRandom(t *testing.T) {
	ctx := setupTest(t)

	f := testwasmlib.ScFuncs.Random(ctx)
	f.Func.TransferIotas(1).Post()
	require.NoError(t, ctx.Err)

	v := testwasmlib.ScFuncs.GetRandom(ctx)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	random := v.Results.Random().Value()
	require.True(t, random >= 0 && random < 1000)
	fmt.Printf("Random value: %d\n", random)
}

func TestMultiRandom(t *testing.T) {
	ctx := setupTest(t)

	numbers := make([]int64, 0)
	for i := 0; i < 10; i++ {
		f := testwasmlib.ScFuncs.Random(ctx)
		f.Func.TransferIotas(1).Post()
		require.NoError(t, ctx.Err)

		v := testwasmlib.ScFuncs.GetRandom(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		random := v.Results.Random().Value()
		require.True(t, random >= 0 && random < 1000)
		numbers = append(numbers, random)
	}

	for _, number := range numbers {
		fmt.Printf("Random value: %d\n", number)
	}
}

// hardcoded seed and chain ID, taken from wasp-cli.json
// note that normally the chain has already been set up and
// the contract has already been deployed in some way, so
// these values are usually available from elsewhere
const (
	mySeed    = "6C6tRksZDWeDTCzX4Q7R2hbpyFV86cSGLVxdkFKSB3sv"
	myChainID = "jn52vSuUUYY22T1mV2ny14EADYBu3ofyewLRSsVRnjpz"
)

func setupClient(t *testing.T) *testwasmlibclient.TestWasmLibService {
	require.True(t, wasmclient.SeedIsValid(mySeed))
	require.True(t, wasmclient.ChainIsValid(myChainID))

	// we're testing against wasp-cluster, so defaults will do
	svcClient := wasmclient.DefaultServiceClient()

	// create the service for the testwasmlib smart contract
	svc, err := testwasmlibclient.NewTestWasmLibService(svcClient, myChainID)
	require.NoError(t, err)

	// we'll use the first address in the seed to sign requests
	svc.SignRequests(wasmclient.SeedToKeyPair(mySeed, 0))
	return svc
}

func TestClientEvents(t *testing.T) {
	svc := setupClient(t)

	// get new triggerEvent interface, pass params, and post the request
	f := svc.TriggerEvent()
	f.Name("Lala")
	f.Address(wasmclient.SeedToAddress(mySeed, 0))
	req1 := f.Post()
	require.NoError(t, req1.Error())

	// err := svc.WaitRequest(req1)
	// require.NoError(t, err)

	// get new triggerEvent interface, pass params, and post the request
	f = svc.TriggerEvent()
	f.Name("Trala")
	f.Address(wasmclient.SeedToAddress(mySeed, 1))
	req2 := f.Post()
	require.NoError(t, req2.Error())

	err := svc.WaitRequest(req2)
	require.NoError(t, err)
}

func TestClientRandom(t *testing.T) {
	svc := setupClient(t)

	// generate new random value
	f := svc.Random()
	req := f.Post()
	require.NoError(t, req.Error())

	err := svc.WaitRequest(req)
	require.NoError(t, err)

	// get current random value
	v := svc.GetRandom()
	res := v.Call()
	require.NoError(t, v.Error())
	require.GreaterOrEqual(t, res.Random(), int64(0))
	fmt.Println("Random: ", res.Random())
}

func TestClientArray(t *testing.T) {
	svc := setupClient(t)

	v := svc.ArrayLength()
	v.Name("Bands")
	res := v.Call()
	require.NoError(t, v.Error())
	require.EqualValues(t, 0, res.Length())

	f := svc.ArraySet()
	f.Name("Bands")
	f.Index(0)
	f.Value("Dire Straits")
	req := f.Post()
	require.NoError(t, req.Error())
	err := svc.WaitRequest(req)
	require.NoError(t, err)

	v = svc.ArrayLength()
	v.Name("Bands")
	res = v.Call()
	require.NoError(t, v.Error())
	require.EqualValues(t, 1, res.Length())

	c := svc.ArrayClear()
	c.Name("Bands")
	req = c.Post()
	require.NoError(t, req.Error())
	err = svc.WaitRequest(req)
	require.NoError(t, err)

	v = svc.ArrayLength()
	v.Name("Bands")
	res = v.Call()
	require.NoError(t, v.Error())
	require.EqualValues(t, 0, res.Length())
}
