package testcore

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accounts"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/testcore/test_sandbox"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func printLogRecords(t *testing.T, recs []datatypes.TimestampedLogRecord, title string) {
	t.Logf("------- Log records for '%s'", title)
	for i := range recs {
		t.Logf("%d: '%s'", recs[i].Timestamp, string(recs[i].Data))
	}
}

func TestChainLogBasic1(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	recs, err := chain.GetChainLogRecords(root.Interface.Name)
	require.NoError(t, err)
	require.Len(t, recs, 1) // 1 root::init request

	num := chain.GetChainLogNumRecords(root.Interface.Name)
	require.EqualValues(t, 1, num)

	num = chain.GetChainLogNumRecords(accounts.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetChainLogNumRecords(blob.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetChainLogNumRecords(chainlog.Interface.Name)
	require.EqualValues(t, 0, num)
}

func TestChainLogDeploy(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	hwasm, err := chain.UploadWasmFromFile(nil, wasmFile)
	require.NoError(t, err)

	num := chain.GetChainLogNumRecords(root.Interface.Name)
	require.EqualValues(t, 1, num)

	num = chain.GetChainLogNumRecords(accounts.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetChainLogNumRecords(chainlog.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetChainLogNumRecords(blob.Interface.Name)
	require.EqualValues(t, 2, num)

	recs, err := chain.GetChainLogRecords(blob.Interface.Name)
	require.NoError(t, err)
	require.Len(t, recs, 2) // 1 root::init request
	printLogRecords(t, recs, "blob")

	name := "testInccounter"
	err = chain.DeployContract(nil, name, hwasm)
	require.NoError(t, err)

	num = chain.GetChainLogNumRecords(root.Interface.Name)
	require.EqualValues(t, 3, num)

	num = chain.GetChainLogNumRecords(accounts.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetChainLogNumRecords(chainlog.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetChainLogNumRecords(blob.Interface.Name)
	require.EqualValues(t, 2, num)

	num = chain.GetChainLogNumRecords(name)
	require.EqualValues(t, 0, num)

	recs, err = chain.GetChainLogRecords(name)
	require.NoError(t, err)
	require.Len(t, recs, 0)
}

func TestChainlogGetLast3(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	for i := 1; i < 6; i++ {
		req := solo.NewCall(test_sandbox.Interface.Name,
			test_sandbox.FuncChainLogGenericData,
			test_sandbox.VarCounter, i,
		)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}

	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamMaxLastRecords, 3,
		chainlog.ParamContractHname, test_sandbox.Interface.Hname(),
	)
	require.NoError(t, err)

	array := datatypes.NewMustArray(res, chainlog.ParamRecords)
	require.EqualValues(t, 3, array.Len())
}

func TestChainlogGetBetweenTs(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	for i := 1; i < 6; i++ {
		req := solo.NewCall(test_sandbox.Interface.Name,
			test_sandbox.FuncChainLogGenericData,
			test_sandbox.VarCounter, i,
		)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}

	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamFromTs, 0,
		chainlog.ParamToTs, chain.State.Timestamp()-int64(1500*time.Millisecond),
		chainlog.ParamMaxLastRecords, 2,
		chainlog.ParamContractHname, test_sandbox.Interface.Hname(),
	)
	require.NoError(t, err)

	array := datatypes.NewMustArray(res, chainlog.ParamRecords)
	require.EqualValues(t, 2, array.Len())
}

func TestChainLogEventData(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(test_sandbox.Interface.Name,
		test_sandbox.FuncChainLogEventData,
	)
	_, err = chain.PostRequest(req, nil)

	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamFromTs, 0,
		chainlog.ParamToTs, chain.State.Timestamp(),
		chainlog.ParamContractHname, test_sandbox.Interface.Hname(),
	)
	require.NoError(t, err)
	array := datatypes.NewMustArray(res, chainlog.ParamRecords)

	require.EqualValues(t, 2, array.Len())

	str, err := chain.GetChainLogRecordsString(test_sandbox.Interface.Name)
	require.NoError(t, err)
	t.Log(str)
}

func TestChainlogDifferentCalls(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	glb.SetTimeStep(500 * time.Millisecond)

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	count := 1
	// events
	for i := 1; i <= 3; i++ {
		req := solo.NewCall(test_sandbox.Interface.Name,
			test_sandbox.FuncChainLogEventData,
			test_sandbox.VarCounter, count,
		)
		count++
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}
	// generic
	for i := 1; i <= 2; i++ {
		req := solo.NewCall(test_sandbox.Interface.Name,
			test_sandbox.FuncChainLogGenericData,
			test_sandbox.VarCounter, count,
		)
		count++
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}
	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamFromTs, 0,
		chainlog.ParamToTs, chain.State.Timestamp(),
		chainlog.ParamContractHname, test_sandbox.Interface.Hname(),
	)
	require.NoError(t, err)
	array := datatypes.NewMustArray(res, chainlog.ParamRecords)
	require.EqualValues(t, (2+3)*2, array.Len())

	str, err := chain.GetChainLogRecordsString(test_sandbox.Interface.Name)
	require.NoError(t, err)
	t.Log(str)

	require.EqualValues(t, 2+3, strings.Count(str, "[req]"))
	require.EqualValues(t, 2, strings.Count(str, "[GenericData]"))
	require.EqualValues(t, 3, strings.Count(str, "[Event]"))
}

func TestChainOwnerID(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	originator := chain.OriginatorAgentID.Bytes()

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(test_sandbox.Interface.Name,
		test_sandbox.FuncChainOwnerID,
	)
	ret, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	c := ret.MustGet(test_sandbox.VarChainOwner)

	require.EqualValues(t, originator, c)
}

func TestChainID(t *testing.T) {
	glb := solo.New(t, false, false)

	chain := glb.NewChain(nil, "chain1")

	chainID := chain.ChainID.Bytes()

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(test_sandbox.Interface.Name,
		test_sandbox.FuncChainID,
	)
	ret, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	c := ret.MustGet(test_sandbox.VarChainID)

	require.EqualValues(t, chainID, c)
}

func TestSandboxCall(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	//Description of root (solo tool)
	desc := "'solo' testing chain"

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(test_sandbox.Interface.Name,
		test_sandbox.FuncSandboxCall,
	)
	ret, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	d := ret.MustGet(test_sandbox.VarSandboxCall)

	require.EqualValues(t, desc, string(d))
}

func TestChainLogGetNumRecords(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(test_sandbox.Interface.Name,
		test_sandbox.FuncChainLogGenericData,
		test_sandbox.VarCounter, 1337,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetNumRecords,
		chainlog.ParamContractHname, test_sandbox.Interface.Hname(),
	)
	require.NoError(t, err)

	v, ok, err := codec.DecodeInt64(res.MustGet(chainlog.ParamNumRecords))

	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, 2, v)

	str, err := chain.GetChainLogRecordsString(test_sandbox.Interface.Name)
	require.NoError(t, err)
	t.Log(str)

	require.EqualValues(t, 1, strings.Count(str, "[req]"))
	require.EqualValues(t, 1, strings.Count(str, "[GenericData]"))
	require.EqualValues(t, 0, strings.Count(str, "[Event]"))
	require.EqualValues(t, 1, strings.Count(str, "1337"))
}

func TestChainLogSandboxDeploy(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(test_sandbox.Interface.Name,
		test_sandbox.FuncChainlogDeploy,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	//This call should return only one record which should be the type of TRDeploy
	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamContractHname, root.Interface.Hname(),
	)
	require.NoError(t, err)
	array := datatypes.NewMustArray(res, chainlog.ParamRecords)

	require.EqualValues(t, 4, array.Len())

	str, err := chain.GetChainLogRecordsString(root.Interface.Name)
	require.NoError(t, err)
	t.Log(str)

	require.EqualValues(t, 2, strings.Count(str, "[deploy]"))
	require.EqualValues(t, 2, strings.Count(str, "[req]"))
}

func TestChainLogMultiple(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	//Deploy of the test sandbox contract
	err := chain.DeployContract(nil, test_sandbox.Interface.Name, test_sandbox.Interface.ProgramHash)
	require.NoError(t, err)

	req := solo.NewCall(test_sandbox.Interface.Name,
		test_sandbox.FuncChainLogEventData,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	req = solo.NewCall(test_sandbox.Interface.Name,
		test_sandbox.FuncChainLogGenericData,
		test_sandbox.VarCounter, 333,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	/////Should return 4 logs records/////
	res, err := chain.CallView(chainlog.Interface.Name, chainlog.FuncGetLogRecords,
		chainlog.ParamContractHname, test_sandbox.Interface.Hname(),
	)
	require.NoError(t, err)
	array := datatypes.NewMustArray(res, chainlog.ParamRecords)
	require.EqualValues(t, 4, array.Len())
	//////////////////////////////////////

	strRoot, err := chain.GetChainLogRecordsString(root.Interface.Name)
	require.NoError(t, err)
	t.Log(strRoot)
	require.EqualValues(t, 2, strings.Count(strRoot, "[req]"))
	require.EqualValues(t, 1, strings.Count(strRoot, "[deploy]"))

	strTest, err := chain.GetChainLogRecordsString(test_sandbox.Interface.Name)
	require.NoError(t, err)
	t.Log(strTest)
	require.EqualValues(t, 2, strings.Count(strTest, "[req]"))
	require.EqualValues(t, 1, strings.Count(strTest, "[GenericData]"))
	require.EqualValues(t, 1, strings.Count(strTest, "[Event]"))
	require.EqualValues(t, 1, strings.Count(strTest, "333"))
}
