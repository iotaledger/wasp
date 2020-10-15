package wasptest2

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"testing"
	"time"
)

const noRequests = 3

func TestKillNode(t *testing.T) {
	wasps := setup(t, "TestKillNode")

	err := requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	scAddr, scColor, err := startSmartContract(wasps, inccounter.ProgramHash, inccounter.Description)
	checkSuccess(err, t, "smart contract has been created and activated")

	wasps.StopNode(3)

	for i := 0; i < noRequests; i++ {
		_, err = waspapi.CreateRequestTransaction(waspapi.CreateRequestTransactionParams{
			NodeClient:      wasps.NodeClient,
			SenderSigScheme: scOwner.SigScheme(),
			BlockParams: []waspapi.RequestBlockParams{
				{
					TargetSCAddress: scAddr,
					RequestCode:     inccounter.RequestInc,
				},
			},
			Post:                true,
			WaitForConfirmation: true,
			WaitForCompletion:   true,
			PublisherHosts:      wasps.PublisherHosts(),
			PublisherQuorum:     3,
			Timeout:             30 * time.Second,
		})
		checkSuccess(err, t, fmt.Sprintf("request #%d has been sent and completed", i))
	}

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

	ph, err := hashing.HashValueFromBase58(inccounter.ProgramHash)
	check(err, t)

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  ph[:],
		inccounter.VarCounter:       noRequests,
	}) {
		t.Fail()
	}
}
