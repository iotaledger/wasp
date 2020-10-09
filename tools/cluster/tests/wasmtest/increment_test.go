package wasmtest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/tools/cluster/tests/wasptest"
	"testing"
)

const wasmPath = "increment_bg.wasm"
const scDescription = "Increment, a PoC smart contract"

func TestDeployment(t *testing.T) {
	preamble(t, wasmPath, scDescription, "TestDeployment")
	startSmartContract(t, scProgramHash, scDescription)

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 1, map[balance.Color]int64{
		*scColor: 1,
	}, "sc in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  scProgramHash[:],
		vmconst.VarNameDescription:  scDescription,
	}) {
		t.Fail()
	}
}

func TestNothing(t *testing.T) {
	testNothing(t, "TestNothing", 1)
}

func Test5xNothing(t *testing.T) {
	testNothing(t, "Test5xNothing", 5)
}

func testNothing(t *testing.T, testName string, numRequests int) {
	preamble(t, wasmPath, scDescription, testName)

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + numRequests,
		"request_out":         2 + numRequests,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	startSmartContract(t, scProgramHash, scDescription)

	for i := 0; i < numRequests; i++ {
		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress: scAddr,
			Vars: map[string]interface{}{
				"fn": "nothing",
			},
		})
		check(err, t)
	}

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		*scColor:          1,
	}, "sc in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  scProgramHash[:],
		vmconst.VarNameDescription:  scDescription,
	}) {
		t.Fail()
	}
}

func TestIncrement(t *testing.T) {
	testIncrement(t, "TestIncrement", 1)
}

func Test5xIncrement(t *testing.T) {
	testIncrement(t, "Test5xIncrement", 5)
}

func testIncrement(t *testing.T, testName string, increments int) {
	preamble(t, wasmPath, scDescription, testName)

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + increments,
		"request_out":         2 + increments,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	startSmartContract(t, scProgramHash, scDescription)

	for i := 0; i < increments; i++ {
		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress: scAddr,
			Vars: map[string]interface{}{
				"fn": "increment",
			},
		})
		check(err, t)
	}

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		*scColor:          1,
	}, "sc in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  scProgramHash[:],
		vmconst.VarNameDescription:  scDescription,
		"counter":                   util.Uint64To8Bytes(uint64(increments)),
	}) {
		t.Fail()
	}
}

func TestRepeatIncrement(t *testing.T) {
	preamble(t, wasmPath, scDescription, "TestRepeatIncrement")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 2,
		"request_out":         2 + 2,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	startSmartContract(t, scProgramHash, scDescription)

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress: scAddr,
		Vars: map[string]interface{}{
			"fn": "incrementRepeat1",
		},
		// also send 1i to the SC address to use as request token
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 1,
		},
	})
	check(err, t)

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
		vmconst.VarNameDescription:  scDescription,
		"counter":                   util.Uint64To8Bytes(uint64(2)),
	}) {
		t.Fail()
	}
}

func TestRepeatManyIncrement(t *testing.T) {
	const numRepeats = 5
	preamble(t, wasmPath, scDescription, "TestRepeatManyIncrement")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 1 + numRepeats,
		"request_out":         2 + 1 + numRepeats,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	startSmartContract(t, scProgramHash, scDescription)

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress: scAddr,
		Vars: map[string]interface{}{
			"fn":         "incrementRepeatMany",
			"numRepeats": numRepeats,
		},
		// also send 5i to the SC address to use as request tokens
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 5,
		},
	})
	check(err, t)

	if !wasps.WaitUntilExpectationsMet() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-6, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 6,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 6, map[balance.Color]int64{
		balance.ColorIOTA: 5,
		*scColor:          1,
	}, "sc in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  scProgramHash[:],
		vmconst.VarNameDescription:  scDescription,
		"counter":                   util.Uint64To8Bytes(uint64(numRepeats + 1)),
		"numRepeats":                util.Uint64To8Bytes(0),
	}) {
		t.Fail()
	}
}
