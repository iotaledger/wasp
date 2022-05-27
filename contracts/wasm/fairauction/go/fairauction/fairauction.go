// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package fairauction

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

const (
	// default duration is 60 min
	DurationDefault = 60
	// minimum duration is 1 min
	DurationMin = 1
	// maximum duration is 120 min
	DurationMax          = 120
	MaxDescriptionLength = 150
	OwnerMarginDefault   = 50
	OwnerMarginMin       = 5
	OwnerMarginMax       = 100
	DustDeposit          = 1
)

func funcStartAuction(ctx wasmlib.ScFuncContext, f *StartAuctionContext) {
	auctionNFT := f.Params.Nft().Value()
	nfts := ctx.Allowance().NftIDs()
	for _, nft := range nfts {
		if nft.String() == auctionNFT.String() {
			goto exist
		}
	}
	ctx.Panic("Missing auction tokens")

exist:
	// Any transfer to the Smart Contract is implemented as a transfer of assets to the caller's L2 account,
	// with an allowance to those assets passed to the Smart Contract, so Smart Contract needs to actively
	// transfer the for sale NFT to Smart Contract account
	transfer := wasmlib.NewScTransferIotas(DustDeposit)
	transfer.AddNFT(&auctionNFT)
	ctx.TransferAllowed(ctx.AccountID(), transfer, false)

	minimumBid := f.Params.MinimumBid().Value()

	// duration in minutes
	duration := f.Params.Duration().Value()
	if duration == 0 {
		duration = DurationDefault
	}
	if duration < DurationMin {
		duration = DurationMin
	}
	if duration > DurationMax {
		duration = DurationMax
	}

	description := f.Params.Description().Value()
	if description == "" {
		description = "N/A"
	}
	if len(description) > MaxDescriptionLength {
		ss := description[:MaxDescriptionLength]
		description = ss + "[...]"
	}

	ownerMargin := f.State.OwnerMargin().Value()
	if ownerMargin == 0 {
		ownerMargin = OwnerMarginDefault
	}

	// need at least 1 iota (dust deposit) to run SC
	margin := minimumBid * ownerMargin / 1000
	if margin == 0 {
		margin = DustDeposit
	}
	deposit := ctx.Allowance().Iotas()
	if deposit < margin {
		ctx.Panic("Insufficient deposit")
	}

	currentAuction := f.State.Auctions().GetAuction(auctionNFT)
	if currentAuction.Exists() {
		ctx.Panic("Auction for this token already exists")
	}

	auction := &Auction{
		Creator:       ctx.Caller(),
		Deposit:       deposit,
		Description:   description,
		Duration:      duration,
		HighestBid:    0,
		HighestBidder: ctx.Caller(),
		MinimumBid:    minimumBid,
		OwnerMargin:   ownerMargin,
		Nft:           auctionNFT,
		WhenStarted:   ctx.Timestamp(),
	}
	currentAuction.SetValue(auction)

	fa := ScFuncs.FinalizeAuction(ctx)
	fa.Params.Nft().SetValue(auction.Nft)
	fa.Func.Delay(duration * 60).Post()
}

func funcFinalizeAuction(ctx wasmlib.ScFuncContext, f *FinalizeAuctionContext) {
	auctionNFT := f.Params.Nft().Value()
	currentAuction := f.State.Auctions().GetAuction(auctionNFT)
	ctx.Require(currentAuction.Exists(), "Missing auction info")
	auction := currentAuction.Value()
	if auction.HighestBid == 0 {
		ctx.Log("No one bid on " + auctionNFT.String())
		ownerFee := auction.MinimumBid * auction.OwnerMargin / 1000
		if ownerFee == 0 {
			ownerFee = 1
		}
		// finalizeAuction request token was probably not confirmed yet
		transferIotas(ctx, ctx.ContractCreator(), ownerFee-1)
		transferNFT(ctx, auction.Creator, auction.Nft)
		transferIotas(ctx, auction.Creator, auction.Deposit-ownerFee)
		return
	}

	ownerFee := auction.HighestBid * auction.OwnerMargin / 1000
	if ownerFee == 0 {
		ownerFee = 1
	}

	// return staked bids to losers
	bids := f.State.Bids().GetBids(auctionNFT)
	bidderList := f.State.BidderList().GetBidderList(auctionNFT)
	size := bidderList.Length()
	for i := uint32(0); i < size; i++ {
		loser := bidderList.GetAgentID(i).Value()
		if loser != auction.HighestBidder {
			bid := bids.GetBid(loser).Value()
			transferIotas(ctx, loser, bid.Amount)
		}
	}

	// finalizeAuction request token was probably not confirmed yet
	transferIotas(ctx, ctx.ContractCreator(), ownerFee-1)
	transferNFT(ctx, auction.HighestBidder, auction.Nft)
	transferIotas(ctx, auction.Creator, auction.Deposit+auction.HighestBid-ownerFee)
}

