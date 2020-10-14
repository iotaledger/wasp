package wasptest2

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client/scclient"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback/dwfclient"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback/dwfimpl"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/mr-tron/base58"
)

func TestDeployDWF(t *testing.T) {
	var seed [32]byte
	rand.Read(seed[:])
	seed58 := base58.Encode(seed[:])
	wallet1 := testutil.NewWallet(seed58)
	scOwner = wallet1.WithIndex(0)
	scOwnerAddr := scOwner.Address()

	programHash, err := hashing.HashValueFromBase58(dwfimpl.ProgramHash)
	check(err, t)

	// setup
	wasps := setup(t, "test_cluster2", "TestDeployDWF")

	err = wasps.NodeClient.RequestFunds(scOwnerAddr)
	check(err, t)

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "sc owner in the beginning") {
		t.Fail()
		return
	}

	scAddr, scColor, err := waspapi.CreateSC(waspapi.CreateSCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHash,
		Description:           dwfimpl.Description,
		Textout:               os.Stdout,
		Prefix:                "[deploy] ",
	})
	checkSuccess(err, t, "smart contract has been created")

	err = waspapi.ActivateSCMulti(waspapi.ActivateSCParams{
		Addresses:         []*address.Address{scAddr},
		ApiHosts:          wasps.ApiHosts(),
		WaitForCompletion: true,
		PublisherHosts:    wasps.PublisherHosts(),
		Timeout:           20 * time.Second,
	})
	checkSuccess(err, t, "smart contract has been activated and initialized")

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
		return
	}

	if !wasps.VerifyAddressBalances(scAddr, 1, map[balance.Color]int64{
		*scColor: 1,
	}, "sc in the end") {
		t.Fail()
		return
	}

	ph, err := hashing.HashValueFromBase58(dwfimpl.ProgramHash)
	check(err, t)

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  ph[:],
		vmconst.VarNameDescription:  dwfimpl.Description,
	}) {
		t.Fail()
	}
}

const numDonations = 5

func TestDWFDonateNTimes(t *testing.T) {
	var seed [32]byte
	rand.Read(seed[:])
	seed58 := base58.Encode(seed[:])
	wallet := testutil.NewWallet(seed58)
	scOwner := wallet.WithIndex(0)
	scOwnerAddr := scOwner.Address()
	donor := wallet.WithIndex(1)
	donorAddr := donor.Address()

	programHash, err := hashing.HashValueFromBase58(dwfimpl.ProgramHash)
	check(err, t)

	// setup
	wasps := setup(t, "test_cluster2", "TestDeployDWF")

	err = wasps.NodeClient.RequestFunds(scOwnerAddr)
	check(err, t)

	err = wasps.NodeClient.RequestFunds(donorAddr)
	check(err, t)

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "sc owner in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "donor in the beginning") {
		t.Fail()
		return
	}

	scAddr, scColor, err := waspapi.CreateSC(waspapi.CreateSCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHash,
		Description:           dwfimpl.Description,
		Textout:               os.Stdout,
		Prefix:                "[deploy] ",
	})
	checkSuccess(err, t, "smart contract has been created")

	err = waspapi.ActivateSCMulti(waspapi.ActivateSCParams{
		Addresses:         []*address.Address{scAddr},
		ApiHosts:          wasps.ApiHosts(),
		WaitForCompletion: true,
		PublisherHosts:    wasps.PublisherHosts(),
		Timeout:           20 * time.Second,
	})
	checkSuccess(err, t, "smart contract has been activated and initialized")

	dwfClient := dwfclient.NewClient(scclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scAddr,
		donor.SigScheme(),
	))

	for i := 0; i < numDonations; i++ {
		_, err = dwfClient.Donate(42, fmt.Sprintf("Donation #%d: well done, I give you 42 iotas", i))
		check(err, t)
		time.Sleep(1 * time.Second)
		checkSuccess(err, t, "donated 42")
	}

	status, err := dwfClient.FetchStatus()
	text := ""
	if err == nil {
		text = fmt.Sprintf("[test] Status fetched succesfully: num rec: %d, "+
			"total donations: %d, max donation: %d, last donation: %v, num rec returned: %d\n",
			status.NumRecords,
			status.TotalDonations,
			status.MaxDonation,
			status.LastDonated,
			len(status.LastRecordsDesc),
		)
		for i, di := range status.LastRecordsDesc {
			text += fmt.Sprintf("           %d: ts: %s, amount: %d, fb: '%s', donor: %s, err:%v\n",
				i,
				di.When.Format("2006-01-02 15:04:05"),
				di.Amount,
				di.Feedback,
				di.Sender.String(),
				di.Error,
			)
		}
	}
	checkSuccess(err, t, text)

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddr, 1+42*numDonations, map[balance.Color]int64{
		*scColor:          1,
		balance.ColorIOTA: 42 * numDonations,
	}, "sc in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount-42*numDonations, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 42*numDonations,
	}, "donor in the end") {
		t.Fail()
	}

	ph, err := hashing.HashValueFromBase58(dwfimpl.ProgramHash)
	check(err, t)

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress:               scOwnerAddr[:],
		vmconst.VarNameProgramHash:                ph[:],
		vmconst.VarNameDescription:                dwfimpl.Description,
		donatewithfeedback.VarStateMaxDonation:    42,
		donatewithfeedback.VarStateTotalDonations: 42 * numDonations,
	}) {
		t.Fail()
	}
}

