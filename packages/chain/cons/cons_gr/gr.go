// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// The purpose of this package is to run the consensus protocol
// as a goroutine and communicate with all the related components.
package cons_gr

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/chain/cons"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

const (
	msgTypeCons byte = iota
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces required from other components (MP, SM)

type Mempool interface {
	ConsensusProposalAsync(ctx context.Context, aliasOutput *isc.AliasOutputWithID) <-chan []*isc.RequestRef
	ConsensusRequestsAsync(ctx context.Context, requestRefs []*isc.RequestRef) <-chan []isc.Request
}

// State manager has to implement this interface.
type StateMgr interface {
	// State manager has to return a signal via the return channel when it
	// ensures all the needed blocks for the specified AliasOutput is present
	// in the database. Context is used to cancel a request.
	ConsensusStateProposal(
		ctx context.Context,
		aliasOutput *isc.AliasOutputWithID,
	) <-chan interface{}
	// State manager has to ensure all the data needed for the specified alias
	// output (presented as aliasOutputID+stateCommitment) is present in the DB.
	ConsensusDecidedState(
		ctx context.Context,
		aliasOutput *isc.AliasOutputWithID,
	) <-chan state.State
	// State manager has to persistently store the block and respond only after
	// the block was flushed to the disk. A WAL can be used for that as well.
	ConsensusProducedBlock(
		ctx context.Context,
		block state.StateDraft,
	) <-chan state.Block
}

type VM interface {
	ConsensusRunTask(ctx context.Context, task *vm.VMTask) <-chan *vm.VMTaskResult
}

////////////////////////////////////////////////////////////////////////////////
// Implementation.

type Output struct {
	Status cons.OutputStatus // Can only be Completed | Skipped.
	Result *cons.Result      // Result of the consensus.
}

func (o *Output) String() string {
	return fmt.Sprintf("{cons.gr.Output, status=%v, result=%v}", o.Status, o.Result)
}

type input struct {
	baseAliasOutput *isc.AliasOutputWithID
	outputCB        func(*Output)
	recoverCB       func()
}

type ConsGr struct {
	me                          gpa.NodeID
	consInst                    gpa.AckHandler
	inputCh                     chan *input
	inputReceived               *atomic.Bool
	inputTimeCh                 chan time.Time
	outputCB                    func(*Output) // For sending output to the user.
	outputReady                 bool          // Set to true, if we provided output already.
	recoverCB                   func()        // For sending recovery hint to the user.
	recoveryTimeout             time.Duration
	redeliveryPeriod            time.Duration
	printStatusPeriod           time.Duration
	mempool                     Mempool
	mempoolProposalsRespCh      <-chan []*isc.RequestRef
	mempoolProposalsAsked       bool
	mempoolRequestsRespCh       <-chan []isc.Request
	mempoolRequestsAsked        bool
	stateMgr                    StateMgr
	stateMgrStateProposalRespCh <-chan interface{}
	stateMgrStateProposalAsked  bool
	stateMgrDecidedStateRespCh  <-chan state.State
	stateMgrDecidedStateAsked   bool
	stateMgrSaveBlockRespCh     <-chan state.Block
	stateMgrSaveBlockAsked      bool
	vm                          VM
	vmRespCh                    <-chan *vm.VMTaskResult
	vmAsked                     bool
	netRecvPipe                 pipe.Pipe[*peering.PeerMessageIn]
	netPeeringID                peering.PeeringID
	netPeerPubs                 map[gpa.NodeID]*cryptolib.PublicKey
	netDisconnect               context.CancelFunc
	net                         peering.NetworkProvider
	ctx                         context.Context
	pipeMetrics                 *metrics.ChainPipeMetrics
	log                         *logger.Logger
}

func New(
	ctx context.Context,
	chainID isc.ChainID,
	chainStore state.Store,
	dkShare tcrypto.DKShare,
	logIndex *cmt_log.LogIndex,
	myNodeIdentity *cryptolib.KeyPair,
	procCache *processors.Cache,
	mempool Mempool,
	stateMgr StateMgr,
	net peering.NetworkProvider,
	validatorAgentID isc.AgentID,
	recoveryTimeout time.Duration,
	redeliveryPeriod time.Duration,
	printStatusPeriod time.Duration,
	chainMetrics *metrics.ChainConsensusMetrics,
	pipeMetrics *metrics.ChainPipeMetrics,
	log *logger.Logger,
) *ConsGr {
	cmtPubKey := dkShare.GetSharedPublic()
	netPeeringID := peering.HashPeeringIDFromBytes(chainID.Bytes(), cmtPubKey.AsBytes(), logIndex.Bytes()) // ChainID × Committee PubKey × LogIndex
	netPeerPubs := map[gpa.NodeID]*cryptolib.PublicKey{}
	for _, peerPubKey := range dkShare.GetNodePubKeys() {
		netPeerPubs[gpa.NodeIDFromPublicKey(peerPubKey)] = peerPubKey
	}
	me := gpa.NodeIDFromPublicKey(myNodeIdentity.GetPublicKey())
	cgr := &ConsGr{
		me:                me,
		consInst:          nil, // Set bellow.
		inputCh:           make(chan *input, 1),
		inputReceived:     atomic.NewBool(false),
		inputTimeCh:       make(chan time.Time, 1),
		recoveryTimeout:   recoveryTimeout,
		redeliveryPeriod:  redeliveryPeriod,
		printStatusPeriod: printStatusPeriod,
		mempool:           mempool,
		stateMgr:          stateMgr,
		vm:                NewVMAsync(chainMetrics, log),
		netRecvPipe:       pipe.NewInfinitePipe[*peering.PeerMessageIn](),
		netPeeringID:      netPeeringID,
		netPeerPubs:       netPeerPubs,
		netDisconnect:     nil, // Set bellow.
		net:               net,
		ctx:               ctx,
		pipeMetrics:       pipeMetrics,
		log:               log,
	}

	pipeMetrics.TrackPipeLenMax("cons-gr-netRecvPipe", netPeeringID.String(), cgr.netRecvPipe.Len)

	consInstRaw := cons.New(chainID,
		chainStore,
		me,
		myNodeIdentity.GetPrivateKey(),
		dkShare,
		procCache,
		netPeeringID[:],
		gpa.NodeIDFromPublicKey,
		validatorAgentID,
		log,
	).AsGPA()
	cgr.consInst = gpa.NewAckHandler(me, consInstRaw, redeliveryPeriod)

	unhook := net.Attach(&netPeeringID, peering.ReceiverChainCons, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != msgTypeCons {
			cgr.log.Warnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		cgr.netRecvPipe.TryAdd(recv, cgr.log.Debugf)
	})
	cgr.netDisconnect = unhook

	go cgr.run()
	return cgr
}

func (cgr *ConsGr) Input(baseAliasOutput *isc.AliasOutputWithID, outputCB func(*Output), recoverCB func()) {
	wasReceivedBefore := cgr.inputReceived.Swap(true)
	if wasReceivedBefore {
		panic(fmt.Errorf("duplicate input: %v", baseAliasOutput))
	}
	inp := &input{
		baseAliasOutput: baseAliasOutput,
		outputCB:        outputCB,
		recoverCB:       recoverCB,
	}
	cgr.inputCh <- inp
	close(cgr.inputCh)
}

func (cgr *ConsGr) Time(t time.Time) {
	cgr.inputTimeCh <- t
}

func (cgr *ConsGr) run() { //nolint:gocyclo,funlen
	defer util.ExecuteIfNotNil(cgr.netDisconnect)
	defer func() {
		cgr.pipeMetrics.ForgetPipeLenMax("cons-gr-netRecvPipe", cgr.netPeeringID.String())
		cgr.netRecvPipe.Discard()
	}()

	ctxClose := cgr.ctx.Done()
	netRecvPipeOutCh := cgr.netRecvPipe.Out()
	redeliveryTickCh := time.After(cgr.redeliveryPeriod)
	var recoveryTimeoutCh <-chan time.Time
	var printStatusCh <-chan time.Time
	for {
		select {
		case recv, ok := <-netRecvPipeOutCh:
			if !ok {
				netRecvPipeOutCh = nil
				continue
			}
			cgr.handleNetMessage(recv)
		case inp, ok := <-cgr.inputCh:
			if !ok {
				cgr.inputCh = nil
				continue
			}
			// The recoveryTimeout was before initiated here. The problem is that in the case
			// of an idle chain, the recovery timeout can be fired even with no reason for that.
			// So we moved the recovery timeout to the MempoolProposal event.
			// Not sure, if that's safe, because mempool can fail to get the correct state and
			// return the proposal. Maybe we can give it some timeout, but then there would be
			// yet another aspect breaking the asynchrony.
			printStatusCh = time.After(cgr.printStatusPeriod)
			cgr.outputCB = inp.outputCB
			cgr.recoverCB = inp.recoverCB
			cgr.handleConsInput(cons.NewInputProposal(inp.baseAliasOutput))
		case t, ok := <-cgr.inputTimeCh:
			if !ok {
				cgr.inputTimeCh = nil
				continue
			}
			cgr.handleConsInput(cons.NewInputTimeData(t))
		case resp, ok := <-cgr.mempoolProposalsRespCh:
			if !ok {
				cgr.mempoolProposalsRespCh = nil
				continue
			}
			recoveryTimeoutCh = time.After(cgr.recoveryTimeout) // See comment for the InputProposal.
			cgr.handleConsInput(cons.NewInputMempoolProposal(resp))
		case resp, ok := <-cgr.mempoolRequestsRespCh:
			if !ok {
				cgr.mempoolRequestsRespCh = nil
				continue
			}
			cgr.handleConsInput(cons.NewInputMempoolRequests(resp))
		case _, ok := <-cgr.stateMgrStateProposalRespCh:
			if !ok {
				cgr.stateMgrStateProposalRespCh = nil
				continue
			}
			cgr.handleConsInput(cons.NewInputStateMgrProposalConfirmed())
		case resp, ok := <-cgr.stateMgrDecidedStateRespCh:
			if !ok {
				cgr.stateMgrDecidedStateRespCh = nil
				continue
			}
			cgr.handleConsInput(cons.NewInputStateMgrDecidedVirtualState(resp))
		case resp, ok := <-cgr.stateMgrSaveBlockRespCh:
			if !ok {
				cgr.stateMgrSaveBlockRespCh = nil
				continue
			}
			if resp == nil {
				panic(fmt.Errorf("cannot save produced block"))
			}
			cgr.handleConsInput(cons.NewInputStateMgrBlockSaved(resp))
		case resp, ok := <-cgr.vmRespCh:
			if !ok {
				cgr.vmRespCh = nil
				continue
			}
			cgr.handleConsInput(cons.NewInputVMResult(resp))
		case t, ok := <-redeliveryTickCh:
			if !ok {
				redeliveryTickCh = nil
				continue
			}
			redeliveryTickCh = time.After(cgr.redeliveryPeriod)
			cgr.handleRedeliveryTick(t)
		case _, ok := <-recoveryTimeoutCh:
			if !ok {
				recoveryTimeoutCh = nil
				continue
			}
			if cgr.outputReady || cgr.recoverCB == nil {
				continue
			}
			cgr.log.Warn("Recovery timeout reached.")
			cgr.recoverCB()
			cgr.recoverCB = nil
			// Don't terminate, maybe output is still needed. // TODO: Reconsider it.
		case <-printStatusCh:
			printStatusCh = time.After(cgr.printStatusPeriod)
			cgr.log.Debugf("Consensus Instance: %v", cgr.consInst.StatusString())
		case <-ctxClose:
			cgr.log.Debugf("Closing ConsGr because context closed.")
			return
		}
	}
}

func (cgr *ConsGr) handleConsInput(inp gpa.Input) {
	outMsgs := cgr.consInst.Input(inp)
	cgr.sendMessages(outMsgs)
	cgr.tryHandleOutput()
}

func (cgr *ConsGr) handleRedeliveryTick(t time.Time) {
	outMsgs := cgr.consInst.Input(cgr.consInst.MakeTickInput(t))
	cgr.sendMessages(outMsgs)
	cgr.tryHandleOutput()
}

func (cgr *ConsGr) handleNetMessage(recv *peering.PeerMessageIn) {
	msg, err := cgr.consInst.UnmarshalMessage(recv.MsgData)
	if err != nil {
		cgr.log.Warnf("cannot parse message: %v", err)
		return
	}
	msg.SetSender(gpa.NodeIDFromPublicKey(recv.SenderPubKey))
	outMsgs := cgr.consInst.Message(msg)
	cgr.sendMessages(outMsgs)
	cgr.tryHandleOutput()
}

func (cgr *ConsGr) tryHandleOutput() { //nolint:gocyclo
	outputUntyped := cgr.consInst.Output()
	if outputUntyped == nil {
		return
	}
	output := outputUntyped.(*cons.Output)
	if output.NeedMempoolProposal != nil && !cgr.mempoolProposalsAsked {
		cgr.mempoolProposalsRespCh = cgr.mempool.ConsensusProposalAsync(cgr.ctx, output.NeedMempoolProposal)
		cgr.mempoolProposalsAsked = true
	}
	if output.NeedMempoolRequests != nil && !cgr.mempoolRequestsAsked {
		cgr.mempoolRequestsRespCh = cgr.mempool.ConsensusRequestsAsync(cgr.ctx, output.NeedMempoolRequests)
		cgr.mempoolRequestsAsked = true
	}
	if output.NeedStateMgrStateProposal != nil && !cgr.stateMgrStateProposalAsked {
		cgr.stateMgrStateProposalRespCh = cgr.stateMgr.ConsensusStateProposal(cgr.ctx, output.NeedStateMgrStateProposal)
		cgr.stateMgrStateProposalAsked = true
	}
	if output.NeedStateMgrDecidedState != nil && !cgr.stateMgrDecidedStateAsked {
		cgr.stateMgrDecidedStateRespCh = cgr.stateMgr.ConsensusDecidedState(cgr.ctx, output.NeedStateMgrDecidedState)
		cgr.stateMgrDecidedStateAsked = true
	}
	if output.NeedStateMgrSaveBlock != nil && !cgr.stateMgrSaveBlockAsked {
		cgr.stateMgrSaveBlockRespCh = cgr.stateMgr.ConsensusProducedBlock(cgr.ctx, output.NeedStateMgrSaveBlock)
		cgr.stateMgrSaveBlockAsked = true
	}
	if output.NeedVMResult != nil && !cgr.vmAsked {
		cgr.vmRespCh = cgr.vm.ConsensusRunTask(cgr.ctx, output.NeedVMResult)
		cgr.vmAsked = true
	}
	if output.Status != cons.Running && !cgr.outputReady && cgr.outputCB != nil {
		cgr.provideOutput(output)
		cgr.outputReady = true
	}
}

func (cgr *ConsGr) provideOutput(output *cons.Output) {
	switch output.Status {
	case cons.Skipped:
		cgr.outputCB(&Output{Status: output.Status})
	case cons.Completed:
		cgr.outputCB(&Output{Status: output.Status, Result: output.Result})
	default:
		panic(fmt.Errorf("unexpected cons.Output.Status=%v", output.Status))
	}
}

func (cgr *ConsGr) sendMessages(outMsgs gpa.OutMessages) {
	if outMsgs == nil {
		return
	}
	outMsgs.MustIterate(func(msg gpa.Message) {
		pm := peering.NewPeerMessageData(cgr.netPeeringID, peering.ReceiverChainCons, msgTypeCons, msg)
		cgr.net.SendMsgByPubKey(cgr.netPeerPubs[msg.Recipient()], pm)
	})
}
