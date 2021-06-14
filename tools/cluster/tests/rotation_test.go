package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	clutest "github.com/iotaledger/wasp/tools/cluster/testutil"
)

func TestRotation(t *testing.T) {
	cmt1 := []int{0, 1, 2, 3}
	cmt2 := []int{2, 3, 4, 5}

	clu := clutest.NewCluster(t, 6)
	addr1, err := clu.RunDKG(cmt1, 3)
	require.NoError(t, err)
	addr2, err := clu.RunDKG(cmt2, 3)
	require.NoError(t, err)

	t.Logf("addr1: %s", addr1.Base58())
	t.Logf("addr2: %s", addr2.Base58())

	chain, err := clu.DeployChain("chain", cmt1, 3, addr1)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain.ChainID.Base58())
}
