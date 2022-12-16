// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/wasm/fairauction/go/fairauction"
	"github.com/iotaledger/wasp/contracts/wasm/fairauction/go/fairauctionimpl"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

const (
	description = "Cool NFTs for sale!"
	deposit     = uint64(1000)
	minBid      = uint64(500)
)

func startAuction(t *testing.T) (*wasmsolo.SoloContext, *wasmsolo.SoloAgent, wasmtypes.ScNftID) {
	ctx := wasmsolo.NewSoloContext(t, fairauction.ScName, fairauctionimpl.OnDispatch)
	auctioneer := ctx.NewSoloAgent()
	nftID := ctx.MintNFT(auctioneer, []byte("NFT metadata"))
	require.NoError(t, ctx.Err)

	// start the auction
	sa := fairauction.ScFuncs.StartAuction(ctx.Sign(auctioneer))
	sa.Params.MinimumBid().SetValue(minBid)
	sa.Params.Description().SetValue(description)
	transfer := wasmlib.NewScTransferBaseTokens(deposit) // deposit, must be >=minimum*margin
	transfer.AddNFT(&nftID)
	sa.Func.Transfer(transfer).Post()
	require.NoError(t, ctx.Err)

	return ctx, auctioneer, nftID
}

func TestDeploy(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, fairauction.ScName, fairauctionimpl.OnDispatch)
	require.NoError(t, ctx.ContractExists(fairauction.ScName))
}

func TestStartAuction(t *testing.T) {
	ctx, auctioneer, nftID := startAuction(t)

	nfts := ctx.Chain.L2NFTs(auctioneer.AgentID())
	require.Len(t, nfts, 0)
	nfts = ctx.Chain.L2NFTs(ctx.Account().AgentID())
	require.Len(t, nfts, 1)
	require.Equal(t, nftID, ctx.Cvt.ScNftID(&nfts[0]))

	// remove pending finalize_auction from backlog
	ctx.AdvanceClockBy(61 * time.Minute)
	require.True(t, ctx.WaitForPendingRequests(1))
}

func TestGetAuctionInfo(t *testing.T) {
	ctx, auctioneer, nftID := startAuction(t)

	info := fairauction.ScFuncs.GetAuctionInfo(ctx)
	info.Params.Nft().SetValue(nftID)
	info.Func.Call()
	require.NoError(t, ctx.Err)

	// no bidder since auction just started
	require.EqualValues(t, 0, info.Results.Bidders().Value())
	require.EqualValues(t, nftID, info.Results.Nft().Value())
	require.EqualValues(t, auctioneer.ScAgentID(), info.Results.Creator().Value())
	require.Equal(t, deposit, info.Results.Deposit().Value())
	require.EqualValues(t, description, info.Results.Description().Value())
	require.EqualValues(t, fairauctionimpl.DurationDefault, info.Results.Duration().Value())

	// initial highest bid is 0
	require.EqualValues(t, 0, info.Results.HighestBid().Value())

	// initial highest bidder is set to auctioneer itself
	require.EqualValues(t, auctioneer.ScAgentID(), info.Results.HighestBidder().Value())
	require.EqualValues(t, minBid, info.Results.MinimumBid().Value())
	require.EqualValues(t, fairauctionimpl.OwnerMarginDefault, info.Results.OwnerMargin().Value())

	// expect timestamp should have difference less than 1 second to the `auction.WhenStarted`
	state, err := ctx.Chain.GetStateReader().LatestState()
	require.NoError(t, err)
	require.InDelta(t, uint64(state.Timestamp().UnixNano()), info.Results.WhenStarted().Value(), float64(1*time.Second.Nanoseconds()))

	// remove pending finalize_auction from backlog
	ctx.AdvanceClockBy(61 * time.Minute)
	require.True(t, ctx.WaitForPendingRequests(1))
}

func TestFinalizedNoBids(t *testing.T) {
	ctx, _, nftID := startAuction(t)

	// wait for finalize_auction
	ctx.AdvanceClockBy(61 * time.Minute)
	require.True(t, ctx.WaitForPendingRequests(1))

	info := fairauction.ScFuncs.GetAuctionInfo(ctx)
	info.Params.Nft().SetValue(nftID)
	info.Func.Call()

	require.NoError(t, ctx.Err)
	require.EqualValues(t, 0, info.Results.Bidders().Value())
}

func TestFinalizedOneBidTooLow(t *testing.T) {
	ctx, _, nftID := startAuction(t)

	bidder := ctx.NewSoloAgent()
	placeBid := fairauction.ScFuncs.PlaceBid(ctx.Sign(bidder))
	placeBid.Params.Nft().SetValue(nftID)
	placeBid.Func.TransferBaseTokens(100).Post()
	require.Error(t, ctx.Err)

	// wait for finalize_auction
	ctx.AdvanceClockBy(61 * time.Minute)
	require.True(t, ctx.WaitForPendingRequests(1))

	info := fairauction.ScFuncs.GetAuctionInfo(ctx)
	info.Params.Nft().SetValue(nftID)
	info.Func.Call()

	require.NoError(t, ctx.Err)
	require.EqualValues(t, 0, info.Results.Bidders().Value())
	require.EqualValues(t, 0, info.Results.HighestBid().Value())
}

func TestFinalizedOneBid(t *testing.T) {
	ctx, _, nftID := startAuction(t)

	bidder0 := ctx.NewSoloAgent()
	placeBid := fairauction.ScFuncs.PlaceBid(ctx.Sign(bidder0))
	placeBid.Params.Nft().SetValue(nftID)
	placeBid.Func.TransferBaseTokens(5000).Post()
	require.NoError(t, ctx.Err)

	bidder1 := ctx.NewSoloAgent()
	placeBid = fairauction.ScFuncs.PlaceBid(ctx.Sign(bidder1))
	placeBid.Params.Nft().SetValue(nftID)
	placeBid.Func.TransferBaseTokens(5001).Post()
	require.NoError(t, ctx.Err)

	// wait for finalize_auction
	ctx.AdvanceClockBy(61 * time.Minute)
	require.True(t, ctx.WaitForPendingRequests(1))

	info := fairauction.ScFuncs.GetAuctionInfo(ctx)
	info.Params.Nft().SetValue(nftID)
	info.Func.Call()

	require.NoError(t, ctx.Err)
	require.EqualValues(t, 2, info.Results.Bidders().Value())
	require.EqualValues(t, 5001, info.Results.HighestBid().Value())
	require.Equal(t, bidder1.ScAgentID(), info.Results.HighestBidder().Value())
}
