// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// The purpose of this package is to run the consensus protocol
// as a goroutine and communicate with all the related components.
package consGR

// TODO: Disconnect.

import (
	"context"
	"time"

	"go.uber.org/atomic"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
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
	ConsensusProposalsAsync(ctx context.Context, aliasOutput *isc.AliasOutputWithID) <-chan []*isc.RequestRef
	ConsensusRequestsAsync(ctx context.Context, requestRefs []*isc.RequestRef) <-chan []isc.Request
}

type StateMgrDecidedState struct {
	AliasOutput        *isc.AliasOutputWithID
	StateBaseline      coreutil.StateBaseline
	VirtualStateAccess state.VirtualStateAccess
}

type StateMgr interface {
	// State manager has to implement this function. It has to return a signal via
	// the return channel when it ensures all the needed blocks for the specified
	// AliasOutput is present in the database. Context is used to cancel a request.
	ConsensusStateProposal(
		ctx context.Context,
		aliasOutput *isc.AliasOutputWithID,
	) <-chan interface{}
	// State manager has to ensure all the data needed for the specified alias
	// output (presented as aliasOutputID+stateCommitment) is present in the DB.
	ConsensusDecidedState(
		ctx context.Context,
		aliasOutputID *iotago.OutputID,
		stateCommitment *state.L1Commitment,
	) <-chan *StateMgrDecidedState
}

type VM interface {
	ConsensusRunTask(ctx context.Context, task *vm.VMTask) <-chan *vm.VMTask
}

////////////////////////////////////////////////////////////////////////////////
// Implementation.

type Output struct {
	State cons.OutputState // Can only be Completed | Skipped.
	TX    *iotago.Transaction
	Block state.Block
}

type input struct {
	baseAliasOutput *isc.AliasOutputWithID
	outputCh        chan<- *Output
	recoverCh       chan<- time.Time
}

type ConsGr struct {
	me                          gpa.NodeID
	consInst                    gpa.GPA
	inputCh                     chan *input
	inputReceived               *atomic.Bool
	outputCh                    chan<- *Output   // For sending output to the user.
	outputReady                 bool             // Set to true, if we provided output already.
	recoverCh                   chan<- time.Time // For sending recovery hint to the user.
	recoveryTimeout             time.Duration
	mempool                     Mempool
	mempoolProposalsRespCh      <-chan []*isc.RequestRef
	mempoolProposalsAsked       bool
	mempoolRequestsRespCh       <-chan []isc.Request
	mempoolRequestsAsked        bool
	stateMgr                    StateMgr
	stateMgrStateProposalRespCh <-chan interface{}
	stateMgrStateProposalAsked  bool
	stateMgrDecidedStateRespCh  <-chan *StateMgrDecidedState
	stateMgrDecidedStateAsked   bool
	vm                          VM
	vmRespCh                    <-chan *vm.VMTask
	vmAsked                     bool
	netRecvPipe                 pipe.Pipe
	netPeeringID                peering.PeeringID
	netPeerPubs                 map[gpa.NodeID]*cryptolib.PublicKey
	netAttach                   interface{}
	net                         peering.NetworkProvider
	ctx                         context.Context
	log                         *logger.Logger
}

