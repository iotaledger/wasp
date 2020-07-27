package wasptest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"testing"
	"time"
)

const scNumFairAuction = 5

func TestFASetOwnerMargin(t *testing.T) {
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

	scAddr, err := address.FromBase58(sc.Address)
	check(err, t)

	// send request SetOwnerMargin
	err = SendSimpleRequest(wasps, sc.OwnerIndexUtxodb, waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddr,
		RequestCode: fairauction.RequestSetOwnerMargin,
		Vars: map[string]interface{}{
			fairauction.VarReqOwnerMargin: 100,
		},
	})
	check(err, t)

	wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
}

func TestFA1Color0Bids(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFairAuction5Requests5Sec1")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          3,
		"request_out":         4,
		"state":               -1,
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

	ownerAddr := utxodb.GetAddress(sc.OwnerIndexUtxodb)
	if !wasps.VerifyAddressBalances(ownerAddr, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 3,
		*color1:           1,
	}) {
		t.Fail()
		return
	}

	scAddr, err := address.FromBase58(sc.Address)
	check(err, t)

	// send request StartAuction
	err = SendSimpleRequest(wasps, sc.OwnerIndexUtxodb, waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddr,
		RequestCode: fairauction.RequestStartAuction,
		Vars: map[string]interface{}{
			fairauction.VarReqAuctionColor:                color1,
			fairauction.VarReqStartAuctionMinimumBid:      100,
			fairauction.VarReqStartAuctionDurationMinutes: 1,
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 5, // 5% from 100
			*color1:           1, // token for sale
		},
	})
	check(err, t)

	wasps.CollectMessages(70 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddr, map[balance.Color]int64{
		balance.ColorIOTA: 7,
		// +1 SC token
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddr, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 8 - 1,
		*color1:           1,
	}) {
		t.Fail()
	}
}

func TestFA2Color0Bids(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFairAuction5Requests5Sec1")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          5,
		"request_out":         6,
		"state":               -1,
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
	color2, err := mintNewColoredTokens(wasps, sc.OwnerIndexUtxodb, 1)
	check(err, t)

	ownerAddr := utxodb.GetAddress(sc.OwnerIndexUtxodb)
	if !wasps.VerifyAddressBalances(ownerAddr, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 4, // 2 to SC, 2 to new color
		*color1:           1,
		*color2:           1,
	}) {
		t.Fail()
		return
	}

	scAddr, err := address.FromBase58(sc.Address)
	check(err, t)

	// send request StartAuction for color1
	err = SendSimpleRequest(wasps, sc.OwnerIndexUtxodb, waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddr,
		RequestCode: fairauction.RequestStartAuction,
		Vars: map[string]interface{}{
			fairauction.VarReqAuctionColor:                color1,
			fairauction.VarReqStartAuctionMinimumBid:      100,
			fairauction.VarReqStartAuctionDurationMinutes: 1,
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 5, // 5% from 100
			*color1:           1, // token for sale
		},
	})
	check(err, t)

	time.Sleep(2 * time.Second)

	// send request StartAuction for color2
	err = SendSimpleRequest(wasps, sc.OwnerIndexUtxodb, waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddr,
		RequestCode: fairauction.RequestStartAuction,
		Vars: map[string]interface{}{
			fairauction.VarReqAuctionColor:                color2,
			fairauction.VarReqStartAuctionMinimumBid:      100,
			fairauction.VarReqStartAuctionDurationMinutes: 1,
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 5, // 5% from 100
			*color2:           1, // token for sale
		},
	})
	check(err, t)

	wasps.CollectMessages(70 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddr, map[balance.Color]int64{
		balance.ColorIOTA: 13,
		// +1 SC token
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddr, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 14 - 2,
		*color1:           1,
		*color2:           1,
	}) {
		t.Fail()
	}
}

const bidderUtxodbIndex = 7

func TestFA1Color1NonWinningBid(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFairAuction5Requests5Sec1")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          4,
		"request_out":         5,
		"state":               -1,
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

	ownerAddr := utxodb.GetAddress(sc.OwnerIndexUtxodb)
	scAddr, err := address.FromBase58(sc.Address)
	bidderAddr := utxodb.GetAddress(bidderUtxodbIndex)
	check(err, t)

	if !wasps.VerifyAddressBalances(ownerAddr, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 3,
		*color1:           1,
	}) {
		t.Fail()
		return
	}

	// send request StartAuction
	err = SendSimpleRequest(wasps, sc.OwnerIndexUtxodb, waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddr,
		RequestCode: fairauction.RequestStartAuction,
		Vars: map[string]interface{}{
			fairauction.VarReqAuctionColor:                color1,
			fairauction.VarReqStartAuctionMinimumBid:      100,
			fairauction.VarReqStartAuctionDurationMinutes: 1,
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 5, // 5% from 100
			*color1:           1, // token for sale
		},
	})
	check(err, t)

	// send 1 non wining bid PlaceBid
	err = SendSimpleRequest(wasps, bidderUtxodbIndex, waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddr,
		RequestCode: fairauction.RequestPlaceBid,
		Vars: map[string]interface{}{
			fairauction.VarReqAuctionColor: color1,
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 42,
		},
	})
	check(err, t)

	wasps.CollectMessages(70 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddr, map[balance.Color]int64{
		balance.ColorIOTA: 3,
		// +1 SC token
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddr, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 3 - 1, // all fees return to auction owner == smart contract owner
		*color1:           1,
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(bidderAddr, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 1,
		//*color1:           0,
	}) {
		t.Fail()
	}
}
