package wasptest

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPutBootupRecords(t *testing.T) {
	// setup
	wasps := setup(t, "TestPutBootupRecords")

	allNodesNanomsg := wasps.WaspHosts(wasps.AllWaspNodes(), (*cluster.WaspNodeConfig).NanomsgHost)

	messages := make(chan *subscribe.HostMessage)
	done := make(chan bool)
	defer close(done)
	err := subscribe.SubscribeMulti(allNodesNanomsg, messages, done,
		"bootuprec", "active_committee", "dismissed_committee")
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

	fmt.Printf("[cluster] waiting 10 sec...\n")
	assert.Equal(t, expected, received)
}

func TestActivate1SC(t *testing.T) {
	// setup
	wasps := setup(t, "TestActivate1SC")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 1})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)

	// exercise
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	wasps.CollectMessages(5 * time.Second)
	wasps.Report()
}

func TestActivate3SC(t *testing.T) {
	// setup
	wasps := setup(t, "TestActivate3SC")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    3,
		"dismissed_committee": 3})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)

	// exercise
	err = Activate3SC(wasps)
	check(err, t)

	wasps.CollectMessages(5 * time.Second)
	wasps.Report()
}

func TestCreateOrigin(t *testing.T) {
	// setup
	wasps := setup(t, "TestCreateOrigin")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 1,
		"state":               1,
		"request_in":          0,
		"request_out":         1,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	// exercise
	err = CreateOrigin1SC(wasps)
	check(err, t)

	wasps.CollectMessages(5 * time.Second)
	wasps.Report()
}
