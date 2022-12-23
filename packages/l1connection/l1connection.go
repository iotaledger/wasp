// to be used by utilities like: cluster-tool, wasp-cli, apilib, etc
package l1connection

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/utxodb"
)

const (
	pollConfirmedBlockInterval = 200 * time.Millisecond
	promoteBlockCooldown       = 5 * time.Second
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
	// sends a tx (including tipselection and local PoW if necessary) and waits for confirmation
	PostTxAndWaitUntilConfirmation(tx *iotago.Transaction, timeout ...time.Duration) (*iotago.Block, error)
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
			outputs, err := res.Outputs()
			if err != nil {
				return nil, fmt.Errorf("failed to fetch address outputs: %w", err)
			}

			outputIDs := res.Response.Items.MustOutputIDs()
			for i := range outputs {
				result[outputIDs[i]] = outputs[i]
			}
		}
	}
	return result, nil
}

// postBlock sends a block (including tipselection and local PoW if necessary).
func (c *l1client) postBlock(ctx context.Context, block *iotago.Block) (*iotago.Block, error) {
	if !c.config.UseRemotePoW {
		if err := doBlockPow(ctx, block, c.nodeAPIClient); err != nil {
			return nil, fmt.Errorf("failed during local PoW: %w", err)
		}
	}

	block, err := c.nodeAPIClient.SubmitBlock(ctx, block, parameters.L1().Protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to submit block: %w", err)
	}

	blockID, err := block.ID()
	if err != nil {
		return nil, err
	}
	c.log.Infof("Posted blockID %v", blockID.ToHex())

	return block, nil
}

// PostBlock sends a block (including tipselection and local PoW if necessary).
func (c *l1client) PostBlock(block *iotago.Block, timeout ...time.Duration) (*iotago.Block, error) {
	ctxWithTimeout, cancelContext := newCtx(c.ctx, timeout...)
	defer cancelContext()

	return c.postBlock(ctxWithTimeout, block)
}

// PostTx sends a tx (including tipselection and local PoW if necessary).
func (c *l1client) postTx(ctx context.Context, tx *iotago.Transaction) (*iotago.Block, error) {
	// Build a Block and post it.
	block, err := builder.NewBlockBuilder().Payload(tx).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build block: %w", err)
	}

	block, err = c.postBlock(ctx, block)
	if err != nil {
		return nil, err
	}

	txID, err := tx.ID()
	if err != nil {
		return nil, err
	}
	c.log.Infof("Posted transaction id %v", txID.ToHex())

	return block, nil
}

// PostTx sends a tx (including tipselection and local PoW if necessary).
func (c *l1client) PostTx(tx *iotago.Transaction, timeout ...time.Duration) (*iotago.Block, error) {
	ctxWithTimeout, cancelContext := newCtx(c.ctx, timeout...)
	defer cancelContext()

	return c.postTx(ctxWithTimeout, tx)
}

// PostTxAndWaitUntilConfirmation sends a tx (including tipselection and local PoW if necessary) and waits for confirmation.
func (c *l1client) PostTxAndWaitUntilConfirmation(tx *iotago.Transaction, timeout ...time.Duration) (*iotago.Block, error) {
	ctxWithTimeout, cancelContext := newCtx(c.ctx, timeout...)
	defer cancelContext()

	block, err := c.postTx(ctxWithTimeout, tx)
	if err != nil {
		return nil, err
	}

	return c.waitUntilBlockConfirmed(ctxWithTimeout, block)
}

