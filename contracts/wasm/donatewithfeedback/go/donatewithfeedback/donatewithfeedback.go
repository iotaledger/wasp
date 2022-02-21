// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package donatewithfeedback

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

func funcDonate(ctx wasmlib.ScFuncContext, f *DonateContext) {
	amount := ctx.Allowance().Balance(wasmtypes.IOTA)
	transfer := wasmlib.NewScTransfer(wasmtypes.IOTA, amount)
	ctx.TransferAllowed(ctx.AccountID(), transfer, false)
	donation := &Donation{
		Amount:    amount,
		Donator:   ctx.Caller(),
		Error:     "",
		Feedback:  f.Params.Feedback().Value(),
		Timestamp: ctx.Timestamp(),
	}
	if donation.Amount == 0 || donation.Feedback == "" {
		donation.Error = "error: empty feedback or donated amount = 0"
	}
	log := f.State.Log()
	log.AppendDonation().SetValue(donation)

	largestDonation := f.State.MaxDonation()
	totalDonated := f.State.TotalDonation()
	if donation.Amount > largestDonation.Value() {
		largestDonation.SetValue(donation.Amount)
	}
	totalDonated.SetValue(totalDonated.Value() + donation.Amount)
}

func funcWithdraw(ctx wasmlib.ScFuncContext, f *WithdrawContext) {
	balance := ctx.Balances().Balance(wasmtypes.IOTA)
	amount := f.Params.Amount().Value()
	if amount == 0 || amount > balance {
		amount = balance
	}
	if amount == 0 {
		ctx.Log("dwf.withdraw: nothing to withdraw")
		return
	}

	scCreator := ctx.ContractCreator().Address()
	ctx.Send(scCreator, wasmlib.NewScTransferIotas(amount))
}

func viewDonation(ctx wasmlib.ScViewContext, f *DonationContext) {
	nr := f.Params.Nr().Value()
	donation := f.State.Log().GetDonation(nr).Value()
	f.Results.Amount().SetValue(donation.Amount)
	f.Results.Donator().SetValue(donation.Donator)
	f.Results.Error().SetValue(donation.Error)
	f.Results.Feedback().SetValue(donation.Feedback)
	f.Results.Timestamp().SetValue(donation.Timestamp)
}

func viewDonationInfo(ctx wasmlib.ScViewContext, f *DonationInfoContext) {
	f.Results.MaxDonation().SetValue(f.State.MaxDonation().Value())
	f.Results.TotalDonation().SetValue(f.State.TotalDonation().Value())
	f.Results.Count().SetValue(f.State.Log().Length())
}
