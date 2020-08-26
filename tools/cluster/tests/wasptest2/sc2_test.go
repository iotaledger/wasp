package wasptest2

import (
	"crypto/rand"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction/faclient"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/mr-tron/base58"
	"os"
	"testing"
	"time"
)

func Test2SC(t *testing.T) {
	var seed [32]byte
	rand.Read(seed[:])
	seed58 := base58.Encode(seed[:])
	wallet := testutil.NewWallet(seed58)
	scOwner := wallet.WithIndex(0) // owner of 2 SCs
	auctionOwner := wallet.WithIndex(2)

	// setup
	wasps := setup(t, "test_cluster2", "TestFA")

	programHashTR, err := hashing.HashValueFromBase58(tokenregistry.ProgramHash)
	check(err, t)

	programHashFA, err := hashing.HashValueFromBase58(fairauction.ProgramHash)
	check(err, t)

	scOwnerAddr := scOwner.Address()
	err = wasps.NodeClient.RequestFunds(&scOwnerAddr)
	check(err, t)

	auctionOwnerAddr := auctionOwner.Address()
	err = wasps.NodeClient.RequestFunds(&auctionOwnerAddr)
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
	scTRAddr, scTRColor, err := waspapi.CreateAndDeploySC(waspapi.CreateAndDeploySCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHashTR,
		Textout:               os.Stdout,
		Prefix:                "[deploy " + tokenregistry.ProgramHash + "]",
	})
	checkSuccess(err, t, "TokenRegistry created")

	scFAAddr, scFAColor, err := waspapi.CreateAndDeploySC(waspapi.CreateAndDeploySCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHashFA,
		Textout:               os.Stdout,
		Prefix:                "[deploy " + fairauction.ProgramHash + "]",
	})
	checkSuccess(err, t, "FaiAuction created")

	err = waspapi.WaitSmartContractInitialized(wasps.PublisherHosts(), scTRAddr, scTRColor, 20*time.Second)
	checkSuccess(err, t, "TokenRegistry initialized")

	err = waspapi.WaitSmartContractInitialized(wasps.PublisherHosts(), scFAAddr, scFAColor, 20*time.Second)
	checkSuccess(err, t, "FairAuction initialized")

	// minting 1 token with TokenRegistry
	mintedColor, err := tokenregistry.MintAndRegister(wasps.NodeClient, tokenregistry.MintAndRegisterParams{
		SenderSigScheme:   auctionOwner.SigScheme(),
		Supply:            1,
		MintTarget:        auctionOwnerAddr,
		RegistryAddr:      *scTRAddr,
		Description:       "Non-fungible coin 1. Very expensive",
		WaitToBeProcessed: true,
		PublisherHosts:    wasps.PublisherHosts(),
		Timeout:           20 * time.Second,
	})
	checkSuccess(err, t, "token minted")

	if !wasps.VerifyAddressBalances(*scFAAddr, 1, map[balance.Color]int64{
		*scFAColor: 1, // sc token
	}, "SC FairAuction address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(*scTRAddr, 1, map[balance.Color]int64{
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
		*mintedColor:      1,
	}, "Auction owner in the beginning") {
		t.Fail()
		return
	}
}

