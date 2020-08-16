package wasptest2

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/testutil"
	_ "github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"os"
	"testing"
)

func TestDeploySC(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster2", "TestDeploySC")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           1,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          1,
		"request_out":         2,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	programHash, err := hashing.HashValueFromBase58(tokenregistry.ProgramHash)
	check(err, t)

	scOwnerAddr := scOwner.Address()
	err = wasps.NodeClient.RequestFunds(&scOwnerAddr)
	check(err, t)

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "sc owner in the beginning") {
		t.Fail()
		return
	}

	t.Logf("peering hosts: %+v", wasps.PeeringHosts())
	scAddr, scColor, err := apilib.CreateAndDeploySC(apilib.CreateAndDeploySCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHash,
		Textout:               os.Stdout,
		Prefix:                "[deploy] ",
	})
	check(err, t)
	//wasps.CollectMessages(20 * time.Second)
	wasps.WaitUntilExpectationsMet()
	if !wasps.Report() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(*scAddr, 1, map[balance.Color]int64{
		*scColor: 1,
	}, "sc in the end") {
		t.Fail()
		return
	}
}
