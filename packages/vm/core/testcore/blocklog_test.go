// excluded temporarily because of compilation errors

package testcore

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
)

func TestBlocklog_BlockInfoLatest(t *testing.T) {
	corecontracts.PrintWellKnownHnames()
	env := solo.New(t)
	chain := env.NewChain()

	bi := chain.GetLatestBlockInfo()
	require.NotNil(t, bi)
	require.EqualValues(t, 1, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 0, bi.NumOffLedgerRequests)
	t.Logf("%s", bi.String())
}

func TestBlocklog_BlockInfo(t *testing.T) {
	corecontracts.PrintWellKnownHnames()
	env := solo.New(t)
	chain := env.NewChain()

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

func TestBlocklog_BlockInfoLatestWithRequest(t *testing.T) {
	env := solo.New(t)

	ch := env.NewChain()

	err := ch.DepositBaseTokensToL2(100_000, nil)
	require.NoError(t, err)

	bi := ch.GetLatestBlockInfo()
	t.Logf("after ch deployment:\n%s", bi.String())

	bi = ch.GetLatestBlockInfo()
	require.NotNil(t, bi)
	require.EqualValues(t, 2, bi.BlockIndex)
	require.EqualValues(t, 1, bi.TotalRequests)
	require.EqualValues(t, 1, bi.NumSuccessfulRequests)
	require.EqualValues(t, 0, bi.NumOffLedgerRequests)
	t.Logf("%s", bi.String())
}

func TestBlocklog_BlockInfoSeveral(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	const numReqs = 6
	for range numReqs {
		err := ch.DepositBaseTokensToL2(100_000, nil)
		require.NoError(t, err)
	}

	bi := ch.GetLatestBlockInfo()
	require.EqualValues(t, 1+numReqs, int(bi.BlockIndex))

	for blockIndex := uint32(0); blockIndex <= bi.BlockIndex; blockIndex++ {
		bi1, err := ch.GetBlockInfo(blockIndex)
		require.NoError(t, err)
		require.NotNil(t, bi1)
		t.Logf("%s", bi1.String())
		require.EqualValues(t, blockIndex, bi1.BlockIndex)
		require.EqualValues(t, 1, bi1.TotalRequests)
		require.EqualValues(t, 1, bi1.NumSuccessfulRequests)
		require.LessOrEqual(t, bi1.NumOffLedgerRequests, bi1.TotalRequests)
	}
}

func TestBlocklog_RequestIsProcessed(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	ch.MustDepositBaseTokensToL2(10_000, nil)

	publicURL := "foo"
	req, _, _, _, err := ch.PostRequestSyncTx(
		solo.NewCallParams(governance.FuncSetMetadata.Message(&publicURL, nil)).
			WithGasBudget(100_000),
		nil,
	)
	require.NoError(t, err)

	bi := ch.GetLatestBlockInfo()
	require.NoError(t, err)
	require.True(t, ch.IsRequestProcessed(req.ID()))
	t.Logf("%s", bi.String())
}

func TestBlocklog_RequestReceipt(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	ch.MustDepositBaseTokensToL2(10_000, nil)

	publicURL := "foo"
	req, _, _, _, err := ch.PostRequestSyncTx(
		solo.NewCallParams(governance.FuncSetMetadata.Message(&publicURL, nil)).
			WithGasBudget(100_000),
		nil,
	)
	require.NoError(t, err)
	require.True(t, ch.IsRequestProcessed(req.ID()))

	receipt, _ := ch.GetRequestReceipt(req.ID())
	a := req.Bytes()
	b := receipt.Request.Bytes()
	require.Equal(t, a, b)
	require.Nil(t, receipt.Error)
	require.EqualValues(t, 3, int(receipt.BlockIndex))
	require.EqualValues(t, 0, receipt.RequestIndex)
	t.Logf("%s", receipt.String())
}

func TestBlocklog_RequestReceiptsForBlocks(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	ch.MustDepositBaseTokensToL2(10_000, nil)

	publicURL := "foo"
	req, _, _, _, err := ch.PostRequestSyncTx(
		solo.NewCallParams(governance.FuncSetMetadata.Message(&publicURL, nil)).
			WithGasBudget(100_000),
		nil,
	)
	require.NoError(t, err)
	require.True(t, ch.IsRequestProcessed(req.ID()))

	require.EqualValues(t, 3, ch.GetLatestAnchor().GetStateIndex())
	recs := ch.GetRequestReceiptsForBlock(ch.GetLatestAnchor().GetStateIndex())
	require.EqualValues(t, 1, len(recs))
	require.EqualValues(t, req.ID(), recs[0].Request.ID())
	t.Logf("%s\n", recs[0].String())
}

func TestBlocklog_RequestIDsForBlocks(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()

	ch.MustDepositBaseTokensToL2(10_000, nil)

	publicURL := "foo"
	req, _, _, _, err := ch.PostRequestSyncTx(
		solo.NewCallParams(governance.FuncSetMetadata.Message(&publicURL, nil)).
			WithGasBudget(100_000),
		nil,
	)
	require.NoError(t, err)
	require.True(t, ch.IsRequestProcessed(req.ID()))

	ids := ch.GetRequestIDsForBlock(ch.GetLatestAnchor().GetStateIndex())
	require.EqualValues(t, 1, len(ids))
	require.EqualValues(t, req.ID(), ids[0])
}

func TestBlocklog_ViewGetRequestReceipt(t *testing.T) {
	env := solo.New(t)
	ch := env.NewChain()
	// try to get a receipt for a request that does not exist
	_, ok := ch.GetRequestReceipt(isc.RequestID{})
	require.False(t, ok)
}

func TestBlocklog_Pruning(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{Debug: true})

	ch, _ := env.NewChainExt(nil, 0, "chain1", 0, 10, nil, nil)
	for i := 1; i <= 20; i++ {
		ch.DepositBaseTokensToL2(1000, nil)
	}
	// at this point blocks 0..10 have been pruned, and blocks 11..20 are available
	require.EqualValues(t, 20, ch.LatestBlock().StateIndex())
	require.EqualValues(t, 20, ch.EVM().BlockNumber().Uint64())
	for i := uint32(0); i <= 10; i++ {
		_, err := ch.GetBlockInfo(i)
		require.ErrorContains(t, err, "not found")
		// evm has the jsonrpcindex
		_, err = ch.EVM().BlockByNumber(big.NewInt(int64(i)))
		require.ErrorContains(t, err, "not found")
	}
	for i := uint32(11); i <= 20; i++ {
		bi, err := ch.GetBlockInfo(i)
		require.NoError(t, err)
		require.EqualValues(t, i, bi.BlockIndex)
		evmBlock, err := ch.EVM().BlockByNumber(big.NewInt(int64(i)))
		require.NoError(t, err)
		require.EqualValues(t, i, evmBlock.Number().Uint64())
	}
}
