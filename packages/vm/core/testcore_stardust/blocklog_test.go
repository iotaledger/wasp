package testcore

import (
	"fmt"
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/stretchr/testify/require"
)

func TestBlockInfoLatest(t *testing.T) {
	core.PrintWellKnownHnames()
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")

	bi := chain.GetLatestBlockInfo()
	require.NotNil(t, bi)
	require.EqualValues(t, 1, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 0, bi.NumOffLedgerRequests)
	t.Logf("%s", bi.String())
}

func TestBlockInfo(t *testing.T) {
	core.PrintWellKnownHnames()
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")

	bi, err := chain.GetBlockInfo(0)
	require.NoError(t, err)
	require.NotNil(t, bi)
	require.EqualValues(t, 0, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 0, bi.NumOffLedgerRequests)
	t.Logf("%s", bi.String())

	bi, err = chain.GetBlockInfo(1)
	require.NoError(t, err)
	require.NotNil(t, bi)
	require.EqualValues(t, 1, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 0, bi.NumOffLedgerRequests)
	t.Logf("%s", bi.String())
}

func TestBlockInfoLatestWithRequest(t *testing.T) {
	env := solo.New(t)

	chain := env.NewChain(nil, "chain1")
	bi := chain.GetLatestBlockInfo()
	t.Logf("after chain deployment:\n%s", bi.String())
	// uploading one blob
	_, err := chain.UploadBlob(nil, "field", "dummy blob data")
	require.NoError(t, err)

	bi = chain.GetLatestBlockInfo()
	require.NotNil(t, bi)
	require.EqualValues(t, 3, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 1, bi.NumOffLedgerRequests)
	t.Logf("%s", bi.String())
}

func TestBlockInfoSeveral(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")

	const numReqs = 5
	for i := 0; i < numReqs; i++ {
		_, err := chain.UploadBlob(nil, "field", fmt.Sprintf("dummy blob data #%d", i))
		require.NoError(t, err)
	}

	bi := chain.GetLatestBlockInfo()
	require.EqualValues(t, 1+2*numReqs, int(bi.BlockIndex))

	for blockIndex := uint32(0); blockIndex <= bi.BlockIndex; blockIndex++ {
		bi1, err := chain.GetBlockInfo(blockIndex)
		require.NoError(t, err)
		require.NotNil(t, bi1)
		t.Logf("%s", bi1.String())
		require.EqualValues(t, blockIndex, bi1.BlockIndex)
		require.EqualValues(t, 1, bi1.TotalRequests)
		require.EqualValues(t, 1, bi1.NumSuccessfulRequests)
		require.LessOrEqual(t, bi1.NumOffLedgerRequests, bi1.TotalRequests)
	}
}

func TestRequestIsProcessed(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name).
		WithGasBudget(30)
	tx, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	reqs, err := env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(reqs))

	bi := chain.GetLatestBlockInfo()
	require.NoError(t, err)
	require.True(t, chain.IsRequestProcessed(reqs[0].ID()))
	t.Logf("%s", bi.String())
}

func TestRequestReceipt(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name).
		WithGasBudget(30)
	tx, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	reqs, err := env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(reqs))
	require.True(t, chain.IsRequestProcessed(reqs[0].ID()))

	receipt, ok := chain.GetRequestReceipt(reqs[0].ID())
	require.True(t, ok)
	a := reqs[0].Bytes()
	b := receipt.RequestData.Bytes()
	require.Equal(t, a, b)
	require.NoError(t, receipt.Error())
	require.EqualValues(t, 2, receipt.BlockIndex)
	require.EqualValues(t, 0, receipt.RequestIndex)
	t.Logf("%s", receipt.String())
}

func TestRequestReceiptsForBlocks(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name).
		WithGasBudget(30)
	tx, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	reqs, err := env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(reqs))

	require.True(t, chain.IsRequestProcessed(reqs[0].ID()))

	recs := chain.GetRequestReceiptsForBlock(2)
	require.EqualValues(t, 1, len(recs))
	require.EqualValues(t, reqs[0].ID(), recs[0].RequestData.ID())
	t.Logf("%s\n", recs[0].String())
}

func TestRequestIDsForBlocks(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(governance.Contract.Name, governance.FuncSetChainInfo.Name).
		WithGasBudget(30)
	tx, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	reqs, err := env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(reqs))

	require.True(t, chain.IsRequestProcessed(reqs[0].ID()))

	ids := chain.GetRequestIDsForBlock(2)
	require.EqualValues(t, 1, len(ids))
	require.EqualValues(t, reqs[0].ID(), ids[0])
}
