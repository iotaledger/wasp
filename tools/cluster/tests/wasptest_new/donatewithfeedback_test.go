package wasptest

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/stretchr/testify/require"
	"testing"
)

const dwfName = "donatewithfeedback"

var dwfFile = wasmhost.WasmPath("donatewithfeedback_bg.wasm")

const dwfDescription = "Donate with feedback, a PoC smart contract"

var dwfHname = coretypes.Hn(dwfName)

func TestDwfDonateOnce(t *testing.T) {
	const numDonations = 1
	al := solo.New(t, false, false)
	chain := al.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, dwfName, dwfFile)
	require.NoError(t, err)

	for i := 0; i < numDonations; i++ {
		feedback := fmt.Sprintf("Donation #%d: well done, I give you 42 iotas", i)
		req := solo.NewCallParams(dwfName, "donate", "feedback", feedback).
			WithTransfer(balance.ColorIOTA, 42)
		_, err = chain.PostRequest(req, nil)
		require.NoError(t, err)
	}

	ret, err := chain.CallView(dwfName, "donations")
	require.NoError(t, err)
	largest, _, err := codec.DecodeInt64(ret.MustGet("maxDonation"))
	check(err, t)
	require.EqualValues(t, 42, largest)
	total, _, err := codec.DecodeInt64(ret.MustGet("totalDonation"))
	check(err, t)
	require.EqualValues(t, 42*numDonations, total)

	//donations := collections.NewMustArray(ret, "donations")
	//for i := uint16(0); i < donations.Len(); i++ {
	//	donation := donations.GetAt(i)
	//	_ = donation
	//	check(err, t)
	//}

	////TODO make sure record encoding is the same
	//if *useWasp {
	//	status, err := dwfClient.FetchStatus()
	//	text := ""
	//	if err == nil {
	//		text = fmt.Sprintf("[test] Status fetched succesfully: num rec: %d, "+
	//			"total donations: %d, max donation: %d, last donation: %v, num rec returned: %d\n",
	//			status.NumRecords,
	//			status.TotalDonations,
	//			status.MaxDonation,
	//			status.LastDonated,
	//			len(status.LastRecordsDesc),
	//		)
	//		for i, di := range status.LastRecordsDesc {
	//			text += fmt.Sprintf("           %d: ts: %s, amount: %d, fb: '%s', donor: %s, err:%v\n",
	//				i,
	//				di.When.Format("2006-01-02 15:04:05"),
	//				di.Amount,
	//				di.Feedback,
	//				di.Sender.String(),
	//				di.Error,
	//			)
	//		}
	//	}
	//	checkSuccess(err, t, text)
	//}
	//
	//if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
	//	balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	//}, "sc owner in the end") {
	//	t.Fail()
	//}
	//
	//if !wasps.VerifyAddressBalances(scAddr, 1+42*numDonations, map[balance.Color]int64{
	//	*scColor:          1,
	//	balance.ColorIOTA: 42 * numDonations,
	//}, "sc in the end") {
	//	t.Fail()
	//}
	//
	//if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount-42*numDonations, map[balance.Color]int64{
	//	balance.ColorIOTA: testutil.RequestFundsAmount - 42*numDonations,
	//}, "donor in the end") {
	//	t.Fail()
	//}
	//
	//if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
	//	vmconst.VarNameOwnerAddress:               scOwnerAddr[:],
	//	vmconst.VarNameProgramData:                programHash[:],
	//	vmconst.VarNameDescription:                dwfDescription,
	//	donatewithfeedback.VarStateMaxDonation:    42,
	//	donatewithfeedback.VarStateTotalDonations: 42 * numDonations,
	//}) {
	//	t.Fail()
	//}
}

