package testcore

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/stretchr/testify/require"
	"testing"
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
