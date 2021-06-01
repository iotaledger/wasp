package chainclient

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/goshimmer"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/transaction"
)

// Client allows to send webapi requests to a specific chain in the node
type Client struct {
	GoshimmerClient *goshimmer.Client
	WaspClient      *client.WaspClient
	ChainID         coretypes.ChainID
	KeyPair         *ed25519.KeyPair
}

// New creates a new chainclient.Client
func New(
	goshimmerClient *goshimmer.Client,
	waspClient *client.WaspClient,
	chainID coretypes.ChainID,
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

// Post1Request sends one request transaction with one request on it to the chain
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
