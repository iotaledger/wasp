package testcore

import (
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/stretchr/testify/require"
	"testing"
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

	num = chain.GetChainLogNumRecords(accountsc.Interface.Name)
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

	num = chain.GetChainLogNumRecords(accountsc.Interface.Name)
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

	num = chain.GetChainLogNumRecords(accountsc.Interface.Name)
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
