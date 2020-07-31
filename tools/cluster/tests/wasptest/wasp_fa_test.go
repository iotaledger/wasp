package wasptest

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
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

	_, err = PutBootupRecords(wasps)
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
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
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
		"request_in":          4,
		"request_out":         5,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	_, err = PutBootupRecords(wasps)
	check(err, t)

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scNumFairAuction]

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(1 * time.Second)

	// create 1 colored token
	color1, err := mintNewColoredTokens(wasps, sc.OwnerSigScheme(), 1)
	check(err, t)

	scAddress := sc.SCAddress()
	ownerAddr := sc.OwnerAddress()

	// send 1i to the SC address. It is needed to send the request to self to start new auction ("operating capital")
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
		RequestCode: vmconst.RequestCodeNOP,
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 1,
		},
	})
	time.Sleep(1 * time.Second)

	if !wasps.VerifyAddressBalances(ownerAddr, iotasFromTheFaucet-3+1, map[balance.Color]int64{
		balance.ColorIOTA: iotasFromTheFaucet - 3,
		*color1:           1,
	}) {
		t.Fail()
		return
	}

	// send request StartAuction
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
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

	scColor := sc.GetColor()

	if !wasps.VerifyAddressBalances(scAddress, 2, map[balance.Color]int64{
		balance.ColorIOTA: 1,
		scColor:           1,
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddr, iotasFromTheFaucet-2-1+1, map[balance.Color]int64{
		balance.ColorIOTA: iotasFromTheFaucet - 2 - 1,
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
		"request_in":          6,
		"request_out":         7,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	_, err = PutBootupRecords(wasps)
	check(err, t)

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scNumFairAuction]

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(1 * time.Second)

	// create 1 colored token
	color1, err := mintNewColoredTokens(wasps, sc.OwnerSigScheme(), 1)
	check(err, t)
	time.Sleep(1 * time.Second)

	color2, err := mintNewColoredTokens(wasps, sc.OwnerSigScheme(), 1)
	check(err, t)
	time.Sleep(1 * time.Second)

	scAddress := sc.SCAddress()

	// send 2i to the SC address. It is needed to send the request to self to start 2 new auctions ("operating capital")
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
		RequestCode: vmconst.RequestCodeNOP,
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 2,
		},
	})
	time.Sleep(1 * time.Second)

	ownerAddr := sc.OwnerAddress()
	if !wasps.VerifyAddressBalances(ownerAddr, iotasFromTheFaucet-5+2, map[balance.Color]int64{
		balance.ColorIOTA: iotasFromTheFaucet - 5, // 1 to origin, 2 to new color, 2 for operating capital of SC
		*color1:           1,
		*color2:           1,
	}) {
		t.Fail()
		return
	}

	// send request StartAuction for color1
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
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
	time.Sleep(1 * time.Second)

	// send request StartAuction for color2
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParams{
		SCAddress:   &scAddress,
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

	if !wasps.VerifyAddressBalances(scAddress, 3, map[balance.Color]int64{
		balance.ColorIOTA: 2,
		sc.GetColor():     1,
	}) {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(ownerAddr, iotasFromTheFaucet-3, map[balance.Color]int64{
		balance.ColorIOTA: iotasFromTheFaucet - 5,
		*color1:           1,
		*color2:           1,
	}) {
		t.Fail()
	}
}

const (
	// TODO: adapt for testing in "real" goshimmer:
	// use hardcoded (or random) seed values and request funds
	auctionOwnerUtxodbIndex = 7
	bidderUtxodbIndex1      = 8
	bidderUtxodbIndex2      = 9
)

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

	scColors, err := PutBootupRecords(wasps)
	check(err, t)

	sc := &wasps.SmartContractConfig[scNumFairAuction]

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(1 * time.Second)

	auctionOwnerAddr := utxodb.GetAddress(auctionOwnerUtxodbIndex)
	auctionOwnerSigScheme := utxodb.GetSigScheme(auctionOwnerAddr)

	// create 1 colored token
	color1, err := mintNewColoredTokens(wasps, auctionOwnerSigScheme, 1)
	check(err, t)

	scOwnerAddr := sc.OwnerAddress()
	scAddr, err := address.FromBase58(sc.Address)
	check(err, t)

	scColor := *scColors[sc.Address]

	bidder1Addr := utxodb.GetAddress(bidderUtxodbIndex1)
	bidder1SigScheme := utxodb.GetSigScheme(bidder1Addr)

	if !wasps.VerifyAddressBalances(scAddr, 1+1, map[balance.Color]int64{
		scColor:           1,
		balance.ColorIOTA: 1,
	}) {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scOwnerAddr, 1000000000-2, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 2,
	}) {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, 1000000000-1+1, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 1,
		*color1:           1,
	}) {
		t.Fail()
	}

	// send request StartAuction. Selling 1 token of color1
	err = SendSimpleRequest(wasps, auctionOwnerSigScheme, waspapi.CreateSimpleRequestParams{
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

	// send 1 non wining bid PlaceBid on color1, sum 42
	err = SendSimpleRequest(wasps, bidder1SigScheme, waspapi.CreateSimpleRequestParams{
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
	// check SC address
	if !wasps.VerifyAddressBalances(scAddr, 3+1, map[balance.Color]int64{
		balance.ColorIOTA: 3,
		scColor:           1,
	}) {
		t.Fail()
	}
	// check SC owner address
	if !wasps.VerifyAddressBalances(scOwnerAddr, 1000000000-2+5, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 2 + 5,
	}) {
		t.Fail()
	}
	// check bidder1 address
	if !wasps.VerifyAddressBalances(bidder1Addr, 1000000000-1-42+42, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 1 - 42 + 42,
	}) {
		t.Fail()
	}
	// check auction owner address
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, 1000000000-1-1-5+1, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 1 - 1 - 5,
		*color1:           1,
	}) {
		t.Fail()
	}
}

