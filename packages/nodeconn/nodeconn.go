// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// nodeconn package provides an interface to the L1 node (Hornet).
// This component is responsible for:
//   - Protocol details.
//   - Block reattachments and promotions.
package nodeconn

import (
	"context"
	"sync"
	"time"

	hivecore "github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/inx-app/nodebridge"
	inx "github.com/iotaledger/inx/go"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

type ChainL1Config struct {
	INXAddress   string
	UseRemotePoW bool
}

type LedgerUpdateHandler func(*nodebridge.LedgerUpdate)

// nodeconn implements chain.NodeConnection.
// Single Wasp node is expected to connect to a single L1 node, thus
// we expect to have a single instance of this structure.
type nodeConn struct {
	ctx           context.Context
	ctxCancel     context.CancelFunc
	chains        map[string]*ncChain // key = iotago.Address.Key()
	chainsLock    sync.RWMutex
	indexerClient nodeclient.IndexerClient
	milestones    *events.Event
	metrics       nodeconnmetrics.NodeConnectionMetrics
	log           *logger.Logger
	config        ChainL1Config
	nodeBridge    *nodebridge.NodeBridge
	nodeClient    *nodeclient.Client
}

var _ chain.NodeConnection = &nodeConn{}

func setL1ProtocolParams(info *nodeclient.InfoResponse) {
	parameters.InitL1(&parameters.L1Params{
		// There are no limits on how big from a size perspective an essence can be, so it is just derived from 32KB - Block fields without payload = max size of the payload
		MaxTransactionSize: 32000, // TODO should this value come from the API in the future? or some const in iotago?
		Protocol:           &info.Protocol,
		BaseToken:          (*parameters.BaseToken)(info.BaseToken),
	})
}

const defaultTimeout = 1 * time.Minute

func newCtx(ctx context.Context, timeout ...time.Duration) (context.Context, context.CancelFunc) {
	t := defaultTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return context.WithTimeout(ctx, t)
}

func New(config ChainL1Config, log *logger.Logger, timeout ...time.Duration) chain.NodeConnection {
	ctx, ctxCancel := context.WithCancel(context.Background())

	ctxWithTimeout, cancelContext := newCtx(ctx, timeout...)
	defer cancelContext()

	nb, err := nodebridge.NewNodeBridge(ctxWithTimeout, config.INXAddress, log.Named("NodeBridge"))
	if err != nil {
		panic(err)
	}

	go nb.Run(context.Background())
	inxNodeClient := nb.INXNodeClient()

	nodeInfo, err := inxNodeClient.Info(ctxWithTimeout)
	if err != nil {
		panic(xerrors.Errorf("error getting node info: %w", err))
	}

	setL1ProtocolParams(nodeInfo)

	indexerClient, err := inxNodeClient.Indexer(ctxWithTimeout)
	if err != nil {
		panic(xerrors.Errorf("failed to get nodeclient indexer: %v", err))
	}

	nc := nodeConn{
		ctx:           ctx,
		ctxCancel:     ctxCancel,
		chains:        make(map[string]*ncChain),
		chainsLock:    sync.RWMutex{},
		indexerClient: indexerClient,
		milestones: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(chain.NodeConnectionMilestonesHandlerFun)(params[0].(*nodeclient.MilestoneInfo))
		}),
		metrics:    nodeconnmetrics.NewEmptyNodeConnectionMetrics(),
		log:        log.Named("nc"),
		config:     config,
		nodeBridge: nb,
		nodeClient: inxNodeClient,
	}

	go nc.subscribeToLedgerUpdates()
	nc.enableMilestoneTrigger()

	return &nc
}

func (nc *nodeConn) subscribeToLedgerUpdates() {
	err := nc.nodeBridge.ListenToLedgerUpdates(nc.ctx, 0, 0, nc.handleLedgerUpdate)
	if err != nil {
		nc.log.Panic(err)
	}
}

func (nc *nodeConn) handleLedgerUpdate(update *nodebridge.LedgerUpdate) error {
	nc.chainsLock.RLock()
	defer nc.chainsLock.RUnlock()

	for _, ledgerOutput := range update.Created {
		output, err := ledgerOutput.UnwrapOutput(serializer.DeSeriModeNoValidation, nc.nodeBridge.ProtocolParameters())
		if err != nil {
			return err
		}

		// notify chains about state updates
		if aliasOutput, ok := output.(*iotago.AliasOutput); ok {
			outputID := ledgerOutput.GetOutputId().Unwrap()
			aliasID := util.AliasIDFromAliasOutput(aliasOutput, outputID)
			chainID := isc.ChainIDFromAliasID(aliasID)
			ncChain := nc.chains[chainID.Key()]
			if ncChain != nil {
				ncChain.HandleStateUpdate(outputID, aliasOutput)
			}
		}

		// notify chains about new UTXOS owned by them
		unlockAddr := output.UnlockConditionSet().Address()
		if unlockAddr == nil {
			continue
		}
		if unlockAliasAddr, ok := unlockAddr.Address.(*iotago.AliasAddress); ok {
			chainID := isc.ChainIDFromAliasID(unlockAliasAddr.AliasID())
			ncChain := nc.chains[chainID.Key()]
			if ncChain != nil {
				ncChain.HandleUnlockableOutput(ledgerOutput.GetOutputId().Unwrap(), output)
			}
		}
	}
	return nil
}

