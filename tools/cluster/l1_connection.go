package cluster

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/parameters"
	"golang.org/x/xerrors"
)

// requirements for a client connection to L1 (hornet or bee node)
type L1Connection interface {
	// requests funds from faucet, waits for confirmation
	RequestFunds(addr iotago.Address, timeout ...time.Duration) error
	// sends a tx, waits for confirmation
	PostTx(tx *iotago.Transaction, timeout ...time.Duration) (*iotago.Message, error)
	// returns the outputs owned by a given address
	OutputMap(myAddress iotago.Address, timeout ...time.Duration) (map[iotago.OutputID]iotago.Output, error)

	// --- for testing (TODO maybe move to a separate interface)
	// sends an HTTP request to the faucet (just for testing, normal usage should use `RequestFunds`)
	FaucetRequestHttp(addr iotago.Address, timeout ...time.Duration) error
	// sends a simple value Tx (just for testing)
	PostSimpleValueTX(
		sender *cryptolib.KeyPair,
		recipientAddr iotago.Address,
		amount uint64,
	) (*iotago.Message, error)
	// creates a simple valueTx (just for testing)
	MakeSimpleValueTX(
		sender *cryptolib.KeyPair,
		recipientAddr iotago.Address,
		amount uint64,
	) (*iotago.Transaction, error)
}

type l1Client struct {
	hostname      string
	apiPort       int
	faucetPort    int
	faucetKeyPair *cryptolib.KeyPair
	l1params      *parameters.L1
	client        *nodeclient.Client
}

func NewL1Client(config L1Config) L1Connection {
	nodeClient := nodeclient.New(
		fmt.Sprintf("http://%s:%d", config.Hostname, config.APIPort),
		parameters.DeSerializationParametersForTesting(),
		nodeclient.WithIndexer(),
	)

	// TODO
	// /api/v2/info // func (client *Client) Info(ctx context.Context) (*InfoResponse, error)
	// how to get protocol params via the client if we need them to start the client lol...?

	l1params := &parameters.L1{
		NetworkID:                 iotago.NetworkIDFromString(config.NetworkID),
		MaxTransactionSize:        32000, // TODO should be some const from iotago
		DeSerializationParameters: parameters.DeSerializationParametersForTesting(),
	}

	return &l1Client{
		hostname:      config.Hostname,
		apiPort:       config.APIPort,
		faucetPort:    config.FaucetPort,
		client:        nodeClient,
		faucetKeyPair: config.FaucetKey,
		l1params:      l1params,
	}
}

const defaultTimeout = 20 * time.Second

func newCtx(timeout ...time.Duration) (context.Context, context.CancelFunc) {
	t := defaultTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return context.WithTimeout(context.Background(), t)
}

func (c *l1Client) OutputMap(myAddress iotago.Address, timeout ...time.Duration) (map[iotago.OutputID]iotago.Output, error) {
	ctxWithTimeout, cancelContext := newCtx(timeout...)
	defer cancelContext()

	res, err := c.client.Indexer().Outputs(ctxWithTimeout, &nodeclient.OutputsQuery{
		AddressBech32: myAddress.Bech32(iscp.NetworkPrefix),
	})
	if err != nil {
		return nil, xerrors.Errorf("failed to query address outputs: %w", err)
	}
	result := make(map[iotago.OutputID]iotago.Output)
	for res.Next() {
		outs, err := res.Outputs()
		if err != nil {
			return nil, xerrors.Errorf("failed to fetch address outputs: %w", err)
		}
		oids := res.Response.Items.MustOutputIDs()
		for i, o := range outs {
			result[oids[i]] = o
		}
	}
	return result, nil
}

func (c *l1Client) PostTx(tx *iotago.Transaction, timeout ...time.Duration) (*iotago.Message, error) {
	ctxWithTimeout, cancelContext := newCtx(timeout...)
	defer cancelContext()

	// Build a message and post it.
	txMsg, err := builder.NewMessageBuilder().Payload(tx).Build()
	if err != nil {
		return nil, xerrors.Errorf("failed to build a tx message: %w", err)
	}
	txMsg, err = c.client.SubmitMessage(ctxWithTimeout, txMsg)
	if err != nil {
		return nil, xerrors.Errorf("failed to submit a tx message: %w", err)
	}
	return txMsg, nil
}

