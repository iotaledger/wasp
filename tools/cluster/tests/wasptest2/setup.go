package wasptest2

import (
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/assert"
	"path"
	"runtime"
	"testing"
)

func check(err error, t *testing.T) {
	t.Helper()
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
}

var (
	wallet  = testutil.NewWallet("C6hPhCS2E2dKUGS3qj4264itKXohwgL3Lm2fNxayAKr")
	scOwner = wallet.WithIndex(0)
)

func setup(t *testing.T, configPath string, testName string) *cluster.Cluster {

	_, filename, _, _ := runtime.Caller(0)

	wasps, err := cluster.New(path.Join(path.Dir(filename), "..", configPath), "cluster-data")
	check(err, t)

	err = wasps.Init(true, testName)
	check(err, t)

	err = wasps.Start()
	check(err, t)

	t.Cleanup(wasps.Stop)

	return wasps
}
