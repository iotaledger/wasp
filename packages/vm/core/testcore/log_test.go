package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func printLogRecords(t *testing.T, recs []collections.TimestampedLogRecord, title string) {
	t.Logf("------- Log records for '%s'", title)
	for i := range recs {
		t.Logf("%d: '%s'", recs[i].Timestamp, string(recs[i].Data))
	}
}

func TestEventLogBasicEmpty(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	recs, err := chain.GetEventLogRecords(root.Contract.Name)
	require.NoError(t, err)
	require.Len(t, recs, 0)

	num := chain.GetEventLogNumRecords(root.Contract.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(accounts.Contract.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(blob.Contract.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(eventlog.Contract.Name)
	require.EqualValues(t, 0, num)

	reqRecs := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)
	require.EqualValues(t, 1, len(reqRecs))

	for _, s := range reqRecs {
		t.Logf("%s", s)
	}
}

func TestChainLogDeploy(t *testing.T) {
	env := solo.New(t, false, false)
	env.EnablePublisher(true)
	chain := env.NewChain(nil, "chain1")
	hwasm, err := chain.UploadWasmFromFile(nil, wasmFile)
	require.NoError(t, err)

	num := chain.GetEventLogNumRecords(root.Contract.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(accounts.Contract.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(eventlog.Contract.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(blob.Contract.Name)
	require.EqualValues(t, 1, num)

	recs, err := chain.GetEventLogRecords(blob.Contract.Name)
	require.NoError(t, err)
	require.Len(t, recs, 1)
	printLogRecords(t, recs, "blob")

	name := "testInccounter"
	err = chain.DeployContract(nil, name, hwasm)
	require.NoError(t, err)

	num = chain.GetEventLogNumRecords(root.Contract.Name)
	require.EqualValues(t, 1, num)

	num = chain.GetEventLogNumRecords(accounts.Contract.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(eventlog.Contract.Name)
	require.EqualValues(t, 0, num)

	num = chain.GetEventLogNumRecords(blob.Contract.Name)
	require.EqualValues(t, 1, num)

	num = chain.GetEventLogNumRecords(name)
	require.EqualValues(t, 0, num)

	recs, err = chain.GetEventLogRecords(name)
	require.NoError(t, err)
	require.Len(t, recs, 0)

	reqRecs := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)

	for _, s := range reqRecs {
		t.Logf("%s", s)
	}
	require.EqualValues(t, 3, len(reqRecs))
	env.WaitPublisher()
}
