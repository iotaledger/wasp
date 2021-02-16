package testcore

import (
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
	"testing"
)

func printLogRecords(t *testing.T, recs []collections.TimestampedLogRecord, title string) {
	t.Logf("------- Log records for '%s'", title)
	for i := range recs {
		t.Logf("%d: '%s'", recs[i].Timestamp, string(recs[i].Data))
	}
}

func TestChainLogBasic1(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	recs, err := chain.GetEventLogRecords(root.Interface.Name)
	require.NoError(t, err)
	require.Len(t, recs, 1) // 1 root::init request

	num := chain.GetEventLogNumRecords(root.Interface.Name)
	require.EqualValues(t, 1, num)

	num = chain.GetEventLogNumRecords(accounts.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(blob.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(eventlog.Interface.Name)
	require.EqualValues(t, 0, num)
}

func TestChainLogDeploy(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	hwasm, err := chain.UploadWasmFromFile(nil, wasmFile)
	require.NoError(t, err)

	num := chain.GetEventLogNumRecords(root.Interface.Name)
	require.EqualValues(t, 1, num)

	num = chain.GetEventLogNumRecords(accounts.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(eventlog.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(blob.Interface.Name)
	require.EqualValues(t, 2, num)

	recs, err := chain.GetEventLogRecords(blob.Interface.Name)
	require.NoError(t, err)
	require.Len(t, recs, 2) // 1 root::init request
	printLogRecords(t, recs, "blob")

	name := "testInccounter"
	err = chain.DeployContract(nil, name, hwasm)
	require.NoError(t, err)

	num = chain.GetEventLogNumRecords(root.Interface.Name)
	require.EqualValues(t, 3, num)

	num = chain.GetEventLogNumRecords(accounts.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(eventlog.Interface.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(blob.Interface.Name)
	require.EqualValues(t, 2, num)

	num = chain.GetEventLogNumRecords(name)
	require.EqualValues(t, 0, num)

	recs, err = chain.GetEventLogRecords(name)
	require.NoError(t, err)
	require.Len(t, recs, 0)
}