func TestFA1Color1Bidder5WinningBids(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFairAuction5Requests5Sec1")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          8,
		"request_out":         9,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scColors, err := PutBootupRecords(wasps)
	check(err, t)

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scNumFairAuction]

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(1 * time.Second)

	auctionOwnerAddr := utxodb.GetAddress(auctionOwnerUtxodbIndex)
	auctionOwnerSigScheme := utxodb.GetSigScheme(auctionOwnerAddr)

	// create 1 colored token
	color1, err := mintNewColoredTokens(wasps, auctionOwnerSigScheme, 1)
	check(err, t)

	scOwnerAddr := sc.OwnerAddress()
	scAddr, err := address.FromBase58(sc.Address)
	check(err, t)

	scColor := *scColors[sc.Address]

	bidder1Addr := utxodb.GetAddress(bidderUtxodbIndex1)
	bidder1SigScheme := utxodb.GetSigScheme(bidder1Addr)

	if !wasps.VerifyAddressBalances(scAddr, 1+1, map[balance.Color]int64{
		scColor:           1,
		balance.ColorIOTA: 1,
	}) {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scOwnerAddr, 1000000000-2, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 2,
	}) {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, 1000000000-1, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 1,
		*color1:           1,
	}) {
		t.Fail()
	}

	// send request StartAuction. Selling 1 token of color1
	err = SendSimpleRequest(wasps, auctionOwnerSigScheme, waspapi.CreateSimpleRequestParams{
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

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, bidder1SigScheme, waspapi.CreateSimpleRequestParams{
			SCAddress:   &scAddr,
			RequestCode: fairauction.RequestPlaceBid,
			Vars: map[string]interface{}{
				fairauction.VarReqAuctionColor: color1,
			},
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: 25,
			},
		})
		check(err, t)
	}

	wasps.CollectMessages(70 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	// check SC address
	if !wasps.VerifyAddressBalances(scAddr, 7+1, map[balance.Color]int64{
		balance.ColorIOTA: 7,
		scColor:           1,
	}) {
		t.Fail()
	}
	// check SC owner address
	if !wasps.VerifyAddressBalances(scOwnerAddr, 1000000000-2+6, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 2 + 6,
	}) {
		t.Fail()
	}
	// check bidder1 address
	if !wasps.VerifyAddressBalances(bidder1Addr, 1000000000-5-5*25+1, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 5 - 5*25,
		*color1:           1,
	}) {
		t.Fail()
	}
	// check auction owner address
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, 1000000000-1-1-5+125-1, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 1 - 1 - 5 + 125 - 1,
	}) {
		t.Fail()
	}
}

