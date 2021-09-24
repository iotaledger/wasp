package chainclient

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/client/goshimmer"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/transaction"
)

// Client allows to interact with a specific chain in the node, for example to send on-ledger or off-ledger requests
type Client struct {
	GoshimmerClient *goshimmer.Client
	WaspClient      *client.WaspClient
	ChainID         iscp.ChainID
	KeyPair         *ed25519.KeyPair
	nonces          map[ed25519.PublicKey]uint64
}

// New creates a new chainclient.Client
func New(
	goshimmerClient *goshimmer.Client,
	waspClient *client.WaspClient,
	chainID iscp.ChainID,
	keyPair *ed25519.KeyPair,
) *Client {
	return &Client{
		GoshimmerClient: goshimmerClient,
		WaspClient:      waspClient,
		ChainID:         chainID,
		KeyPair:         keyPair,
		nonces:          make(map[ed25519.PublicKey]uint64),
	}
}

type PostRequestParams struct {
	Transfer colored.Balances
	Args     requestargs.RequestArgs
	Nonce    uint64
}

// Post1Request sends an on-ledger transaction with one request on it to the chain
func (c *Client) Post1Request(
	contractHname iscp.Hname,
	entryPoint iscp.Hname,
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
	contractHname iscp.Hname,
	entrypoint iscp.Hname,
	params ...PostRequestParams,
) (*request.OffLedger, error) {
	par := PostRequestParams{}
	if len(params) > 0 {
		par = params[0]
	}
	if par.Nonce == 0 {
		c.nonces[c.KeyPair.PublicKey]++
		par.Nonce = c.nonces[c.KeyPair.PublicKey]
	}
	offledgerReq := request.NewOffLedger(contractHname, entrypoint, par.Args).WithTransfer(par.Transfer)
	offledgerReq.WithNonce(par.Nonce)
	offledgerReq.Sign(c.KeyPair)
	return offledgerReq, c.WaspClient.PostOffLedgerRequest(&c.ChainID, offledgerReq)
}

// NewPostRequestParams simplifies encoding of request parameters
func NewPostRequestParams(p ...interface{}) *PostRequestParams {
	return &PostRequestParams{
		Args: requestargs.New(nil).AddEncodeSimpleMany(parseParams(p)),
	}
}

func (par *PostRequestParams) WithTransfer(transfer colored.Balances) *PostRequestParams {
	par.Transfer = transfer
	return par
}

func (par *PostRequestParams) WithTransferEncoded(colval ...interface{}) *PostRequestParams {
	if len(colval) == 0 {
		return par
	}
	if len(colval)%2 != 0 {
		panic("WithTransferEncode: len(params) % 2 != 0")
	}
	par.Transfer = colored.NewBalances()
	for i := 0; i < len(colval)/2; i++ {
		key, ok := colval[2*i].(colored.Color)
		if !ok {
			panic("toMap: color.Color expected")
		}
		par.Transfer.Set(key, encodeIntToUint64(colval[2*i+1]))
	}
	return par
}

func (par *PostRequestParams) WithIotas(i uint64) *PostRequestParams {
	return par.WithTransferEncoded(colored.IOTA, i)
}

func encodeIntToUint64(i interface{}) uint64 {
	switch i := i.(type) {
	case int:
	case byte:
	case int8:
	case int16:
	case uint16:
	case int32:
	case uint32:
	case int64:
	case uint64:
		return i
	}
	panic("integer type expected")
}

func parseParams(params []interface{}) dict.Dict {
	if len(params) == 1 {
		return params[0].(dict.Dict)
	}
	return codec.MakeDict(toMap(params))
}

// makes map without hashing
func toMap(params []interface{}) map[string]interface{} {
	par := make(map[string]interface{})
	if len(params) == 0 {
		return par
	}
	if len(params)%2 != 0 {
		panic("toMap: len(params) % 2 != 0")
	}
	for i := 0; i < len(params)/2; i++ {
		key, ok := params[2*i].(string)
		if !ok {
			panic("toMap: string expected")
		}
		par[key] = params[2*i+1]
	}
	return par
}
