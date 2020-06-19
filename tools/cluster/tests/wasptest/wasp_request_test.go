package wasptest

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
)

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

func TestSend10Requests0Sec(t *testing.T) {
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

	err = SendRequests(wasps, &wasps.SmartContractConfig[0], 10, 0*time.Second)
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
