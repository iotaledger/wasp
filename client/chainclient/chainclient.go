package chainclient

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

// Client allows to interact with a specific chain in the node, for example to send on-ledger or off-ledger requests
type Client struct {
	Layer1Client interface{} // TODO implement
	WaspClient   *client.WaspClient
	ChainID      *iscp.ChainID
	KeyPair      *cryptolib.KeyPair
	nonces       map[[cryptolib.PublicKeySize]byte]uint64
}

// New creates a new chainclient.Client
func New(
	layer1Client interface{},
	waspClient *client.WaspClient,
	chainID *iscp.ChainID,
	keyPair *cryptolib.KeyPair,
) *Client {
	return &Client{
		Layer1Client: layer1Client,
		WaspClient:   waspClient,
		ChainID:      chainID,
		KeyPair:      keyPair,
		nonces:       make(map[[cryptolib.PublicKeySize]byte]uint64),
	}
}

type PostRequestParams struct {
	Transfer *iscp.Assets
	Args     dict.Dict
	Nonce    uint64
}

// Post1Request sends an on-ledger transaction with one request on it to the chain
func (c *Client) Post1Request(
	contractHname iscp.Hname,
	entryPoint iscp.Hname,
	params ...PostRequestParams,
) (*iotago.Transaction, error) {
	panic("TODO implement")
	// par := PostRequestParams{}
	// if len(params) > 0 {
	// 	par = params[0]
	// }
	// return c.Layer1Client.PostRequestTransaction(transaction.NewRequestTransactionParams{
	// 	SenderKeyPair: c.KeyPair,
	// 	Requests: []transaction.RequestParams{{
	// 		ChainID:    c.ChainID,
	// 		Contract:   contractHname,
	// 		EntryPoint: entryPoint,
	// 		Transfer:   par.Transfer,
	// 		Args:       par.Args,
	// 	}},
	// })
}

// PostOffLedgerRequest sends an off-ledger tx via the wasp node web api
func (c *Client) PostOffLedgerRequest(
	contractHname iscp.Hname,
	entrypoint iscp.Hname,
	params ...PostRequestParams,
) (*iscp.OffLedgerRequestData, error) {
	panic("TODO implement")
	// par := PostRequestParams{}
	// if len(params) > 0 {
	// 	par = params[0]
	// }
	// if par.Nonce == 0 {
	// 	c.nonces[c.KeyPair.PublicKey]++
	// 	par.Nonce = c.nonces[c.KeyPair.PublicKey]
	// }
	// offledgerReq := request.NewOffLedger(c.ChainID, contractHname, entrypoint, par.Args).WithTransfer(par.Transfer)
	// offledgerReq.WithNonce(par.Nonce)
	// offledgerReq.Sign(c.KeyPair)
	// return offledgerReq, c.WaspClient.PostOffLedgerRequest(c.ChainID, offledgerReq)
}

func (c *Client) DepositFunds(n uint64) (*iotago.Transaction, error) {
	return c.Post1Request(accounts.Contract.Hname(), accounts.FuncDeposit.Hname(), PostRequestParams{
		Transfer: iscp.NewAssets(n, nil),
	})
}

// NewPostRequestParams simplifies encoding of request parameters
func NewPostRequestParams(p ...interface{}) *PostRequestParams {
	return &PostRequestParams{
		Args: parseParams(p),
	}
}

func (par *PostRequestParams) WithTransfer(transfer *iscp.Assets) *PostRequestParams {
	par.Transfer = transfer
	return par
}

func (par *PostRequestParams) WithTransferEncoded(colval ...interface{}) *PostRequestParams {
	panic("TODO not implemented")
	// if len(colval) == 0 {
	// 	return par
	// }
	// if len(colval)%2 != 0 {
	// 	panic("WithTransferEncode: len(params) % 2 != 0")
	// }
	// par.Transfer = colored.NewBalances()
	// for i := 0; i < len(colval)/2; i++ {
	// 	key, ok := colval[2*i].(colored.Color)
	// 	if !ok {
	// 		panic("toMap: color.Color expected")
	// 	}
	// 	par.Transfer.Set(key, encodeIntToUint64(colval[2*i+1]))
	// }
	// return par
}

func (par *PostRequestParams) WithIotas(i uint64) *PostRequestParams {
	par.Transfer.AddIotas(i)
	return par
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
