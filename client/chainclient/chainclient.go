package chainclient

import (
	"github.com/iotaledger/wasp/packages/coretypes"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/nodeclient"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

type Client struct {
	NodeClient nodeclient.NodeClient
	WaspClient *client.WaspClient
	ChainID    coretypes.ChainID
	SigScheme  signaturescheme.SignatureScheme
}

func New(
	nodeClient nodeclient.NodeClient,
	waspClient *client.WaspClient,
	chainID coretypes.ChainID,
	sigScheme signaturescheme.SignatureScheme,
) *Client {
	return &Client{
		NodeClient: nodeClient,
		WaspClient: waspClient,
		ChainID:    chainID,
		SigScheme:  sigScheme,
	}
}

func (c *Client) PostRequest(
	contractHname coretypes.Hname,
	entryPoint coretypes.Hname,
	mint map[address.Address]int64, // TODO
	transfer map[balance.Color]int64,
	vars map[string]interface{},
) (*sctransaction.Transaction, error) {
	return apilib.CreateRequestTransaction(apilib.CreateRequestTransactionParams{
		NodeClient:      c.NodeClient,
		SenderSigScheme: c.SigScheme,
		Mint:            mint,
		RequestSectionParams: []apilib.RequestSectionParams{{
			TargetContractID: coretypes.NewContractID(c.ChainID, contractHname),
			EntryPointCode:   entryPoint,
			Transfer:         transfer,
			Vars:             vars,
		}},
		Post: true,
	})
}
