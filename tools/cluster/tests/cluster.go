package tests

import (
	"os"
	"path"
	"testing"

	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

type waspClusterOpts struct {
	nNodes       int
	modifyConfig cluster.ModifyNodesConfigFn
}

// newCluster starts a new cluster environment for tests.
// It is a private function because cluster tests cannot be run in parallel,
// so all cluster tests MUST be in this same package.
func newCluster(t *testing.T, opt ...waspClusterOpts) *cluster.Cluster {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}

	config := cluster.DefaultConfig()
	config.L1 = ClustL1Config

	var modifyNodesConfig cluster.ModifyNodesConfigFn
	if len(opt) > 0 {
		config.Wasp.NumNodes = opt[0].nNodes
		modifyNodesConfig = opt[0].modifyConfig
	}

	clu := cluster.New(t.Name(), config, t)

	dataPath := path.Join(os.TempDir(), "wasp-cluster")
	err := clu.InitDataPath(".", dataPath, true, modifyNodesConfig)
	require.NoError(t, err)

	err = clu.Start(dataPath)
	require.NoError(t, err)

	t.Cleanup(clu.Stop)

	return clu
}
