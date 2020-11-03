package chainclient

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"net/url"
	"time"

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

	timeout time.Duration

	publisherHost string
}

func New(
	nodeClient nodeclient.NodeClient,
	waspClient *client.WaspClient,
	chainID *coretypes.ChainID,
	sigScheme signaturescheme.SignatureScheme,
	waitForCompletionTimeout ...time.Duration,
) *Client {
	var t time.Duration
	if len(waitForCompletionTimeout) > 0 {
		t = waitForCompletionTimeout[0]
	}
	return &Client{
		NodeClient: nodeClient,
		WaspClient: waspClient,
		ChainID:    *chainID,
		SigScheme:  sigScheme,
		timeout:    t,
	}
}

func (c *Client) PostRequest(
	contractIndex uint16,
	entryPoint coretypes.EntryPointCode,
	mint map[address.Address]int64, // TODO
	transfer map[balance.Color]int64,
	vars map[string]interface{},
) (*sctransaction.Transaction, error) {
	if c.timeout > 0 && len(c.publisherHost) == 0 {
		info, err := c.WaspClient.Info()
		if err != nil {
			return nil, err
		}
		u, err := url.Parse(c.WaspClient.BaseURL())
		if err != nil {
			return nil, err
		}
		c.publisherHost = fmt.Sprintf("%s:%d", u.Hostname(), info.PublisherPort)
	}

	return apilib.CreateRequestTransaction(apilib.CreateRequestTransactionParams{
		NodeClient:      c.NodeClient,
		SenderSigScheme: c.SigScheme,
		Mint:            mint,
		BlockParams: []apilib.RequestBlockParams{{
			TargetContractID: coretypes.NewContractID(c.ChainID, contractIndex),
			EntryPointCode:   entryPoint,
			Transfer:         transfer,
			Vars:             vars,
		}},
		Post:                true,
		WaitForConfirmation: c.timeout > 0,
		WaitForCompletion:   c.timeout > 0,
		PublisherHosts:      []string{c.publisherHost},
		PublisherQuorum:     1,
		Timeout:             c.timeout,
	})
}