//
//func TestDwfDonateNTimes(t *testing.T) {
//	const numDonations = 5
//
//	wasps := setup(t, "TestDwfDonateNTimes")
//
//	err := loadWasmIntoWasps(wasps, dwfWasmPath, dwfDescription)
//	check(err, t)
//
//	err = requestFunds(wasps, scOwnerAddr, "sc owner")
//	check(err, t)
//
//	donor := wallet.WithIndex(1)
//	donorAddr := donor.Address()
//	err = requestFunds(wasps, donorAddr, "donor")
//	check(err, t)
//
//	scChain, scAddr, scColor, err := startSmartContract(wasps, dwfimpl.ProgramHash, dwfDescription)
//	checkSuccess(err, t, "smart contract has been created and activated")
//
//	dwfClient := dwfclient.NewClient(chainclient.New(
//		wasps.Level1Client,
//		wasps.WaspClient(0),
//		scChain,
//		donor.SigScheme(),
//		15*time.Second,
//	), 0)
//
//	for i := 0; i < numDonations; i++ {
//		_, err = dwfClient.Donate(42, fmt.Sprintf("Donation #%d: well done, I give you 42 iotas", i))
//		check(err, t)
//		time.Sleep(1 * time.Second)
//		checkSuccess(err, t, "donated 42")
//	}
//
//	//TODO make sure record encoding is the same
//	if *useWasp {
//		status, err := dwfClient.FetchStatus()
//		text := ""
//		if err == nil {
//			text = fmt.Sprintf("[test] Status fetched succesfully: num rec: %d, "+
//				"total donations: %d, max donation: %d, last donation: %v, num rec returned: %d\n",
//				status.NumRecords,
//				status.TotalDonations,
//				status.MaxDonation,
//				status.LastDonated,
//				len(status.LastRecordsDesc),
//			)
//			for i, di := range status.LastRecordsDesc {
//				text += fmt.Sprintf("           %d: ts: %s, amount: %d, fb: '%s', donor: %s, err:%v\n",
//					i,
//					di.When.Format("2006-01-02 15:04:05"),
//					di.Amount,
//					di.Feedback,
//					di.Sender.String(),
//					di.Error,
//				)
//			}
//		}
//		checkSuccess(err, t, text)
//	}
//
//	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
//	}, "sc owner in the end") {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(scAddr, 1+42*numDonations, map[balance.Color]int64{
//		*scColor:          1,
//		balance.ColorIOTA: 42 * numDonations,
//	}, "sc in the end") {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount-42*numDonations, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 42*numDonations,
//	}, "donor in the end") {
//		t.Fail()
//	}
//
//	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
//		vmconst.VarNameOwnerAddress:               scOwnerAddr[:],
//		vmconst.VarNameProgramData:                programHash[:],
//		vmconst.VarNameDescription:                dwfDescription,
//		donatewithfeedback.VarStateMaxDonation:    42,
//		donatewithfeedback.VarStateTotalDonations: 42 * numDonations,
//	}) {
//		t.Fail()
//	}
//}
//
//func TestDwfDonateWithdrawAuthorised(t *testing.T) {
//	wasps := setup(t, "TestDwfDonateWithdrawAuthorised")
//
//	err := loadWasmIntoWasps(wasps, dwfWasmPath, dwfDescription)
//	check(err, t)
//
//	err = requestFunds(wasps, scOwnerAddr, "sc owner")
//	check(err, t)
//
//	donor := wallet.WithIndex(1)
//	donorAddr := donor.Address()
//	err = requestFunds(wasps, donorAddr, "donor")
//	check(err, t)
//
//	scChain, scAddr, scColor, err := startSmartContract(wasps, dwfimpl.ProgramHash, dwfDescription)
//	checkSuccess(err, t, "smart contract has been created and activated")
//
//	dwfDonorClient := dwfclient.NewClient(chainclient.New(
//		wasps.Level1Client,
//		wasps.WaspClient(0),
//		scChain,
//		donor.SigScheme(),
//		15*time.Second,
//	), 0)
//	_, err = dwfDonorClient.Donate(42, "well done, I give you 42i")
//	check(err, t)
//	checkSuccess(err, t, "donated 42")
//
//	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
//	}, "sc owner after donation") {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(scAddr, 1+42, map[balance.Color]int64{
//		*scColor:          1,
//		balance.ColorIOTA: 42,
//	}, "sc after donation") {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount-42, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 42,
//	}, "donor after donation") {
//		t.Fail()
//	}
//
//	dwfOwnerClient := dwfclient.NewClient(chainclient.New(
//		wasps.Level1Client,
//		wasps.WaspClient(0),
//		scChain,
//		scOwner.SigScheme(),
//		15*time.Second,
//	), 0)
//	_, err = dwfOwnerClient.Withdraw(40)
//	check(err, t)
//	checkSuccess(err, t, "harvested 40")
//
//	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1+40, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 1 + 40,
//	}, "sc owner after withdraw") {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(scAddr, 1+2, map[balance.Color]int64{
//		*scColor:          1,
//		balance.ColorIOTA: 2,
//	}, "sc after withdraw") {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount-42, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 42,
//	}, "donor after withdraw") {
//		t.Fail()
//	}
//
//	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
//		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
//		vmconst.VarNameProgramData:  programHash[:],
//		vmconst.VarNameDescription:  dwfDescription,
//	}) {
//		t.Fail()
//	}
//}
//
//func TestDwfDonateWithdrawNotAuthorised(t *testing.T) {
//	wasps := setup(t, "TestDwfDonateWithdrawNotAuthorised")
//
//	err := loadWasmIntoWasps(wasps, dwfWasmPath, dwfDescription)
//	check(err, t)
//
//	err = requestFunds(wasps, scOwnerAddr, "sc owner")
//	check(err, t)
//
//	donor := wallet.WithIndex(1)
//	donorAddr := donor.Address()
//	err = requestFunds(wasps, donorAddr, "donor")
//	check(err, t)
//
//	scChain, scAddr, scColor, err := startSmartContract(wasps, dwfimpl.ProgramHash, dwfDescription)
//	checkSuccess(err, t, "smart contract has been created and activated")
//
//	dwfDonorClient := dwfclient.NewClient(chainclient.New(
//		wasps.Level1Client,
//		wasps.WaspClient(0),
//		scChain,
//		donor.SigScheme(),
//		15*time.Second,
//	), 0)
//	_, err = dwfDonorClient.Donate(42, "well done, I give you 42i")
//	check(err, t)
//	checkSuccess(err, t, "donated 42")
//
//	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
//	}, "sc owner in the end") {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(scAddr, 1+42, map[balance.Color]int64{
//		*scColor:          1,
//		balance.ColorIOTA: 42,
//	}, "sc in the end") {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount-42, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 42,
//	}, "donor in the end") {
//		t.Fail()
//	}
//
//	// donor want to take back. Not authorised
//	_, err = dwfDonorClient.Withdraw(40)
//	check(err, t)
//	checkSuccess(err, t, "sent harvest 40")
//
//	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
//	}, "sc owner in the end") {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(scAddr, 1+42, map[balance.Color]int64{
//		*scColor:          1,
//		balance.ColorIOTA: 42,
//	}, "sc in the end") {
//		t.Fail()
//	}
//
//	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount-42, map[balance.Color]int64{
//		balance.ColorIOTA: testutil.RequestFundsAmount - 42,
//	}, "donor in the end") {
//		t.Fail()
//	}
//
//	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
//		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
//		vmconst.VarNameProgramData:  programHash[:],
//		vmconst.VarNameDescription:  dwfDescription,
//	}) {
//		t.Fail()
//	}
//}
