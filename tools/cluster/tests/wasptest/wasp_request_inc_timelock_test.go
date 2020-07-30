package wasptest

import (
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"testing"
	"time"
)

func TestSend1ReqIncTimelock(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1ReqIncTimelock")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
		"state":               -1,
	})
	check(err, t)

	_, err = PutBootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[2]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	reqs := []*waspapi.RequestBlockJson{{
		Address:     sc.Address,
		RequestCode: inccounter.RequestInc,
		Timelock:    util.UnixAfterSec(3),
	}}
	err = SendRequestsNTimes(wasps, sc.OwnerSigScheme(), 1, reqs, 0*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(20 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	if !wasps.VerifySCState(sc, 2, map[kv.Key][]byte{
		"counter": util.Uint64To8Bytes(uint64(1)),
	}) {
		t.Fail()
	}
}

func TestSend1ReqIncRepeatTimelock(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1ReqIncRepeatTimelock")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          3,
		"request_out":         4,
		"state":               -1,
	})
	check(err, t)

	_, err = PutBootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[2]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	reqs := []*waspapi.RequestBlockJson{{
		Address:     sc.Address,
		RequestCode: inccounter.RequestIncAndRepeatOnceAfter5s,
	}}
	err = SendRequestsNTimes(wasps, sc.OwnerSigScheme(), 1, reqs, 0*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter": util.Uint64To8Bytes(uint64(2)),
	}) {
		t.Fail()
	}
}

const chainOfRequestsLength = 5

func TestChainIncTimelock(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestChainIncTimelock")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          chainOfRequestsLength + 2,
		"request_out":         chainOfRequestsLength + 3,
		"state":               -1,
	})
	check(err, t)

	_, err = PutBootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[2]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	reqs := []*waspapi.RequestBlockJson{{
		Address:     sc.Address,
		RequestCode: inccounter.RequestIncAndRepeatMany,
		Vars: map[string]interface{}{
			inccounter.ArgNumRepeats: chainOfRequestsLength,
		},
	}}
	err = SendRequestsNTimes(wasps, sc.OwnerSigScheme(), 1, reqs, 0*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(30 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter":                util.Uint64To8Bytes(uint64(chainOfRequestsLength + 1)),
		inccounter.ArgNumRepeats: util.Uint64To8Bytes(0),
	}) {
		t.Fail()
	}
}
