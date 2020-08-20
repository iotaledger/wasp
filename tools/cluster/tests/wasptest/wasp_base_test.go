package wasptest

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"testing"
)

func TestPutBootupRecords(t *testing.T) {
	wasps := setup(t, "test_cluster", "TestPutBootupRecords")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           1,
		"active_committee":    0,
		"dismissed_committee": 0,
		"request_in":          0,
		"request_out":         0,
		"state":               0,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[0]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
}

func TestActivate1SC(t *testing.T) {
	wasps := setup(t, "test_cluster", "TestActivate1SC")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           1,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          0,
		"request_out":         0,
		"state":               0,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[0]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
}

func TestActivateAllSC(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestActivateAllSC")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    wasps.NumSmartContracts(),
		"dismissed_committee": 0,
		"request_in":          0,
		"request_out":         0,
		"state":               0,
	})
	check(err, t)

	for _, sc := range wasps.SmartContractConfig {
		_, err = PutBootupRecord(wasps, &sc)
		check(err, t)
	}

	// exercise
	err = ActivateAllSC(wasps)
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
}

func TestCreateOrigin(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestCreateOrigin")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           1,
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               2,
		"request_in":          1,
		"request_out":         2,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[0]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	// exercise
	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
	if !wasps.VerifySCState(sc, 1, map[kv.Key][]byte{
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
	}) {
		t.Fail()
	}
}
