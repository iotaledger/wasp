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

	err = SendRequests(wasps, &wasps.SmartContractConfig[0], 1, vmconst.RequestCodeNOP, nil, 0)
	check(err, t)

	wasps.CollectMessages(30 * time.Second)

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
		"request_in":          5,
		"request_out":         6,
		"state":               6,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = SendRequests(wasps, &wasps.SmartContractConfig[0], 1, vmconst.RequestCodeNOP, nil, 1*time.Second)
	check(err, t)

	wasps.CollectMessages(20 * time.Second)

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
		"request_in":          10,
		"request_out":         11,
		"state":               -1,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = SendRequests(wasps, &wasps.SmartContractConfig[0], 10, vmconst.RequestCodeNOP, nil, 0*time.Second)
	check(err, t)

	wasps.CollectMessages(20 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestSend60Requests(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend60Requests")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          60,
		"request_out":         61,
		"state":               -1,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = SendRequests(wasps, &wasps.SmartContractConfig[0], 60, vmconst.RequestCodeNOP, nil, 500*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(1 * time.Minute)

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
		"request_in":          60,
		"request_out":         61,
		"state":               -1,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	err = SendRequests(wasps, &wasps.SmartContractConfig[0], 60, vmconst.RequestCodeNOP, nil, 0*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(1 * time.Minute)

	if !wasps.Report() {
		t.Fail()
	}
}
