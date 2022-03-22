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

// interface for the cluster tool to interact with a L1 node (hornet or bee node)
type L1Connection interface {
	// requests funds from faucet, waits for confirmation
	RequestFunds(addr iotago.Address, timeout ...time.Duration) error
	// sends a tx, waits for confirmation
	PostTx(tx *iotago.Transaction, timeout ...time.Duration) (*iotago.Message, error)
	// returns the outputs owned by a given address
	OutputMap(myAddress iotago.Address, timeout ...time.Duration) (map[iotago.OutputID]iotago.Output, error)
}

// implementation of the L1Connection (hornet specific for now using the REST API)
type L1Client struct {
	hostname      string
	apiPort       int
	faucetPort    int
	faucetKeyPair *cryptolib.KeyPair
	l1params      *parameters.L1
	client        *nodeclient.Client
}

const defaultTimeout = 1 * time.Minute

func newCtx(timeout ...time.Duration) (context.Context, context.CancelFunc) {
	t := defaultTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return context.WithTimeout(context.Background(), t)
}

func NewL1Client(config L1Config, timeout ...time.Duration) L1Connection {
	nc := nodeclient.New(fmt.Sprintf("http://%s:%d", config.Hostname, config.APIPort))

	ctxWithTimeout, cancelContext := newCtx(timeout...)
	defer cancelContext()
	l1Info, err := nc.Info(ctxWithTimeout)
	if err != nil {
		panic(xerrors.Errorf("error getting L1 connection info: %w", err))
	}

	l1params := &parameters.L1{
		NetworkName:        l1Info.Protocol.NetworkName,
		NetworkID:          iotago.NetworkIDFromString(l1Info.Protocol.NetworkName),
		Bech32Prefix:       l1Info.Protocol.Bech32HRP,
		MaxTransactionSize: 32000, // TODO should be some const from iotago
		DeSerializationParameters: &iotago.DeSerializationParameters{
			RentStructure: &l1Info.Protocol.RentStructure,
		},
	}

	return &L1Client{
		hostname:      config.Hostname,
		apiPort:       config.APIPort,
		faucetPort:    config.FaucetPort,
		client:        nc,
		faucetKeyPair: config.FaucetKey,
		l1params:      l1params,
	}
}

func (c *L1Client) OutputMap(myAddress iotago.Address, timeout ...time.Duration) (map[iotago.OutputID]iotago.Output, error) {
	ctxWithTimeout, cancelContext := newCtx(timeout...)
	defer cancelContext()

	indexerClient, err := c.client.Indexer(ctxWithTimeout)
	if err != nil {
		return nil, xerrors.Errorf("failed getting the indexer client: %w", err)
	}
	res, err := indexerClient.Outputs(ctxWithTimeout, &nodeclient.OutputsQuery{
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

const pollConfirmedTxInterval = 200 * time.Millisecond

// PostTx sends any tx to the L1 node, then waits until the tx is confirmed.
func (c *L1Client) PostTx(tx *iotago.Transaction, timeout ...time.Duration) (*iotago.Message, error) {
	ctxWithTimeout, cancelContext := newCtx(timeout...)
	defer cancelContext()

	// Build a message and post it.
	txMsg, err := builder.NewMessageBuilder().Payload(tx).Build()
	if err != nil {
		return nil, xerrors.Errorf("failed to build a tx message: %w", err)
	}
	txMsg, err = c.client.SubmitMessage(ctxWithTimeout, txMsg, c.l1params.DeSerializationParameters)
	if err != nil {
		return nil, xerrors.Errorf("failed to submit a tx message: %w", err)
	}

	// wait until tx is confirmed
	msgID, err := txMsg.ID()
	if err != nil {
		return nil, xerrors.Errorf("failed to get msg ID: %w", err)
	}

	// poll the node by getting `MessageMetadataByMessageID`
	for {
		metadataResp, err := c.client.MessageMetadataByMessageID(ctxWithTimeout, *msgID)
		if err != nil {
			return nil, xerrors.Errorf("failed to get msg metadata: %w", err)
		}

		if metadataResp.ReferencedByMilestoneIndex != nil {
			if metadataResp.LedgerInclusionState != nil && *metadataResp.LedgerInclusionState == "included" {
				return txMsg, nil
			}
			return nil, xerrors.Errorf("tx was not included in the ledger")
		}
		// reattach or promote if needed
		if metadataResp.ShouldPromote != nil && *metadataResp.ShouldPromote {
			// create an empty message and the messageID as one of the parents
			promotionMsg, err := builder.NewMessageBuilder().Parents([][]byte{msgID[:]}).Build()
			if err != nil {
				return nil, xerrors.Errorf("failed to build promotion message: %w", err)
			}
			_, err = c.client.SubmitMessage(ctxWithTimeout, promotionMsg, c.l1params.DeSerializationParameters)
			if err != nil {
				return nil, xerrors.Errorf("failed to promote msg: %w", err)
			}
		}
		if metadataResp.ShouldReattach != nil && *metadataResp.ShouldReattach {
			// remote PoW: Take the message, clear parents, clear nonce, send to node
			txMsg.Parents = nil
			txMsg.Nonce = 0
			txMsg, err = c.client.SubmitMessage(ctxWithTimeout, txMsg, c.l1params.DeSerializationParameters)
			if err != nil {
				return nil, xerrors.Errorf("failed to get re-attach msg: %w", err)
			}
		}

		if err = ctxWithTimeout.Err(); err != nil {
			return nil, xerrors.Errorf("failed to wait for tx confimation within timeout: %s", err)
		}
		time.Sleep(pollConfirmedTxInterval)
	}
}

const faucetAmountToSend = uint64(123)

func (c *L1Client) RequestFunds(addr iotago.Address, timeout ...time.Duration) error {
	if c.faucetKeyPair == nil {
		return c.FaucetRequestHTTP(addr, timeout...)
	}
	_, err := c.PostSimpleValueTX(c.faucetKeyPair, addr, faucetAmountToSend)
	return err
}

// PostFaucetRequest makes a faucet request.
// Simple value TX is processed faster, and should be used in cases where we are using a private testnet and have the genesis key available.
func (c *L1Client) FaucetRequestHTTP(addr iotago.Address, timeout ...time.Duration) error {
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
func (c *L1Client) PostSimpleValueTX(
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

func (c *L1Client) MakeSimpleValueTX(
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