func TestDWFDonateWithdrawAuthorised(t *testing.T) {
	var seed [32]byte
	rand.Read(seed[:])
	seed58 := base58.Encode(seed[:])
	wallet := testutil.NewWallet(seed58)
	scOwner := wallet.WithIndex(0)
	scOwnerAddr := scOwner.Address()
	donor := wallet.WithIndex(1)
	donorAddr := donor.Address()

	programHash, err := hashing.HashValueFromBase58(dwfimpl.ProgramHash)
	check(err, t)

	// setup
	wasps := setup(t, "test_cluster2", "TestDeployDWF")

	err = wasps.NodeClient.RequestFunds(scOwnerAddr)
	check(err, t)

	err = wasps.NodeClient.RequestFunds(donorAddr)
	check(err, t)

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "sc owner in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "donor in the beginning") {
		t.Fail()
		return
	}

	scAddr, scColor, err := waspapi.CreateSC(waspapi.CreateSCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHash,
		Description:           dwfimpl.Description,
		Textout:               os.Stdout,
		Prefix:                "[deploy] ",
	})
	checkSuccess(err, t, "smart contract has been created")

	err = waspapi.ActivateSCMulti(waspapi.ActivateSCParams{
		Addresses:         []*address.Address{scAddr},
		ApiHosts:          wasps.ApiHosts(),
		WaitForCompletion: true,
		PublisherHosts:    wasps.PublisherHosts(),
		Timeout:           30 * time.Second,
	})
	checkSuccess(err, t, "smart contract has been activated and initialized")

	dwfDonorClient := dwfclient.NewClient(scclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scAddr,
		donor.SigScheme(),
	))
	_, err = dwfDonorClient.Donate(42, "well done, I give you 42i")
	check(err, t)
	checkSuccess(err, t, "donated 42")

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddr, 1+42, map[balance.Color]int64{
		*scColor:          1,
		balance.ColorIOTA: 42,
	}, "sc in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount-42, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 42,
	}, "donor in the end") {
		t.Fail()
	}

	dwfOwnerClient := dwfclient.NewClient(scclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scAddr,
		scOwner.SigScheme(),
	))
	_, err = dwfOwnerClient.Withdraw(40)
	check(err, t)
	checkSuccess(err, t, "harvested 40")

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1+40, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1 + 40,
	}, "sc owner in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddr, 1+2, map[balance.Color]int64{
		*scColor:          1,
		balance.ColorIOTA: 2,
	}, "sc in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount-42, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 42,
	}, "donor in the end") {
		t.Fail()
	}

	ph, err := hashing.HashValueFromBase58(dwfimpl.ProgramHash)
	check(err, t)

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  ph[:],
		vmconst.VarNameDescription:  dwfimpl.Description,
	}) {
		t.Fail()
	}
}

