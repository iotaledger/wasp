package wasptest

import (
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"testing"
	"time"
)

func TestSend1ReqIncSimple(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1ReqIncSimple")

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
	}}
	err = SendRequestsNTimes(wasps, sc.OwnerSigScheme(), 1, reqs, 0*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	if !wasps.VerifySCState(sc, 2, map[kv.Key][]byte{
		"counter": util.Uint64To8Bytes(uint64(1)),
	}) {
		t.Fail()
	}
}

func TestSend5ReqInc1SecSimple(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend5ReqInc1SecSimple")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          6,
		"request_out":         7,
		"state":               -1,
		"vmmsg":               -1,
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
	}}
	err = SendRequestsNTimes(wasps, sc.OwnerSigScheme(), 5, reqs, 1*time.Second)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter": util.Uint64To8Bytes(uint64(5)),
	}) {
		t.Fail()
	}
}

func TestSend10ReqIncrease0SecSimple(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend10ReqIncrease0SecSimple")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          11,
		"request_out":         12,
		"vmready":             -1,
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
	}}
	err = SendRequestsNTimes(wasps, sc.OwnerSigScheme(), 10, reqs, 0*time.Second)
	check(err, t)

	wasps.CollectMessages(20 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter": util.Uint64To8Bytes(uint64(10)),
	}) {
		t.Fail()
	}
}

func TestSend60ReqIncrease500msecSimple(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend60ReqIncrease500msecSimple")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          61,
		"request_out":         62,
		"state":               -1,
		"vmmsg":               -1, // 60 or less
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
	}}
	err = SendRequestsNTimes(wasps, sc.OwnerSigScheme(), 60, reqs, 500*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(40 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter": util.Uint64To8Bytes(uint64(60)),
	}) {
		t.Fail()
	}
}

func TestSend60ReqInc0SecSimple(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend60ReqInc0SecSimple")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          61,
		"request_out":         62,
		"state":               -1,
		"vmmsg":               -1,
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
	}}
	err = SendRequestsNTimes(wasps, sc.OwnerSigScheme(), 60, reqs, 0*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(40 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter": util.Uint64To8Bytes(uint64(60)),
	}) {
		t.Fail()
	}
}
