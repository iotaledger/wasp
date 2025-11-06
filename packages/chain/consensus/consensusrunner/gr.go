// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package consensusrunner runs the consensus protocol for IOTA Smart Contracts.
// as a goroutine and communicate with all the related components.
package consensusrunner

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
	"github.com/iotaledger/wasp/v2/packages/chain/consensus"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/util/pipe"
	"github.com/iotaledger/wasp/v2/packages/vm"
	"github.com/iotaledger/wasp/v2/packages/vm/processors"
)

const (
	msgTypeCons byte = iota
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces required from other components (MP, SM)

type ConsensusID [iotago.AddressLen + 4]byte

func NewConsensusID(committeeAddr *cryptolib.Address, logIndex *committeelog.LogIndex) ConsensusID {
	ret := ConsensusID{}
	copy(ret[:], committeeAddr.Bytes())
	copy(ret[iotago.AddressLen:], codec.Encode[uint32](logIndex.AsUint32()))
	return ret
}

type Mempool interface {
	ConsensusProposalAsync(ctx context.Context, anchor *isc.StateAnchor, consensusID ConsensusID) <-chan []*isc.RequestRef
	ConsensusRequestsAsync(ctx context.Context, requestRefs []*isc.RequestRef) <-chan []isc.Request
}

// StateMgr is an interface that state managers have to implement.
type StateMgr interface {
	// State manager has to return a signal via the return channel when it
	// ensures all the needed blocks for the specified Anchor is present
	// in the database. Context is used to cancel a request.
	ConsensusStateProposal(
		ctx context.Context,
		anchor *isc.StateAnchor,
	) <-chan interface{}
	// State manager has to ensure all the data needed for the specified alias
	// output (presented as anchorID+stateCommitment) is present in the DB.
	ConsensusDecidedState(
		ctx context.Context,
		anchor *isc.StateAnchor,
	) <-chan state.State
	// State manager has to persistently store the block and respond only after
	// the block was flushed to the disk. A WAL can be used for that as well.
	ConsensusProducedBlock(
		ctx context.Context,
		block state.StateDraft,
	) <-chan state.Block
}

type NodeConnL1Info interface {
	GetGasCoins() []*coin.CoinWithRef
	GetL1Params() *parameters.L1Params
}

type NodeConn interface {
	ConsensusL1InfoProposal(
		ctx context.Context,
		anchor *isc.StateAnchor,
	) <-chan NodeConnL1Info
}

type VM interface {
	ConsensusRunTask(ctx context.Context, task *vm.VMTask) <-chan *vm.VMTaskResult
}

////////////////////////////////////////////////////////////////////////////////
// Implementation.

type Output struct {
	Status consensus.OutputStatus // Can only be Completed | Skipped.
	Result *consensus.Result      // Result of the consensus.
}

func (o *Output) String() string {
	return fmt.Sprintf("{cons.gr.Output, status=%v, result=%v}", o.Status, o.Result)
}

type input struct {
	baseAnchor *isc.StateAnchor
	outputCB   func(*Output)
	recoverCB  func()
}

type ConsensusRunner struct {
	me                          gpa.NodeID
	consInst                    gpa.AckHandler
	inputCh                     chan *input
	inputReceived               *atomic.Bool
	inputRotateToCh             chan *iotago.Address
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
	nodeConn                    NodeConn
	nodeConnL1InfoRespCh        <-chan NodeConnL1Info
	nodeConnL1InfoAsked         bool
	vm                          VM
	vmRespCh                    <-chan *vm.VMTaskResult
	vmAsked                     bool
	netRecvPipe                 pipe.Pipe[*peering.PeerMessageIn]
	netPeeringID                peering.PeeringID
	netPeerPubs                 map[gpa.NodeID]*cryptolib.PublicKey
	netDisconnect               context.CancelFunc
	net                         peering.NetworkProvider
	consensusID                 ConsensusID
	ctx                         context.Context
	pipeMetrics                 *metrics.ChainPipeMetrics
	log                         log.Logger
}

func New(
	ctx context.Context,
	chainID isc.ChainID,
	chainStore state.Store,
	dkShare tcrypto.DKShare,
	logIndex *committeelog.LogIndex,
	myNodeIdentity *cryptolib.KeyPair,
	procCache *processors.Config,
	mempool Mempool,
	stateMgr StateMgr,
	nodeConn NodeConn,
	net peering.NetworkProvider,
	rotateTo *iotago.Address,
	validatorAgentID isc.AgentID,
	recoveryTimeout time.Duration,
	redeliveryPeriod time.Duration,
	printStatusPeriod time.Duration,
	chainMetrics *metrics.ChainConsensusMetrics,
	pipeMetrics *metrics.ChainPipeMetrics,
	log log.Logger,
) *ConsensusRunner {
	committeePubKey := dkShare.GetSharedPublic()
	netPeeringID := peering.HashPeeringIDFromBytes(chainID.Bytes(), committeePubKey.AsBytes(), logIndex.Bytes()) // ChainID × Committee PubKey × LogIndex
	netPeerPubs := map[gpa.NodeID]*cryptolib.PublicKey{}
	for _, peerPubKey := range dkShare.GetNodePubKeys() {
		netPeerPubs[gpa.NodeIDFromPublicKey(peerPubKey)] = peerPubKey
	}
	me := gpa.NodeIDFromPublicKey(myNodeIdentity.GetPublicKey())
	runner := &ConsensusRunner{
		me:                me,
		consInst:          nil, // Set bellow.
		inputCh:           make(chan *input, 1),
		inputReceived:     atomic.NewBool(false),
		inputRotateToCh:   make(chan *iotago.Address, 1),
		inputTimeCh:       make(chan time.Time, 1),
		recoveryTimeout:   recoveryTimeout,
		redeliveryPeriod:  redeliveryPeriod,
		printStatusPeriod: printStatusPeriod,
		mempool:           mempool,
		stateMgr:          stateMgr,
		nodeConn:          nodeConn,
		vm:                NewVMAsync(chainMetrics, log),
		netRecvPipe:       pipe.NewInfinitePipe[*peering.PeerMessageIn](),
		netPeeringID:      netPeeringID,
		netPeerPubs:       netPeerPubs,
		netDisconnect:     nil, // Set bellow.
		net:               net,
		consensusID:       NewConsensusID(committeePubKey.AsAddress(), logIndex),
		ctx:               ctx,
		pipeMetrics:       pipeMetrics,
		log:               log,
	}

	pipeMetrics.TrackPipeLenMax("cons-gr-netRecvPipe", netPeeringID.String(), runner.netRecvPipe.Len)

	consInstRaw := consensus.New(chainID,
		chainStore,
		me,
		myNodeIdentity.GetPrivateKey(),
		dkShare,
		rotateTo,
		procCache,
		netPeeringID[:],
		gpa.NodeIDFromPublicKey,
		validatorAgentID,
		log,
	).AsGPA()
	runner.consInst = gpa.NewAckHandler(me, consInstRaw, redeliveryPeriod)

	unhook := net.Attach(&netPeeringID, peering.ReceiverChainCons, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != msgTypeCons {
			runner.log.LogWarnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		runner.netRecvPipe.TryAdd(recv, runner.log.LogDebugf)
	})
	runner.netDisconnect = unhook

	go runner.run()
	return runner
}

func (r *ConsensusRunner) Input(baseAnchor *isc.StateAnchor, outputCB func(*Output), recoverCB func()) {
	wasReceivedBefore := r.inputReceived.Swap(true)
	if wasReceivedBefore {
		panic(fmt.Errorf("duplicate input: %v", baseAnchor))
	}
	inp := &input{
		baseAnchor: baseAnchor,
		outputCB:   outputCB,
		recoverCB:  recoverCB,
	}
	r.inputCh <- inp
	close(r.inputCh)
}

func (r *ConsensusRunner) RotateTo(address *iotago.Address) {
	r.inputRotateToCh <- address
}

func (r *ConsensusRunner) Time(t time.Time) {
	r.inputTimeCh <- t
}

func (r *ConsensusRunner) run() { //nolint:gocyclo,funlen
	defer util.ExecuteIfNotNil(r.netDisconnect)
	defer func() {
		r.pipeMetrics.ForgetPipeLenMax("cons-gr-netRecvPipe", r.netPeeringID.String())
		r.netRecvPipe.Discard()
	}()

	ctxClose := r.ctx.Done()
	netRecvPipeOutCh := r.netRecvPipe.Out()
	redeliveryTickCh := time.After(r.redeliveryPeriod)
	var recoveryTimeoutCh <-chan time.Time
	var printStatusCh <-chan time.Time
	for {
		select {
		case recv, ok := <-netRecvPipeOutCh:
			if !ok {
				netRecvPipeOutCh = nil
				continue
			}
			r.handleNetMessage(recv)
		case inp, ok := <-r.inputCh:
			if !ok {
				r.inputCh = nil
				continue
			}
			// The recoveryTimeout was before initiated here. The problem is that in the case
			// of an idle chain, the recovery timeout can be fired even with no reason for that.
			// So we moved the recovery timeout to the MempoolProposal event.
			// Not sure, if that's safe, because mempool can fail to get the correct state and
			// return the proposal. Maybe we can give it some timeout, but then there would be
			// yet another aspect breaking the asynchrony.
			printStatusCh = time.After(r.printStatusPeriod)
			r.outputCB = inp.outputCB
			r.recoverCB = inp.recoverCB
			r.handleConsInput(consensus.NewInputProposal(inp.baseAnchor))

		case a, ok := <-r.inputRotateToCh:
			if !ok {
				r.inputRotateToCh = nil
				continue
			}
			r.handleConsInput(consensus.NewInputRotateTo(a))

		case t, ok := <-r.inputTimeCh:
			if !ok {
				r.inputTimeCh = nil
				continue
			}
			r.handleConsInput(consensus.NewInputTimeData(t))

		case resp, ok := <-r.mempoolProposalsRespCh:
			if !ok {
				r.mempoolProposalsRespCh = nil
				continue
			}
			recoveryTimeoutCh = time.After(r.recoveryTimeout) // See comment for the InputProposal.
			r.handleConsInput(consensus.NewInputMempoolProposal(resp))
		case resp, ok := <-r.mempoolRequestsRespCh:
			if !ok {
				r.mempoolRequestsRespCh = nil
				continue
			}
			r.handleConsInput(consensus.NewInputMempoolRequests(resp))
		case _, ok := <-r.stateMgrStateProposalRespCh:
			if !ok {
				r.stateMgrStateProposalRespCh = nil
				continue
			}
			r.handleConsInput(consensus.NewInputStateMgrProposalConfirmed())
		case resp, ok := <-r.stateMgrDecidedStateRespCh:
			if !ok {
				r.stateMgrDecidedStateRespCh = nil
				continue
			}
			r.handleConsInput(consensus.NewInputStateMgrDecidedVirtualState(resp))
		case resp, ok := <-r.stateMgrSaveBlockRespCh:
			if !ok {
				r.stateMgrSaveBlockRespCh = nil
				continue
			}
			if resp == nil {
				panic(fmt.Errorf("cannot save produced block"))
			}
			r.handleConsInput(consensus.NewInputStateMgrBlockSaved(resp))

		case t, ok := <-r.nodeConnL1InfoRespCh:
			r.log.LogDebugf("ConsensusL1InfoProposal received, respCh=%v, response=%v", r.nodeConnL1InfoRespCh, t)
			if !ok {
				r.nodeConnL1InfoRespCh = nil
				continue
			}
			r.handleConsInput(consensus.NewInputL1Info(t.GetGasCoins(), t.GetL1Params()))

		case resp, ok := <-r.vmRespCh:
			if !ok {
				r.vmRespCh = nil
				continue
			}
			r.handleConsInput(consensus.NewInputVMResult(resp))
		case t, ok := <-redeliveryTickCh:
			if !ok {
				redeliveryTickCh = nil
				continue
			}
			redeliveryTickCh = time.After(r.redeliveryPeriod)
			r.handleRedeliveryTick(t)
		case _, ok := <-recoveryTimeoutCh:
			if !ok {
				recoveryTimeoutCh = nil
				continue
			}
			if r.outputReady || r.recoverCB == nil {
				continue
			}
			r.log.LogWarn("Recovery timeout reached.")
			r.recoverCB()
			r.recoverCB = nil
			// Don't terminate, maybe output is still needed. // TODO: Reconsider it.
		case <-printStatusCh:
			printStatusCh = time.After(r.printStatusPeriod)
			r.log.LogDebugf("Consensus Instance: %v", r.consInst.StatusString())
		case <-ctxClose:
			r.log.LogDebugf("Closing ConsGr because context closed.")
			return
		}
	}
}

func (r *ConsensusRunner) handleConsInput(inp gpa.Input) {
	outMsgs := r.consInst.Input(inp)
	r.sendMessages(outMsgs)
	r.tryHandleOutput()
}

func (r *ConsensusRunner) handleRedeliveryTick(t time.Time) {
	outMsgs := r.consInst.Input(r.consInst.MakeTickInput(t))
	r.sendMessages(outMsgs)
	r.tryHandleOutput()
}

func (r *ConsensusRunner) handleNetMessage(recv *peering.PeerMessageIn) {
	msg, err := r.consInst.UnmarshalPayload(recv.MsgData)
	if err != nil {
		r.log.LogWarnf("cannot parse message: %v", err)
		return
	}
	outMsgs := r.consInst.Message(gpa.NewMessageIn(gpa.NodeIDFromPublicKey(recv.SenderPubKey), msg))
	r.sendMessages(outMsgs)
	r.tryHandleOutput()
}

func (r *ConsensusRunner) tryHandleOutput() {
	outputUntyped := r.consInst.Output()
	if outputUntyped == nil {
		return
	}
	output := outputUntyped.(*consensus.Output)
	if output.NeedMempoolProposal != nil && !r.mempoolProposalsAsked {
		r.mempoolProposalsRespCh = r.mempool.ConsensusProposalAsync(r.ctx, output.NeedMempoolProposal, r.consensusID)
		r.mempoolProposalsAsked = true
	}
	if output.NeedMempoolRequests != nil && !r.mempoolRequestsAsked {
		r.mempoolRequestsRespCh = r.mempool.ConsensusRequestsAsync(r.ctx, output.NeedMempoolRequests)
		r.mempoolRequestsAsked = true
	}
	if output.NeedStateMgrStateProposal != nil && !r.stateMgrStateProposalAsked {
		r.stateMgrStateProposalRespCh = r.stateMgr.ConsensusStateProposal(r.ctx, output.NeedStateMgrStateProposal)
		r.stateMgrStateProposalAsked = true
	}
	if output.NeedStateMgrDecidedState != nil && !r.stateMgrDecidedStateAsked {
		r.stateMgrDecidedStateRespCh = r.stateMgr.ConsensusDecidedState(r.ctx, output.NeedStateMgrDecidedState)
		r.stateMgrDecidedStateAsked = true
	}
	if output.NeedStateMgrSaveBlock != nil && !r.stateMgrSaveBlockAsked {
		r.stateMgrSaveBlockRespCh = r.stateMgr.ConsensusProducedBlock(r.ctx, output.NeedStateMgrSaveBlock)
		r.stateMgrSaveBlockAsked = true
	}
	if output.NeedNodeConnL1Info != nil && !r.nodeConnL1InfoAsked {
		r.nodeConnL1InfoRespCh = r.nodeConn.ConsensusL1InfoProposal(r.ctx, output.NeedNodeConnL1Info)
		r.log.LogDebugf("ConsensusL1InfoProposal asked, respCh=%v", r.nodeConnL1InfoRespCh)
		r.nodeConnL1InfoAsked = true
	}
	if output.NeedVMResult != nil && !r.vmAsked {
		r.vmRespCh = r.vm.ConsensusRunTask(r.ctx, output.NeedVMResult)
		r.vmAsked = true
	}
	if output.Status != consensus.Running && !r.outputReady && r.outputCB != nil {
		r.provideOutput(output)
		r.outputReady = true
	}
}

func (r *ConsensusRunner) provideOutput(output *consensus.Output) {
	switch output.Status {
	case consensus.Skipped:
		r.outputCB(&Output{Status: output.Status})
	case consensus.Completed:
		r.outputCB(&Output{Status: output.Status, Result: output.Result})
	default:
		panic(fmt.Errorf("unexpected cons.Output.Status=%v", output.Status))
	}
}

func (r *ConsensusRunner) sendMessages(outMsgs []gpa.MessageOut) {
	if outMsgs == nil {
		return
	}
	for _, msg := range outMsgs {
		msgBytes := lo.Must(gpa.MarshalPayload(msg.Payload))
		pm := peering.NewPeerMessageData(r.netPeeringID, peering.ReceiverChainCons, msgTypeCons, msgBytes)
		r.net.SendMsgByPubKey(r.netPeerPubs[msg.Recipient], pm)
	}
}
