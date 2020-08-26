package wasptest2

import (
	"crypto/rand"
	"os"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry/trclient"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/mr-tron/base58"
)

func TestTRTest(t *testing.T) {
	var seed [32]byte
	rand.Read(seed[:])
	seed58 := base58.Encode(seed[:])
	wallet := testutil.NewWallet(seed58)
	scOwner := wallet.WithIndex(0)
	minter := wallet.WithIndex(1)

	// setup
	wasps := setup(t, "test_cluster2", "TestTRTest")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           1,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          2,
		"request_out":         3,
		"state":               -1, // must be 6 or 7
		"vmmsg":               -1,
	})
	check(err, t)

	programHash, err := hashing.HashValueFromBase58(tokenregistry.ProgramHash)
	check(err, t)

	scOwnerAddr := scOwner.Address()
	err = wasps.NodeClient.RequestFunds(&scOwnerAddr)
	check(err, t)

	minterAddr := minter.Address()
	err = wasps.NodeClient.RequestFunds(&minterAddr)
	check(err, t)

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "sc owner in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(minterAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "minter in the beginning") {
		t.Fail()
		return
	}

	scAddr, scColor, err := waspapi.CreateAndDeploySC(waspapi.CreateAndDeploySCParams{
		Node:                    wasps.NodeClient,
		CommitteeApiHosts:       wasps.ApiHosts(),
		CommitteePeeringHosts:   wasps.PeeringHosts(),
		N:                       4,
		T:                       3,
		OwnerSigScheme:          scOwner.SigScheme(),
		ProgramHash:             programHash,
		Textout:                 os.Stdout,
		Prefix:                  "[deploy " + tokenregistry.ProgramHash + "]",
		WaitForInitialization:   true,
		CommitteePublisherHosts: wasps.PublisherHosts(),
		Timeout:                 20 * time.Second,
	})
	check(err, t)

	if !wasps.VerifyAddressBalances(*scAddr, 1, map[balance.Color]int64{
		*scColor: 1, // sc token
	}, "SC address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "owner in the beginning") {
		t.Fail()
		return
	}

	tc := trclient.NewClient(wasps.NodeClient, wasps.Config.Nodes[0].ApiHost(), scAddr, minter.SigScheme())

	tx1, err := tc.MintAndRegister(trclient.MintAndRegisterParams{
		Supply:      1,
		MintTarget:  minterAddr,
		Description: "Non-fungible coin 1",
	})
	check(err, t)

	mintedColor1 := balance.Color(tx1.ID())

	//wasps.CollectMessages(30 * time.Second)
	wasps.WaitUntilExpectationsMet()

	if !wasps.Report() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(*scAddr, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		*scColor:          1,
	}, "SC address in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(minterAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		mintedColor1:      1,
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "minter1 in the end") {
		t.Fail()
		return
	}
	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  programHash[:],
	}) {
		t.Fail()
	}

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress:      scOwnerAddr[:],
		vmconst.VarNameProgramHash:       programHash[:],
		tokenregistry.VarStateListColors: []byte(mintedColor1.String()),
	}) {
		t.Fail()
	}
}
