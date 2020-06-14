package wasptest

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/assert"
	"path"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
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
	wasps := setup(t)

	err := wasps.ListenToMessages("bootuprec", "active_committee", "dismissed_committee")
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)

	// exercise
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	// count
	_, counters := wasps.CountMessages(5 * time.Second)

	fmt.Printf("[cluster] ++++++++++ counters\n")
	keys := make([]string, 0)
	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("[cluster] ++++++++++ %s :%d\n", k, counters[k])
	}

	// verify
}

func TestActivate3SC(t *testing.T) {
	// setup
	wasps := setup(t)

	err := wasps.ListenToMessages("bootuprec", "active_committee", "dismissed_committee", "request")
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)

	// exercise
	err = Activate3SC(wasps)
	check(err, t)

	// count
	allMsg, counters := wasps.CountMessages(5 * time.Second)

	fmt.Printf("[cluster] ++++++++++ counters\n")
	keys := make([]string, 0)
	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("[cluster] ++++++++++ %s :%d\n", k, counters[k])
	}

	for _, msg := range allMsg {
		fmt.Printf("%s => '%s'\n", msg.Sender, strings.Join(msg.Message, " "))
	}
	// verify
}

func TestCreateOrigin(t *testing.T) {
	// setup
	wasps := setup(t)

	err := wasps.ListenToMessages("bootuprec", "active_committee", "dismissed_committee", "state", "request")
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	// exercise
	err = CreateOrigin1SC(wasps)
	check(err, t)

	allMsg, counters := wasps.CountMessages(5 * time.Second)

	fmt.Printf("[cluster] ++++++++++ counters\n")
	keys := make([]string, 0)
	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("[cluster] ++++++++++ %s :%d\n", k, counters[k])
	}

	for _, msg := range allMsg {
		fmt.Printf("%s => '%s'\n", msg.Sender, strings.Join(msg.Message, " "))
	}
	// verify
	// TODO
}

func TestSend1Request(t *testing.T) {
	// setup
	wasps := setup(t)

	err := wasps.ListenToMessages("bootuprec", "active_committee", "dismissed_committee", "state", "request")
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = CreateOrigin1SC(wasps)
	check(err, t)

	err = SendRequests(wasps, &wasps.SmartContractConfig[0], 1, 0)
	check(err, t)

	allMsg, counters := wasps.CountMessages(15 * time.Second)

	fmt.Printf("[cluster] ++++++++++ counters\n")
	keys := make([]string, 0)
	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("[cluster] ++++++++++ %s :%d\n", k, counters[k])
	}

	for _, msg := range allMsg {
		fmt.Printf("%s => '%s'\n", msg.Sender, strings.Join(msg.Message, " "))
	}
	// verify
}

func TestSend5Requests1Sec(t *testing.T) {
	// setup
	wasps := setup(t)

	err := wasps.ListenToMessages("bootuprec", "active_committee", "dismissed_committee", "state", "request")
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = CreateOrigin1SC(wasps)
	check(err, t)

	err = SendRequests(wasps, &wasps.SmartContractConfig[0], 5, 1*time.Second)
	check(err, t)

	allMsg, counters := wasps.CountMessages(20 * time.Second)

	fmt.Printf("[cluster] ++++++++++ counters\n")
	keys := make([]string, 0)
	for k := range counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("[cluster] ++++++++++ %s :%d\n", k, counters[k])
	}

	for _, msg := range allMsg {
		fmt.Printf("%s => '%s'\n", msg.Sender, strings.Join(msg.Message, " "))
	}
	// verify
}
