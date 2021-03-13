package chainclient

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/level1"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/sctransaction_old"
)

// Client allows to send webapi requests to a specific chain in the node
type Client struct {
	Level1Client level1.Level1Client
	WaspClient   *client.WaspClient
	ChainID      coretypes.ChainID
	SigScheme    signaturescheme.SignatureScheme
}

// New creates a new chainclient.Client
func New(
	level1Client level1.Level1Client,
	waspClient *client.WaspClient,
	chainID coretypes.ChainID,
	sigScheme signaturescheme.SignatureScheme,
) *Client {
	return &Client{
		Level1Client: level1Client,
		WaspClient:   waspClient,
		ChainID:      chainID,
		SigScheme:    sigScheme,
	}
}

type PostRequestParams struct {
	Transfer coretypes.ColoredBalancesOld
	Args     requestargs.RequestArgs
}

// PostRequest sends a request transaction to the chain
func (c *Client) PostRequest(
	contractHname coretypes.Hname,
	entryPoint coretypes.Hname,
	params ...PostRequestParams,
) (*sctransaction_old.TransactionEssence, error) {
	par := PostRequestParams{}
	if len(params) > 0 {
		par = params[0]
	}

	return apilib.CreateRequestTransaction(apilib.CreateRequestTransactionParams{
		Level1Client:    c.Level1Client,
		SenderSigScheme: c.SigScheme,
		RequestSectionParams: []apilib.RequestSectionParams{{
			TargetContractID: coretypes.NewContractID(c.ChainID, contractHname),
			EntryPointCode:   entryPoint,
			Transfer:         par.Transfer,
			Args:             par.Args,
		}},
		Post: true,
	})
}
