package wasptest

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

func TestPutBootupRecord(t *testing.T) {
	wasps := setup(t, "test_cluster", "TestPutBootupRecord")

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

// TODO deploy cycle with chainclient

func TestDeployChain(t *testing.T) {
	wasps := setup(t, "test_cluster", "TestDeployChain")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"state":               2,
		"request_in":          1,
		"request_out":         1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[0]

	err = wasps.NodeClient.RequestFunds(sc.OwnerAddress())
	check(err, t)

	_, _, _, err = wasps.DeployChain(sc, wasps.Config.SmartContracts[0].Quorum)
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
	time.Sleep(5 * time.Second)

	if !wasps.VerifySCState(sc, 1, map[kv.Key][]byte{
		root.VarChainID:          sc.SCAddress().Bytes(),
		root.VarDescription:      []byte(sc.Description),
		root.VarStateInitialized: []byte{0xFF},
	}) {
		t.Fail()
	}
}

//
//func TestActivate1Chain(t *testing.T) {
//	wasps := setup(t, "test_cluster", "TestActivate1Chain")
//
//	err := wasps.ListenToMessages(map[string]int{
//		"bootuprec":           2,
//		"active_committee":    1,
//		"dismissed_committee": 0,
//		"request_in":          0,
//		"request_out":         0,
//		"state":               0,
//	})
//	check(err, t)
//
//	sc := &wasps.SmartContractConfig[0]
//
//	_, err = PutBootupRecord(wasps, sc)
//	check(err, t)
//
//	err = Activate1Chain(wasps, sc)
//	check(err, t)
//
//	if !wasps.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//}
//
//func TestActivateAllChains(t *testing.T) {
//	wasps := setup(t, "test_cluster", "TestActivateAllSC")
//
//	err := wasps.ListenToMessages(map[string]int{
//		"bootuprec":           wasps.NumSmartContracts() * 2,
//		"active_committee":    wasps.NumSmartContracts(),
//		"dismissed_committee": 0,
//		"request_in":          0,
//		"request_out":         0,
//		"state":               0,
//	})
//	check(err, t)
//
//	for _, sc := range wasps.SmartContractConfig {
//		_, err = PutBootupRecord(wasps, &sc)
//		check(err, t)
//	}
//
//	err = ActivateAllSC(wasps)
//	check(err, t)
//
//	if !wasps.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//}
//
//func TestDeactivate1Chain(t *testing.T) {
//	wasps := setup(t, "test_cluster", "TestDeactivate1SC")
//
//	err := wasps.ListenToMessages(map[string]int{
//		"bootuprec":           3,
//		"active_committee":    1,
//		"dismissed_committee": 1,
//		"request_in":          0,
//		"request_out":         0,
//		"state":               0,
//	})
//	check(err, t)
//
//	sc := &wasps.SmartContractConfig[0]
//
//	_, err = PutBootupRecord(wasps, sc)
//	check(err, t)
//
//	err = Activate1Chain(wasps, sc)
//	check(err, t)
//
//	time.Sleep(5 * time.Second)
//
//	err = Deactivate1Chain(wasps, sc)
//	check(err, t)
//
//	if !wasps.WaitUntilExpectationsMet() {
//		t.Fail()
//	}
//}
