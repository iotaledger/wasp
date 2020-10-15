package wasptest2

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction/faclient"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry/trclient"
)

func Test2SC(t *testing.T) {
	wasps := setup(t, "Test2SC")

	err := requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	auctionOwner := wallet.WithIndex(1)
	auctionOwnerAddr := auctionOwner.Address()
	err = requestFunds(wasps, auctionOwnerAddr, "auction owner")
	check(err, t)

	scTRAddr, scTRColor, err:= startSmartContract(wasps, tokenregistry.ProgramHash, tokenregistry.Description)
	checkSuccess(err, t, "TokenRegistry has been created and activated")

	scFAAddr, scFAColor, err:= startSmartContract(wasps, fairauction.ProgramHash, fairauction.Description)
	checkSuccess(err, t, "FairAuction has been created and activated")

	succ := waspapi.CheckSC(wasps.ApiHosts(), scTRAddr)
	assert.True(t, succ)

	succ = waspapi.CheckSC(wasps.ApiHosts(), scFAAddr)
	assert.True(t, succ)

	tc := trclient.NewClient(scclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scTRAddr,
		auctionOwner.SigScheme(),
		20*time.Second,
	))

	// minting 1 token with TokenRegistry
	tx, err := tc.MintAndRegister(trclient.MintAndRegisterParams{
		Supply:      1,
		MintTarget:  *auctionOwnerAddr,
		Description: "Non-fungible coin 1. Very expensive",
	})
	checkSuccess(err, t, "token minted")

	mintedColor := balance.Color(tx.ID())

	if !wasps.VerifyAddressBalances(scFAAddr, 1, map[balance.Color]int64{
		*scFAColor: 1, // sc token
	}, "SC FairAuction address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scTRAddr, 1, map[balance.Color]int64{
		*scTRColor: 1, // sc token
	}, "SC TokenRegistry address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-2, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 2,
	}, "SC owner in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
		mintedColor:       1,
	}, "Auction owner in the beginning") {
		t.Fail()
		return
	}
}

// scenario with 2 smart contracts
func TestPlus2SC(t *testing.T) {
	wasps := setup(t, "TestPlus2SC")

	err := requestFunds(wasps, scOwnerAddr, "sc owner")
	check(err, t)

	auctionOwner := wallet.WithIndex(1)
	auctionOwnerAddr := auctionOwner.Address()
	err = requestFunds(wasps, auctionOwnerAddr, "auction owner")
	check(err, t)

	bidder1 := wallet.WithIndex(2)
	bidder1Addr := bidder1.Address()
	err = requestFunds(wasps, bidder1Addr, "bidder 1")
	check(err, t)

	bidder2 := wallet.WithIndex(3)
	bidder2Addr := bidder2.Address()
	err = requestFunds(wasps, bidder2Addr, "bidder 2")
	check(err, t)

	scTRAddr, scTRColor, err:= startSmartContract(wasps, tokenregistry.ProgramHash, tokenregistry.Description)
	checkSuccess(err, t, "TokenRegistry has been created and activated")

	scFAAddr, scFAColor, err:= startSmartContract(wasps, fairauction.ProgramHash, fairauction.Description)
	checkSuccess(err, t, "FairAuction has been created and activated")

	tc := trclient.NewClient(scclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scTRAddr,
		auctionOwner.SigScheme(),
		20*time.Second,
	))

	// minting 1 token with TokenRegistry
	tx, err := tc.MintAndRegister(trclient.MintAndRegisterParams{
		Supply:      1,
		MintTarget:  *auctionOwnerAddr,
		Description: "Non-fungible coin 1. Very expensive",
	})
	checkSuccess(err, t, "token minted")

	mintedColor := balance.Color(tx.ID())

	if !wasps.VerifyAddressBalances(scFAAddr, 1, map[balance.Color]int64{
		*scFAColor: 1, // sc token
	}, "SC FairAuction address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scTRAddr, 1, map[balance.Color]int64{
		*scTRColor: 1, // sc token
	}, "SC TokenRegistry address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-2, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 2,
	}, "SC owner in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
		mintedColor:       1,
	}, "Auction owner in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(bidder1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "Bidder1 in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(bidder2Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "Bidder2 in the beginning") {
		t.Fail()
		return
	}

	faclientOwner := faclient.NewClient(scclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scFAAddr,
		auctionOwner.SigScheme(),
		20*time.Second,
	))

	_, err = faclientOwner.StartAuction("selling my only token", &mintedColor, 1, 100, 1)
	checkSuccess(err, t, "StartAuction created")

	faclientBidder1 := faclient.NewClient(scclient.New(wasps.NodeClient, wasps.WaspClient(0), scFAAddr, bidder1.SigScheme()))
	faclientBidder2 := faclient.NewClient(scclient.New(wasps.NodeClient, wasps.WaspClient(0), scFAAddr, bidder2.SigScheme()))

	subs, err := subscribe.SubscribeMulti(wasps.PublisherHosts(), "request_out")
	check(err, t)

	tx1, err := faclientBidder1.PlaceBid(&mintedColor, 110)
	check(err, t)
	tx2, err := faclientBidder2.PlaceBid(&mintedColor, 110)
	check(err, t)

	patterns := [][]string{
		{"request_out", scFAAddr.String(), tx1.ID().String(), "0"},
		{"request_out", scFAAddr.String(), tx2.ID().String(), "0"},
	}
	err = nil
	if !subs.WaitForPatterns(patterns, 40*time.Second) {
		err = fmt.Errorf("didn't receive completion message in time")
	}
	checkSuccess(err, t, "2 bids have been placed")

	// wait for auction to finish
	time.Sleep(65 * time.Second)

	if !wasps.VerifyAddressBalances(scFAAddr, 2, map[balance.Color]int64{
		*scFAColor:        1, // sc token
		balance.ColorIOTA: 1, // 1 i for sending request to itself
	}, "SC FairAuction address in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scTRAddr, 1, map[balance.Color]int64{
		*scTRColor: 1, // sc token
	}, "SC TokenRegistry address in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-2+4, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 2 + 4,
	}, "SC owner in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount-2+110-4, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 2 + 110 - 4,
	}, "Auction owner in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(bidder1Addr, testutil.RequestFundsAmount-110+1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 110,
		mintedColor:       1,
	}, "Bidder1 in the end") {
		t.Fail()
	}
	if !wasps.VerifyAddressBalances(bidder2Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "Bidder2 in the end") {
		t.Fail()
	}
}