func New(
	ctx context.Context,
	chainID *isc.ChainID,
	journalID *journal.ID,
	logIndex *journal.LogIndex,
	myNodeIdentity *cryptolib.KeyPair,
	myDKShare tcrypto.DKShare,
	procCache *processors.Cache,
	mempool Mempool,
	stateMgr StateMgr,
	net peering.NetworkProvider,
	recoveryTimeout time.Duration,
	log *logger.Logger,
) *ConsGr {
	consInstID := hashing.HashDataBlake2b(chainID.Bytes(), []byte(logIndex.AsStringKey(*journalID)))
	netPeeringID := peering.PeeringIDFromBytes(consInstID.Bytes())
	netPeerPubs := map[gpa.NodeID]*cryptolib.PublicKey{}
	for _, peerPubKey := range myDKShare.GetNodePubKeys() {
		netPeerPubs[pubKeyAsNodeID(peerPubKey)] = peerPubKey
	}
	me := pubKeyAsNodeID(myNodeIdentity.GetPublicKey())
	cgr := &ConsGr{
		me:              me,
		consInst:        nil, // Set bellow.
		inputReceived:   atomic.NewBool(false),
		recoveryTimeout: recoveryTimeout,
		mempool:         mempool,
		stateMgr:        stateMgr,
		vm:              NewVMAsync(),
		netRecvPipe:     pipe.NewDefaultInfinitePipe(),
		netPeeringID:    netPeeringID,
		netPeerPubs:     netPeerPubs,
		netAttach:       nil, // Set bellow.
		net:             net,
		ctx:             ctx,
		log:             log,
	}
	cgr.consInst = cons.New(*chainID, me, myNodeIdentity.GetPrivateKey(), myDKShare, procCache, consInstID.Bytes(), pubKeyAsNodeID, log).AsGPA()
	netRecvPipeInCh := cgr.netRecvPipe.In()
	cgr.netAttach = net.Attach(&netPeeringID, peering.PeerMessageReceiverChainCons, func(recv *peering.PeerMessageIn) {
		if recv.MsgType != msgTypeCons {
			cgr.log.Warnf("Unexpected message, type=%v", recv.MsgType)
			return
		}
		netRecvPipeInCh <- recv
	})

	go cgr.run()
	return cgr
}

func (cgr *ConsGr) Input(baseAliasOutput *isc.AliasOutputWithID) (<-chan *Output, <-chan time.Time) {
	wasReceivedBefore := cgr.inputReceived.Swap(true)
	if wasReceivedBefore {
		panic(xerrors.Errorf("duplicate input: %v", baseAliasOutput))
	}
	outputCh := make(chan *Output, 1)
	recoverCh := make(chan time.Time)
	inp := &input{
		baseAliasOutput: baseAliasOutput,
		outputCh:        outputCh,
		recoverCh:       recoverCh,
	}
	cgr.inputCh <- inp
	close(cgr.inputCh)
	return outputCh, recoverCh
}

func (cgr *ConsGr) run() { //nolint:gocyclo
	ctxClose := cgr.ctx.Done()
	netRecvPipeOutCh := cgr.netRecvPipe.Out()
	var recoveryTimeoutCh <-chan time.Time
	for {
		select {
		case recv, ok := <-netRecvPipeOutCh:
			if !ok {
				netRecvPipeOutCh = nil
				continue
			}
			cgr.handleNetMessage(recv.(*peering.PeerMessageIn))
		case inp, ok := <-cgr.inputCh:
			if !ok {
				cgr.inputCh = nil
				continue
			}
			recoveryTimeoutCh = time.After(cgr.recoveryTimeout)
			cgr.outputCh = inp.outputCh
			cgr.recoverCh = inp.recoverCh
			cgr.handleInput(inp.baseAliasOutput)
		case resp, ok := <-cgr.mempoolProposalsRespCh:
			if !ok {
				cgr.mempoolProposalsRespCh = nil
				continue
			}
			cgr.handleMessage(cons.NewMsgMempoolProposal(cgr.me, resp))
		case resp, ok := <-cgr.mempoolRequestsRespCh:
			if !ok {
				cgr.mempoolRequestsRespCh = nil
				continue
			}
			cgr.handleMessage(cons.NewMsgMempoolRequests(cgr.me, resp))
		case _, ok := <-cgr.stateMgrStateProposalRespCh:
			if !ok {
				cgr.stateMgrStateProposalRespCh = nil
				continue
			}
			cgr.handleMessage(cons.NewMsgStateMgrProposalConfirmed(cgr.me))
		case resp, ok := <-cgr.stateMgrDecidedStateRespCh:
			if !ok {
				cgr.stateMgrDecidedStateRespCh = nil
				continue
			}
			cgr.handleMessage(cons.NewMsgStateMgrDecidedVirtualState(cgr.me, resp.AliasOutput, resp.StateBaseline, resp.VirtualStateAccess))
		case resp, ok := <-cgr.vmRespCh:
			if !ok {
				cgr.vmRespCh = nil
				continue
			}
			cgr.handleMessage(cons.NewMsgVMResult(cgr.me, resp))
		case t, ok := <-recoveryTimeoutCh:
			if !ok || cgr.recoverCh == nil {
				recoveryTimeoutCh = nil
				continue
			}
			cgr.recoverCh <- t
			close(cgr.recoverCh)
			cgr.recoverCh = nil
			cgr.log.Warnf("Recovery timeout reached.")
			// Don't terminate, maybe output is still needed. // TODO: Reconsider it.
		case <-ctxClose:
			cgr.log.Debugf("Closing ConsGr because context closed.")
			return
		}
	}
}

