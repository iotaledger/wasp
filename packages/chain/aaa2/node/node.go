// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This runs single chain will all the committees, mempool, state mgr etc.
// The main task for this package to run the protocol as in a threaded environment,
// communicate between ChainMgr, Mempool, StateMgr, NodeConn and ConsensusInstances.
//
// The following threads (goroutines) are running for a chain:
//   - ChainMgr (the main synchronizing thread)
//   - Mempool
//   - StateMgr
//   - Consensus (a thread for each instance).
//
// This object interacts with:
//   - NodeConn.
//   - Administrative functions.
package node

import (
	"context"
	"time"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/aaa2/chainMgr"
	consGR "github.com/iotaledger/wasp/packages/chain/aaa2/cons/gr"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

const (
	recoveryTimeout         time.Duration = 15 * time.Minute // TODO: Make it configurable?
	redeliveryPeriod        time.Duration = 3 * time.Second  // TODO: Make it configurable?
	printStatusPeriod       time.Duration = 10 * time.Second // TODO: Make it configurable?
	consensusInstsInAdvance int           = 3                // TODO: Make it configurable?
)

type ChainNode interface {
	// TODO: All the public administrative functions.
}

type ChainMempool interface {
	consGR.Mempool
	ReceiveOnLedgerRequest(request isc.OnLedgerRequest)
}

type ChainStateMgr interface {
	consGR.StateMgr
	ReceiveConfirmedAliasOutput(aliasOutput *isc.AliasOutputWithID)
}

type OutputHandler = func(outputID iotago.OutputID, output iotago.Output)

type TxPostHandler = func(tx *iotago.Transaction, confirmed bool)

type MilestoneHandler = func(timestamp time.Time)

type ChainNodeConn interface {
	// Publishing can be canceled via the context.
	// The result must be returned via the callback, unless ctx is canceled first.
	PublishTX(
		ctx context.Context,
		chainID *isc.ChainID,
		tx *iotago.Transaction,
		callback TxPostHandler,
	)
	// Alias outputs are expected to be returned in order. Considering the Hornet node, the rules are:
	//   - Upon Attach -- existing unspent alias output is returned FIRST.
	//   - Upon receiving a spent/unspent AO from L1 they are returned in
	//     the same order, as the milestones are issued.
	//   - If a single milestone has several alias outputs, they have to be ordered
	//     according to the chain of TXes.
	//
	// NOTE: Any out-of-order AO will be considered as a rollback or AO by the chain impl.
	AttachChain(
		ctx context.Context,
		chainID *isc.ChainID,
		recvRequestCB,
		recvAliasOutput OutputHandler,
		recvMilestone MilestoneHandler,
	)
}

type chainNodeImpl struct {
	me                  gpa.NodeID
	nodeIdentity        *cryptolib.KeyPair
	chainID             *isc.ChainID
	chainMgr            chainMgr.ChainMgr
	nodeConn            ChainNodeConn
	mempool             ChainMempool  // TODO: ...
	stateMgr            ChainStateMgr // TODO: ...
	recvAliasOutputPipe pipe.Pipe
	recvTxPublishedPipe pipe.Pipe
	recvMilestonePipe   pipe.Pipe
	consensusInsts      map[chainMgr.CommitteeID]map[journal.LogIndex]*consensusInst // Running consensus instances.
	publishingTXes      map[iotago.TransactionID]context.CancelFunc                  // TX'es now being published.
	procCache           *processors.Cache                                            // TODO: ...
	net                 peering.NetworkProvider
	log                 *logger.Logger
}

type consensusInst struct {
	aliasOutput *isc.AliasOutputWithID
	cancelFunc  context.CancelFunc
	consensus   *consGR.ConsGr
	outputCh    <-chan *consGR.Output
	recoverCh   <-chan time.Time
}

// This is event received from the NodeConn as response to PublishTX
type txPublished struct {
	committeeID chainMgr.CommitteeID
	txID        iotago.TransactionID
	confirmed   bool
}

var _ ChainNode = &chainNodeImpl{}

func New(
	ctx context.Context,
	chainID *isc.ChainID,
	nodeConn ChainNodeConn,
	nodeIdentity *cryptolib.KeyPair,
	net peering.NetworkProvider,
	log *logger.Logger,
) ChainNode {
	cni := &chainNodeImpl{
		me:                  pubKeyAsNodeID(nodeIdentity.GetPublicKey()),
		nodeIdentity:        nodeIdentity,
		chainID:             chainID,
		chainMgr:            nil, // TODO: chainMgr.New(), // TODO: AckHandler and ticks.
		nodeConn:            nodeConn,
		mempool:             nil, // TODO: ...
		stateMgr:            nil, // TODO: ...
		recvAliasOutputPipe: pipe.NewDefaultInfinitePipe(),
		recvTxPublishedPipe: pipe.NewDefaultInfinitePipe(),
		recvMilestonePipe:   pipe.NewDefaultInfinitePipe(),
		consensusInsts:      map[chainMgr.CommitteeID]map[journal.LogIndex]*consensusInst{},
		publishingTXes:      map[iotago.TransactionID]context.CancelFunc{},
		net:                 net,
		log:                 log,
	}
	recvAliasOutputPipeInCh := cni.recvAliasOutputPipe.In()
	recvAliasOutputCB := func(outputID iotago.OutputID, output iotago.Output) {
		aliasOutput := isc.NewAliasOutputWithID(output.(*iotago.AliasOutput), outputID.UTXOInput())
		cni.stateMgr.ReceiveConfirmedAliasOutput(aliasOutput)
		recvAliasOutputPipeInCh <- aliasOutput
	}
	recvRequestCB := func(outputID iotago.OutputID, output iotago.Output) {
		req, err := isc.OnLedgerFromUTXO(output, outputID.UTXOInput())
		if err != nil {
			cni.log.Warnf("Cannot create OnLedgerRequest from output: %v", err)
			return
		}
		cni.mempool.ReceiveOnLedgerRequest(req)
	}
	recvMilestonePipeInCh := cni.recvMilestonePipe.In()
	recvMilestoneCB := func(timestamp time.Time) {
		recvMilestonePipeInCh <- timestamp
	}
	nodeConn.AttachChain(ctx, chainID, recvRequestCB, recvAliasOutputCB, recvMilestoneCB)
	go cni.run(ctx)
	return cni
}

func (cni *chainNodeImpl) run(ctx context.Context) {
	ctxDone := ctx.Done()
	recvAliasOutputPipeOutCh := cni.recvAliasOutputPipe.Out()
	recvTxPublishedPipeOutCh := cni.recvTxPublishedPipe.Out()
	recvMilestonePipeOutCh := cni.recvMilestonePipe.Out()
	for {
		select {
		case txPublishResult, ok := <-recvTxPublishedPipeOutCh:
			if !ok {
				recvTxPublishedPipeOutCh = nil
				continue
			}
			cni.handleTxPublished(ctx, txPublishResult.(*txPublished))
		case aliasOutput, ok := <-recvAliasOutputPipeOutCh:
			if !ok {
				recvAliasOutputPipeOutCh = nil
				continue
			}
			cni.handleAliasOutput(ctx, aliasOutput.(*isc.AliasOutputWithID))
		case timestamp, ok := <-recvMilestonePipeOutCh:
			if !ok {
				recvMilestonePipeOutCh = nil
			}
			cni.handleMilestoneTimestamp(timestamp.(time.Time))
		case <-ctxDone:
			return
		}
	}
}

func (cni *chainNodeImpl) handleTxPublished(ctx context.Context, txPubResult *txPublished) {
	if _, ok := cni.publishingTXes[txPubResult.txID]; !ok {
		return
	}
	delete(cni.publishingTXes, txPubResult.txID)
	msg := chainMgr.NewMsgChainTxPublishResult(cni.me, txPubResult.committeeID, txPubResult.txID, txPubResult.confirmed)
	outMsgs := cni.chainMgr.AsGPA().Message(msg)
	if outMsgs.Count() != 0 { // TODO: Wrong, NextLI will be exchanged.
		panic("unexpected messages from the chainMgr")
	}
	cni.handleChainMgrOutput(ctx, cni.chainMgr.AsGPA().Output())
}

func (cni *chainNodeImpl) handleAliasOutput(ctx context.Context, aliasOutput *isc.AliasOutputWithID) {
	msg := chainMgr.NewMsgAliasOutputConfirmed(cni.me, aliasOutput)
	outMsgs := cni.chainMgr.AsGPA().Message(msg)
	if outMsgs.Count() != 0 { // TODO: Wrong, NextLI will be exchanged.
		panic(xerrors.Errorf("unexpected messaged from chainMgr: %+v", outMsgs))
	}
	cni.handleChainMgrOutput(ctx, cni.chainMgr.AsGPA().Output())
}

func (cni *chainNodeImpl) handleMilestoneTimestamp(timestamp time.Time) {
	for ji := range cni.consensusInsts {
		for li := range cni.consensusInsts[ji] {
			ci := cni.consensusInsts[ji][li]
			if ci.cancelFunc != nil {
				ci.consensus.Time(timestamp)
			}
		}
	}
}

func (cni *chainNodeImpl) handleChainMgrOutput(ctx context.Context, outputUntyped gpa.Output) {
	if outputUntyped == nil {
		cni.cleanupConsensusInsts(nil, nil)
		cni.cleanupPublishingTXes([]chainMgr.NeedPostedTX{})
		return
	}
	output := outputUntyped.(*chainMgr.Output)
	//
	// Start new consensus instances, if needed.
	if output.NeedConsensus != nil {
		ci := cni.ensureConsensusInst(ctx, output.NeedConsensus)
		if ci.aliasOutput == nil {
			outputCh, recoverCh := ci.consensus.Input(output.NeedConsensus.BaseAliasOutput)
			ci.aliasOutput = output.NeedConsensus.BaseAliasOutput
			ci.outputCh = outputCh   // TODO: Read from these.
			ci.recoverCh = recoverCh // TODO: Read from these.
		}
	}
	//
	// Start publishing TX'es, if there not being posted already.
	for i := range output.NeedPostTXes {
		txToPost := output.NeedPostTXes[i] // Have to take a copy to be used in callback.
		if _, ok := cni.publishingTXes[txToPost.TxID]; !ok {
			subCtx, subCancel := context.WithCancel(ctx)
			cni.publishingTXes[txToPost.TxID] = subCancel
			cni.nodeConn.PublishTX(subCtx, cni.chainID, txToPost.Tx, func(tx *iotago.Transaction, confirmed bool) {
				cni.recvTxPublishedPipe.In() <- &txPublished{
					committeeID: txToPost.CommitteeID, // TODO: Was journalID: txToPost.JournalID,
					txID:        txToPost.TxID,
					confirmed:   confirmed,
				}
			})
		}
	}
	cni.cleanupPublishingTXes(output.NeedPostTXes)
}

func (cni *chainNodeImpl) ensureConsensusInst(ctx context.Context, needConsensus *chainMgr.NeedConsensus) *consensusInst {
	committeeID := needConsensus.CommitteeID
	logIndex := needConsensus.LogIndex
	dkShare := needConsensus.DKShare
	if _, ok := cni.consensusInsts[committeeID]; !ok {
		cni.consensusInsts[committeeID] = map[journal.LogIndex]*consensusInst{}
	}
	addLogIndex := logIndex
	added := false
	for i := 0; i < consensusInstsInAdvance; i++ {
		if _, ok := cni.consensusInsts[committeeID][addLogIndex]; !ok {
			consGrCtx, consGrCancel := context.WithCancel(ctx)
			logIndexCopy := addLogIndex
			cgr := consGR.New(
				consGrCtx, cni.chainID, dkShare, &logIndexCopy, cni.nodeIdentity,
				cni.procCache, cni.mempool, cni.stateMgr, cni.net,
				recoveryTimeout, redeliveryPeriod, printStatusPeriod, cni.log,
			)
			cni.consensusInsts[committeeID][addLogIndex] = &consensusInst{ // TODO: Handle terminations somehow.
				cancelFunc: consGrCancel,
				consensus:  cgr,
			}
			added = true
		}
		addLogIndex = addLogIndex.Next()
	}
	if added {
		cni.cleanupConsensusInsts(&committeeID, &logIndex)
	}
	return cni.consensusInsts[committeeID][logIndex]
}

func (cni *chainNodeImpl) cleanupConsensusInsts(keepCommitteeID *chainMgr.CommitteeID, keepLogIndex *journal.LogIndex) {
	for cmtID := range cni.consensusInsts {
		for li := range cni.consensusInsts[cmtID] {
			if keepCommitteeID != nil && keepLogIndex != nil && cmtID == *keepCommitteeID && li >= *keepLogIndex {
				continue
			}
			ci := cni.consensusInsts[cmtID][li]
			if ci.aliasOutput == nil && ci.cancelFunc != nil {
				// We can cancel an instance, if input was not yet provided.
				ci.cancelFunc() // TODO: Somehow cancel hanging instances, maybe with old LogIndex, etc.
				ci.cancelFunc = nil
			}
		}
	}
}

// Cleanup TX'es that are not needed to be posted anymore.
func (cni *chainNodeImpl) cleanupPublishingTXes(neededPostTXes []chainMgr.NeedPostedTX) {
	for txID, cancelFunc := range cni.publishingTXes {
		found := false
		for _, npt := range neededPostTXes {
			if npt.TxID == txID {
				found = true
				break
			}
		}
		if !found {
			cancelFunc()
			delete(cni.publishingTXes, txID)
		}
	}
}

func pubKeyAsNodeID(pubKey *cryptolib.PublicKey) gpa.NodeID {
	return gpa.NodeID(pubKey.String())
}
