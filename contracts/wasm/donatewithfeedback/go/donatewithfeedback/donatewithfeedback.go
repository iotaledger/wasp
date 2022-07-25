// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package donatewithfeedback

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
)

func funcDonate(ctx wasmlib.ScFuncContext, f *DonateContext) {
	amount := ctx.Allowance().Iotas()
	transfer := wasmlib.NewScTransferIotas(amount)
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
	balance := ctx.Balances().Iotas()
	amount := f.Params.Amount().Value()
	if amount == 0 || amount > balance {
		amount = balance
	}
	if amount == 0 {
		ctx.Log("dwf.withdraw: nothing to withdraw")
		return
	}

	scOwner := f.State.Owner().Value().Address()
	ctx.Send(scOwner, wasmlib.NewScTransferIotas(amount))
}

func viewDonation(_ wasmlib.ScViewContext, f *DonationContext) {
	nr := f.Params.Nr().Value()
	donation := f.State.Log().GetDonation(nr).Value()
	f.Results.Amount().SetValue(donation.Amount)
	f.Results.Donator().SetValue(donation.Donator)
	f.Results.Error().SetValue(donation.Error)
	f.Results.Feedback().SetValue(donation.Feedback)
	f.Results.Timestamp().SetValue(donation.Timestamp)
}

func viewDonationInfo(_ wasmlib.ScViewContext, f *DonationInfoContext) {
	f.Results.MaxDonation().SetValue(f.State.MaxDonation().Value())
	f.Results.TotalDonation().SetValue(f.State.TotalDonation().Value())
	f.Results.Count().SetValue(f.State.Log().Length())
}

func funcInit(ctx wasmlib.ScFuncContext, f *InitContext) {
	if f.Params.Owner().Exists() {
		f.State.Owner().SetValue(f.Params.Owner().Value())
		return
	}
	f.State.Owner().SetValue(ctx.RequestSender())
}
