package wasptest

import (
	"path"
	"runtime"
	"testing"

	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/assert"
)

func check(err error, t *testing.T) {
	t.Helper()
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
}

func setup(t *testing.T) *cluster.Cluster {
	_, filename, _, _ := runtime.Caller(0)

	wasps, err := cluster.New(path.Join(path.Dir(filename), "../test_cluster"), "cluster-data")
	check(err, t)

	err = wasps.Init(true)
	check(err, t)

	err = wasps.Start()
	check(err, t)

	t.Cleanup(wasps.Stop)

	return wasps
}

func TestPutBootupRecords(t *testing.T) {
	// setup
	wasps := setup(t)

	allNodesNanomsg := wasps.WaspHosts(wasps.AllWaspNodes(), (*cluster.WaspNodeConfig).NanomsgHost)

	messages := make(chan *publisher.HostMessage)
	done := make(chan bool)
	defer close(done)
	err := publisher.SubscribeMulti(allNodesNanomsg, messages, done, "bootuprec")
	check(err, t)

	// exercise
	err = Put3BootupRecords(wasps)
	check(err, t)

	// verify
	expected := make(map[string]int)
	for _, sc := range wasps.SmartContractConfig {
		for _, wasp := range allNodesNanomsg {
			expected[wasp+sc.Address] = 1
		}
	}
	received := make(map[string]int)
	for i := 0; i < len(allNodesNanomsg)*len(wasps.SmartContractConfig); i++ {
		msg := <-messages
		received[msg.Sender+msg.Message[1]] += 1
	}
	assert.Equal(t, expected, received)
}

func TestActivate1SC(t *testing.T) {
	// setup
	wasps := setup(t)
	err := Put3BootupRecords(wasps)
	check(err, t)

	// exercise
	err = Activate1SC(wasps)
	check(err, t)

	// verify
	// TODO
}

func TestActivate3SC(t *testing.T) {
	// setup
	wasps := setup(t)
	err := Put3BootupRecords(wasps)
	check(err, t)

	// exercise
	err = Activate3SC(wasps)
	check(err, t)

	// verify
	// TODO
}

func TestCreateOrigin(t *testing.T) {
	// setup
	wasps := setup(t)
	err := Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps)
	check(err, t)

	// exercise
	err = CreateOrigin1SC(wasps)
	check(err, t)

	// verify
	// TODO
}

func TestSendRequest(t *testing.T) {
	// setup
	wasps := setup(t)
	err := Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps)
	check(err, t)

	// exercise
	err = Send1Request(wasps)
	check(err, t)

	// verify
	// TODO
}
