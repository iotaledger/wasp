package nodeconn

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/utxodb"
	"golang.org/x/xerrors"
)

type L1Config struct {
	Hostname   string
	APIPort    int
	FaucetPort int
	FaucetKey  *cryptolib.KeyPair
}

// nodeconn implements L1Client
// interface to be used by utilities like: cluster-tool, wasp-cli, apilib, etc
type L1Client interface {
	// requests funds from faucet, waits for confirmation
	RequestFunds(addr iotago.Address, timeout ...time.Duration) error
	// sends a tx, waits for confirmation
	PostTx(tx *iotago.Transaction, timeout ...time.Duration) error
	// returns the outputs owned by a given address
	OutputMap(myAddress iotago.Address, timeout ...time.Duration) (map[iotago.OutputID]iotago.Output, error)
	// returns the l1 parameters used by the node
	L1Params() *parameters.L1
	// used to query the health endpoint of the node
	Health(timeout ...time.Duration) (bool, error)
}

var _ L1Client = &nodeConn{}

func NewL1Client(config L1Config, log *logger.Logger, timeout ...time.Duration) L1Client {
	return newNodeConn(config, log, timeout...)
}

const defaultTimeout = 1 * time.Minute

// OutputMap implements L1Connection
func (nc *nodeConn) OutputMap(myAddress iotago.Address, timeout ...time.Duration) (map[iotago.OutputID]iotago.Output, error) {
	ctxWithTimeout, cancelContext := newCtx(timeout...)
	defer cancelContext()

	indexerClient, err := nc.nodeClient.Indexer(ctxWithTimeout)
	if err != nil {
		return nil, xerrors.Errorf("failed getting the indexer client: %w", err)
	}
	res, err := indexerClient.Outputs(ctxWithTimeout, &nodeclient.OutputsQuery{
		AddressBech32: myAddress.Bech32(nc.l1params.Bech32Prefix),
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

// TODO refactor attach/promotion to be reusable
// PostTx implements L1Connection
// sends any tx to the L1 node, then waits until the tx is confirmed.
func (nc *nodeConn) PostTx(tx *iotago.Transaction, timeout ...time.Duration) error {
	ctxWithTimeout, cancelContext := newCtx(timeout...)
	defer cancelContext()

	// Build a message and post it.
	txMsg, err := builder.NewMessageBuilder().Payload(tx).Build()
	if err != nil {
		return xerrors.Errorf("failed to build a tx message: %w", err)
	}
	txMsg, err = nc.nodeClient.SubmitMessage(ctxWithTimeout, txMsg, nc.l1params.DeSerializationParameters)
	if err != nil {
		return xerrors.Errorf("failed to submit a tx message: %w", err)
	}

	// wait until tx is confirmed
	msgID, err := txMsg.ID()
	if err != nil {
		return xerrors.Errorf("failed to get msg ID: %w", err)
	}

	// poll the node by getting `MessageMetadataByMessageID`
	for {
		metadataResp, err := nc.nodeClient.MessageMetadataByMessageID(ctxWithTimeout, *msgID)
		if err != nil {
			return xerrors.Errorf("failed to get msg metadata: %w", err)
		}

		if metadataResp.ReferencedByMilestoneIndex != nil {
			if metadataResp.LedgerInclusionState != nil && *metadataResp.LedgerInclusionState == "included" {
				return nil
			}
			return xerrors.Errorf("tx was not included in the ledger")
		}
		// reattach or promote if needed
		if metadataResp.ShouldPromote != nil && *metadataResp.ShouldPromote {
			// create an empty message and the messageID as one of the parents
			promotionMsg, err := builder.NewMessageBuilder().Parents([][]byte{msgID[:]}).Build()
			if err != nil {
				return xerrors.Errorf("failed to build promotion message: %w", err)
			}
			_, err = nc.nodeClient.SubmitMessage(ctxWithTimeout, promotionMsg, nc.l1params.DeSerializationParameters)
			if err != nil {
				return xerrors.Errorf("failed to promote msg: %w", err)
			}
		}
		if metadataResp.ShouldReattach != nil && *metadataResp.ShouldReattach {
			// remote PoW: Take the message, clear parents, clear nonce, send to node
			txMsg.Parents = nil
			txMsg.Nonce = 0
			txMsg, err = nc.nodeClient.SubmitMessage(ctxWithTimeout, txMsg, nc.l1params.DeSerializationParameters)
			if err != nil {
				return xerrors.Errorf("failed to get re-attach msg: %w", err)
			}
		}

		if err = ctxWithTimeout.Err(); err != nil {
			return xerrors.Errorf("failed to wait for tx confimation within timeout: %s", err)
		}
		time.Sleep(pollConfirmedTxInterval)
	}
}

// RequestFunds implements L1Connection
func (nc *nodeConn) RequestFunds(addr iotago.Address, timeout ...time.Duration) error {
	if nc.config.FaucetKey == nil {
		return nc.FaucetRequestHTTP(addr, timeout...)
	}
	return nc.PostSimpleValueTX(nc.config.FaucetKey, addr, utxodb.FundsFromFaucetAmount)
}

// PostFaucetRequest makes a faucet request.
// Simple value TX is processed faster, and should be used in cases where we are using a private testnet and have the genesis key available.
func (nc *nodeConn) FaucetRequestHTTP(addr iotago.Address, timeout ...time.Duration) error {
	ctxWithTimeout, cancelContext := newCtx(timeout...)
	defer cancelContext()

	faucetReq := fmt.Sprintf("{\"address\":%q}", addr.Bech32(nc.L1Params().Bech32Prefix))
	faucetURL := fmt.Sprintf("http://%s:%d/api/plugins/faucet/v1/enqueue", nc.config.Hostname, nc.config.APIPort)
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
func (nc *nodeConn) PostSimpleValueTX(
	sender *cryptolib.KeyPair,
	recipientAddr iotago.Address,
	amount uint64,
) error {
	tx, err := MakeSimpleValueTX(nc, sender, recipientAddr, amount)
	if err != nil {
		return xerrors.Errorf("failed to build a tx: %w", err)
	}
	return nc.PostTx(tx)
}

func MakeSimpleValueTX(
	client L1Client,
	sender *cryptolib.KeyPair,
	recipientAddr iotago.Address,
	amount uint64,
) (*iotago.Transaction, error) {
	senderAddr := sender.GetPublicKey().AsEd25519Address()
	senderOuts, err := client.OutputMap(senderAddr)
	if err != nil {
		return nil, xerrors.Errorf("failed to get address outputs: %w", err)
	}
	txBuilder := builder.NewTransactionBuilder(client.L1Params().NetworkID)
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
		client.L1Params().DeSerializationParameters,
		sender.AsAddressSigner(),
	)
	if err != nil {
		return nil, xerrors.Errorf("failed to build a tx: %w", err)
	}
	return tx, nil
}

// Health implements L1Client
func (nc *nodeConn) Health(timeout ...time.Duration) (bool, error) {
	ctxWithTimeout, cancelContext := newCtx(timeout...)
	defer cancelContext()

	return nc.nodeClient.Health(ctxWithTimeout)
}
