package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func TestBlockInfoLatest(t *testing.T) {
	core.PrintWellKnownHnames()
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	bi := chain.GetLatestBlockInfo()
	require.NotNil(t, bi)
	require.EqualValues(t, 1, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 0, bi.NumOffLedgerRequests)
}

func TestBlockInfo(t *testing.T) {
	core.PrintWellKnownHnames()
	env := solo.New(t, false, false)
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

func TestBlockInfoLatestSeveral(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	hwasm, err := chain.UploadWasmFromFile(nil, wasmFile)
	require.NoError(t, err)

	err = chain.DeployContract(nil, "testCore", hwasm)
	require.NoError(t, err)

	bi := chain.GetLatestBlockInfo()
	require.NotNil(t, bi)
	require.EqualValues(t, 3, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 0, bi.NumOffLedgerRequests)
}

func TestBlockInfoSeveral(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	hwasm, err := chain.UploadWasmFromFile(nil, wasmFile)
	require.NoError(t, err)

	err = chain.DeployContract(nil, "testCore", hwasm)
	require.NoError(t, err)

	bi := chain.GetLatestBlockInfo()

	for blockIndex := uint32(0); blockIndex <= bi.BlockIndex; blockIndex++ {
		bi1, err := chain.GetBlockInfo(blockIndex)
		require.NoError(t, err)
		require.NotNil(t, bi1)
		require.EqualValues(t, blockIndex, bi1.BlockIndex)
		require.EqualValues(t, 1, bi1.TotalRequests)
		require.EqualValues(t, 1, bi1.NumSuccessfulRequests)
		require.EqualValues(t, 0, bi1.NumOffLedgerRequests)
		t.Logf("%s", bi1.String())
	}
}

func TestRequestIsProcessed(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Contract.Name, root.FuncSetDefaultFee.Name,
		root.ParamOwnerFee, 0, root.ParamValidatorFee, 0).WithIotas(1)
	tx, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	reqs, err := env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(reqs))

	require.True(t, chain.IsRequestProcessed(reqs[0].ID()))
}

func TestRequestLogRecord(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Contract.Name, root.FuncSetDefaultFee.Name,
		root.ParamOwnerFee, 0, root.ParamValidatorFee, 0).WithIotas(1)
	tx, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	reqs, err := env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(reqs))

	require.True(t, chain.IsRequestProcessed(reqs[0].ID()))

	rec, blockIndex, requestIndex, ok := chain.GetRequestLogRecord(reqs[0].ID())
	require.True(t, ok)
	require.EqualValues(t, reqs[0].ID(), rec.RequestID)
	require.False(t, rec.OffLedger)
	require.EqualValues(t, 0, len(rec.LogData))
	require.EqualValues(t, 2, blockIndex)
	require.EqualValues(t, 0, requestIndex)
}

func TestRequestLogRecordsForBlocks(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Contract.Name, root.FuncSetDefaultFee.Name,
		root.ParamOwnerFee, 0, root.ParamValidatorFee, 0).WithIotas(1)
	tx, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	reqs, err := env.RequestsForChain(tx, chain.ChainID)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(reqs))

	require.True(t, chain.IsRequestProcessed(reqs[0].ID()))

	recs := chain.GetRequestLogRecordsForBlock(2)
	require.EqualValues(t, 1, len(recs))
	require.EqualValues(t, reqs[0].ID(), recs[0].RequestID)
}

func TestRequestIDsForBlocks(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	req := solo.NewCallParams(root.Contract.Name, root.FuncSetDefaultFee.Name,
		root.ParamOwnerFee, 0, root.ParamValidatorFee, 0).WithIotas(1)
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
