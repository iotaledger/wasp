package wasptest

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/vm/vmconst"
)

func TestSend1Request(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1Request")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
		"state":               2,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = SendRequestNTimes(wasps, &wasps.SmartContractConfig[0], 1, vmconst.RequestCodeNOP, nil, 0)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestSend5Requests1Sec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend5Requests1Sec")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          6,
		"request_out":         7,
		"state":               7,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = SendRequestNTimes(wasps, &wasps.SmartContractConfig[0], 5, vmconst.RequestCodeNOP, nil, 1*time.Second)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestSend10Requests0Sec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend10Requests0Sec")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          11,
		"request_out":         12,
		"vmready":             -1,
		"state":               -1,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = SendRequestNTimes(wasps, &wasps.SmartContractConfig[0], 10, vmconst.RequestCodeNOP, nil, 0*time.Second)
	check(err, t)

	wasps.CollectMessages(20 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestSend60Requests500msec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend60Requests")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          61,
		"request_out":         62,
		"state":               60,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = SendRequestNTimes(wasps, &wasps.SmartContractConfig[0], 60, vmconst.RequestCodeNOP, nil, 500*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(40 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestSend60Requests0Sec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend10Requests0Sec")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          61,
		"request_out":         62,
		"state":               -1,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = SendRequestNTimes(wasps, &wasps.SmartContractConfig[0], 60, vmconst.RequestCodeNOP, nil, 0*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(40 * time.Minute)

	if !wasps.Report() {
		t.Fail()
	}
}
