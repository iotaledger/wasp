package dwfclient

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
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