func TestDWFDonateWithdrawNotAuthorised(t *testing.T) {
	var seed [32]byte
	rand.Read(seed[:])
	seed58 := base58.Encode(seed[:])
	wallet := testutil.NewWallet(seed58)
	scOwner := wallet.WithIndex(0)
	scOwnerAddr := scOwner.Address()
	donor := wallet.WithIndex(1)
	donorAddr := donor.Address()

	programHash, err := hashing.HashValueFromBase58(dwfimpl.ProgramHash)
	check(err, t)

	// setup
	wasps := setup(t, "test_cluster2", "TestDeployDWF")

	err = wasps.NodeClient.RequestFunds(scOwnerAddr)
	check(err, t)

	err = wasps.NodeClient.RequestFunds(donorAddr)
	check(err, t)

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "sc owner in the beginning") {
		t.Fail()
		return
	}
	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount,
	}, "donor in the beginning") {
		t.Fail()
		return
	}

	scAddr, scColor, err := waspapi.CreateSC(waspapi.CreateSCParams{
		Node:                  wasps.NodeClient,
		CommitteeApiHosts:     wasps.ApiHosts(),
		CommitteePeeringHosts: wasps.PeeringHosts(),
		N:                     4,
		T:                     3,
		OwnerSigScheme:        scOwner.SigScheme(),
		ProgramHash:           programHash,
		Description:           dwfimpl.Description,
		Textout:               os.Stdout,
		Prefix:                "[deploy] ",
	})
	checkSuccess(err, t, "smart contract has been created")

	err = waspapi.ActivateSCMulti(waspapi.ActivateSCParams{
		Addresses:         []*address.Address{scAddr},
		ApiHosts:          wasps.ApiHosts(),
		WaitForCompletion: true,
		PublisherHosts:    wasps.PublisherHosts(),
		Timeout:           20 * time.Second,
	})
	checkSuccess(err, t, "smart contract has been activated and initialized")

	dwfDonorClient := dwfclient.NewClient(scclient.New(
		wasps.NodeClient,
		wasps.WaspClient(0),
		scAddr,
		donor.SigScheme(),
	))
	_, err = dwfDonorClient.Donate(42, "well done, I give you 42i")
	check(err, t)
	checkSuccess(err, t, "donated 42")

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddr, 1+42, map[balance.Color]int64{
		*scColor:          1,
		balance.ColorIOTA: 42,
	}, "sc in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount-42, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 42,
	}, "donor in the end") {
		t.Fail()
	}

	// donor want to take back. Not authorised
	_, err = dwfDonorClient.Withdraw(40)
	check(err, t)
	checkSuccess(err, t, "sent harvest 40")

	if !wasps.VerifyAddressBalances(scOwnerAddr, testutil.RequestFundsAmount-1, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 1,
	}, "sc owner in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(scAddr, 1+42, map[balance.Color]int64{
		*scColor:          1,
		balance.ColorIOTA: 42,
	}, "sc in the end") {
		t.Fail()
	}

	if !wasps.VerifyAddressBalances(donorAddr, testutil.RequestFundsAmount-42, map[balance.Color]int64{
		balance.ColorIOTA: testutil.RequestFundsAmount - 42,
	}, "donor in the end") {
		t.Fail()
	}

	ph, err := hashing.HashValueFromBase58(dwfimpl.ProgramHash)
	check(err, t)

	if !wasps.VerifySCStateVariables2(scAddr, map[kv.Key]interface{}{
		vmconst.VarNameOwnerAddress: scOwnerAddr[:],
		vmconst.VarNameProgramHash:  ph[:],
		vmconst.VarNameDescription:  dwfimpl.Description,
	}) {
		t.Fail()
	}
}