// waitUntilBlockConfirmed waits until a given block is confirmed, it takes care of promotions/re-attachments for that block
//
//nolint:gocyclo,funlen
func (c *l1client) waitUntilBlockConfirmed(ctx context.Context, block *iotago.Block) (*iotago.Block, error) {
	blockID, err := block.ID()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate blockID: %w", err)
	}

	_, isTransactionPayload := block.Payload.(*iotago.Transaction)
	var lastPromotionTime time.Time

	checkContext := func() error {
		if err = ctx.Err(); err != nil {
			return fmt.Errorf("failed to wait for block confimation within timeout: %s", err)
		}

		return nil
	}

	checkAndPromote := func(metadata *nodeclient.BlockMetadataResponse) error {
		if err := checkContext(); err != nil {
			return err
		}

		if metadata.ShouldPromote != nil && *metadata.ShouldPromote {
			// check if the cooldown time for the next promotion is due
			if !lastPromotionTime.IsZero() && time.Since(lastPromotionTime) < promoteBlockCooldown {
				return nil
			}
			lastPromotionTime = time.Now()

			c.log.Debugf("promoting blockID: %s", blockID.ToHex())
			// create an empty Block and the BlockID as one of the parents
			tipsResp, err := c.nodeAPIClient.Tips(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch tips: %w", err)
			}
			tips, err := tipsResp.Tips()
			if err != nil {
				return fmt.Errorf("failed to get tips from tips response: %w", err)
			}
			if len(tips) > 7 {
				tips = tips[:7] // max 8 parents
			}

			parents := []iotago.BlockID{
				blockID,
			}
			parents = append(parents, tips...)

			promotionBlock, err := builder.NewBlockBuilder().Parents(parents).Build()
			if err != nil {
				return fmt.Errorf("failed to build promotion Block: %w", err)
			}

			if _, err := c.postBlock(ctx, promotionBlock); err != nil {
				return fmt.Errorf("failed to promote block: %w", err)
			}
		}

		return nil
	}

	checkAndReattach := func(metadata *nodeclient.BlockMetadataResponse) error {
		if err := checkContext(); err != nil {
			return err
		}

		if metadata.ShouldReattach != nil && *metadata.ShouldReattach {
			c.log.Debugf("reattaching block: %v", block)

			// build new block with same payload
			block, err = builder.NewBlockBuilder().Payload(block.Payload).Build()
			if err != nil {
				return fmt.Errorf("failed to reattach block: %w", err)
			}

			// reattach the block
			block, err = c.postBlock(ctx, block)
			if err != nil {
				return err
			}

			// update the tracked blockID
			blockID, err = block.ID()
			if err != nil {
				return fmt.Errorf("failed to calculate blockID: %w", err)
			}
		}

		return nil
	}

	for {
		if err := checkContext(); err != nil {
			return nil, err
		}

		// poll the node for block confirmation state
		metadata, err := c.nodeAPIClient.BlockMetadataByBlockID(ctx, blockID)
		if err != nil {
			return nil, fmt.Errorf("failed to get block metadata: %w", err)
		}

		// check if block was included
		if metadata.ReferencedByMilestoneIndex != 0 {
			if metadata.LedgerInclusionState != "" {
				if isTransactionPayload {
					if metadata.LedgerInclusionState == "included" {
						return block, nil // success
					}
				} else {
					if metadata.LedgerInclusionState == "noTransaction" {
						return block, nil // success
					}
				}
			}

			return nil, fmt.Errorf("block was not included in the ledger. IsTransaction: %t, LedgerInclusionState: %s, ConflictReason: %d",
				isTransactionPayload, metadata.LedgerInclusionState, metadata.ConflictReason)
		}

		// promote if needed
		if err := checkAndPromote(metadata); err != nil {
			return nil, err
		}

		// reattach if needed
		if err := checkAndReattach(metadata); err != nil {
			return nil, err
		}

		time.Sleep(pollConfirmedBlockInterval)
	}
}

func (c *l1client) GetAliasOutput(aliasID iotago.AliasID, timeout ...time.Duration) (iotago.OutputID, iotago.Output, error) {
	ctxWithTimeout, cancelContext := newCtx(c.ctx, timeout...)
	outputID, stateOutput, _, err := c.indexerClient.Alias(ctxWithTimeout, aliasID)
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

	_, err = c.PostTxAndWaitUntilConfirmation(tx)
	return err
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
		MaxPayloadSize: parameters.MaxPayloadSize,
		Protocol:       &info.Protocol,
		BaseToken:      (*parameters.BaseToken)(info.BaseToken),
	})
}
