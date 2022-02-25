package sbtestsc

import (
	"fmt"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

// testSplitFunds calls Send in a loop by sending 200 iotas back to the caller
func testSplitFunds(ctx iscp.Sandbox) dict.Dict {
	for !ctx.AllowanceAvailable().IsEmpty() && ctx.AllowanceAvailable().Assets.Iotas >= 200 {
		// claim 200 iotas from allowance at a time
		// send back to caller's address
		// depending on the amount of iotas, it will exceed number of outputs or not
		ctx.TransferAllowedFunds(ctx.AccountID(), iscp.NewAllowance(200, nil, nil))
		ctx.Send(
			iscp.RequestParameters{
				TargetAddress: ctx.Caller().Address(),
				Assets:        iscp.NewAssetsIotas(200),
			},
		)
	}
	return nil
}

// testSplitFundsNativeTokens calls Send for each Native token
func testSplitFundsNativeTokens(ctx iscp.Sandbox) dict.Dict {
	// claims all iotas from allowance
	ctx.TransferAllowedFunds(ctx.AccountID(), iscp.NewAllowance(ctx.AllowanceAvailable().Assets.Iotas, nil, nil))
	for _, token := range ctx.AllowanceAvailable().Assets.Tokens {
		for ctx.AllowanceAvailable().Assets.AmountNativeToken(&token.ID).Cmp(util.Big0) > 0 {
			// claim 1 token from allowance at a time
			// send back to caller's address
			// depending on the amount of tokens, it will exceed number of outputs or not
			assets := iscp.NewEmptyAssets().AddNativeTokens(token.ID, 1)
			transfer := iscp.NewAllowanceFromAssets(assets, nil)
			rem := ctx.TransferAllowedFunds(ctx.AccountID(), transfer)
			fmt.Printf("%s\n", rem)
			ctx.Send(
				iscp.RequestParameters{
					TargetAddress:              ctx.Caller().Address(),
					Assets:                     assets,
					AdjustToMinimumDustDeposit: true,
				},
			)
		}
	}
	return nil
}

func pingAllowanceBack(ctx iscp.Sandbox) dict.Dict {
	// assert caller is L1 address, not a SC
	ctx.Requiref(!ctx.Caller().Address().Equal(ctx.ChainID().AsAddress()) && ctx.Caller().Hname() == 0,
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
		iscp.RequestParameters{
			TargetAddress: ctx.Caller().Address(),
			Assets:        toSend,
		},
	)
	return nil
}

// testEstimateMinimumDust returns true if the provided allowance is enough to pay for a L1 request, panics otherwise
func testEstimateMinimumDust(ctx iscp.Sandbox) dict.Dict {
	provided := ctx.AllowanceAvailable().Assets.Iotas

	requestParams := iscp.RequestParameters{
		TargetAddress: tpkg.RandEd25519Address(),
		Metadata: &iscp.SendMetadata{
			EntryPoint:     iscp.Hn("foo"),
			TargetContract: iscp.Hn("bar"),
		},
		AdjustToMinimumDustDeposit: true,
	}

	required := ctx.EstimateRequiredDustDeposit(requestParams)
	if provided < required {
		panic("not enough funds")
	}
	return nil
}

// tries to sendback whaever NFTs are specified in allowance
func sendNFTsBack(ctx iscp.Sandbox) dict.Dict {
	allowance := ctx.AllowanceAvailable()
	ctx.TransferAllowedFunds(ctx.AccountID())
	for _, nftID := range allowance.NFTs {
		ctx.Send(iscp.RequestParameters{
			TargetAddress:              ctx.Caller().Address(),
			Assets:                     &iscp.Assets{},
			AdjustToMinimumDustDeposit: true,
			Metadata:                   &iscp.SendMetadata{},
			Options:                    iscp.SendOptions{},
			NFTID:                      nftID,
		})
	}
	return nil
}
