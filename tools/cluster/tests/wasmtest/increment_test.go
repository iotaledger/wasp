package wasmtest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
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

const inc_wasmPath = "increment_bg.wasm"
const inc_description = "Increment, a PoC smart contract"

func TestDeployment(t *testing.T) {
	preamble(t, inc_wasmPath, inc_description, "TestDeployment")
	startSmartContract(t, scProgramHash, inc_description)

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
		vmconst.VarNameDescription:  inc_description,
	}) {
		t.Fail()
	}
}

func TestIncNothing(t *testing.T) {
	testNothing(t, "TestIncNothing", inc_wasmPath, inc_description, 1)
}

func Test5xIncNothing(t *testing.T) {
	testNothing(t, "Test5xIncNothing", inc_wasmPath, inc_description, 5)
}

func testNothing(t *testing.T, testName string, wasmPath string, descr string, numRequests int) {
	preamble(t, wasmPath, descr, testName)

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

	startSmartContract(t, scProgramHash, inc_description)

	for i := 0; i < numRequests; i++ {
		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress: scAddr,
			RequestCode: inc_code_nothing,
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
		vmconst.VarNameDescription:  inc_description,
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
	preamble(t, inc_wasmPath, inc_description, testName)

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

	startSmartContract(t, scProgramHash, inc_description)

	for i := 0; i < increments; i++ {
		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress: scAddr,
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
		vmconst.VarNameProgramHash:  scProgramHash[:],
		vmconst.VarNameDescription:  inc_description,
		"counter":                   util.Uint64To8Bytes(uint64(increments)),
	}) {
		t.Fail()
	}
}

func TestRepeatIncrement(t *testing.T) {
	preamble(t, inc_wasmPath, inc_description, "TestRepeatIncrement")

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

	startSmartContract(t, scProgramHash, inc_description)

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress: scAddr,
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
		vmconst.VarNameProgramHash:  scProgramHash[:],
		vmconst.VarNameDescription:  inc_description,
		"counter":                   util.Uint64To8Bytes(uint64(2)),
	}) {
		t.Fail()
	}
}

func TestRepeatManyIncrement(t *testing.T) {
	const numRepeats = 5
	preamble(t, inc_wasmPath, inc_description, "TestRepeatManyIncrement")

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

	startSmartContract(t, scProgramHash, inc_description)

	err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress: scAddr,
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
		vmconst.VarNameProgramHash:  scProgramHash[:],
		vmconst.VarNameDescription:  inc_description,
		"counter":                   util.Uint64To8Bytes(uint64(numRepeats + 1)),
		"numRepeats":                util.Uint64To8Bytes(0),
	}) {
		t.Fail()
	}
}
