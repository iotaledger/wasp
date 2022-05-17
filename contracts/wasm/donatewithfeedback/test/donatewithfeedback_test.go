// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/donatewithfeedback/go/donatewithfeedback"
	"github.com/iotaledger/wasp/packages/iscp"
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

	donator1 := ctx.NewSoloAgent()
	donator1L1 := donator1.Balance()
	bal := ctx.Balances(donator1)

	donate := donatewithfeedback.ScFuncs.Donate(ctx.Sign(donator1))
	donate.Params.Feedback().SetValue("Nice work!")
	iotasToSend := 1 * iscp.Mi
	donate.Func.TransferIotas(iotasToSend).Post()
	require.NoError(t, ctx.Err)

	bal.Account += iotasToSend
	bal.Chain += ctx.GasFee
	bal.Add(donator1, -ctx.GasFee)
	bal.VerifyBalances(t)
	require.EqualValues(t, donator1L1-iotasToSend, donator1.Balance())

	donationInfo := donatewithfeedback.ScFuncs.DonationInfo(ctx)
	donationInfo.Func.Call()

	require.EqualValues(t, 1, donationInfo.Results.Count().Value())
	require.EqualValues(t, iotasToSend, donationInfo.Results.MaxDonation().Value())
	require.EqualValues(t, iotasToSend, donationInfo.Results.TotalDonation().Value())
}

func TestDonateTwice(t *testing.T) {
	ctx := setupTest(t)

	donator1 := ctx.NewSoloAgent()
	donator2 := ctx.NewSoloAgent()
	donator1L1 := donator1.Balance()
	donator2L1 := donator2.Balance()
	bal := ctx.Balances(donator1, donator2)

	donate := donatewithfeedback.ScFuncs.Donate(ctx.Sign(donator1))
	donate.Params.Feedback().SetValue("Nice work!")
	iotasToSend1 := 1 * iscp.Mi
	donate.Func.TransferIotas(iotasToSend1).Post()
	require.NoError(t, ctx.Err)

	bal.Account += iotasToSend1
	bal.Chain += ctx.GasFee
	bal.Add(donator1, -ctx.GasFee)
	bal.VerifyBalances(t)
	require.EqualValues(t, donator1L1-iotasToSend1, donator1.Balance())

	donate2 := donatewithfeedback.ScFuncs.Donate(ctx.Sign(donator2))
	donate2.Params.Feedback().SetValue("Nice work!")
	iotasToSend2 := 2 * iscp.Mi
	donate2.Func.TransferIotas(iotasToSend2).Post()
	require.NoError(t, ctx.Err)

	bal.Account += iotasToSend2
	bal.Chain += ctx.GasFee
	bal.Add(donator2, -ctx.GasFee)
	bal.VerifyBalances(t)
	require.EqualValues(t, donator2L1-iotasToSend2, donator2.Balance())

	donationInfo := donatewithfeedback.ScFuncs.DonationInfo(ctx)
	donationInfo.Func.Call()

	require.EqualValues(t, 2, donationInfo.Results.Count().Value())
	require.EqualValues(t, iotasToSend2, donationInfo.Results.MaxDonation().Value())
	require.EqualValues(t, iotasToSend1+iotasToSend2, donationInfo.Results.TotalDonation().Value())
}
