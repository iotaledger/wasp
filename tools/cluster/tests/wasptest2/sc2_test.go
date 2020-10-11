package wasptest2

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction/faclient"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry/trclient"
	"github.com/mr-tron/base58"
)

func Test2SC(t *testing.T) {
	var seed [32]byte
	rand.Read(seed[:])
	seed58 := base58.Encode(seed[:])
	wallet := testutil.NewWallet(seed58)
	scOwner := wallet.WithIndex(0) // owner of 2 SCs
	auctionOwner := wallet.WithIndex(2)

	// setup
	wasps := setup(t, "test_cluster2", "Test2SC")

	programHashTR, err := hashing.HashValueFromBase58(tokenregistry.ProgramHash)
	check(err, t)

	programHashFA, err := hashing.HashValueFromBase58(fairauction.ProgramHash)
	check(err, t)

	scOwnerAddr := scOwner.Address()
	err = wasps.NodeClient.RequestFunds(scOwnerAddr)
	check(err, t)

	auctionOwnerAddr := auctionOwner.Address()
	err = wasps.NodeClient.RequestFunds(auctionOwnerAddr)
	check(err, t)
	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "sc owner in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "auctionOwner in the beginning") {
		t.Fail()
		return
	}

	// create TokenRegistry
	scTRAddr, scTRColor, err := waspapi.CreateSC(waspapi.CreateSCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHashTR,
		Description:           "TokenRegistry, a PoC smart contract",
		Textout:               os.Stdout,
		Prefix:                "[deploy " + tokenregistry.ProgramHash + "]",
	})
	checkSuccess(err, t, "TokenRegistry created")

	scFAAddr, scFAColor, err := waspapi.CreateSC(waspapi.CreateSCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHashFA,
		Description:           "FairAuction, a PoC smart contract",
		Textout:               os.Stdout,
		Prefix:                "[deploy " + fairauction.ProgramHash + "]",
	})
	checkSuccess(err, t, "FairAuction created")

	err = waspapi.ActivateSCMulti(waspapi.ActivateSCParams{
		Addresses:         []*address.Address{scFAAddr, scTRAddr},
		ApiHosts:          wasps.ApiHosts(),
		WaitForCompletion: true,
		PublisherHosts:    wasps.PublisherHosts(),
		Timeout:           60 * time.Second,
	})
	checkSuccess(err, t, "2 smart contracts activated and initialized")

	succ := waspapi.CheckSC(wasps.ApiHosts(), scTRAddr)
	assert.True(t, succ)

	succ = waspapi.CheckSC(wasps.ApiHosts(), scFAAddr)
	assert.True(t, succ)

	tc := trclient.NewClient(scclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scTRAddr,
		auctionOwner.SigScheme(),
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
	var seed [32]byte
	rand.Read(seed[:])
	seed58 := base58.Encode(seed[:])
	wallet := testutil.NewWallet(seed58)
	scOwner := wallet.WithIndex(0) // owner of 2 SCs
	auctionOwner := wallet.WithIndex(1)
	bidder1 := wallet.WithIndex(2)
	bidder2 := wallet.WithIndex(3)

	// setup
	wasps := setup(t, "test_cluster2", "TestFA")

	programHashTR, err := hashing.HashValueFromBase58(tokenregistry.ProgramHash)
	check(err, t)

	programHashFA, err := hashing.HashValueFromBase58(fairauction.ProgramHash)
	check(err, t)

	scOwnerAddr := scOwner.Address()
	err = wasps.NodeClient.RequestFunds(scOwnerAddr)
	check(err, t)

	auctionOwnerAddr := auctionOwner.Address()
	err = wasps.NodeClient.RequestFunds(auctionOwnerAddr)
	check(err, t)

	bidder1Addr := bidder1.Address()
	err = wasps.NodeClient.RequestFunds(bidder1Addr)
	check(err, t)

	bidder2Addr := bidder2.Address()
	err = wasps.NodeClient.RequestFunds(bidder2Addr)
	check(err, t)

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "sc owner in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "auctionOwner in the beginning") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(bidder1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "bidder1 in the beginning") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(bidder2Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "bidder2 in the beginning") {
		t.Fail()
		return
	}
	t.Logf("bidder1 addr: %s", bidder1Addr.String())
	t.Logf("bidder2 addr: %s", bidder2Addr.String())

	// create TokenRegistry
	scTRAddr, scTRColor, err := waspapi.CreateSC(waspapi.CreateSCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHashTR,
		Description:           "TokenRegistry, a PoC smart contract",
		Textout:               os.Stdout,
		Prefix:                "[deploy " + tokenregistry.ProgramHash + "]",
	})
	checkSuccess(err, t, "TokenRegistry created")

	scFAAddr, scFAColor, err := waspapi.CreateSC(waspapi.CreateSCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHashFA,
		Description:           "FairAuction, a PoC smart contract",
		Textout:               os.Stdout,
		Prefix:                "[deploy " + fairauction.ProgramHash + "]",
	})
	checkSuccess(err, t, "FairAuction created")

	err = waspapi.ActivateSCMulti(waspapi.ActivateSCParams{
		Addresses:         []*address.Address{scFAAddr, scTRAddr},
		ApiHosts:          wasps.ApiHosts(),
		WaitForCompletion: true,
		PublisherHosts:    wasps.PublisherHosts(),
		Timeout:           60 * time.Second,
	})
	checkSuccess(err, t, "2 smart contracts have been activated and initialized")

	tc := trclient.NewClient(scclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scTRAddr,
		auctionOwner.SigScheme(),
	))

	// minting 1 token with TokenRegistry
	tx, err := tc.MintAndRegister(trclient.MintAndRegisterParams{
		Supply:      1,
		MintTarget:  *auctionOwnerAddr,
		Description: "Non-fungible coin 1. Very expensive",
	})
	checkSuccess(err, t, "token minted")

	time.Sleep(2 * time.Second)

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

	faclientOwner := faclient.NewClient(scclient.New(wasps.NodeClient, wasps.WaspClient(0), scFAAddr, auctionOwner.SigScheme()))

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
	time.Sleep(90 * time.Second)

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
