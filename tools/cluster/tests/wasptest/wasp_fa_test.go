package wasptest

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
)

const scNumFairAuction = 5

func TestFASetOwnerMargin(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFASetOwnerMargin")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          -1,
		"request_out":         -1,
		"state":               -1, // must be 6 or 7
		"vmmsg":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[scNumFairAuction]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	ownerAddr := sc.OwnerAddress()
	scAddr := sc.SCAddress()
	scColor := sc.GetColor()

	auctionOwnerAddr := auctionOwner.Address()

	err = wasps.NodeClient.RequestFunds(auctionOwnerAddr)
	check(err, t)

	time.Sleep(2 * time.Second)

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "owner address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scAddr, 1, map[balance.Color]int64{
		scColor: 1,
	}, "SC address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "auction owner address in the beginning") {
		t.Fail()
		return
	}

	// send request SetOwnerMargin
	err = SendSimpleRequest(wasps, sc.OwnerSigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddr,
		RequestCode: fairauction.RequestSetOwnerMargin,
		Vars: map[string]interface{}{
			fairauction.VarReqOwnerMargin: 100, // 10%
		},
	})
	check(err, t)

	wasps.WaitUntilExpectationsMet()
	//wasps.CollectMessages(15 * time.Second)

	if !wasps.Report() {
		t.Fail()
	}

	time.Sleep(1 * time.Second)

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "owner address in the end") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scAddr, 1, map[balance.Color]int64{
		scColor: 1,
	}, "SC address in the end") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "auction address in the end") {
		t.Fail()
		return
	}
	if !wasps.VerifySCState(sc, 0, map[kv.Key][]byte{
		fairauction.VarStateOwnerMarginPromille: util.Uint64To8Bytes(uint64(100)),
	}) {
		t.Fail()
		return
	}
}

// creating auction which expires without bids
func TestFA1Color0Bids(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFA1Color0Bids")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          3,
		"request_out":         4,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scNumFairAuction]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(2 * time.Second)

	ownerAddr := sc.OwnerAddress()
	scAddress := sc.SCAddress()
	scColor := sc.GetColor()

	auctionOwnerAddr := auctionOwner.Address()

	err = wasps.NodeClient.RequestFunds(auctionOwnerAddr)
	check(err, t)

	// create 1 colored token
	color1, err := mintNewColoredTokens(wasps, auctionOwner.SigScheme(), 1)
	check(err, t)

	time.Sleep(5 * time.Second)

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "owner address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		scColor: 1,
	}, "SC address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
		*color1:           1,
	}, "auction owner address in the beginning") {
		t.Fail()
		return
	}

	// send request StartAuction from the auction owner
	err = SendSimpleRequest(wasps, auctionOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
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

	wasps.WaitUntilExpectationsMet()

	//wasps.CollectMessages(90 * time.Second)
	//if !wasps.Report() {
	//	t.Fail()
	//}

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-1+4, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 + 4, // sc origin, service fee - 1
	}, "owner address in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scAddress, 1+1, map[balance.Color]int64{
		scColor:           1,
		balance.ColorIOTA: 1, // from request token
	}, "SC address in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount-5, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 - 5,
		*color1:           1,
	}, "auction owner address in the end") {
		t.Fail()
	}
}

