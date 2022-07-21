// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// nodeconn package provides an interface to the L1 node (Hornet).
// This component is responsible for:
//   - Protocol details.
//   - Block reattachments and promotions.
//   - Management of PoW.
//
package nodeconn

import (
	"context"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"golang.org/x/xerrors"
)

// nodeconn implements chain.NodeConnection.
// Single Wasp node is expected to connect to a single L1 node, thus
// we expect to have a single instance of this structure.
type nodeConn struct {
	ctx           context.Context
	ctxCancel     context.CancelFunc
	chains        map[string]*ncChain // key = iotago.Address.Key()
	chainsLock    sync.RWMutex
	nodeAPIClient *nodeclient.Client
	mqttClient    *nodeclient.EventAPIClient
	indexerClient nodeclient.IndexerClient
	milestones    *events.Event
	metrics       nodeconnmetrics.NodeConnectionMetrics
	log           *logger.Logger
	config        L1Config
}

var _ chain.NodeConnection = &nodeConn{}

func setL1ProtocolParams(info *nodeclient.InfoResponse) {
	parameters.L1 = &parameters.L1Params{
		// There are no limits on how big from a size perspective an essence can be, so it is just derived from 32KB - Block fields without payload = max size of the payload
		MaxTransactionSize: 32000, // TODO should this value come from the API in the future? or some const in iotago?
		Protocol:           &info.Protocol,
		BaseToken:          (*parameters.BaseToken)(info.BaseToken),
	}
}

func newCtx(ctx context.Context, timeout ...time.Duration) (context.Context, context.CancelFunc) {
	t := defaultTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return context.WithTimeout(ctx, t)
}

func New(config L1Config, log *logger.Logger, timeout ...time.Duration) chain.NodeConnection {
	return newNodeConn(config, log, true, timeout...)
}

func newNodeConn(config L1Config, log *logger.Logger, initMqttClient bool, timeout ...time.Duration) *nodeConn {
	ctx, ctxCancel := context.WithCancel(context.Background())
	nodeAPIClient := nodeclient.New(config.APIAddress)

	ctxWithTimeout, cancelContext := newCtx(ctx, timeout...)
	defer cancelContext()
	l1Info, err := nodeAPIClient.Info(ctxWithTimeout)
	if err != nil {
		panic(xerrors.Errorf("error getting L1 connection info: %w", err))
	}
	setL1ProtocolParams(l1Info)

	mqttClient, err := nodeAPIClient.EventAPI(ctxWithTimeout)
	if err != nil {
		panic(xerrors.Errorf("error getting node event client: %w", err))
	}

	indexerClient, err := nodeAPIClient.Indexer(ctxWithTimeout)
	if err != nil {
		panic(xerrors.Errorf("failed to get nodeclient indexer: %v", err))
	}
	nc := nodeConn{
		ctx:           ctx,
		ctxCancel:     ctxCancel,
		chains:        make(map[string]*ncChain),
		chainsLock:    sync.RWMutex{},
		nodeAPIClient: nodeAPIClient,
		mqttClient:    mqttClient,
		indexerClient: indexerClient,
		milestones: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(chain.NodeConnectionMilestonesHandlerFun)(params[0].(*nodeclient.MilestoneInfo))
		}),
		metrics: nodeconnmetrics.NewEmptyNodeConnectionMetrics(),
		log:     log.Named("nc"),
		config:  config,
	}

	if initMqttClient {
		go nc.run()
		for {
			if nc.mqttClient.MQTTClient.IsConnected() {
				break
			}
			nc.log.Debugf("waiting until mqtt client is connected")
			time.Sleep(1 * time.Second)
		}
	}
	return &nc
}

func (nc *nodeConn) SetMetrics(metrics nodeconnmetrics.NodeConnectionMetrics) {
	nc.metrics = metrics
}

// RegisterChain implements chain.NodeConnection.
func (nc *nodeConn) RegisterChain(
	chainID *iscp.ChainID,
	stateOutputHandler,
	outputHandler func(iotago.OutputID, iotago.Output),
) {
	nc.metrics.SetRegistered(chainID)
	ncc := newNCChain(nc, chainID, stateOutputHandler, outputHandler)
	nc.chainsLock.Lock()
	defer nc.chainsLock.Unlock()
	nc.chains[chainID.Key()] = ncc
	nc.log.Debugf("nodeconn: chain registered: %s", chainID)
}

