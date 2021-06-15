package testutil

import (
	"flag"
	"os"
	"path"
	"testing"

	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

var numNodes = flag.Int("num-nodes", 4, "amount of wasp nodes")

func NewCluster(t *testing.T, nNodes ...int) *cluster.Cluster {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}

	if len(nNodes) > 0 {
		*numNodes = nNodes[0]
	}
	config := cluster.DefaultConfig()
	config.Wasp.NumNodes = *numNodes
	clu := cluster.New(t.Name(), config)

	dataPath := path.Join(os.TempDir(), "wasp-cluster")
	err := clu.InitDataPath(".", dataPath, true)
	require.NoError(t, err)

	err = clu.Start(dataPath)
	require.NoError(t, err)

	t.Cleanup(clu.Stop)

	return clu
}