func (cgr *ConsGr) handleInput(inp gpa.Input) {
	outMsgs := cgr.consInst.Input(inp)
	cgr.sendMessages(outMsgs)
	cgr.tryHandleOutput()
}

func (cgr *ConsGr) handleMessage(msg gpa.Message) {
	outMsgs := cgr.consInst.Message(msg)
	cgr.sendMessages(outMsgs)
	cgr.tryHandleOutput()
}

func (cgr *ConsGr) handleNetMessage(recv *peering.PeerMessageIn) {
	msg, err := cgr.consInst.UnmarshalMessage(recv.MsgData)
	if err != nil {
		cgr.log.Warnf("cannot parse message: %v", err)
	}
	msg.SetSender(pubKeyAsNodeID(recv.SenderPubKey))
	cgr.handleMessage(msg)
}

func (cgr *ConsGr) tryHandleOutput() {
	output := cgr.consInst.Output().(*cons.Output)
	if output.NeedMempoolProposal != nil && !cgr.mempoolProposalsAsked {
		cgr.mempoolProposalsRespCh = cgr.mempool.ConsensusProposalsAsync(cgr.ctx, output.NeedMempoolProposal)
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
		cgr.stateMgrDecidedStateRespCh = cgr.stateMgr.ConsensusDecidedState(cgr.ctx, output.NeedStateMgrDecidedState.AliasOutputID, output.NeedStateMgrDecidedState.StateCommitment)
		cgr.stateMgrDecidedStateAsked = true
	}
	if output.NeedVMResult != nil && !cgr.vmAsked {
		cgr.vmRespCh = cgr.vm.ConsensusRunTask(cgr.ctx, output.NeedVMResult)
		cgr.vmAsked = true
	}
	if output.State != cons.Running && !cgr.outputReady {
		cgr.provideOutput(output)
		cgr.outputReady = true
	}
}

func (cgr *ConsGr) provideOutput(output *cons.Output) {
	switch output.State {
	case cons.Skipped:
		cgr.outputCh <- &Output{State: output.State}
	case cons.Completed:
		block, err := output.ResultState.ExtractBlock()
		if err != nil {
			panic(xerrors.Errorf("cannot extract block from virtual state: %w", err))
		}
		cgr.outputCh <- &Output{State: output.State, TX: output.ResultTransaction, Block: block}
	default:
		panic(xerrors.Errorf("unexpected cons.Output.State=%v", output.State))
	}
}

func (cgr *ConsGr) sendMessages(outMsgs gpa.OutMessages) {
	outMsgs.MustIterate(func(m gpa.Message) {
		msgData, err := m.MarshalBinary()
		if err != nil {
			cgr.log.Warnf("Failed to send a message: %v", err)
			return
		}
		pm := &peering.PeerMessageData{
			PeeringID:   cgr.netPeeringID,
			MsgReceiver: peering.PeerMessageReceiverChainCons,
			MsgType:     msgTypeCons,
			MsgData:     msgData,
		}
		cgr.net.SendMsgByPubKey(cgr.netPeerPubs[m.Recipient()], pm)
	})
}

func pubKeyAsNodeID(pubKey *cryptolib.PublicKey) gpa.NodeID {
	return gpa.NodeID(pubKey.String())
}
