// +build ignore

package wasptest

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
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

	var err error

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scTokenRegistryNum]

	_, err = PutChainRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(2 * time.Second)

	scOwnerAddr := sc.OwnerAddress()
	scAddress := sc.SCAddress()
	scColor := sc.GetColor()
	minterAddr := minter1.Address()
	progHash, err := hashing.HashValueFromBase58(tokenregistry.ProgramHash)
	check(err, t)

	err = wasps.Level1Client.RequestFunds(minterAddr)
	check(err, t)

	time.Sleep(2 * time.Second)

	if !wasps.VerifyAddressBalances(minterAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
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

	tc := trclient.NewClient(wasps.SCClient(sc, minter1.SigScheme()))

	tx, err := tc.MintAndRegister(trclient.MintAndRegisterParams{
		Supply:      1,
		MintTarget:  *minterAddr,
		Description: "Non-fungible coin 1",
	})
	check(err, t)

	mintedColor := balance.Color(tx.ID())

	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		balance.ColorIOTA: 0,
		sc.GetColor():     1,
	}, "SC address in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(minterAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		mintedColor:       1,
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "minter1 in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifySCStateVariables(sc, map[kv.Key][]byte{
		vmconst.VarNameOwnerAddress:      scOwnerAddr.Bytes(),
		vmconst.VarNameProgramData:       progHash.Bytes(),
		tokenregistry.VarStateListColors: []byte(mintedColor.String()),
	}) {
		t.Fail()
	}
}
