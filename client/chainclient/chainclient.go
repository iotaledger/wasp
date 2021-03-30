package chainclient

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/level1"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

// Client allows to send webapi requests to a specific chain in the node
type Client struct {
	Level1Client level1.Level1Client
	WaspClient   *client.WaspClient
	ChainID      coretypes.ChainID
	KeyPair      *ed25519.KeyPair
}

// New creates a new chainclient.Client
func New(
	level1Client level1.Level1Client,
	waspClient *client.WaspClient,
	chainID coretypes.ChainID,
	keyPair *ed25519.KeyPair,
) *Client {
	return &Client{
		Level1Client: level1Client,
		WaspClient:   waspClient,
		ChainID:      chainID,
		KeyPair:      keyPair,
	}
}

type PostRequestParams struct {
	Transfer *ledgerstate.ColoredBalances
	Args     requestargs.RequestArgs
}

// PostRequest sends a request transaction to the chain
func (c *Client) PostRequest(
	contractHname coretypes.Hname,
	entryPoint coretypes.Hname,
	params ...PostRequestParams,
) (*ledgerstate.Transaction, error) {
	par := PostRequestParams{}
	if len(params) > 0 {
		par = params[0]
	}

	return apilib.PostRequestTransaction(c.Level1Client, sctransaction.NewRequestTransactionParams{
		SenderKeyPair: c.KeyPair,
		Requests: []sctransaction.RequestParams{{
			ChainID:    c.ChainID,
			Contract:   contractHname,
			EntryPoint: entryPoint,
			Args:       par.Args,
		}},
	})
}
