package wasptest

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const scNumFairAuction = 5

func TestFairAuctionSetOwnerMargin(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFairAuction5Requests5Sec1")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          -1,
		"request_out":         -1,
		"state":               -1, // must be 6 or 7
		"vmmsg":               -1,
	})
	check(err, t)

	err = PutBootupRecords(wasps)
	check(err, t)

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scNumFairAuction]

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	// send request SetOwnerMargin
	reqs := []*waspapi.RequestBlockJson{
		{Address: sc.Address,
			RequestCode: fairauction.RequestSetOwnerMargin,
			Vars: map[string]interface{}{
				fairauction.VarReqOwnerMargin: 100,
			},
		},
	}
	err = SendRequests(wasps, sc.OwnerIndexUtxodb, reqs)
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestFairAuction1Color(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFairAuction5Requests5Sec1")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          -1,
		"request_out":         -1,
		"state":               -1, // must be 6 or 7
		"vmmsg":               -1,
	})
	check(err, t)

	err = PutBootupRecords(wasps)
	check(err, t)

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scNumFairAuction]

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(1 * time.Second)

	// create 1 colored token
	color1, err := mintNewColoredTokens(wasps, sc.OwnerIndexUtxodb, 1)
	check(err, t)
	// owner margin = 3%
	ownerAddr := utxodb.GetAddress(sc.OwnerIndexUtxodb)
	if !wasps.VerifyAddressBalances(ownerAddr, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 3,
		*color1:           1,
	}) {
		t.Fail()
		return
	}

	// create auction
	req := &waspapi.RequestBlockJson{
		Address:     sc.Address,
		RequestCode: fairauction.RequestStartAuction,
		AmountIotas: 3 + 1, // TODO add colored token
		Vars: map[string]interface{}{
			fairauction.VarReqAuctionColor:                string(color1.Bytes()),
			fairauction.VarReqStartAuctionMinimumBid:      100,
			fairauction.VarReqStartAuctionDurationMinutes: 1,
		},
	}
	err = SendRequests(wasps, sc.OwnerIndexUtxodb, []*waspapi.RequestBlockJson{req})
	check(err, t)

	wasps.CollectMessages(70 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	scAddr, err := address.FromBase58(sc.Address)
	assert.NoError(t, err)

	if !wasps.VerifyAddressBalances(scAddr, map[balance.Color]int64{
		balance.ColorIOTA: 7,
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddr, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 7,
		*color1:           1,
	}) {
		t.Fail()
	}
}

func mintNewColoredTokens(wasps *cluster.Cluster, ownerIdx int, amount int64) (*balance.Color, error) {
	ownerAddr := utxodb.GetAddress(ownerIdx)
	ownerSigScheme := utxodb.GetSigScheme(ownerAddr)
	tx, err := waspapi.NewColoredTokensTransaction(wasps.Config.GoshimmerApiHost(), ownerSigScheme, amount)
	if err != nil {
		return nil, err
	}
	err = wasps.PostAndWaitForConfirmation(tx)
	if err != nil {
		return nil, err
	}
	ret := (balance.Color)(tx.ID())

	fmt.Printf("[cluster] minted %d new tokens of color %s\n", amount, ret.String())
	//fmt.Printf("Value ts: %s\n", tx.String())
	return &ret, nil
}
