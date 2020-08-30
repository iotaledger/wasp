package dwfclient

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback"
	"time"
)

type DWFClient struct {
	nodeClient nodeclient.NodeClient
	waspHost   string
	scAddress  *address.Address
	sigScheme  signaturescheme.SignatureScheme
}

func NewClient(nodeClient nodeclient.NodeClient, waspHost string, scAddress *address.Address, sigScheme signaturescheme.SignatureScheme) *DWFClient {
	return &DWFClient{nodeClient, waspHost, scAddress, sigScheme}
}

type DonateParams struct {
	Amount            int64
	Feedback          string
	WaitForCompletion bool
	PublisherHosts    []string
	PublisherQuorum   int
	Timeout           time.Duration
}

func (client *DWFClient) Donate(par DonateParams) (*sctransaction.Transaction, error) {
	return apilib.CreateRequestTransaction(apilib.CreateRequestTransactionParams{
		NodeClient:      client.nodeClient,
		SenderSigScheme: client.sigScheme,
		BlockParams: []apilib.RequestBlockParams{{
			TargetSCAddress: client.scAddress,
			RequestCode:     donatewithfeedback.RequestDonate,
			Transfer: map[balance.Color]int64{
				balance.ColorIOTA: par.Amount,
			},
			Vars: map[string]interface{}{
				donatewithfeedback.VarReqFeedback: par.Feedback,
			}},
		},
		Post:                true,
		WaitForConfirmation: par.WaitForCompletion,
		WaitForCompletion:   par.WaitForCompletion,
		PublisherHosts:      par.PublisherHosts,
		PublisherQuorum:     par.PublisherQuorum,
		Timeout:             par.Timeout,
	})
}

type HarvestParams struct {
	Amount            int64
	WaitForCompletion bool
	PublisherHosts    []string
	PublisherQuorum   int
	Timeout           time.Duration
}

func (client *DWFClient) Harvest(par HarvestParams) (*sctransaction.Transaction, error) {
	return apilib.CreateRequestTransaction(apilib.CreateRequestTransactionParams{
		NodeClient:      client.nodeClient,
		SenderSigScheme: client.sigScheme,
		BlockParams: []apilib.RequestBlockParams{{
			TargetSCAddress: client.scAddress,
			RequestCode:     donatewithfeedback.RequestHarvest,
			Vars: map[string]interface{}{
				donatewithfeedback.VarReqHarvestSum: par.Amount,
			},
		}},
		Post:                true,
		WaitForConfirmation: par.WaitForCompletion,
		WaitForCompletion:   par.WaitForCompletion,
		PublisherHosts:      par.PublisherHosts,
		PublisherQuorum:     par.PublisherQuorum,
		Timeout:             par.Timeout,
	})
}

type Status struct {
	SCBalance map[balance.Color]int64
	FetchedAt time.Time

	NumRecords     int64
	FirstDonated   time.Time
	LastDonated    time.Time
	MaxDonation    int64
	MinDonation    int64
	TotalDonations int64
}

func (client *DWFClient) FetchStatus() (*Status, error) {
	status := &Status{
		FetchedAt: time.Now().UTC(),
	}

	scBalance, err := client.fetchSCBalance()
	if err != nil {
		return nil, err
	}
	status.SCBalance = scBalance

	return status, nil
}

func (client *DWFClient) fetchSCBalance() (map[balance.Color]int64, error) {
	outs, err := client.nodeClient.GetAccountOutputs(client.scAddress)
	if err != nil {
		return nil, err
	}
	ret, _ := util.OutputBalancesByColor(outs)
	return ret, nil
}
