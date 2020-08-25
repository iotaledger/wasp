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

func TestSend1ReqIncTimelock(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1ReqIncTimelock")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           1,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
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

	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
		RequestCode: inccounter.RequestInc,
		Timelock:    util.UnixAfterSec(3),
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
		"counter":                   util.Uint64To8Bytes(uint64(1)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  sc.GetProgramHash().Bytes(),
	}) {
		t.Fail()
	}
}

func TestSend1ReqIncRepeatFailTimelock(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1ReqIncRepeatFailTimelock")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           1,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
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

	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
		RequestCode: inccounter.RequestIncAndRepeatOnceAfter5s,
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
		"counter":                   util.Uint64To8Bytes(uint64(1)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  sc.GetProgramHash().Bytes(),
	}) {
		t.Fail()
	}
}

func TestSend1ReqIncRepeatSuccessTimelock(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSend1ReqIncRepeatSuccessTimelock")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           1,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          4,
		"request_out":         5,
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

	// send 1i to the SC address. It is needed to send the request to self
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
		RequestCode: vmconst.RequestCodeNOP,
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 1,
		},
	})
	check(err, t)
	time.Sleep(1 * time.Second)

	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
		RequestCode: inccounter.RequestIncAndRepeatOnceAfter5s,
	})
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scAddress, 2, map[balance.Color]int64{
		balance.ColorIOTA: 1,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter":                   util.Uint64To8Bytes(uint64(2)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  sc.GetProgramHash().Bytes(),
	}) {
		t.Fail()
	}
}

const chainOfRequestsLength = 5

func TestChainIncTimelock(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestChainIncTimelock")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           1,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          chainOfRequestsLength + 3,
		"request_out":         chainOfRequestsLength + 4,
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

	// send 5i to the SC address. It is needed to send 5 requests to self at once
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
		RequestCode: vmconst.RequestCodeNOP,
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 5,
		},
	})
	check(err, t)
	time.Sleep(1 * time.Second)

	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
		RequestCode: inccounter.RequestIncAndRepeatMany,
		Vars: map[string]interface{}{
			inccounter.ArgNumRepeats: chainOfRequestsLength,
		},
	})
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scAddress, 6, map[balance.Color]int64{
		balance.ColorIOTA: 5,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		"counter":                   util.Uint64To8Bytes(uint64(chainOfRequestsLength + 1)),
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  sc.GetProgramHash().Bytes(),
		inccounter.VarNumRepeats:    util.Uint64To8Bytes(0),
	}) {
		t.Fail()
	}
}
