package wasmtest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/tools/cluster/tests/wasptest"
	"testing"
)

const (
	inc_code_nothing             = sctransaction.RequestCode(1)
	inc_code_test                = sctransaction.RequestCode(2)
	inc_code_increment           = sctransaction.RequestCode(3)
	inc_code_incrementRepeat1    = sctransaction.RequestCode(4)
	inc_code_incrementRepeatMany = sctransaction.RequestCode(5)
)

const inc_wasmPath = "wasm/increment_bg.wasm"
const inc_description = "Increment, a PoC smart contract"

func TestDeployment(t *testing.T) {
	wasps := setup(t, "TestDeployment")

	err := loadWasmIntoWasps(wasps, inc_wasmPath, inc_description)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHash, inc_description)
	checkSuccess(err, t, "smart contract has been created and activated")

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
		vmconst.VarNameProgramHash:  programHash[:],
		vmconst.VarNameDescription:  inc_description,
	}) {
		t.Fail()
	}
}

func TestIncNothing(t *testing.T) {
	testNothing(t, "TestIncNothing", inccounter.ProgramHash, inc_wasmPath, inc_description, 1)
}

func Test5xIncNothing(t *testing.T) {
	testNothing(t, "Test5xIncNothing", inccounter.ProgramHash, inc_wasmPath, inc_description, 5)
}

func testNothing(t *testing.T, testName string, hash string, wasmPath string, description string, numRequests int) {
	wasps := setup(t, testName)

	err := loadWasmIntoWasps(wasps, wasmPath, description)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + numRequests,
		"request_out":         2 + numRequests,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, hash, description)
	checkSuccess(err, t, "smart contract has been created and activated")

	for i := 0; i < numRequests; i++ {
		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddr,
			RequestCode: 1,
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
		vmconst.VarNameProgramHash:  programHash[:],
		vmconst.VarNameDescription:  description,
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
	wasps := setup(t, testName)

	err := loadWasmIntoWasps(wasps, inc_wasmPath, inc_description)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + increments,
		"request_out":         2 + increments,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHash, inc_description)
	checkSuccess(err, t, "smart contract has been created and activated")

	for i := 0; i < increments; i++ {
		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddr,
			RequestCode: inc_code_increment,
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
		vmconst.VarNameProgramHash:  programHash[:],
		vmconst.VarNameDescription:  inc_description,
		"counter":                   util.Uint64To8Bytes(uint64(increments)),
	}) {
		t.Fail()
	}
}

func TestRepeatIncrement(t *testing.T) {
	wasps := setup(t, "TestRepeatIncrement")

	err := loadWasmIntoWasps(wasps, inc_wasmPath, inc_description)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 2,
		"request_out":         2 + 2,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHash, inc_description)
	checkSuccess(err, t, "smart contract has been created and activated")

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddr,
		RequestCode: inc_code_incrementRepeat1,
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
		vmconst.VarNameProgramHash:  programHash[:],
		vmconst.VarNameDescription:  inc_description,
		"counter":                   util.Uint64To8Bytes(uint64(2)),
	}) {
		t.Fail()
	}
}

func TestRepeatManyIncrement(t *testing.T) {
	const numRepeats = 5

	wasps := setup(t, "TestRepeatManyIncrement")

	err := loadWasmIntoWasps(wasps, inc_wasmPath, inc_description)
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	err = wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + 1 + numRepeats,
		"request_out":         2 + 1 + numRepeats,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHash, inc_description)
	checkSuccess(err, t, "smart contract has been created and activated")

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddr,
		RequestCode: inc_code_incrementRepeatMany,
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
		vmconst.VarNameProgramHash:  programHash[:],
		vmconst.VarNameDescription:  inc_description,
		"counter":                   util.Uint64To8Bytes(uint64(numRepeats + 1)),
		"numRepeats":                util.Uint64To8Bytes(0),
	}) {
		t.Fail()
	}
}