func TestPlus2SC(t *testing.T) {
	var seed [32]byte
	rand.Read(seed[:])
	seed58 := base58.Encode(seed[:])
	wallet := testutil.NewWallet(seed58)
	scOwner := wallet.WithIndex(0) // owner of 2 SCs
	auctionOwner := wallet.WithIndex(2)
	bidder1 := wallet.WithIndex(3)
	bidder2 := wallet.WithIndex(4)

	// setup
	wasps := setup(t, "test_cluster2", "TestFA")

	programHashTR, err := hashing.HashValueFromBase58(tokenregistry.ProgramHash)
	check(err, t)

	programHashFA, err := hashing.HashValueFromBase58(fairauction.ProgramHash)
	check(err, t)

	scOwnerAddr := scOwner.Address()
	err = wasps.NodeClient.RequestFunds(&scOwnerAddr)
	check(err, t)

	auctionOwnerAddr := auctionOwner.Address()
	err = wasps.NodeClient.RequestFunds(&auctionOwnerAddr)
	check(err, t)

	bidder1Addr := bidder1.Address()
	err = wasps.NodeClient.RequestFunds(&bidder1Addr)
	check(err, t)

	bidder2Addr := bidder2.Address()
	err = wasps.NodeClient.RequestFunds(&bidder2Addr)
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

	// create TokenRegistry
	scTRAddr, scTRColor, err := waspapi.CreateAndDeploySC(waspapi.CreateAndDeploySCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHashTR,
		Textout:               os.Stdout,
		Prefix:                "[deploy " + tokenregistry.ProgramHash + "]",
	})
	checkSuccess(err, t, "TokenRegistry created")

	scFAAddr, scFAColor, err := waspapi.CreateAndDeploySC(waspapi.CreateAndDeploySCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHashFA,
		Textout:               os.Stdout,
		Prefix:                "[deploy " + fairauction.ProgramHash + "]",
	})
	checkSuccess(err, t, "FaiAuction created")

	err = waspapi.WaitSmartContractInitialized(wasps.PublisherHosts(), scTRAddr, scTRColor, 20*time.Second)
	checkSuccess(err, t, "TokenRegistry initialized")

	err = waspapi.WaitSmartContractInitialized(wasps.PublisherHosts(), scFAAddr, scFAColor, 20*time.Second)
	checkSuccess(err, t, "FairAuction initialized")

	// minting 1 token with TokenRegistry
	mintedColor, err := tokenregistry.MintAndRegister(wasps.NodeClient, tokenregistry.MintAndRegisterParams{
		SenderSigScheme:   auctionOwner.SigScheme(),
		Supply:            1,
		MintTarget:        auctionOwnerAddr,
		RegistryAddr:      *scTRAddr,
		Description:       "Non-fungible coin 1. Very expensive",
		WaitToBeProcessed: true,
		PublisherHosts:    wasps.PublisherHosts(),
		Timeout:           20 * time.Second,
	})
	checkSuccess(err, t, "token minted")

	if !wasps.VerifyAddressBalances(*scFAAddr, 1, map[balance.Color]int64{
		*scFAColor: 1, // sc token
	}, "SC FairAuction address in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(*scTRAddr, 1, map[balance.Color]int64{
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
		*mintedColor:      1,
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
	//
	faclientOwner := faclient.NewClient(wasps.NodeClient, wasps.ApiHosts()[0], scFAAddr, auctionOwner.SigScheme())
	reqTxId, err := faclientOwner.StartAuction("selling my only token", mintedColor, 1, 100, 1)
	checkSuccess(err, t, "StartAuction request created")

	err = waspapi.WaitForRequestProcessedMulti(wasps.PublisherHosts(), scFAAddr, reqTxId, 0, 15*time.Second)
	checkSuccess(err, t, "StartAuction request processed")

	//
	//time.Sleep(3 * time.Second)
	//
	//faclientBidder1 := faclient.NewClient(wasps.NodeClient, wasps.ApiHosts()[0], scFAAddr, bidder1.SigScheme())
	//faclientBidder2 := faclient.NewClient(wasps.NodeClient, wasps.ApiHosts()[0], scFAAddr, bidder2.SigScheme())
	//
	//err = faclientBidder1.PlaceBid(mintedColor, 110)
	//check(err, t)
	//
	//err = faclientBidder2.PlaceBid(mintedColor, 120)
	//check(err, t)
	//
	//time.Sleep(10 * time.Second)
	//
	//wasps.CollectMessages(15 * time.Second)
	//
	//if !wasps.Report() {
	//	t.Fail()
	//}
	//if !wasps.VerifyAddressBalances(*scFAAddr, 1, map[balance.Color]int64{
	//	*scFAColor: 1, // sc token
	//}, "SC FairAuction address in the end") {
	//	t.Fail()
	//	return
	//}
	//if !wasps.VerifyAddressBalances(*scTRAddr, 1, map[balance.Color]int64{
	//	*scTRColor: 1, // sc token
	//}, "SC TokenRegistry address in the end") {
	//	t.Fail()
	//	return
	//}
	//if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-2 + 6, map[balance.Color]int64{
	//	balance.ColorIOTA: testutil.RequestFundsAmount - 2 + 6,
	//}, "SC owner in the end") {
	//	t.Fail()
	//	return
	//}
	//if !wasps.VerifyAddressBalances(auctionOwnerAddr, testutil.RequestFundsAmount - 1 + 120 - 6, map[balance.Color]int64{
	//	balance.ColorIOTA: testutil.RequestFundsAmount - 1 + 120 - 6,
	//}, "Auction owner in the end") {
	//	t.Fail()
	//	return
	//}
	//if !wasps.VerifyAddressBalances(bidder1Addr, testutil.RequestFundsAmount, map[balance.Color]int64{
	//	balance.ColorIOTA: testutil.RequestFundsAmount,
	//}, "Bidder1 in the end") {
	//	t.Fail()
	//	return
	//}
	//if !wasps.VerifyAddressBalances(bidder2Addr, testutil.RequestFundsAmount-120, map[balance.Color]int64{
	//	balance.ColorIOTA: testutil.RequestFundsAmount-120,
	//	*mintedColor: 1,
	//}, "Bidder2 in the end") {
	//	t.Fail()
	//	return
	//}
}