// two auctions from same auction owner, no bids
func TestFA2Color0Bids(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFairAuction5Requests5Sec1")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          5,
		"request_out":         6,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scNumFairAuction]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(4 * time.Second)

	ownerAddr := sc.OwnerAddress()
	scAddress := sc.SCAddress()
	scColor := sc.GetColor()

	auctionOwnerAddr := auctionOwner.Address()

	err = wasps.NodeClient.RequestFunds(auctionOwnerAddr)
	check(err, t)

	// create 2 colored tokens
	color1, err := mintNewColoredTokens(wasps, auctionOwner.SigScheme(), 1)
	check(err, t)
	time.Sleep(1 * time.Second)

	color2, err := mintNewColoredTokens(wasps, auctionOwner.SigScheme(), 1)
	check(err, t)

	time.Sleep(5 * time.Second)

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "owner address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		scColor: 1,
	}, "SC address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 2,
		*color1:           1,
		*color2:           1,
	}, "auction owner address in the beginning") {
		t.Fail()
		return
	}

	// send request StartAuction for color1
	err = SendSimpleRequest(wasps, auctionOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
		RequestCode: fairauction.RequestStartAuction,
		Vars: map[string]interface{}{
			fairauction.VarReqAuctionColor:                color1,
			fairauction.VarReqStartAuctionMinimumBid:      100,
			fairauction.VarReqStartAuctionDurationMinutes: 1,
			fairauction.VarReqStartAuctionDescription:     "Auction for color1",
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 5, // 5% from 100
			*color1:           1, // token for sale
		},
	})
	check(err, t)
	time.Sleep(1 * time.Second)

	// send request StartAuction for color2
	err = SendSimpleRequest(wasps, auctionOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
		RequestCode: fairauction.RequestStartAuction,
		Vars: map[string]interface{}{
			fairauction.VarReqAuctionColor:                color2,
			fairauction.VarReqStartAuctionMinimumBid:      100,
			fairauction.VarReqStartAuctionDurationMinutes: 1,
			fairauction.VarReqStartAuctionDescription:     "Auction for color2",
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 5, // 5% from 100
			*color2:           1, // token for sale
		},
	})
	check(err, t)

	wasps.WaitUntilExpectationsMet()
	//wasps.CollectMessages(70 * time.Second)
	//if !wasps.Report() {
	//	t.Fail()
	//}

	if !wasps.VerifyAddressBalances(ownerAddr, testutil.RequestFundsAmount-1+4+4, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 + 4 + 4, // sc origin, service fee x 2
	}, "owner address in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scAddress, 1+2, map[balance.Color]int64{
		scColor:           1,
		balance.ColorIOTA: 2, // from request token
	}, "SC address in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount-5-5, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 2 - 5 - 5,
		*color1:           1,
		*color2:           1,
	}, "auction owner address in the end") {
		t.Fail()
	}
}

// create 1 non winning bid in 1 auction
func TestFA1Color1NonWinningBid(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFairAuction5Requests5Sec1")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          4,
		"request_out":         5,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[scNumFairAuction]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(2 * time.Second)

	auctionOwnerAddr := auctionOwner.Address()

	err = wasps.NodeClient.RequestFunds(auctionOwnerAddr)
	check(err, t)

	bidder1Addr := bidder1.Address()

	err = wasps.NodeClient.RequestFunds(bidder1Addr)
	check(err, t)

	scOwnerAddr := sc.OwnerAddress()
	scAddress := sc.SCAddress()
	scColor := sc.GetColor()

	// create 1 colored token
	color1, err := mintNewColoredTokens(wasps, auctionOwner.SigScheme(), 1)
	check(err, t)

	time.Sleep(2 * time.Second)

	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		*color1:           1,
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "auction owner in the beginning") {
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
	if !wasps.VerifyAddressBalances(bidder1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "bidder1 in the beginning") {
		t.Fail()
		return
	}

	// send request StartAuction. Selling 1 token of color1
	err = SendSimpleRequest(wasps, auctionOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
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

	err = SendSimpleRequest(wasps, bidder1.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
		RequestCode: fairauction.RequestPlaceBid,
		Vars: map[string]interface{}{
			fairauction.VarReqAuctionColor: color1,
		},
		Transfer: map[balance.Color]int64{
			balance.ColorIOTA: 42,
		},
	})
	check(err, t)

	wasps.WaitUntilExpectationsMet()
	//wasps.CollectMessages(70 * time.Second)
	//if !wasps.Report() {
	//	t.Fail()
	//}

	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount-5, map[balance.Color]int64{
		*color1:           1,
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 - 5,
	}, "auction owner in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scAddress, 2, map[balance.Color]int64{
		scColor:           1, // sc token
		balance.ColorIOTA: 1,
	}, "SC address in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1+4, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 + 4,
	}, "owner address in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(bidder1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "bidder1 address in the end") {
		t.Fail()
	}
}