// UnregisterChain implements chain.NodeConnection.
func (nc *nodeConn) UnregisterChain(chainID *iscp.ChainID) {
	nc.metrics.SetUnregistered(chainID)
	nccKey := chainID.Key()
	nc.chainsLock.Lock()
	defer nc.chainsLock.Unlock()
	if ncc, ok := nc.chains[nccKey]; ok {
		ncc.Close()
		delete(nc.chains, nccKey)
	}
	nc.log.Debugf("nodeconn: chain unregistered: %s", chainID)
}

// PublishStateTransaction implements chain.NodeConnection.
func (nc *nodeConn) PublishStateTransaction(chainID *iscp.ChainID, stateIndex uint32, tx *iotago.Transaction) error {
	nc.chainsLock.RLock()
	ncc, ok := nc.chains[chainID.Key()]
	nc.chainsLock.RUnlock()
	if !ok {
		return xerrors.Errorf("Chain %v is not connected.", chainID.String())
	}
	return ncc.PublishTransaction(tx)
}

// PublishGovernanceTransaction implements chain.NodeConnection.
// TODO: identical to PublishStateTransaction; needs to be reviewed
func (nc *nodeConn) PublishGovernanceTransaction(chainID *iscp.ChainID, tx *iotago.Transaction) error {
	nc.chainsLock.RLock()
	ncc, ok := nc.chains[chainID.Key()]
	nc.chainsLock.RUnlock()
	if !ok {
		return xerrors.Errorf("Chain %v is not connected.", chainID.String())
	}
	return ncc.PublishTransaction(tx)
}

func (nc *nodeConn) AttachTxInclusionStateEvents(chainID *iscp.ChainID, handler chain.NodeConnectionInclusionStateHandlerFun) (*events.Closure, error) {
	nc.chainsLock.RLock()
	ncc, ok := nc.chains[chainID.Key()]
	nc.chainsLock.RUnlock()
	if !ok {
		return nil, xerrors.Errorf("Chain %v is not connected.", chainID.String())
	}
	closure := events.NewClosure(handler)
	ncc.inclusionStates.Attach(closure)
	return closure, nil
}

func (nc *nodeConn) DetachTxInclusionStateEvents(chainID *iscp.ChainID, closure *events.Closure) error {
	nc.chainsLock.RLock()
	ncc, ok := nc.chains[chainID.Key()]
	nc.chainsLock.RUnlock()
	if !ok {
		return xerrors.Errorf("Chain %v is not connected.", chainID.String())
	}
	ncc.inclusionStates.Detach(closure)
	return nil
}

// AttachMilestones implements chain.NodeConnection.
func (nc *nodeConn) AttachMilestones(handler chain.NodeConnectionMilestonesHandlerFun) *events.Closure {
	closure := events.NewClosure(handler)
	nc.milestones.Attach(closure)
	return closure
}

// DetachMilestones implements chain.NodeConnection.
func (nc *nodeConn) DetachMilestones(attachID *events.Closure) {
	nc.milestones.Detach(attachID)
}

func (nc *nodeConn) Close() {
	nc.ctxCancel()
}

func (nc *nodeConn) PullLatestOutput(chainID *iscp.ChainID) {
	ncc := nc.chains[chainID.Key()]
	if ncc == nil {
		nc.log.Errorf("PullLatestOutput: NCChain not  found for chainID %s", chainID)
		return
	}
	ncc.queryLatestChainStateUTXO()
}

func (nc *nodeConn) PullTxInclusionState(chainID *iscp.ChainID, txid iotago.TransactionID) {
	// TODO - is this needed? - output should come from MQTT subscription
	// we are also constantly polling for confirmation in the promotion/reattachment logic
}

func (nc *nodeConn) PullStateOutputByID(chainID *iscp.ChainID, id *iotago.UTXOInput) {
	ncc := nc.chains[chainID.Key()]
	if ncc == nil {
		nc.log.Errorf("PullOutputByID: NCChain not  found for chainID %s", chainID)
		return
	}
	ncc.PullStateOutputByID(id.ID())
}

func (nc *nodeConn) GetMetrics() nodeconnmetrics.NodeConnectionMetrics {
	return nc.metrics
}

