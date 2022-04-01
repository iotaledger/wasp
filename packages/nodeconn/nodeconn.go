// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// nodeconn package provides an interface to the L1 node (Hornet).
// This component is responsible for:
//   - Protocol details.
//   - Message reattachments and promotions.
//   - Management of PoW.
//
package nodeconn

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/wasp/packages/chain"
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
	l1params      *parameters.L1
	log           *logger.Logger
	config        L1Config
}

var _ chain.NodeConnection = &nodeConn{}

func L1ParamsFromInfoResp(info *nodeclient.InfoResponse) *parameters.L1 {
	return &parameters.L1{
		NetworkName:        info.Protocol.NetworkName,
		NetworkID:          iotago.NetworkIDFromString(info.Protocol.NetworkName),
		Bech32Prefix:       iotago.NetworkPrefix(info.Protocol.Bech32HRP),
		MaxTransactionSize: 32000, // TODO should be some const from iotago
		DeSerializationParameters: &iotago.DeSerializationParameters{
			RentStructure: &info.Protocol.RentStructure,
		},
	}
}

func newCtx(timeout ...time.Duration) (context.Context, context.CancelFunc) {
	t := defaultTimeout
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return context.WithTimeout(context.Background(), t)
}

func New(config L1Config, log *logger.Logger, timeout ...time.Duration) chain.NodeConnection {
	return newNodeConn(config, log, timeout...)
}

func newNodeConn(config L1Config, log *logger.Logger, timeout ...time.Duration) *nodeConn {
	ctx, ctxCancel := context.WithCancel(context.Background())
	nodeAPIClient := nodeclient.New(fmt.Sprintf("http://%s:%d", config.Hostname, config.APIPort))

	ctxWithTimeout, cancelContext := newCtx(timeout...)
	defer cancelContext()
	l1Info, err := nodeAPIClient.Info(ctxWithTimeout)
	if err != nil {
		panic(xerrors.Errorf("error getting L1 connection info: %w", err))
	}

	mqttClient, err := nodeAPIClient.EventAPI(ctx)
	if err != nil {
		panic(xerrors.Errorf("error getting node event client: %w", err))
	}

	indexerClient, err := nodeAPIClient.Indexer(ctx)
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
			handler.(chain.NodeConnectionMilestonesHandlerFun)(params[0].(*nodeclient.MilestonePointer))
		}),
		l1params: L1ParamsFromInfoResp(l1Info),
		log:      log.Named("nc"),
		config:   config,
	}
	go nc.run()
	return &nc
}

func (nc *nodeConn) L1Params() *parameters.L1 {
	return nc.l1params
}

// RegisterChain implements chain.NodeConnection.
func (nc *nodeConn) RegisterChain(
	chainAddr iotago.Address,
	stateOutputHandler func(iotago.OutputID, iotago.Output),
	outputHandler func(iotago.OutputID, iotago.Output),
) {
	ncc := newNCChain(nc, chainAddr, stateOutputHandler, outputHandler)
	nc.chainsLock.Lock()
	defer nc.chainsLock.Unlock()
	nc.chains[ncc.Key()] = ncc
	nc.log.Debugf("nodeconn: chain registered: %s", chainAddr)
}

// UnregisterChain implements chain.NodeConnection.
func (nc *nodeConn) UnregisterChain(chainAddr iotago.Address) {
	nccKey := chainAddr.Key()
	nc.chainsLock.Lock()
	defer nc.chainsLock.Unlock()
	if ncc, ok := nc.chains[nccKey]; ok {
		ncc.Close()
		delete(nc.chains, nccKey)
	}
	nc.log.Debugf("nodeconn: chain unregistered: %s", chainAddr)
}

// PublishTransaction implements chain.NodeConnection.
func (nc *nodeConn) PublishTransaction(chainAddr iotago.Address, stateIndex uint32, tx *iotago.Transaction) error {
	nc.chainsLock.RLock()
	ncc, ok := nc.chains[chainAddr.Key()]
	nc.chainsLock.RUnlock()
	if !ok {
		return xerrors.Errorf("Chain %v is not connected.", chainAddr.String())
	}
	return ncc.PublishTransaction(tx)
}

func (nc *nodeConn) AttachTxInclusionStateEvents(chainAddr iotago.Address, handler chain.NodeConnectionInclusionStateHandlerFun) (*events.Closure, error) {
	nc.chainsLock.RLock()
	ncc, ok := nc.chains[chainAddr.Key()]
	nc.chainsLock.RUnlock()
	if !ok {
		return nil, xerrors.Errorf("Chain %v is not connected.", chainAddr.String())
	}
	closure := events.NewClosure(handler)
	ncc.inclusionStates.Attach(closure)
	return closure, nil
}

func (nc *nodeConn) DetachTxInclusionStateEvents(chainAddr iotago.Address, closure *events.Closure) error {
	nc.chainsLock.RLock()
	ncc, ok := nc.chains[chainAddr.Key()]
	nc.chainsLock.RUnlock()
	if !ok {
		return xerrors.Errorf("Chain %v is not connected.", chainAddr.String())
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

func (nc *nodeConn) PullLatestOutput(chainAddr iotago.Address) {
	// TODO
}

func (nc *nodeConn) PullTxInclusionState(chainAddr iotago.Address, txid iotago.TransactionID) {
	// TODO
}

func (nc *nodeConn) PullOutputByID(chainAddr iotago.Address, id *iotago.UTXOInput) {
	// TODO
}

func (nc *nodeConn) GetMetrics() nodeconnmetrics.NodeConnectionMetrics {
	// TODO
	return nil
}

const pollConfirmedTxInterval = 200 * time.Millisecond

// waitUntilConfirmed waits until a given tx message is confirmed, it takes care of promotions/re-attachments for that message
func (nc *nodeConn) waitUntilConfirmed(ctx context.Context, txMsg *iotago.Message) error {
	// wait until tx is confirmed
	msgID, err := txMsg.ID()
	if err != nil {
		return xerrors.Errorf("failed to get msg ID: %w", err)
	}

	// poll the node by getting `MessageMetadataByMessageID`
	for {
		metadataResp, err := nc.nodeAPIClient.MessageMetadataByMessageID(ctx, *msgID)
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
			_, err = nc.nodeAPIClient.SubmitMessage(ctx, promotionMsg, nc.l1params.DeSerializationParameters)
			if err != nil {
				return xerrors.Errorf("failed to promote msg: %w", err)
			}
		}
		if metadataResp.ShouldReattach != nil && *metadataResp.ShouldReattach {
			// remote PoW: Take the message, clear parents, clear nonce, send to node
			txMsg.Parents = nil
			txMsg.Nonce = 0
			txMsg, err = nc.nodeAPIClient.SubmitMessage(ctx, txMsg, nc.l1params.DeSerializationParameters)
			if err != nil {
				return xerrors.Errorf("failed to get re-attach msg: %w", err)
			}
		}

		if err = ctx.Err(); err != nil {
			return xerrors.Errorf("failed to wait for tx confimation within timeout: %s", err)
		}
		time.Sleep(pollConfirmedTxInterval)
	}
}
