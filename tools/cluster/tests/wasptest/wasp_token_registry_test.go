package wasptest

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry/trclient"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
)

const scTokenRegistryNum = 6

// sending 5 NOP requests with 1 sec sleep between each
func TestTRMint1Token(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSC6Requests5Sec1")

	//err := wasps.ListenToMessages(map[string]int{
	//	"bootuprec":           1, // wasps.NumSmartContracts(),
	//	"active_committee":    1,
	//	"dismissed_committee": 0,
	//	"request_in":          2,
	//	"request_out":         3,
	//	"state":               -1, // must be 6 or 7
	//	"vmmsg":               -1,
	//})
	//check(err, t)
	var err error

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scTokenRegistryNum]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(2 * time.Second)

	scOwnerAddr := sc.OwnerAddress()
	scAddress := sc.SCAddress()
	scColor := sc.GetColor()
	minter1Addr := minter1.Address()
	progHash, err := hashing.HashValueFromBase58(tokenregistry.ProgramHash)
	check(err, t)

	err = wasps.NodeClient.RequestFunds(&minter1Addr)
	check(err, t)

	time.Sleep(2 * time.Second)

	if !wasps.VerifyAddressBalances(minter1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "minter1 in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		scColor: 1, // sc token
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

	tc := trclient.NewClient(wasps.NodeClient, wasps.Config.Nodes[0].ApiHost(), &scAddress, minter1.SigScheme())

	tx1, err := waspapi.RunAndWaitForRequestProcessedMulti(wasps.PublisherHosts(), &scAddress, 0, 15*time.Second, func() (*sctransaction.Transaction, error) {
		return tc.MintAndRegister(trclient.MintAndRegisterParams{
			Supply:      1,
			MintTarget:  minter1Addr,
			Description: "Non-fungible coin 1",
		})
	})
	check(err, t)
	mintedColor1 := balance.Color(tx1.ID())

	//wasps.CollectMessages(30 * time.Second)
	wasps.WaitUntilExpectationsMet()

	if !wasps.Report() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		sc.GetColor():     1,
	}, "SC address in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(minter1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		mintedColor1:      1,
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "minter1 in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables(sc, map[kv.Key][]byte{
		vmconst.VarNameOwnerAddress:      scOwnerAddr.Bytes(),
		vmconst.VarNameProgramHash:       progHash.Bytes(),
		tokenregistry.VarStateListColors: []byte(mintedColor1.String()),
	}) {
		t.Fail()
	}
}