func (nc *nodeConn) doPostTx(ctx context.Context, tx *iotago.Transaction) (*iotago.Block, error) {
	// Build a Block and post it.
	block, err := builder.NewBlockBuilder().
		Payload(tx).
		Tips(ctx, nc.nodeAPIClient).
		Build()
	if err != nil {
		return nil, xerrors.Errorf("failed to build a tx: %w", err)
	}
	err = nc.doPoW(ctx, block)
	if err != nil {
		return nil, xerrors.Errorf("failed duing PoW: %w", err)
	}
	block, err = nc.nodeAPIClient.SubmitBlock(ctx, block, parameters.L1.Protocol)
	if err != nil {
		return nil, xerrors.Errorf("failed to submit a tx: %w", err)
	}
	txID, err := tx.ID()
	if err == nil {
		nc.log.Debugf("Posted transaction id %v", iscp.TxID(txID))
	} else {
		nc.log.Warnf("Posted transaction; failed to calculate its id: %v", err)
	}
	return block, nil
}

const pollConfirmedTxInterval = 200 * time.Millisecond

// waitUntilConfirmed waits until a given tx Block is confirmed, it takes care of promotions/re-attachments for that Block
func (nc *nodeConn) waitUntilConfirmed(ctx context.Context, block *iotago.Block) error {
	// wait until tx is confirmed
	msgID, err := block.ID()
	if err != nil {
		return xerrors.Errorf("failed to get msg ID: %w", err)
	}

	// poll the node by getting `BlockMetadataByBlockID`
	for {
		metadataResp, err := nc.nodeAPIClient.BlockMetadataByBlockID(ctx, msgID)
		if err != nil {
			return xerrors.Errorf("failed to get msg metadata: %w", err)
		}

		if metadataResp.ReferencedByMilestoneIndex != 0 {
			if metadataResp.LedgerInclusionState != "" && metadataResp.LedgerInclusionState == "included" {
				return nil // success
			}
			return xerrors.Errorf("tx was not included in the ledger. LedgerInclusionState: %s, ConflictReason: %d",
				metadataResp.LedgerInclusionState, metadataResp.ConflictReason)
		}
		// reattach or promote if needed
		if metadataResp.ShouldPromote != nil && *metadataResp.ShouldPromote {
			nc.log.Debugf("promoting msgID: %s", msgID.ToHex())
			// create an empty Block and the BlockID as one of the parents
			tipsResp, err := nc.nodeAPIClient.Tips(ctx)
			if err != nil {
				return xerrors.Errorf("failed to fetch Tips: %w", err)
			}
			tips, err := tipsResp.Tips()
			if err != nil {
				return xerrors.Errorf("failed to get Tips from tips response: %w", err)
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
				return xerrors.Errorf("failed to build promotion Block: %w", err)
			}
			_, err = nc.nodeAPIClient.SubmitBlock(ctx, promotionMsg, parameters.L1.Protocol)
			if err != nil {
				return xerrors.Errorf("failed to promote msg: %w", err)
			}
		}
		if metadataResp.ShouldReattach != nil && *metadataResp.ShouldReattach {
			nc.log.Debugf("reattaching block: %s", block)
			err = nc.doPoW(ctx, block)
			if err != nil {
				return err
			}
		}
		if err = ctx.Err(); err != nil {
			return xerrors.Errorf("failed to wait for tx confimation within timeout: %s", err)
		}
		time.Sleep(pollConfirmedTxInterval)
	}
}

const refreshTipsDuringPoWInterval = 5 * time.Second

func (nc *nodeConn) doPoW(ctx context.Context, block *iotago.Block) error {
	if nc.config.UseRemotePoW {
		// remote PoW: Take the Block, clear parents, clear nonce, send to node
		block.Parents = nil
		block.Nonce = 0
		_, err := nc.nodeAPIClient.SubmitBlock(ctx, block, parameters.L1.Protocol)
		return err
	}
	// do the PoW
	refreshTipsFn := func() (tips iotago.BlockIDs, err error) {
		// refresh tips if PoW takes longer than a configured duration.
		resp, err := nc.nodeAPIClient.Tips(ctx)
		if err != nil {
			return nil, err
		}
		return resp.Tips()
	}

	return doPoW(
		ctx,
		block,
		refreshTipsDuringPoWInterval,
		refreshTipsFn,
	)
}
