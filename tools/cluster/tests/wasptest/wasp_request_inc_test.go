package wasptest

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"testing"
	"time"
)

func TestSend1ReqInc(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1ReqInc")

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

	sc := &wasps.SmartContractConfig[2]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	err = SendRequestNTimes(wasps, sc, 1, inccounter.RequestInc, nil, 0)
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

func TestSend1ReqIncTimelock(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1ReqIncTimelock")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          11,
		"request_out":         10,
		"state":               10,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[2]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	err = SendRequestNTimes(wasps, sc, 1, inccounter.RequestIncAndRepeat, nil, 0)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	if !wasps.VerifySCState(sc, 10, map[kv.Key][]byte{
		"counter": util.Uint64To8Bytes(uint64(10)),
	}) {
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
		"vmmsg":               5,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[2]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	err = SendRequestNTimes(wasps, sc, 5, inccounter.RequestInc, nil, 1*time.Second)
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

	sc := &wasps.SmartContractConfig[2]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	err = SendRequestNTimes(wasps, sc, 10, inccounter.RequestInc, nil, 0*time.Second)
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

func TestSend60ReqIncrease500msec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend60Requests500msec")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          61,
		"request_out":         62,
		"state":               61,
		"vmmsg":               60,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[2]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	err = SendRequestNTimes(wasps, sc, 60, inccounter.RequestInc, nil, 500*time.Millisecond)
	check(err, t)

	wasps.CollectMessages(40 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	if !wasps.VerifySCState(sc, 60, map[kv.Key][]byte{
		"counter": util.Uint64To8Bytes(uint64(60)),
	}) {
		t.Fail()
	}
}

func TestSend60ReqInc0Sec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend60ReqInc0Sec")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           3,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          61,
		"request_out":         62,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	err = Put3BootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[2]
	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	err = SendRequestNTimes(wasps, sc, 60, inccounter.RequestInc, nil, 0*time.Millisecond)
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
