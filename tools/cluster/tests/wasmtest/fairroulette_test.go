package wasmtest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/tools/cluster/tests/wasptest"
	"testing"
)

const (
	fr_code_nothing    = sctransaction.RequestCode(1)
	fr_code_placeBet   = sctransaction.RequestCode(2)
	fr_code_lockBets   = sctransaction.RequestCode(3)
	fr_code_payWinners = sctransaction.RequestCode(4)
	fr_code_playPeriod = sctransaction.RequestCode(5 | sctransaction.RequestCodeProtected)
)

const fr_wasmPath = "fairroulette_bg.wasm"
const fr_description = "Fair roulette, a PoC smart contract"

func TestFrNothing(t *testing.T) {
	testNothing(t, "TestFrNothing", fr_wasmPath, fr_description, 1)
}

func Test5xFrNothing(t *testing.T) {
	testNothing(t, "Test5xFrNothing", fr_wasmPath, fr_description, 5)
}

func TestPlaceBet(t *testing.T) {
	preamble(t, fr_wasmPath, fr_description, "TestPlaceBet")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 1,
		"request_out":         2 + 1,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	startSmartContract(t, scProgramHash, fr_description)

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress: scAddr,
		RequestCode: fr_code_placeBet,
		Vars: map[string]interface{}{
			"color":       3,
			"$haltEvents": 1, // do not propagate queued events
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 100,
		},
	})
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1-100, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 - 100,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 1+100, map[balance.Color]int64{
		balance.ColorIOTA: 100,
		*scColor:          1,
	}, "sc in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  scProgramHash[:],
		vmconst.VarNameDescription:  fr_description,
	}) {
		t.Fail()
	}
}

func TestPlace5BetsAndPlay(t *testing.T) {
	preamble(t, fr_wasmPath, fr_description, "TestPlace5BetsAndPlay")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 1 + 5 + 1 + 1,
		"request_out":         2 + 1 + 5 + 1 + 1,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	startSmartContract(t, scProgramHash, fr_description)

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress: scAddr,
		RequestCode: fr_code_playPeriod,
		Vars: map[string]interface{}{
			"playPeriod": 10,
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 1,
		},
	})
	check(err, t)

	for i := 0; i < 5; i++ {
		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress: scAddr,
			RequestCode: fr_code_placeBet,
			Vars: map[string]interface{}{
				"color": i + 1,
			},
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: 100,
			},
		})
		check(err, t)
	}

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-2, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 2,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 2, map[balance.Color]int64{
		balance.ColorIOTA: 1,
		*scColor:          1,
	}, "sc in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  scProgramHash[:],
		vmconst.VarNameDescription:  fr_description,
		"playPeriod":                10,
		//"lastWinningColor": 3,
	}) {
		t.Fail()
	}
}
