// +build ignore

package wasmtest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/tools/cluster/tests/wasptest"
	"testing"
)

const (
	incCodeIncrement           = coretypes.Hname(1)
	incCodeIncrementRepeat1    = coretypes.Hname(2)
	incCodeIncrementRepeatMany = coretypes.Hname(3)
	incCodeTest                = coretypes.Hname(4)
	incCodeNothing             = coretypes.Hname(5)
)

const incWasmPath = "wasm/increment"
const incDescription = "Increment, a PoC smart contract"

func TestIncDeployment(t *testing.T) {
	wasps := setup(t, "TestIncDeployment")

	err := loadWasmIntoWasps(wasps, incWasmPath, incDescription)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	scChain, scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHashStr, incDescription)
	checkSuccess(err, t, "smart contract has been created and activated")
	_ = scChain

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
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  incDescription,
	}) {
		t.Fail()
	}
}

func TestIncNothing(t *testing.T) {
	testNothing(t, "TestIncNothing", inccounter.ProgramHashStr, incWasmPath, incDescription, 1)
}

func TestInc5xNothing(t *testing.T) {
	testNothing(t, "TestInc5xNothing", inccounter.ProgramHashStr, incWasmPath, incDescription, 5)
}

func testNothing(t *testing.T, testName string, hash string, wasmPath string, description string, numRequests int) {
	wasps := setup(t, testName)

	err := loadWasmIntoWasps(wasps, wasmPath, description)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.StartMessageCounter(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + numRequests,
		"request_out":         2 + numRequests,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scChain, scAddr, scColor, err := startSmartContract(wasps, hash, description)
	checkSuccess(err, t, "smart contract has been created and activated")

	for i := 0; i < numRequests; i++ {
		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			TargetContract: coretypes.NewContractID(*scChain, 0),
			RequestCode:    incCodeNothing,
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
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  description,
	}) {
		t.Fail()
	}
}

func TestIncIncrement(t *testing.T) {
	testIncrement(t, "TestIncIncrement", 1)
}

func TestInc5xIncrement(t *testing.T) {
	testIncrement(t, "TestInc5xIncrement", 5)
}

func testIncrement(t *testing.T, testName string, increments int) {
	wasps := setup(t, testName)

	err := loadWasmIntoWasps(wasps, incWasmPath, incDescription)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.StartMessageCounter(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + increments,
		"request_out":         2 + increments,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scChain, scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHashStr, incDescription)
	checkSuccess(err, t, "smart contract has been created and activated")

	for i := 0; i < increments; i++ {
		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			TargetContract: coretypes.NewContractID(*scChain, 0),
			RequestCode:    incCodeIncrement,
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
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  incDescription,
		"counter":                   util.Uint64To8Bytes(uint64(increments)),
	}) {
		t.Fail()
	}
}

func TestIncRepeatIncrement(t *testing.T) {
	wasps := setup(t, "TestIncRepeatIncrement")

	err := loadWasmIntoWasps(wasps, incWasmPath, incDescription)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.StartMessageCounter(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 2,
		"request_out":         2 + 2,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scChain, scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHashStr, incDescription)
	checkSuccess(err, t, "smart contract has been created and activated")

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		TargetContract: coretypes.NewContractID(*scChain, 0),
		RequestCode:    incCodeIncrementRepeat1,
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
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  incDescription,
		"counter":                   util.Uint64To8Bytes(uint64(2)),
	}) {
		t.Fail()
	}
}

func TestIncRepeatManyIncrement(t *testing.T) {
	const numRepeats = 5

	wasps := setup(t, "TestIncRepeatManyIncrement")

	err := loadWasmIntoWasps(wasps, incWasmPath, incDescription)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.StartMessageCounter(map[string]int{
		"chainrec":            2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 1 + numRepeats,
		"request_out":         2 + 1 + numRepeats,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scChain, scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHashStr, incDescription)
	checkSuccess(err, t, "smart contract has been created and activated")

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		TargetContract: coretypes.NewContractID(*scChain, 0),
		RequestCode:    incCodeIncrementRepeatMany,
		Vars: map[string]interface{}{
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
		vmconst.VarNameProgramData:  programHash[:],
		vmconst.VarNameDescription:  incDescription,
		"counter":                   util.Uint64To8Bytes(uint64(numRepeats + 1)),
		"numRepeats":                util.Uint64To8Bytes(0),
	}) {
		t.Fail()
	}
}