func funcPlaceBid(ctx wasmlib.ScFuncContext, f *PlaceBidContext) {
	bidAmount := ctx.Allowance().Iotas()
	ctx.Require(bidAmount > 0, "Missing bid amount")

	token := f.Params.Nft().Value()
	currentAuction := f.State.Auctions().GetAuction(token)
	ctx.Require(currentAuction.Exists(), "Missing auction info")

	auction := currentAuction.Value()
	bids := f.State.Bids().GetBids(token)
	bidderList := f.State.BidderList().GetBidderList(token)
	caller := ctx.Caller()
	currentBid := bids.GetBid(caller)
	if currentBid.Exists() {
		ctx.Log("Upped bid from: " + caller.String())
		bid := currentBid.Value()
		bidAmount += bid.Amount
		bid.Amount = bidAmount
		bid.Timestamp = ctx.Timestamp()
		currentBid.SetValue(bid)
	} else {
		ctx.Require(bidAmount >= auction.MinimumBid, "Insufficient bid amount")
		ctx.Log("New bid from: " + caller.String())
		index := bidderList.Length()
		bidderList.AppendAgentID().SetValue(caller)
		bid := &Bid{
			Index:     index,
			Amount:    bidAmount,
			Timestamp: ctx.Timestamp(),
		}
		currentBid.SetValue(bid)
	}
	if bidAmount > auction.HighestBid {
		ctx.Log("New highest bidder")
		auction.HighestBid = bidAmount
		auction.HighestBidder = caller
		currentAuction.SetValue(auction)
	}
}

func funcSetOwnerMargin(ctx wasmlib.ScFuncContext, f *SetOwnerMarginContext) {
	ownerMargin := f.Params.OwnerMargin().Value()
	if ownerMargin < OwnerMarginMin {
		ownerMargin = OwnerMarginMin
	}
	if ownerMargin > OwnerMarginMax {
		ownerMargin = OwnerMarginMax
	}
	f.State.OwnerMargin().SetValue(ownerMargin)
}

func viewGetAuctionInfo(ctx wasmlib.ScViewContext, f *GetAuctionInfoContext) {
	token := f.Params.Nft().Value()
	currentAuction := f.State.Auctions().GetAuction(token)
	if !currentAuction.Exists() {
		ctx.Panic("Missing auction info")
	}

	auction := currentAuction.Value()
	f.Results.Creator().SetValue(auction.Creator)
	f.Results.Deposit().SetValue(auction.Deposit)
	f.Results.Description().SetValue(auction.Description)
	f.Results.Duration().SetValue(auction.Duration)
	f.Results.HighestBid().SetValue(auction.HighestBid)
	f.Results.HighestBidder().SetValue(auction.HighestBidder)
	f.Results.MinimumBid().SetValue(auction.MinimumBid)
	f.Results.OwnerMargin().SetValue(auction.OwnerMargin)
	f.Results.Nft().SetValue(auction.Nft)
	f.Results.WhenStarted().SetValue(auction.WhenStarted)

	bidderList := f.State.BidderList().GetBidderList(token)
	f.Results.Bidders().SetValue(bidderList.Length())
}

func transferIotas(ctx wasmlib.ScFuncContext, agent wasmtypes.ScAgentID, amount uint64) {
	if agent.IsAddress() {
		// send back to original Tangle address
		ctx.Send(agent.Address(), wasmlib.NewScTransferIotas(amount))
		return
	}

	// TODO not an address, deposit into account on chain
	ctx.Send(agent.Address(), wasmlib.NewScTransferIotas(amount))
}

func transferNFT(ctx wasmlib.ScFuncContext, agent wasmtypes.ScAgentID, nft wasmtypes.ScNftID) {
	if agent.IsAddress() {
		// send back to original Tangle address
		ctx.Send(agent.Address(), wasmlib.NewScTransferNFT(&nft))
		return
	}

	// TODO not an address, deposit into account on chain
	ctx.Send(agent.Address(), wasmlib.NewScTransferNFT(&nft))
}