func (nc *nodeConn) enableMilestoneTrigger() {
	nc.nodeBridge.Events.LatestMilestoneChanged.Hook(hivecore.NewClosure(func(metadata *nodebridge.Milestone) {
		milestone := nodeclient.MilestoneInfo{
			MilestoneID: metadata.MilestoneID.String(),
			Index:       metadata.Milestone.Index,
			Timestamp:   metadata.Milestone.Timestamp,
		}

		nc.log.Debugf("Milestone received, index=%v, timestamp=%v", milestone.Index, milestone.Timestamp)

		nc.metrics.GetInMilestone().CountLastMessage(milestone)
		nc.milestones.Trigger(&milestone)
	}))
}

func (nc *nodeConn) SetMetrics(metrics nodeconnmetrics.NodeConnectionMetrics) {
	nc.metrics = metrics
}

// RegisterChain implements chain.NodeConnection.
func (nc *nodeConn) RegisterChain(
	chainID *isc.ChainID,
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
func (nc *nodeConn) UnregisterChain(chainID *isc.ChainID) {
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
func (nc *nodeConn) PublishStateTransaction(chainID *isc.ChainID, stateIndex uint32, tx *iotago.Transaction) error {
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
func (nc *nodeConn) PublishGovernanceTransaction(chainID *isc.ChainID, tx *iotago.Transaction) error {
	nc.chainsLock.RLock()
	ncc, ok := nc.chains[chainID.Key()]
	nc.chainsLock.RUnlock()
	if !ok {
		return xerrors.Errorf("Chain %v is not connected.", chainID.String())
	}
	return ncc.PublishTransaction(tx)
}

func (nc *nodeConn) AttachTxInclusionStateEvents(chainID *isc.ChainID, handler chain.NodeConnectionInclusionStateHandlerFun) (*events.Closure, error) {
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

func (nc *nodeConn) DetachTxInclusionStateEvents(chainID *isc.ChainID, closure *events.Closure) error {
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

func (nc *nodeConn) PullLatestOutput(chainID *isc.ChainID) {
	ncc := nc.chains[chainID.Key()]
	if ncc == nil {
		nc.log.Errorf("PullLatestOutput: NCChain not  found for chainID %s", chainID)
		return
	}
	ncc.queryLatestChainStateUTXO()
}

func (nc *nodeConn) PullTxInclusionState(chainID *isc.ChainID, txid iotago.TransactionID) {
	// TODO - is this needed? - output should come from INX subscription
	// we are also constantly polling for confirmation in the promotion/reattachment logic
}

func (nc *nodeConn) PullStateOutputByID(chainID *isc.ChainID, id *iotago.UTXOInput) {
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

func (nc *nodeConn) doPostTx(ctx context.Context, tx *iotago.Transaction) (*iotago.BlockID, error) {
	// Build a Block and post it.
	block, err := builder.NewBlockBuilder().
		Payload(tx).
		Build()
	if err != nil {
		return nil, xerrors.Errorf("failed to build a tx: %w", err)
	}

	blockID, err := nc.nodeBridge.SubmitBlock(ctx, block)
	if err != nil {
		return nil, xerrors.Errorf("failed to submit a tx: %w", err)
	}

	return &blockID, nil
}

const pollConfirmedTxInterval = 200 * time.Millisecond

// waitUntilConfirmed waits until a given tx Block is confirmed, it takes care of promotions/re-attachments for that Block
func (nc *nodeConn) waitUntilConfirmed(ctx context.Context, blockID *iotago.BlockID, tx *iotago.Transaction) error {
	// wait until tx is confirmed
	// poll the node by getting `BlockMetadataByBlockID`

	currentBlockID := blockID
	for {
		metadataResp, err := nc.nodeBridge.BlockMetadata(ctx, *currentBlockID)
		if err != nil {
			return xerrors.Errorf("failed to get msg metadata: %w", err)
		}

		if metadataResp.ReferencedByMilestoneIndex != 0 {
			if metadataResp.LedgerInclusionState == inx.BlockMetadata_LEDGER_INCLUSION_STATE_INCLUDED {
				return nil // success
			}
			return xerrors.Errorf("tx was not included in the ledger. LedgerInclusionState: %s, ConflictReason: %d",
				metadataResp.LedgerInclusionState, metadataResp.ConflictReason)
		}
		// reattach or promote if needed

		if metadataResp.ShouldPromote {
			nc.log.Debugf("promoting msgID: %s", currentBlockID.ToHex())
			// create an empty Block and the BlockID as one of the parents
			tipsResp, err := nc.nodeClient.Tips(ctx)
			if err != nil {
				return xerrors.Errorf("failed to fetch Tips: %w", err)
			}
			tips, err := tipsResp.Tips()
			if err != nil {
				return xerrors.Errorf("failed to get Tips from tips response: %w", err)
			}

			parents := []iotago.BlockID{
				*currentBlockID,
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
			_, err = nc.nodeClient.SubmitBlock(ctx, promotionMsg, parameters.L1().Protocol)
			if err != nil {
				return xerrors.Errorf("failed to promote msg: %w", err)
			}
		}

		if metadataResp.ShouldReattach {
			nc.log.Debugf("reattaching tx")
			currentBlockID, err = nc.doPostTx(ctx, tx)
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
