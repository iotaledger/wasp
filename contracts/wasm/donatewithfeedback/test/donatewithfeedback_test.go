// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/donatewithfeedback/go/donatewithfeedback"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) *wasmsolo.SoloContext {
	return wasmsolo.NewSoloContext(t, donatewithfeedback.ScName, donatewithfeedback.OnLoad)
}

func TestDeploy(t *testing.T) {
	ctx := setupTest(t)
	require.NoError(t, ctx.ContractExists(donatewithfeedback.ScName))
}

func TestStateAfterDeploy(t *testing.T) {
	ctx := setupTest(t)

	donationInfo := donatewithfeedback.ScFuncs.DonationInfo(ctx)
	donationInfo.Func.Call()

	require.EqualValues(t, 0, donationInfo.Results.Count().Value())
	require.EqualValues(t, 0, donationInfo.Results.MaxDonation().Value())
	require.EqualValues(t, 0, donationInfo.Results.TotalDonation().Value())
}

func TestDonateOnce(t *testing.T) {
	ctx := setupTest(t)
	accountBalance := ctx.Balance(ctx.Account())

	donator1 := ctx.NewSoloAgent()
	ctx.Chain.MustDepositIotasToL2(1000, donator1.Pair)
	require.EqualValues(t, 1000-100, ctx.Balance(donator1))

	donate := donatewithfeedback.ScFuncs.Donate(ctx.Sign(donator1))
	donate.Params.Feedback().SetValue("Nice work!")
	donate.Func.TransferIotas(1234).Post()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, 1000-100-ctx.GasFee, ctx.Balance(donator1))

	donationInfo := donatewithfeedback.ScFuncs.DonationInfo(ctx)
	donationInfo.Func.Call()

	require.EqualValues(t, 1, donationInfo.Results.Count().Value())
	require.EqualValues(t, 1234, donationInfo.Results.MaxDonation().Value())
	require.EqualValues(t, 1234, donationInfo.Results.TotalDonation().Value())

	// 1234 iota transferred from wallet to contract
	require.EqualValues(t, solo.Saldo-1000-1234, donator1.Balance())
	require.EqualValues(t, accountBalance+1234, ctx.Balance(ctx.Account()))
}

func TestDonateTwice(t *testing.T) {
	ctx := setupTest(t)
	accountBalance := ctx.Balance(ctx.Account())

	donator1 := ctx.NewSoloAgent()
	ctx.Chain.MustDepositIotasToL2(1000, donator1.Pair)
	require.EqualValues(t, 1000-100, ctx.Balance(donator1))

	donate := donatewithfeedback.ScFuncs.Donate(ctx.Sign(donator1))
	donate.Params.Feedback().SetValue("Nice work!")
	donate.Func.TransferIotas(1234).Post()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, 1000-100-ctx.GasFee, ctx.Balance(donator1))

	donator2 := ctx.NewSoloAgent()
	ctx.Chain.MustDepositIotasToL2(1000, donator2.Pair)
	require.EqualValues(t, 1000-100, ctx.Balance(donator2))

	donate2 := donatewithfeedback.ScFuncs.Donate(ctx.Sign(donator2))
	donate2.Params.Feedback().SetValue("Nice work!")
	donate2.Func.TransferIotas(2345).Post()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, 1000-100-ctx.GasFee, ctx.Balance(donator2))

	donationInfo := donatewithfeedback.ScFuncs.DonationInfo(ctx)
	donationInfo.Func.Call()

	require.EqualValues(t, 2, donationInfo.Results.Count().Value())
	require.EqualValues(t, 2345, donationInfo.Results.MaxDonation().Value())
	require.EqualValues(t, 3579, donationInfo.Results.TotalDonation().Value())

	// 1234 iota transferred from wallet to contract
	require.EqualValues(t, solo.Saldo-1000-1234, donator1.Balance())
	require.EqualValues(t, solo.Saldo-1000-2345, donator2.Balance())
	require.EqualValues(t, accountBalance+1234+2345, ctx.Balance(ctx.Account()))
}
