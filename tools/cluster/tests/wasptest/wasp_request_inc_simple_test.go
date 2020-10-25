package wasptest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"testing"
	"time"
)

func TestSend1ReqIncSimple(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1ReqIncSimple")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
		"state":               -1, // we dont know if it will go in same batch with init request or separate
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[2]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress := sc.SCAddress()

	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
		RequestCode: inccounter.RequestInc,
	})
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifySCState(sc, 2, map[kv.Key][]byte{
		"counter":                   util.Uint64To8Bytes(uint64(1)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
	}) {
		t.Fail()
	}
}

func TestSend5ReqInc0SecSimple(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend5ReqInc0SecSimple")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          6,
		"request_out":         7,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[2]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)
	time.Sleep(1 * time.Second)

	scAddress := sc.SCAddress()

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddress,
			RequestCode: inccounter.RequestInc,
		})
		check(err, t)
	}

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
	wasps.Report()

	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter":                   util.Uint64To8Bytes(uint64(5)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
	}) {
		t.Fail()
	}
}

func TestSend10ReqIncrease0SecSimple(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend10ReqIncrease0SecSimple")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          11,
		"request_out":         12,
		"vmready":             -1,
		"state":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[2]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress := sc.SCAddress()

	for i := 0; i < 10; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddress,
			RequestCode: inccounter.RequestInc,
		})
		check(err, t)
	}

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter":                   util.Uint64To8Bytes(uint64(10)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
	}) {
		t.Fail()
	}
}

func TestSend60ReqIncrease500msecSimple(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend60ReqIncrease500msecSimple")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          61,
		"request_out":         62,
		"state":               -1,
		"vmmsg":               -1, // 60 or less
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[2]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress := sc.SCAddress()

	for i := 0; i < 60; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddress,
			RequestCode: inccounter.RequestInc,
		})
		check(err, t)
		time.Sleep(500 * time.Millisecond)
	}

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter":                   util.Uint64To8Bytes(uint64(60)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
	}) {
		t.Fail()
	}
}

func TestSend60ReqInc0SecSimple(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend60ReqInc0SecSimple")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          61,
		"request_out":         62,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[2]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress := sc.SCAddress()

	for i := 0; i < 60; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddress,
			RequestCode: inccounter.RequestInc,
		})
		check(err, t)
	}

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter":                   util.Uint64To8Bytes(uint64(60)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramData:  sc.GetProgramHash().Bytes(),
	}) {
		t.Fail()
	}
}
