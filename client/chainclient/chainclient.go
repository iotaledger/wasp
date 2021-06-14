package chainclient

import (
	"net/http"
	"strings"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/goshimmer"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// Client allows to send on-ledger or off-ledger requests to a specific chain in the node
type Client struct {
	GoshimmerClient *goshimmer.Client
	WaspClient      *client.WaspClient
	ChainID         chainid.ChainID
	KeyPair         *ed25519.KeyPair
}

// New creates a new chainclient.Client
func New(
	goshimmerClient *goshimmer.Client,
	waspClient *client.WaspClient,
	chainID chainid.ChainID,
	keyPair *ed25519.KeyPair,
) *Client {
	return &Client{
		GoshimmerClient: goshimmerClient,
		WaspClient:      waspClient,
		ChainID:         chainID,
		KeyPair:         keyPair,
	}
}

type PostRequestParams struct {
	Transfer *ledgerstate.ColoredBalances
	Args     requestargs.RequestArgs
}

// Post1Request sends an on-ledger transaction with one request on it to the chain
func (c *Client) Post1Request(
	contractHname coretypes.Hname,
	entryPoint coretypes.Hname,
	params ...PostRequestParams,
) (*ledgerstate.Transaction, error) {
	par := PostRequestParams{}
	if len(params) > 0 {
		par = params[0]
	}

	return c.GoshimmerClient.PostRequestTransaction(transaction.NewRequestTransactionParams{
		SenderKeyPair: c.KeyPair,
		Requests: []transaction.RequestParams{{
			ChainID:    c.ChainID,
			Contract:   contractHname,
			EntryPoint: entryPoint,
			Transfer:   par.Transfer,
			Args:       par.Args,
		}},
	})
}

// PostOffLedgerRequest sends an off-ledger tx via the wasp node web api
func (c *Client) PostOffLedgerRequest(
	contractHname coretypes.Hname,
	entrypoint coretypes.Hname,
	args requestargs.RequestArgs,
) (*request.RequestOffLedger, error) {
	offledgerReq := request.NewRequestOffLedger(contractHname, entrypoint, args)
	offledgerReq.Sign(c.KeyPair)

	httpclient := &http.Client{}
	body := "{\"request\": \"" + offledgerReq.Base64() + "\"}"
	apiEndpointURL := c.WaspClient.BaseURL() + routes.NewRequest(c.ChainID.Base58())
	httpReq, err := http.NewRequest("POST", apiEndpointURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("Content-type", `application/json"`)
	resp, err := httpclient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return offledgerReq, nil
}
