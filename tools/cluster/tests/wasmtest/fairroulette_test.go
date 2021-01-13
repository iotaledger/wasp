// +build ignore

package wasmtest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/iotaledger/wasp/tools/cluster/tests/wasptest"
	"testing"
)

const (
	frCodePlaceBet   = coretypes.Hname(1)
	frCodeLockBets   = coretypes.Hname(2)
	frCodePayWinners = coretypes.Hname(3)
	frCodePlayPeriod = coretypes.Hname(4)
	frCodeNothing    = coretypes.Hname(5)
)

const frWasmPath = "wasm/fairroulette"
const frDescription = "Fair roulette, a PoC smart contract"

func TestFrNothing(t *testing.T) {
	testNothing(t, "TestFrNothing", fairroulette.ProgramHash, frWasmPath, frDescription, 1)
}

func TestFr5xNothing(t *testing.T) {
	testNothing(t, "TestFr5xNothing", fairroulette.ProgramHash, frWasmPath, frDescription, 5)
}

func TestFrPlaceBet(t *testing.T) {
	wasps := setup(t, "TestFrPlaceBet")

	err := loadWasmIntoWasps(wasps, frWasmPath, frDescription)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.StartMessageCounter(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 1,
		"request_out":         2 + 0,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scChain, scAddr, scColor, err := startSmartContract(wasps, fairroulette.ProgramHash, frDescription)
	checkSuccess(err, t, "smart contract has been created and activated")

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		TargetContract: coretypes.NewContractID(*scChain, 0),
		RequestCode:    frCodePlaceBet,
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
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  frDescription,
	}) {
		t.Fail()
	}
}

func TestFrPlace1BetAndWin(t *testing.T) {
	wasps := setup(t, "TestFrPlace5BetsAndPlay")
	testFrPlaceBetsAndPlay(t, 1, wasps)
}

func TestFrPlace5BetsAndWin(t *testing.T) {
	wasps := setup(t, "TestFrPlace5BetsAndPlay")
	testFrPlaceBetsAndPlay(t, 5, wasps)
}

func testFrPlaceBetsAndPlay(t *testing.T, nrOfBets int, wasps *cluster.Cluster) {
	err := loadWasmIntoWasps(wasps, frWasmPath, frDescription)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.StartMessageCounter(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 1 + nrOfBets + 1 + 1,
		"request_out":         2 + 1 + nrOfBets + 1 + 1,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scChain, scAddr, scColor, err := startSmartContract(wasps, fairroulette.ProgramHash, frDescription)
	checkSuccess(err, t, "smart contract has been created and activated")

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		TargetContract: coretypes.NewContractID(*scChain, 0),
		RequestCode:    frCodePlayPeriod,
		Vars: map[string]interface{}{
			"playPeriod": 10,
			"testMode":   1,
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 1,
		},
	})
	check(err, t)

	for i := 0; i < nrOfBets; i++ {
		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			TargetContract: coretypes.NewContractID(*scChain, 0),
			RequestCode:    frCodePlaceBet,
			Vars: map[string]interface{}{
				"color": (1+i)%5 + 1,
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
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  frDescription,
		"playPeriod":                10,
		"lastWinningColor":          2,
	}) {
		t.Fail()
	}
}
