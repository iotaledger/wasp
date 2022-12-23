package sbtestsc

import (
	"fmt"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

// testSplitFunds calls Send in a loop by sending 200 base tokens back to the caller
func testSplitFunds(ctx isc.Sandbox) dict.Dict {
	addr, ok := isc.AddressFromAgentID(ctx.Caller())
	ctx.Requiref(ok, "caller must have L1 address")
	// claim 1Mi base tokens from allowance at a time
	baseTokensToTransfer := 1 * isc.Million
	for !ctx.AllowanceAvailable().IsEmpty() && ctx.AllowanceAvailable().Assets.BaseTokens >= baseTokensToTransfer {
		// send back to caller's address
		// depending on the amount of base tokens, it will exceed number of outputs or not
		ctx.TransferAllowedFunds(ctx.AccountID(), isc.NewAllowance(baseTokensToTransfer, nil, nil))
		ctx.Send(
			isc.RequestParameters{
				TargetAddress:  addr,
				FungibleTokens: isc.NewFungibleBaseTokens(baseTokensToTransfer),
			},
		)
	}
	return nil
}

// testSplitFundsNativeTokens calls Send for each Native token
func testSplitFundsNativeTokens(ctx isc.Sandbox) dict.Dict {
	addr, ok := isc.AddressFromAgentID(ctx.Caller())
	ctx.Requiref(ok, "caller must have L1 address")
	// claims all base tokens from allowance
	ctx.TransferAllowedFunds(ctx.AccountID(), isc.NewAllowance(ctx.AllowanceAvailable().Assets.BaseTokens, nil, nil))
	for _, token := range ctx.AllowanceAvailable().Assets.Tokens {
		for ctx.AllowanceAvailable().Assets.AmountNativeToken(&token.ID).Cmp(util.Big0) > 0 {
			// claim 1 token from allowance at a time
			// send back to caller's address
			// depending on the amount of tokens, it will exceed number of outputs or not
			assets := isc.NewEmptyFungibleTokens().AddNativeTokens(token.ID, 1)
			transfer := isc.NewAllowanceFungibleTokens(assets)
			rem := ctx.TransferAllowedFunds(ctx.AccountID(), transfer)
			fmt.Printf("%s\n", rem)
			ctx.Send(
				isc.RequestParameters{
					TargetAddress:                 addr,
					FungibleTokens:                assets,
					AdjustToMinimumStorageDeposit: true,
				},
			)
		}
	}
	return nil
}

func pingAllowanceBack(ctx isc.Sandbox) dict.Dict {
	addr, ok := isc.AddressFromAgentID(ctx.Caller())
	// assert caller is L1 address, not a SC
	ctx.Requiref(ok && !ctx.ChainID().IsSameChain(ctx.Caller()),
		"pingAllowanceBack: caller expected to be a L1 address")
	// save allowance budget because after transfer it will be modified
	toSend := ctx.AllowanceAvailable().Assets
	if toSend.IsEmpty() {
		// nothing to send back, NOP
		return nil
	}
	// claim all transfer to the current account
	left := ctx.TransferAllowedFunds(ctx.AccountID())
	// assert what has left is empty. Only for testing
	ctx.Requiref(left.IsEmpty(), "pingAllowanceBack: inconsistency")

	// send the funds to the caller L1 address on-ledger
	ctx.Send(
		isc.RequestParameters{
			TargetAddress:  addr,
			FungibleTokens: toSend,
		},
	)
	return nil
}

// testEstimateMinimumStorageDeposit returns true if the provided allowance is enough to pay for a L1 request, panics otherwise
func testEstimateMinimumStorageDeposit(ctx isc.Sandbox) dict.Dict {
	addr, ok := isc.AddressFromAgentID(ctx.Caller())
	ctx.Requiref(ok, "caller must have L1 address")

	provided := ctx.AllowanceAvailable().Assets.BaseTokens

	requestParams := isc.RequestParameters{
		TargetAddress: addr,
		Metadata: &isc.SendMetadata{
			EntryPoint:     isc.Hn("foo"),
			TargetContract: isc.Hn("bar"),
		},
		AdjustToMinimumStorageDeposit: true,
	}

	required := ctx.EstimateRequiredStorageDeposit(requestParams)
	ctx.Requiref(provided >= required, "not enough funds")
	return nil
}

// tries to sendback whaever NFTs are specified in allowance
func sendNFTsBack(ctx isc.Sandbox) dict.Dict {
	addr, ok := isc.AddressFromAgentID(ctx.Caller())
	ctx.Requiref(ok, "caller must have L1 address")

	allowance := ctx.AllowanceAvailable()
	ctx.TransferAllowedFunds(ctx.AccountID())
	for _, nftID := range allowance.NFTs {
		ctx.SendAsNFT(isc.RequestParameters{
			TargetAddress:                 addr,
			FungibleTokens:                &isc.FungibleTokens{},
			AdjustToMinimumStorageDeposit: true,
			Metadata:                      &isc.SendMetadata{},
			Options:                       isc.SendOptions{},
		}, nftID)
	}
	return nil
}

// just claims everything from allowance and does nothing with it
// tests the "getData" sandbox call for every NFT sent in allowance
func claimAllowance(ctx isc.Sandbox) dict.Dict {
	initialNFTset := ctx.OwnedNFTs()
	allowance := ctx.AllowanceAvailable()
	ctx.TransferAllowedFunds(ctx.AccountID())
	ctx.Requiref(len(ctx.OwnedNFTs())-len(initialNFTset) == len(allowance.NFTs), "must get all NFTs from allowance")
	for _, id := range allowance.NFTs {
		nftData := ctx.GetNFTData(id)
		ctx.Requiref(!nftData.ID.Empty(), "must have NFTID")
		ctx.Requiref(len(nftData.Metadata) > 0, "must have metadata")
		ctx.Requiref(nftData.Issuer != nil, "must have issuer")
	}

	return nil
}

func sendLargeRequest(ctx isc.Sandbox) dict.Dict {
	req := isc.RequestParameters{
		TargetAddress: tpkg.RandEd25519Address(),
		Metadata: &isc.SendMetadata{
			EntryPoint:     isc.Hn("foo"),
			TargetContract: isc.Hn("bar"),
			Params:         dict.Dict{"x": make([]byte, ctx.Params().MustGetInt32(ParamSize))},
		},
		AdjustToMinimumStorageDeposit: true,
		FungibleTokens:                ctx.AllowanceAvailable().Assets,
	}
	storageDeposit := ctx.EstimateRequiredStorageDeposit(req)
	provided := ctx.AllowanceAvailable().Assets.BaseTokens
	if provided < storageDeposit {
		panic("not enough funds for storage deposit")
	}
	ctx.TransferAllowedFunds(ctx.AccountID(), isc.NewAllowanceBaseTokens(storageDeposit))
	req.FungibleTokens.BaseTokens = storageDeposit
	ctx.Send(req)
	return nil
}
