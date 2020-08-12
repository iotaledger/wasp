package wasptest

import (
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
)

const scTokenRegistryNum = 6

// sending 5 NOP requests with 1 sec sleep between each
func TestTRRequests5Sec1(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestSC6Requests5Sec1")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           1, // wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          6,
		"request_out":         7,
		"state":               -1, // must be 6 or 7
		"vmmsg":               -1,
	})
	check(err, t)

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

	err = wasps.NodeClient.RequestFunds(&minter1Addr)
	check(err, t)

	// create 1 colored token
	color1, err := mintNewColoredTokens(wasps, minter1.SigScheme(), 42)
	check(err, t)

	time.Sleep(2 * time.Second)

	if !wasps.VerifyAddressBalances(minter1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		*color1:           42,
		balance.ColorIOTA: testutil.RequestFundsAmount - 42,
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

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, minter1.SigScheme(), waspapi.CreateSimpleRequestParams{
			SCAddress:   &scAddress,
			RequestCode: tokenregistry.RequestInitSC,
		})
		check(err, t)
		time.Sleep(1 * time.Second)
	}

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(minter1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		*color1:           42,
		balance.ColorIOTA: testutil.RequestFundsAmount - 42,
	}, "minter1 in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		vmconst.VarNameOwnerAddress: sc.GetColor().Bytes(),
		vmconst.VarNameProgramHash:  []byte(tokenregistry.ProgramHash),
	}) {
		t.Fail()
	}
}
