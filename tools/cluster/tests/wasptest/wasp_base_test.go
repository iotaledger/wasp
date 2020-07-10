package wasptest

import (
	"testing"
	"time"
)

func TestPutBootupRecords(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestPutBootupRecords")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    0,
		"dismissed_committee": 0,
		"request_in":          0,
		"request_out":         0,
		"state":               0,
	})
	check(err, t)

	// exercise
	err = PutBootupRecords(wasps)
	check(err, t)

	wasps.CollectMessages(10 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestActivate1SC(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestActivate1SC")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          0,
		"request_out":         0,
		"state":               0,
	})
	check(err, t)

	err = PutBootupRecords(wasps)
	check(err, t)

	// exercise
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	wasps.CollectMessages(5 * time.Second)
	if !wasps.Report() {
		t.Fail()
	}
}

func TestActivateAllSC(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestActivateAllSC")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    6,
		"dismissed_committee": 0,
		"request_in":          0,
		"request_out":         0,
		"state":               0,
	})
	check(err, t)

	err = PutBootupRecords(wasps)
	check(err, t)

	// exercise
	err = ActivateAllSC(wasps)
	check(err, t)

	wasps.CollectMessages(5 * time.Second)
	if !wasps.Report() {
		t.Fail()
	}
}

func TestCreateOrigin(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestCreateOrigin")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               2,
		"request_in":          1,
		"request_out":         2,
	})
	check(err, t)

	err = PutBootupRecords(wasps)
	check(err, t)
	err = Activate1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	// exercise
	err = CreateOrigin1SC(wasps, &wasps.SmartContractConfig[0])
	check(err, t)

	wasps.CollectMessages(10 * time.Second)
	if !wasps.Report() {
		t.Fail()
	}
}
