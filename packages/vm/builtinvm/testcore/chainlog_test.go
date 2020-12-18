package testcore

import (
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/solo"
	"github.com/stretchr/testify/require"
	"testing"
)

func printRequestLogRecords(t *testing.T, recs []datatypes.TimestampedLogRecord, title string) {
	t.Logf("Request log records for '%s'", title)
	for i := range recs {
		r, err := chainlog.DecodeRequestChainLogRecord(recs[i].Data)
		require.NoError(t, err)
		t.Logf("%d: %s/%s -- %s", recs[i].Timestamp, r.RequestID.String(), r.EntryPoint.String(), r.Error)
	}
}

func printGenericLogRecords(t *testing.T, recs []datatypes.TimestampedLogRecord, title string) {
	t.Logf("Generic log records for '%s'", title)
	for i := range recs {
		t.Logf("%d: '%s'", recs[i].Timestamp, string(recs[i].Data))
	}
}

func TestChainLogBasic1(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")

	recs, err := chain.GetChainLogRecords(root.Interface.Name, chainlog.TRRequest)
	require.NoError(t, err)
	require.Len(t, recs, 1) // 1 root::init request

	num := chain.GetChainLogNumRecords(root.Interface.Name)
	require.EqualValues(t, 1, num)

	num = chain.GetChainLogNumRecords(accountsc.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetChainLogNumRecords(blob.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetChainLogNumRecords(chainlog.Interface.Name)
	require.EqualValues(t, 0, num)
}

//func TestChainLogDeploy(t *testing.T) {
//	glb := solo.New(t, false, false)
//	chain := glb.NewChain(nil, "chain1")
//	hwasm, err := chain.UploadWasmFromFile(nil, wasmFile)
//	require.NoError(t, err)
//
//	num := chain.GetChainLogNumRecords(root.Interface.Name)
//	require.EqualValues(t, 1, num)
//
//	num = chain.GetChainLogNumRecords(accountsc.Interface.Name)
//	require.EqualValues(t, 0, num)
//
//	num = chain.GetChainLogNumRecords(chainlog.Interface.Name)
//	require.EqualValues(t, 0, num)
//
//	num = chain.GetChainLogNumRecords(blob.Interface.Name)
//	require.EqualValues(t, 2, num)
//
//	recs, err := chain.GetChainLogRecords(blob.Interface.Name, chainlog.TRRequest)
//	require.NoError(t, err)
//	require.Len(t, recs, 1)   // 1 root::init request
//	printRequestLogRecords(t, recs, "blob")
//	recs, err = chain.GetChainLogRecords(blob.Interface.Name, chainlog.TRGenericData)
//	require.NoError(t, err)
//	require.Len(t, recs, 1)   // 1 root::init request
//	printGenericLogRecords(t, recs, "blob")
//
//	name := "testInccounter"
//	err = chain.DeployContract(nil, name, hwasm)
//	require.NoError(t, err)
//
//	num = chain.GetChainLogNumRecords(root.Interface.Name)
//	require.EqualValues(t, 2, num)
//
//	num = chain.GetChainLogNumRecords(accountsc.Interface.Name)
//	require.EqualValues(t, 0, num)
//
//	num = chain.GetChainLogNumRecords(chainlog.Interface.Name)
//	require.EqualValues(t, 0, num)
//
//	num = chain.GetChainLogNumRecords(blob.Interface.Name)
//	require.EqualValues(t, 2, num)
//
//	num = chain.GetChainLogNumRecords(name)
//	require.EqualValues(t, 0, num)
//
//	recs, err = chain.GetChainLogRecords(name, chainlog.TRRequest)
//	require.NoError(t, err)
//	require.Len(t, recs, 1)   // 1 root::init request
//	printRequestLogRecords(t, recs, "blob")
//}
