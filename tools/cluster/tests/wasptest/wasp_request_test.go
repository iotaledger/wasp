package wasptest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/vm/vmconst"
)

func TestSend1Request(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1Request")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
		"state":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[0]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress := sc.SCAddress()

	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
		RequestCode: vmconst.RequestCodeNOP,
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

	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  sc.GetProgramHash().Bytes(),
	}) {
		t.Fail()
	}
}

func TestSend5Requests1Sec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend5Requests1Sec")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          6,
		"request_out":         7,
		"state":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[0]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	scAddress := sc.SCAddress()

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddress,
			RequestCode: vmconst.RequestCodeNOP,
		})
		check(err, t)
		time.Sleep(1 * time.Second)
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
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  sc.GetProgramHash().Bytes(),
	}) {
		t.Fail()
	}
}

func TestSend10Requests0Sec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend10Requests0Sec")

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

	sc := &wasps.SmartContractConfig[0]

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
			RequestCode: vmconst.RequestCodeNOP,
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
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  sc.GetProgramHash().Bytes(),
	}) {
		t.Fail()
	}
}

func TestSend60Requests500msec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend60Requests")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          61,
		"request_out":         62,
		"state":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[0]

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
			RequestCode: vmconst.RequestCodeNOP,
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
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  sc.GetProgramHash().Bytes(),
	}) {
		t.Fail()
	}
}

func TestSend60Requests0Sec(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend10Requests0Sec")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          61,
		"request_out":         62,
		"state":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[0]

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
			RequestCode: vmconst.RequestCodeNOP,
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
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  sc.GetProgramHash().Bytes(),
	}) {
		t.Fail()
	}
}