func TestFA1Color2Bidders(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFairAuction5Requests5Sec1")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          13,
		"request_out":         14,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	scColors, err := PutBootupRecords(wasps)
	check(err, t)

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scNumFairAuction]

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(1 * time.Second)

	auctionOwnerAddr := utxodb.GetAddress(auctionOwnerUtxodbIndex)
	auctionOwnerSigScheme := utxodb.GetSigScheme(auctionOwnerAddr)

	// create 1 colored token
	color1, err := mintNewColoredTokens(wasps, auctionOwnerSigScheme, 1)
	check(err, t)

	scOwnerAddr := sc.OwnerAddress()
	scAddr, err := address.FromBase58(sc.Address)
	check(err, t)

	scColor := *scColors[sc.Address]

	bidder1Addr := utxodb.GetAddress(bidderUtxodbIndex1)
	bidder1SigScheme := utxodb.GetSigScheme(bidder1Addr)
	bidder2Addr := utxodb.GetAddress(bidderUtxodbIndex2)
	bidder2SigScheme := utxodb.GetSigScheme(bidder2Addr)

	if !wasps.VerifyAddressBalances(scAddr, 1+1, map[balance.Color]int64{
		scColor:           1,
		balance.ColorIOTA: 1,
	}) {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scOwnerAddr, 1000000000-2, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 2,
	}) {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, 1000000000-1+1, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 1,
		*color1:           1,
	}) {
		t.Fail()
	}

	// send request StartAuction. Selling 1 token of color1
	err = SendSimpleRequest(wasps, auctionOwnerSigScheme, waspapi.CreateSimpleRequestParams{
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

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, bidder1SigScheme, waspapi.CreateSimpleRequestParams{
			SCAddress:   &scAddr,
			RequestCode: fairauction.RequestPlaceBid,
			Vars: map[string]interface{}{
				fairauction.VarReqAuctionColor: color1,
			},
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: 25,
			},
		})
		check(err, t)

		err = SendSimpleRequest(wasps, bidder2SigScheme, waspapi.CreateSimpleRequestParams{
			SCAddress:   &scAddr,
			RequestCode: fairauction.RequestPlaceBid,
			Vars: map[string]interface{}{
				fairauction.VarReqAuctionColor: color1,
			},
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: 25,
			},
		})
		check(err, t)

	}

	wasps.CollectMessages(70 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}
	// check SC address
	if !wasps.VerifyAddressBalances(scAddr, 1+1, map[balance.Color]int64{
		balance.ColorIOTA: 12,
		scColor:           1,
	}) {
		t.Fail()
	}
	// check SC owner address
	if !wasps.VerifyAddressBalances(scOwnerAddr, 1000000000-2+6, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 2 + 6,
	}) {
		t.Fail()
	}
	// check bidder1 address (winner)
	if !wasps.VerifyAddressBalances(bidder1Addr, 1000000000-5-5*25+1, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 5 - 5*25,
		*color1:           1,
	}) {
		t.Fail()
	}
	// check bidder2 address (loser)
	if !wasps.VerifyAddressBalances(bidder2Addr, 1000000000-5-5*25+5*25, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 5 - 5*25 + 5*25,
	}) {
		t.Fail()
	}
	// check auction owner address
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, 1000000000-1-1-5+125-1, map[balance.Color]int64{
		balance.ColorIOTA: 1000000000 - 1 - 1 - 5 + 125 - 1,
	}) {
		t.Fail()
	}
}
