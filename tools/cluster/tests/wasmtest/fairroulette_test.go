package wasmtest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/tools/cluster/tests/wasptest"
	"testing"
)

const (
	frCodePlaceBet   = sctransaction.RequestCode(1)
	frCodeLockBets   = sctransaction.RequestCode(2)
	frCodePayWinners = sctransaction.RequestCode(3)
	frCodePlayPeriod = sctransaction.RequestCode(4 | sctransaction.RequestCodeProtected)
	frCodeNothing    = sctransaction.RequestCode(5)
)

const frWasmPath = "wasm/fairroulette"
const frDescription = "Fair roulette, a PoC smart contract"

func TestFrNothing(t *testing.T) {
	testNothing(t, "TestFrNothing", fairroulette.ProgramHash, frWasmPath, frDescription, 1)
}

func Test5xFrNothing(t *testing.T) {
	testNothing(t, "Test5xFrNothing", fairroulette.ProgramHash, frWasmPath, frDescription, 5)
}

func TestPlaceBet(t *testing.T) {
	wasps := setup(t, "TestPlaceBet")

	err := loadWasmIntoWasps(wasps, frWasmPath, frDescription)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 1,
		"request_out":         2 + 0,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, fairroulette.ProgramHash, frDescription)
	checkSuccess(err, t, "smart contract has been created and activated")

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddr,
		RequestCode: frCodePlaceBet,
		Vars: map[string]interface{}{
			"color": 3,
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 100,
		},
	})
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-2-100, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 2 - 100,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 2+100, map[balance.Color]int64{
		balance.ColorIOTA: 100, // 1 extra is locked up in time-locked
		*scColor:          1,   // request that hasn't been processed yet
	}, "sc in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  programHash[:],
		vmconst.VarNameDescription:  frDescription,
	}) {
		t.Fail()
	}
}

func TestPlace5BetsAndPlay(t *testing.T) {
	wasps := setup(t, "TestPlace5BetsAndPlay")

	err := loadWasmIntoWasps(wasps, frWasmPath, frDescription)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 1 + 5 + 1 + 1,
		"request_out":         2 + 1 + 5 + 1 + 1,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, fairroulette.ProgramHash, frDescription)
	checkSuccess(err, t, "smart contract has been created and activated")

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddr,
		RequestCode: frCodePlayPeriod,
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
			SCAddress:   scAddr,
			RequestCode: frCodePlaceBet,
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
		vmconst.VarNameProgramHash:  programHash[:],
		vmconst.VarNameDescription:  frDescription,
		"playPeriod":                10,
		//"lastWinningColor": 3,
	}) {
		t.Fail()
	}
}
