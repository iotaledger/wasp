package wasptest

import (
	"github.com/iotaledger/wasp/packages/subscribe"
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

func setup(t *testing.T, name string) *cluster.Cluster {
	_, filename, _, _ := runtime.Caller(0)

	wasps, err := cluster.New(path.Join(path.Dir(filename), "../test_cluster"), "cluster-data")
	check(err, t)

	err = wasps.Init(true, name)
	check(err, t)

	err = wasps.Start()
	check(err, t)

	t.Cleanup(wasps.Stop)

	return wasps
}

func count(msgs []*subscribe.HostMessage) {

}