func TestFA1Color1Bidder5WinningBids(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFA1Color1Bidder5WinningBids")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2, // wasps.NumSmartContracts(),
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          8,
		"request_out":         9,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	sc := &wasps.SmartContractConfig[scNumFairAuction]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(2 * time.Second)

	auctionOwnerAddr := auctionOwner.Address()

	err = wasps.NodeClient.RequestFunds(auctionOwnerAddr)
	check(err, t)

	// create 1 colored token
	color1, err := mintNewColoredTokens(wasps, auctionOwner.SigScheme(), 1)
	check(err, t)

	scOwnerAddr := sc.OwnerAddress()
	scAddress := sc.SCAddress()
	scColor := sc.GetColor()
	check(err, t)

	bidder1Addr := bidder1.Address()

	err = wasps.NodeClient.RequestFunds(bidder1Addr)
	check(err, t)

	time.Sleep(2 * time.Second)

	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		scColor: 1,
	}, "sc address begin") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "owner address begin") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
		*color1:           1,
	}, "auction owner address begin") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(bidder1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "bidder1 address begin") {
		t.Fail()
		return
	}

	// send request StartAuction. Selling 1 token of color1
	err = SendSimpleRequest(wasps, auctionOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
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
		err = SendSimpleRequest(wasps, bidder1.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddress,
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
	if !wasps.VerifyAddressBalances(scAddress, 1+1, map[balance.Color]int64{
		balance.ColorIOTA: 1,
		scColor:           1,
	}, "SC address end") {
		t.Fail()
	}
	// check SC owner address
	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1+5, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 + 5,
	}, "owner address end") {
		t.Fail()
	}
	// check bidder1 address
	if !wasps.VerifyAddressBalances(bidder1Addr, testutil.RequestFundsAmount-5*25+1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 5*25,
		*color1:           1,
	}, "bidder1 address end") {
		t.Fail()
	}
	// check auction owner address
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount-1-6+125, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 - 6 + 125,
	}, "auction owner address end") {
		t.Fail()
	}
}

// two bidders
func TestFA1Color2Bidders(t *testing.T) {
	// setup
	wasps := setup(t, "test_cluster", "TestFA1Color2Bidders")

	err := wasps.ListenToMessages(map[string]int{
		"bootuprec":           2,
		"active_committee":    1,
		"dismissed_committee": 0,
		"request_in":          13,
		"request_out":         14,
		"state":               -1,
		"vmmsg":               -1,
	})
	check(err, t)

	// number 5 is "Wasm VM PoC program" in cluster.json
	sc := &wasps.SmartContractConfig[scNumFairAuction]

	_, err = PutBootupRecord(wasps, sc)
	check(err, t)

	err = Activate1SC(wasps, sc)
	check(err, t)

	err = CreateOrigin1SC(wasps, sc)
	check(err, t)

	time.Sleep(3 * time.Second)

	auctionOwnerAddr := auctionOwner.Address()

	err = wasps.NodeClient.RequestFunds(auctionOwnerAddr)
	check(err, t)

	// create 1 colored token
	color1, err := mintNewColoredTokens(wasps, auctionOwner.SigScheme(), 1)
	check(err, t)

	scOwnerAddr := sc.OwnerAddress()
	scAddress := sc.SCAddress()

	bidder1Addr := bidder1.Address()

	err = wasps.NodeClient.RequestFunds(bidder1Addr)
	check(err, t)

	bidder2Addr := bidder2.Address()

	err = wasps.NodeClient.RequestFunds(bidder2Addr)
	check(err, t)

	time.Sleep(2 * time.Second)

	if !wasps.VerifyAddressBalances(scAddress, 1, map[balance.Color]int64{
		sc.GetColor(): 1,
	}, "sc address begin") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "owner address begin") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
		*color1:           1,
	}, "auction owner address begin") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(bidder1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "bidder1 address begin") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(bidder2Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "bidder2 address begin") {
		t.Fail()
		return
	}

	// send request StartAuction. Selling 1 token of color1
	err = SendSimpleRequest(wasps, auctionOwner.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
		SCAddress:   scAddress,
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

	for i := 0; i < 5; i++ {
		err = SendSimpleRequest(wasps, bidder1.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddress,
			RequestCode: fairauction.RequestPlaceBid,
			Vars: map[string]interface{}{
				fairauction.VarReqAuctionColor: color1,
			},
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: 22,
			},
		})
		check(err, t)

		err = SendSimpleRequest(wasps, bidder2.SigScheme(), waspapi.CreateSimpleRequestParamsOld{
			SCAddress:   scAddress,
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
	if !wasps.VerifyAddressBalances(scAddress, 2, map[balance.Color]int64{
		sc.GetColor():     1,
		balance.ColorIOTA: 1,
	}, "sc address end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1+5, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 + 5,
	}, "owner address end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount-1+125-6, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 + 125 - 6,
	}, "auction owner address end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(bidder1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "bidder1 address begin") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(bidder2Addr, testutil.RequestFundsAmount-125+1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 125,
		*color1:           1,
	}, "bidder2 address begin") {
		t.Fail()
	}
}
