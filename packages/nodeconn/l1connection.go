package nodeconn

import (
	"bytes"
	"context"
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
)

type L1Config struct {
	APIAddress    string
	INXAddress    string
	FaucetAddress string
	FaucetKey     *cryptolib.KeyPair
	UseRemotePoW  bool
}

// nodeconn implements L1Client
// interface to be used by utilities like: cluster-tool, wasp-cli, apilib, etc
type L1Client interface {
	// requests funds from faucet, waits for confirmation
	RequestFunds(addr iotago.Address, timeout ...time.Duration) error
	// sends a tx, waits for confirmation
	PostTx(tx *iotago.Transaction, timeout ...time.Duration) error
	// returns the outputs owned by a given address
	OutputMap(myAddress iotago.Address, timeout ...time.Duration) (iotago.OutputSet, error)
	// output
	GetAliasOutput(aliasID iotago.AliasID, timeout ...time.Duration) (iotago.OutputID, iotago.Output, error)
	// used to query the health endpoint of the node
	Health(timeout ...time.Duration) (bool, error)
}

var _ L1Client = &nodeConn{}

func NewL1Client(config L1Config, log *logger.Logger, timeout ...time.Duration) L1Client {
	return newNodeConn(config, log, false, timeout...)
}

const defaultTimeout = 1 * time.Minute

// OutputMap implements L1Connection
func (nc *nodeConn) OutputMap(myAddress iotago.Address, timeout ...time.Duration) (iotago.OutputSet, error) {
	ctxWithTimeout, cancelContext := newCtx(nc.ctx, timeout...)
	defer cancelContext()

	bech32Addr := myAddress.Bech32(parameters.L1().Protocol.Bech32HRP)
	queries := []nodeclient.IndexerQuery{
		&nodeclient.BasicOutputsQuery{AddressBech32: bech32Addr},
		&nodeclient.FoundriesQuery{AliasAddressBech32: bech32Addr},
		&nodeclient.NFTsQuery{AddressBech32: bech32Addr},
		&nodeclient.AliasesQuery{StateControllerBech32: bech32Addr},
	}

	result := make(map[iotago.OutputID]iotago.Output)

	for _, query := range queries {
		res, err := nc.indexerClient.Outputs(ctxWithTimeout, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query address outputs: %w", err)
		}
		for res.Next() {
			outs, err := res.Outputs()
			if err != nil {
				return nil, fmt.Errorf("failed to fetch address outputs: %w", err)
			}
			oids := res.Response.Items.MustOutputIDs()
			for i, o := range outs {
				result[oids[i]] = o
			}
		}
	}
	return result, nil
}

// PostTx implements L1Connection
// sends any tx to the L1 node, then waits until the tx is confirmed.
func (nc *nodeConn) PostTx(tx *iotago.Transaction, timeout ...time.Duration) error {
	ctxWithTimeout, cancelContext := newCtx(nc.ctx, timeout...)
	defer cancelContext()

	txId, err := nc.doPostTx(ctxWithTimeout, tx)
	if err != nil {
		return err
	}

	return nc.waitUntilConfirmed(ctxWithTimeout, txId)
}

func (nc *nodeConn) GetAliasOutput(aliasID iotago.AliasID, timeout ...time.Duration) (iotago.OutputID, iotago.Output, error) {
	ctxWithTimeout, cancelContext := newCtx(nc.ctx, timeout...)
	outputID, stateOutput, err := nc.indexerClient.Alias(ctxWithTimeout, aliasID)
	cancelContext()
	return *outputID, stateOutput, err
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
	initialAddrOutputs, err := nc.OutputMap(addr)
	if err != nil {
		return err
	}
	ctxWithTimeout, cancelContext := newCtx(nc.ctx, timeout...)
	defer cancelContext()

	faucetReq := fmt.Sprintf("{\"address\":%q}", addr.Bech32(parameters.L1().Protocol.Bech32HRP))
	faucetURL := fmt.Sprintf("%s/api/enqueue", nc.config.FaucetAddress)
	httpReq, err := http.NewRequestWithContext(ctxWithTimeout, "POST", faucetURL, bytes.NewReader([]byte(faucetReq)))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("unable to call faucet: %w", err)
	}
	if res.StatusCode != 202 {
		resBody, err := io.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			return fmt.Errorf("faucet status=%v, unable to read response body: %w", res.Status, err)
		}
		return fmt.Errorf("faucet call failed, response status=%v, body=%v", res.Status, string(resBody))
	}
	// wait until funds are available
	for {
		select {
		case <-ctxWithTimeout.Done():
			return fmt.Errorf("faucet request timed-out while waiting for funds to be available")
		case <-time.After(1 * time.Second):
			newOutputs, err := nc.OutputMap(addr)
			if err != nil {
				return err
			}
			if len(newOutputs) > len(initialAddrOutputs) {
				return nil // success
			}
		}
	}
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
		return fmt.Errorf("failed to build a tx: %w", err)
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
		return nil, fmt.Errorf("failed to get address outputs: %w", err)
	}
	txBuilder := builder.NewTransactionBuilder(parameters.L1().Protocol.NetworkID())
	inputSum := uint64(0)
	for i, o := range senderOuts {
		if inputSum >= amount {
			break
		}
		oid := i
		out := o
		txBuilder = txBuilder.AddInput(&builder.TxInput{
			UnlockTarget: senderAddr,
			InputID:      oid,
			Input:        out,
		})
		inputSum += out.Deposit()
	}
	if inputSum < amount {
		return nil, fmt.Errorf("not enough funds, have=%v, need=%v", inputSum, amount)
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
		parameters.L1().Protocol,
		sender.AsAddressSigner(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build a tx: %w", err)
	}
	return tx, nil
}

// Health implements L1Client
func (nc *nodeConn) Health(timeout ...time.Duration) (bool, error) {
	ctxWithTimeout, cancelContext := newCtx(context.Background(), timeout...)
	defer cancelContext()
	return nc.nodeBridge.INXNodeClient().Health(ctxWithTimeout)
}
