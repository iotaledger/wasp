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
	for !ctx.AllowanceAvailable().IsEmpty() && ctx.AllowanceAvailable().Iotas >= 200 {
		// claim 200 iotas from allowance at a time
		// send back to caller's address
		// depending on the amount of iotas, it will exceed number of outputs or not
		ctx.TransferAllowedFunds(ctx.AccountID(), iscp.NewAssetsIotas(200))
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
	ctx.TransferAllowedFunds(ctx.AccountID(), iscp.NewAssetsIotas(ctx.AllowanceAvailable().Iotas))
	for _, token := range ctx.AllowanceAvailable().Tokens {
		for ctx.AllowanceAvailable().AmountNativeToken(&token.ID).Cmp(util.Big0) > 0 {
			// claim 1 token from allowance at a time
			// send back to caller's address
			// depending on the amount of tokens, it will exceed number of outputs or not
			transfer := iscp.NewEmptyAssets().AddNativeTokens(token.ID, 1)
			rem := ctx.TransferAllowedFunds(ctx.AccountID(), transfer)
			fmt.Printf("%s\n", rem)
			ctx.Send(
				iscp.RequestParameters{
					TargetAddress:              ctx.Caller().Address(),
					Assets:                     transfer,
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
	toSend := ctx.AllowanceAvailable()
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
	provided := ctx.AllowanceAvailable().Iotas

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

func sendLargeRequest(ctx iscp.Sandbox) dict.Dict {
	req := iscp.RequestParameters{
		TargetAddress: tpkg.RandEd25519Address(),
		Metadata: &iscp.SendMetadata{
			EntryPoint:     iscp.Hn("foo"),
			TargetContract: iscp.Hn("bar"),
			Params:         dict.Dict{"x": make([]byte, ctx.Params().MustGetInt32(ParamSize))},
		},
		AdjustToMinimumDustDeposit: true,
		Assets:                     ctx.AllowanceAvailable(),
	}
	dust := ctx.EstimateRequiredDustDeposit(req)
	provided := ctx.AllowanceAvailable().Iotas
	if provided < dust {
		panic("not enough funds for dust")
	}
	ctx.TransferAllowedFunds(ctx.AccountID(), iscp.NewAssetsIotas(dust))
	req.Assets.Iotas = dust
	ctx.Send(req)
	return nil
}
