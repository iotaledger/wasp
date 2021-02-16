package testutil

import (
	"os"
	"path"
	"testing"

	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

func NewCluster(t *testing.T) *cluster.Cluster {
	if testing.Short() {
		t.Skip("Skipping cluster test in short mode")
	}

	config := cluster.DefaultConfig()
	clu := cluster.New(t.Name(), config)

	dataPath := path.Join(os.TempDir(), "wasp-cluster")
	err := clu.InitDataPath(".", dataPath, true)
	require.NoError(t, err)

	err = clu.Start(dataPath)
	require.NoError(t, err)

	t.Cleanup(clu.Stop)

	return clu
}
