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

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	iotagox "github.com/iotaledger/iota.go/v3/x"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"golang.org/x/xerrors"
)

// nodeconn implements chain.NodeConnection.
// Single Wasp node is expected to connect to a single L1 node, thus
// we expect to have a single instance of this structure.
type nodeConn struct {
	ctx        context.Context
	ctxCancel  context.CancelFunc
	chains     map[string]*ncChain // key = iotago.Address.Key()
	chainsLock sync.RWMutex
	nodeClient *nodeclient.Client
	nodeEvents *iotagox.NodeEventAPIClient
	milestones *events.Event
	net        peering.NetworkProvider
	log        *logger.Logger
}

var _ chain.NodeConnection = &nodeConn{}

func New(nodeHost string, nodePort int, net peering.NetworkProvider, log *logger.Logger) chain.NodeConnection {
	ctx, ctxCancel := context.WithCancel(context.Background())
	nodeClient := nodeclient.New(
		fmt.Sprintf("http://%s:%d", nodeHost, nodePort),
		iotago.ZeroRentParas, // TODO: ...
		nodeclient.WithIndexer(),
	)
	nodeEvents := iotagox.NewNodeEventAPIClient(
		fmt.Sprintf("ws://%s:%d/mqtt", nodeHost, nodePort),
	)
	nc := nodeConn{
		ctx:        ctx,
		ctxCancel:  ctxCancel,
		chains:     make(map[string]*ncChain),
		chainsLock: sync.RWMutex{},
		nodeClient: nodeClient,
		nodeEvents: nodeEvents,
		milestones: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(chain.NodeConnectionMilestonesHandlerFun)(params[0].(*iotagox.MilestonePointer))
		}),
		net: net,
		log: log.Named("nc"),
	}
	go nc.run()
	return &nc
}

// RegisterChain implements chain.NodeConnection. // TODO -> ConnectChain.
func (nc *nodeConn) RegisterChain(chainAddr iotago.Address, outputHandler func(iotago.OutputID, iotago.Output)) {
	ncc := newNCChain(nc, chainAddr, outputHandler)
	nc.chainsLock.Lock()
	defer nc.chainsLock.Unlock()
	nc.chains[ncc.Key()] = ncc
}

// UnregisterChain implements chain.NodeConnection. // TODO -> DisconnectChain.
func (nc *nodeConn) UnregisterChain(chainAddr iotago.Address) {
	nccKey := chainAddr.Key()
	nc.chainsLock.Lock()
	defer nc.chainsLock.Unlock()
	if ncc, ok := nc.chains[nccKey]; ok {
		ncc.Close()
		delete(nc.chains, nccKey)
	}
}

// PublishTransaction implements chain.NodeConnection.
func (nc *nodeConn) PublishTransaction(chainAddr iotago.Address, stateIndex uint32, tx *iotago.Transaction) error {
	nc.chainsLock.RLock()
	ncc, ok := nc.chains[chainAddr.Key()]
	nc.chainsLock.RUnlock()
	if !ok {
		return xerrors.Errorf("Chain %v is not connected.", chainAddr.String())
	}
	return ncc.PublishTransaction(stateIndex, tx)
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
