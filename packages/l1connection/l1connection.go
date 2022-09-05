// to be used by utilities like: cluster-tool, wasp-cli, apilib, etc
package l1connection

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
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/utxodb"
)

type Config struct {
	APIAddress    string
	INXAddress    string
	FaucetAddress string
	FaucetKey     *cryptolib.KeyPair
	UseRemotePoW  bool
}

type Client interface {
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

var _ Client = &l1client{}

type l1client struct {
	ctx           context.Context
	ctxCancel     context.CancelFunc
	indexerClient nodeclient.IndexerClient
	nodeAPIClient *nodeclient.Client
	log           *logger.Logger
	config        Config
}

func NewClient(config Config, log *logger.Logger, timeout ...time.Duration) Client {
	ctx, ctxCancel := context.WithCancel(context.Background())
	nodeAPIClient := nodeclient.New(config.APIAddress)

	ctxWithTimeout, cancelContext := newCtx(ctx, timeout...)
	defer cancelContext()
	l1Info, err := nodeAPIClient.Info(ctxWithTimeout)
	if err != nil {
		panic(fmt.Errorf("error getting L1 connection info: %w", err))
	}
	setL1ProtocolParams(l1Info)

	indexerClient, err := nodeAPIClient.Indexer(ctxWithTimeout)
	if err != nil {
		panic(fmt.Errorf("failed to get nodeclient indexer: %v", err))
	}

	return &l1client{
		ctx:           ctx,
		ctxCancel:     ctxCancel,
		indexerClient: indexerClient,
		nodeAPIClient: nodeAPIClient,
		log:           log.Named("nc"),
		config:        config,
	}
}

// OutputMap implements L1Connection
func (c *l1client) OutputMap(myAddress iotago.Address, timeout ...time.Duration) (iotago.OutputSet, error) {
	ctxWithTimeout, cancelContext := newCtx(c.ctx, timeout...)
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
		res, err := c.indexerClient.Outputs(ctxWithTimeout, query)
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
func (c *l1client) PostTx(tx *iotago.Transaction, timeout ...time.Duration) error {
	ctxWithTimeout, cancelContext := newCtx(c.ctx, timeout...)
	defer cancelContext()

	// Build a Block and post it.
	block, err := builder.NewBlockBuilder().
		Payload(tx).
		Tips(ctxWithTimeout, c.nodeAPIClient).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build a tx: %w", err)
	}
	submitBlock := func(ctx context.Context, block *iotago.Block) error {
		_, err := c.nodeAPIClient.SubmitBlock(ctx, block, parameters.L1().Protocol)
		return err
	}
	err = doBlockPow(ctxWithTimeout, block, c.config.UseRemotePoW, submitBlock, c.nodeAPIClient)
	if err != nil {
		return fmt.Errorf("failed duing PoW: %w", err)
	}
	block, err = c.nodeAPIClient.SubmitBlock(ctxWithTimeout, block, parameters.L1().Protocol)
	if err != nil {
		return fmt.Errorf("failed to submit a tx: %w", err)
	}
	blockID, err := block.ID()
	if err != nil {
		return err
	}
	c.log.Infof("Posted blockID %v", blockID.ToHex())
	txID, err := tx.ID()
	if err != nil {
		return err
	}
	c.log.Infof("Posted transaction id %v", isc.TxID(txID))

	return c.waitUntilConfirmed(ctxWithTimeout, block)
}

const pollConfirmedTxInterval = 200 * time.Millisecond

// waitUntilConfirmed waits until a given tx Block is confirmed, it takes care of promotions/re-attachments for that Block
func (c *l1client) waitUntilConfirmed(ctx context.Context, block *iotago.Block) error {
	// wait until tx is confirmed
	msgID, err := block.ID()
	if err != nil {
		return fmt.Errorf("failed to get msg ID: %w", err)
	}

	// poll the node by getting `BlockMetadataByBlockID`
	for {
		metadataResp, err := c.nodeAPIClient.BlockMetadataByBlockID(ctx, msgID)
		if err != nil {
			return fmt.Errorf("failed to get msg metadata: %w", err)
		}

		if metadataResp.ReferencedByMilestoneIndex != 0 {
			if metadataResp.LedgerInclusionState != "" && metadataResp.LedgerInclusionState == "included" {
				return nil // success
			}
			return fmt.Errorf("tx was not included in the ledger. LedgerInclusionState: %s, ConflictReason: %d",
				metadataResp.LedgerInclusionState, metadataResp.ConflictReason)
		}
		// reattach or promote if needed
		if metadataResp.ShouldPromote != nil && *metadataResp.ShouldPromote {
			c.log.Debugf("promoting msgID: %s", msgID.ToHex())
			// create an empty Block and the BlockID as one of the parents
			tipsResp, err := c.nodeAPIClient.Tips(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch Tips: %w", err)
			}
			tips, err := tipsResp.Tips()
			if err != nil {
				return fmt.Errorf("failed to get Tips from tips response: %w", err)
			}

			parents := []iotago.BlockID{
				msgID,
			}

			if len(tips) > 7 {
				tips = tips[:7] // max 8 parents
			}
			for _, tip := range tips {
				parents = append(parents, tip)
			}
			promotionMsg, err := builder.NewBlockBuilder().Parents(parents).Build()
			if err != nil {
				return fmt.Errorf("failed to build promotion Block: %w", err)
			}
			_, err = c.nodeAPIClient.SubmitBlock(ctx, promotionMsg, parameters.L1().Protocol)
			if err != nil {
				return fmt.Errorf("failed to promote msg: %w", err)
			}
		}
		if metadataResp.ShouldReattach != nil && *metadataResp.ShouldReattach {
			c.log.Debugf("reattaching block: %v", block)
			submitBlock := func(ctx context.Context, block *iotago.Block) error {
				_, err := c.nodeAPIClient.SubmitBlock(ctx, block, parameters.L1().Protocol)
				return err
			}
			err = doBlockPow(ctx, block, c.config.UseRemotePoW, submitBlock, c.nodeAPIClient)
			if err != nil {
				return err
			}
		}
		if err = ctx.Err(); err != nil {
			return fmt.Errorf("failed to wait for tx confimation within timeout: %s", err)
		}
		time.Sleep(pollConfirmedTxInterval)
	}
}

func (c *l1client) GetAliasOutput(aliasID iotago.AliasID, timeout ...time.Duration) (iotago.OutputID, iotago.Output, error) {
	ctxWithTimeout, cancelContext := newCtx(c.ctx, timeout...)
	outputID, stateOutput, err := c.indexerClient.Alias(ctxWithTimeout, aliasID)
	cancelContext()
	return *outputID, stateOutput, err
}

// RequestFunds implements L1Connection
func (c *l1client) RequestFunds(addr iotago.Address, timeout ...time.Duration) error {
	if c.config.FaucetKey == nil {
		return c.FaucetRequestHTTP(addr, timeout...)
	}
	return c.PostSimpleValueTX(c.config.FaucetKey, addr, utxodb.FundsFromFaucetAmount)
}

// PostFaucetRequest makes a faucet request.
// Simple value TX is processed faster, and should be used in cases where we are using a private testnet and have the genesis key available.
func (c *l1client) FaucetRequestHTTP(addr iotago.Address, timeout ...time.Duration) error {
	initialAddrOutputs, err := c.OutputMap(addr)
	if err != nil {
		return err
	}
	ctxWithTimeout, cancelContext := newCtx(c.ctx, timeout...)
	defer cancelContext()

	faucetReq := fmt.Sprintf("{\"address\":%q}", addr.Bech32(parameters.L1().Protocol.Bech32HRP))
	faucetURL := fmt.Sprintf("%s/api/enqueue", c.config.FaucetAddress)
	httpReq, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodPost, faucetURL, bytes.NewReader([]byte(faucetReq)))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("unable to call faucet: %w", err)
	}
	if res.StatusCode != http.StatusAccepted {
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
			newOutputs, err := c.OutputMap(addr)
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
func (c *l1client) PostSimpleValueTX(
	sender *cryptolib.KeyPair,
	recipientAddr iotago.Address,
	amount uint64,
) error {
	tx, err := MakeSimpleValueTX(c, sender, recipientAddr, amount)
	if err != nil {
		return fmt.Errorf("failed to build a tx: %w", err)
	}
	return c.PostTx(tx)
}

func MakeSimpleValueTX(
	client Client,
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
func (c *l1client) Health(timeout ...time.Duration) (bool, error) {
	ctxWithTimeout, cancelContext := newCtx(context.Background(), timeout...)
	defer cancelContext()
	return c.nodeAPIClient.Health(ctxWithTimeout)
}

const defaultTimeout = 1 * time.Minute

func newCtx(ctx context.Context, timeout ...time.Duration) (context.Context, context.CancelFunc) {
	t := defaultTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return context.WithTimeout(ctx, t)
}

func setL1ProtocolParams(info *nodeclient.InfoResponse) {
	parameters.InitL1(&parameters.L1Params{
		// There are no limits on how big from a size perspective an essence can be, so it is just derived from 32KB - Block fields without payload = max size of the payload
		MaxTransactionSize: 32000, // TODO should this value come from the API in the future? or some const in iotago?
		Protocol:           &info.Protocol,
		BaseToken:          (*parameters.BaseToken)(info.BaseToken),
	})
}
