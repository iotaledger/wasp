package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func printLogRecords(t *testing.T, recs []string, title string) {
	t.Logf("------- Log records for '%s'", title)
	for _, text := range recs {
		t.Logf(text)
	}
}

func TestEventLogBasicEmpty(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	recs, err := chain.GetEventsForContract(root.Contract.Name)
	require.NoError(t, err)
	require.Len(t, recs, 0)

	events, _ := chain.GetEventsForContract(root.Contract.Name)
	require.EqualValues(t, 0, len(events))

	events, _ = chain.GetEventsForContract(accounts.Contract.Name)
	require.EqualValues(t, 0, len(events))

	events, _ = chain.GetEventsForContract(blob.Contract.Name)
	require.EqualValues(t, 0, len(events))

	events, _ = chain.GetEventsForContract(blocklog.Contract.Name)
	require.EqualValues(t, 0, len(events))

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

	events, _ := chain.GetEventsForContract(root.Contract.Name)
	require.EqualValues(t, 0, len(events))

	events, _ = chain.GetEventsForContract(accounts.Contract.Name)
	require.EqualValues(t, 0, len(events))

	events, _ = chain.GetEventsForContract(blocklog.Contract.Name)
	require.EqualValues(t, 0, len(events))

	events, _ = chain.GetEventsForContract(blob.Contract.Name)
	require.EqualValues(t, 1, len(events))

	recs, err := chain.GetEventsForContract(blob.Contract.Name)
	require.NoError(t, err)
	require.Len(t, recs, 1)
	printLogRecords(t, recs, "blob")

	name := "testInccounter"
	err = chain.DeployContract(nil, name, hwasm)
	require.NoError(t, err)

	events, _ = chain.GetEventsForContract(root.Contract.Name)
	require.EqualValues(t, 1, len(events))

	events, _ = chain.GetEventsForContract(accounts.Contract.Name)
	require.EqualValues(t, 0, len(events))

	events, _ = chain.GetEventsForContract(blocklog.Contract.Name)
	require.EqualValues(t, 0, len(events))

	events, _ = chain.GetEventsForContract(blob.Contract.Name)
	require.EqualValues(t, 1, len(events))

	events, _ = chain.GetEventsForContract(name)
	require.EqualValues(t, 0, len(events))

	recs, err = chain.GetEventsForContract(name)
	require.NoError(t, err)
	require.Len(t, recs, 0)

	reqRecs := chain.GetRequestReceiptsForBlockRangeAsStrings(0, 0)

	for _, s := range reqRecs {
		t.Logf("%s", s)
	}
	require.EqualValues(t, 3, len(reqRecs))
	env.WaitPublisher()
}
