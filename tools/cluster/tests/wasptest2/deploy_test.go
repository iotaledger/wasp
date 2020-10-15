package wasptest2

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/tools/cluster/tests/wasptest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeploySC(t *testing.T) {
	wasps := setup(t, "TestDeploySC")

	err := requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, tokenregistry.ProgramHash, tokenregistry.Description)
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

	ph, err := hashing.HashValueFromBase58(tokenregistry.ProgramHash)
	check(err, t)

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  ph[:],
		vmconst.VarNameDescription:  tokenregistry.Description,
	}) {
		t.Fail()
	}
}

func TestGetSCData(t *testing.T) {
	wasps := setup(t, "TestDeploySC")

	err := requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, tokenregistry.ProgramHash, tokenregistry.Description)
	checkSuccess(err, t, "smart contract has been created and activated")

	bd, err := wasps.Config.Nodes[0].Client().GetBootupData(scAddr)
	assert.NoError(t, err)
	assert.NotNil(t, bd)
	assert.EqualValues(t, bd.OwnerAddress, *scOwnerAddr)
	assert.True(t, bytes.Equal(bd.Color[:], scColor[:]))

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

	ph, err := hashing.HashValueFromBase58(tokenregistry.ProgramHash)
	check(err, t)

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  ph[:],
		vmconst.VarNameDescription:  tokenregistry.Description,
	}) {
		t.Fail()
	}
}

const numRequests = 5

func TestSend5ReqInc0SecDeploy(t *testing.T) {
	wasps := setup(t, "TestSend5ReqInc0SecDeploy")

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

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHash, inccounter.Description)
	checkSuccess(err, t, "smart contract has been created and activated")

	for i := 0; i < numRequests; i++ {
		err = wasptest.SendSimpleRequest(wasps, scOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddr,
			RequestCode: inccounter.RequestInc,
		})
		check(err, t)
	}

	wasps.WaitUntilExpectationsMet()
	//wasps.CollectMessages(20 * time.Second)
	//if !wasps.Report() {
	//	t.Fail()
	//}

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
		vmconst.VarNameDescription:  inccounter.Description,
	}) {
		t.Fail()
	}
}

const numRequestsInTheBlock = 100

func TestSend100ReqMulti(t *testing.T) {
	wasps := setup(t, "TestSend5ReqInc0SecDeploy")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1 + numRequestsInTheBlock,
		"request_out":         2 + numRequestsInTheBlock,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	err = requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHash, inccounter.Description)
	checkSuccess(err, t, "smart contract has been created and activated")

	pars := make([]waspapi.CreateSimpleRequestParamsOld, numRequestsInTheBlock)
	for i := 0; i < numRequestsInTheBlock; i++ {
		pars[i] = waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddr,
			RequestCode: inccounter.RequestInc,
		}
	}
	err = wasptest.SendSimpleRequestMulti(wasps, scOwner.SigScheme(), pars)
	check(err, t)

	wasps.WaitUntilExpectationsMet()
	//wasps.CollectMessages(20 * time.Second)
	//if !wasps.Report() {
	//	t.Fail()
	//}

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
		vmconst.VarNameDescription:  inccounter.Description,
	}) {
		t.Fail()
	}
}
