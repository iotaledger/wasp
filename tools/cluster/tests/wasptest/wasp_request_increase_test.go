package wasptest

import (
	"github.com/iotaledger/wasp/packages/vm/examples/increasecounter"
	"testing"
	"time"
)

func TestSend1ReqIncrease(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1ReqIncrease")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
		"state":               -1,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[2])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[2])
	check(err, t)

	err = SendRequestNTimes(wasps, &wasps.SmartContractConfig[2], 1, increasecounter.RequestIncrease, nil, 0)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestSend5ReqIncrease1Sec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend5ReqIncrease1Sec")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          6,
		"request_out":         7,
		"state":               -1,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[2])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[2])
	check(err, t)

	err = SendRequestNTimes(wasps, &wasps.SmartContractConfig[2], 5, increasecounter.RequestIncrease, nil, 1*time.Second)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestSend10ReqIncrease0Sec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend10ReqIncrease0Sec")

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
	err = Activate1SC(wasps, &wasps.SmartContractConfig[2])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[2])
	check(err, t)

	err = SendRequestNTimes(wasps, &wasps.SmartContractConfig[2], 10, increasecounter.RequestIncrease, nil, 0*time.Second)
	check(err, t)

	wasps.CollectMessages(20 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestSend60ReqIncrease500msec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend60Requests500msec")

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
	err = Activate1SC(wasps, &wasps.SmartContractConfig[2])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[2])
	check(err, t)

	err = SendRequestNTimes(wasps, &wasps.SmartContractConfig[2], 60, increasecounter.RequestIncrease, nil, 500*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(40 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestSend60ReqIncrease0Sec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend60ReqIncrease0Sec")

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
	err = Activate1SC(wasps, &wasps.SmartContractConfig[2])
	check(err, t)

	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[2])
	check(err, t)

	err = SendRequestNTimes(wasps, &wasps.SmartContractConfig[2], 60, increasecounter.RequestIncrease, nil, 0*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(40 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}