const faucetAmountToSend = uint64(123)

func (c *l1Client) RequestFunds(addr iotago.Address, timeout ...time.Duration) error {
	if c.faucetKeyPair == nil {
		return c.FaucetRequestHttp(addr, timeout...)
	}
	_, err := c.PostSimpleValueTX(c.faucetKeyPair, addr, faucetAmountToSend)
	return err
}

// PostFaucetRequest makes a faucet request.
// Simple value TX is processed faster, and should be used in cases where we are using a private testnet and have the genesis key available.
func (c *l1Client) FaucetRequestHttp(addr iotago.Address, timeout ...time.Duration) error {
	ctxWithTimeout, cancelContext := newCtx(timeout...)
	defer cancelContext()

	faucetReq := fmt.Sprintf("{\"address\":%q}", addr.Bech32(iscp.NetworkPrefix))
	faucetURL := fmt.Sprintf("http://%s:%d/api/plugins/faucet/v1/enqueue", c.hostname, c.faucetPort)
	httpReq, err := http.NewRequestWithContext(ctxWithTimeout, "POST", faucetURL, bytes.NewReader([]byte(faucetReq)))
	if err != nil {
		return xerrors.Errorf("unable to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return xerrors.Errorf("unable to call faucet: %w", err)
	}
	if res.StatusCode == 202 {
		return nil
	}
	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return xerrors.Errorf("faucet status=%v, unable to read response body: %w", res.Status, err)
	}
	return xerrors.Errorf("faucet call failed, responPrivateKeyse status=%v, body=%v", res.Status, resBody)
}

// PostSimpleValueTX submits a simple value transfer TX.
// Can be used instead of the faucet API if the genesis key is known.
func (c *l1Client) PostSimpleValueTX(
	sender *cryptolib.KeyPair,
	recipientAddr iotago.Address,
	amount uint64,
) (*iotago.Message, error) {
	tx, err := c.MakeSimpleValueTX(sender, recipientAddr, amount)
	if err != nil {
		return nil, xerrors.Errorf("failed to build a tx: %w", err)
	}
	return c.PostTx(tx)
}

func (c *l1Client) MakeSimpleValueTX(
	sender *cryptolib.KeyPair,
	recipientAddr iotago.Address,
	amount uint64,
) (*iotago.Transaction, error) {
	senderAddr := sender.GetPublicKey().AsEd25519Address()
	senderOuts, err := c.OutputMap(senderAddr)
	if err != nil {
		return nil, xerrors.Errorf("failed to get address outputs: %w", err)
	}
	txBuilder := builder.NewTransactionBuilder(c.l1params.NetworkID)
	inputSum := uint64(0)
	for i, o := range senderOuts {
		if inputSum >= amount {
			break
		}
		oid := i
		out := o
		txBuilder = txBuilder.AddInput(&builder.ToBeSignedUTXOInput{
			Address:  senderAddr,
			OutputID: oid,
			Output:   out,
		})
		inputSum += out.Deposit()
	}
	if inputSum < amount {
		return nil, xerrors.Errorf("not enough funds, have=%v, need=%v", inputSum, amount)
	}
	txBuilder = txBuilder.AddOutput(&iotago.BasicOutput{
		Amount:     amount,
		Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{Address: recipientAddr}},
	})
	if inputSum > amount {
		txBuilder = txBuilder.AddOutput(&iotago.BasicOutput{
			Amount:     inputSum - amount,
			Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{Address: senderAddr}},
		})
	}
	tx, err := txBuilder.Build(
		c.l1params.DeSerializationParameters,
		sender.AsAddressSigner(),
	)
	if err != nil {
		return nil, xerrors.Errorf("failed to build a tx: %w", err)
	}
	return tx, nil
}
