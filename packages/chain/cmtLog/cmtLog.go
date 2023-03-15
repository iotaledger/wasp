// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package cmtLog is responsible for producing a log of chain's block decisions
// for a particular committee. The main tasks for this module are:
//   - Track the head of the chain log for a committee.
//   - Track which blocks are approved, pending or reverted.
//   - Handle startup and recovery scenarios.
//
// Main principles of the algorithm:
// >   - Propose to go to the next log index when
// >       - a) consensus for the latest known LogIndex has terminated (confirmed | rejected | skip | recover).
// >            All the previous can be left uncompleted.
// >       - b) confirmed alias output is received from L1, that we haven't posted in this node.
// >       - and we have a tip AO (from the local view).
// >   - If there is a single clean chain, we can use pipelining (consider consDone event
// >     instead of waiting for confirmed | rejected).
// >   - We start a consensus instance whenever VLI reports a quorum and we have not started it yet.
// >     Even if we provided no proposal for that LI.
//
// The algorithm at a high level:
//
// > ON Startup:
// >     Let prevLI <- TRY restoring the last started LogIndex ELSE 0
// >     MinLI <- prevLI + 1
// >     LogIndex.Start(prevLI)
// > UPON AliasOutput (AO) {Confirmed | Rejected} by L1:
// >     LocalView.Update(AO)
// >     IF LocalView changed THEN
// >         LogIndex.L1ReplacedBaseAliasOutput()
// >         TryProposeConsensus()
//
// > ON Startup:
// >     Let prevLI <- TRY restoring the last started LogIndex ELSE 0
// >     MinLI <- prevLI + 1
// >     LogIndex.Start(prevLI)
// >     TryProposeConsensus()
// > UPON AliasOutput (AO) {Confirmed | Rejected} by L1:
// >     LocalView.Update(AO)
// >     IF LocalView changed THEN
// >         LogIndex.L1ReplacedBaseAliasOutput()
// >         TryProposeConsensus()
// > ON ConsensusOutput/DONE (CD)
// >     LocalView.Update(CD)
// >     IF LocalView changed THEN
// >         LogIndex.ConsensusOutput(CD.LogIndex)
// >         TryProposeConsensus()
// > ON ConsensusOutput/SKIP (CS)
// >     LogIndex.ConsensusOutput(CS.LogIndex)
// >     TryProposeConsensus()
// > ON ConsensusTimeout (CT)
// >     LogIndex.ConsensusTimeout(CT.LogIndex)
// >     TryProposeConsensus()
// > ON Suspend:
// >     Suspended <- TRUE
// >     TryProposeConsensus()
// > ON Reception of ⟨NextLI, •⟩ message:
// >     LogIndex.Receive(⟨NextLI, •⟩ message).
// >     TryProposeConsensus()
// > PROCEDURE TryProposeConsensus:
// >     IF ∧ LocalView.BaseAO ≠ NIL
// >        ∧ LogIndex > ConsensusLI
// >        ∧ LogIndex ≥ MinLI // ⇒ LogIndex ≠ NIL
// >        ∧ ¬ Suspended
// >     THEN
// >         Persist LogIndex
// >         ConsensusLI <- LogIndex
// >         Propose LocalView.BaseAO for LogIndex
// >     ELSE
// >         Don't propose any consensus.
//
// See `WaspChainRecovery.tla` for more precise specification.
//
// Notes and invariants:
//   - Here only a single consensus instance will be considered needed for this node at a time.
//     Other instances may continue running, but their results will be ignored. That's
//     because a consensus takes an input from the previous consensus output (the base
//     alias ID and other parts that depend on it).
//   - A consensus is started when we have new log index greater than that we have
//     crashed with, and there is an alias output received.
//
// ## Summary.
//
// Inputs expected:
//   - Consensus: Start -> Done | Timeout.
//   - AliasOutput: Confirmed | Rejected -> {}.
//   - Suspend.
//
// Messages exchanged:
//   - NextLogIndex (private, between cmtLog instances).
package cmtLog

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cons"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

// Public interface for this algorithm.
type CmtLog interface {
	AsGPA() gpa.GPA
}

type State struct {
	LogIndex LogIndex
}

// Interface used to store and recover the existing persistent state.
// To be implemented by the registry.
type ConsensusStateRegistry interface {
	Get(chainID isc.ChainID, committeeAddress iotago.Address) (*State, error) // Can return ErrCmtLogStateNotFound.
	Set(chainID isc.ChainID, committeeAddress iotago.Address, state *State) error
}

var ErrCmtLogStateNotFound = errors.New("errCmtLogStateNotFound")

// Output for this protocol indicates, what instance of a consensus
// is currently required to be run. The unique identifier here is the
// logIndex (there will be no different baseAliasOutputs for the same logIndex).
type Output struct {
	logIndex        LogIndex
	baseAliasOutput *isc.AliasOutputWithID
}

func makeOutput(logIndex LogIndex, baseAliasOutput *isc.AliasOutputWithID) *Output {
	return &Output{logIndex: logIndex, baseAliasOutput: baseAliasOutput}
}

func (o *Output) GetLogIndex() LogIndex {
	return o.logIndex
}

func (o *Output) GetBaseAliasOutput() *isc.AliasOutputWithID {
	return o.baseAliasOutput
}

func (o *Output) String() string {
	return fmt.Sprintf("{Output, logIndex=%v, baseAliasOutput=%v}", o.logIndex, o.baseAliasOutput)
}

// Protocol implementation.
type cmtLogImpl struct {
	chainID                isc.ChainID            // Chain, for which this log is maintained by this committee.
	cmtAddr                iotago.Address         // Address of the committee running this chain.
	consensusStateRegistry ConsensusStateRegistry // Persistent storage.
	suspended              bool                   // Is this committee suspended?
	minLI                  LogIndex               // Lowest log index this instance is allowed to participate.
	consensusLI            LogIndex               // Latest LogIndex for which consensus was been started.
	varLogIndex            VarLogIndex            // Calculates the current log index.
	varLocalView           VarLocalView           // Tracks the pending alias outputs.
	varRunning             VarRunning             // Tracks the latest LI.
	output                 *Output                // The current request for a consensus.
	asGPA                  gpa.GPA                // This object, just with all the needed wrappers.
	log                    *logger.Logger
}

var _ gpa.GPA = &cmtLogImpl{}

// Construct new node instance for this protocol.
//
// > ON Startup:
// >     Let prevLI <- TRY restoring the last started LogIndex ELSE 0
// >     MinLI <- prevLI + 1
// >     ...
func New(
	me gpa.NodeID,
	chainID isc.ChainID,
	dkShare tcrypto.DKShare,
	consensusStateRegistry ConsensusStateRegistry,
	nodeIDFromPubKey func(pubKey *cryptolib.PublicKey) gpa.NodeID,
	log *logger.Logger,
) (CmtLog, error) {
	cmtAddr := dkShare.GetSharedPublic().AsEd25519Address()
	//
	// Load the last LogIndex we were working on.
	var prevLI LogIndex
	state, err := consensusStateRegistry.Get(chainID, cmtAddr)
	if err != nil {
		if !errors.Is(err, ErrCmtLogStateNotFound) {
			return nil, fmt.Errorf("cannot load cmtLogState for %v: %w", cmtAddr, err)
		}
		prevLI = NilLogIndex()
	} else {
		// Don't participate in the last stored LI, because maybe we have already sent some messages.
		prevLI = state.LogIndex
	}
	//
	// Make node IDs.
	nodePKs := dkShare.GetNodePubKeys()
	nodeIDs := make([]gpa.NodeID, len(nodePKs))
	for i := range nodeIDs {
		nodeIDs[i] = nodeIDFromPubKey(nodePKs[i])
	}
	//
	// Construct the object.
	n := len(nodeIDs)
	f := dkShare.DSS().MaxFaulty()
	if f > (n-1)/3 {
		log.Panicf("invalid f=%v for n=%v", n, f)
	}
	//
	// Log important info.
	log.Infof("Committee: N=%v, F=%v, address=%v, betch32=%v", n, f, cmtAddr.String(), cmtAddr.Bech32(parameters.L1().Protocol.Bech32HRP))
	for i := range nodePKs {
		log.Infof("Committee node[%v]=%v", i, nodePKs[i])
	}
	//
	// Create it.
	minLogIndex := prevLI.Next()
	cl := &cmtLogImpl{
		chainID:                chainID,
		cmtAddr:                cmtAddr,
		consensusStateRegistry: consensusStateRegistry,
		suspended:              false,
		minLI:                  minLogIndex,
		consensusLI:            NilLogIndex(),
		varLogIndex:            NewVarLogIndex(nodeIDs, n, f, prevLI, func(li LogIndex, ao *isc.AliasOutputWithID) {}, log.Named("VLI")),
		varLocalView:           NewVarLocalView(log.Named("VLV")),
		varRunning:             NewVarRunning(log.Named("VR")),
		log:                    log,
	}
	cl.asGPA = gpa.NewOwnHandler(me, cl)
	return cl, nil
}

// Implements the CmtLog interface.
func (cl *cmtLogImpl) AsGPA() gpa.GPA {
	return cl.asGPA
}

// Implements the gpa.GPA interface.
func (cl *cmtLogImpl) Input(input gpa.Input) gpa.OutMessages {
	cl.log.Debugf("Input %T: %+v", input, input)
	switch input := input.(type) {
	case *inputAliasOutputConfirmed:
		return cl.handleInputAliasOutputConfirmed(input)
	case *inputConsensusOutputDone:
		return cl.handleInputConsensusOutputDone(input)
	case *inputConsensusOutputSkip:
		return cl.handleInputConsensusOutputSkip(input)
	case *inputConsensusOutputConfirmed:
		return cl.handleInputConsensusOutputConfirmed(input)
	case *inputConsensusOutputRejected:
		return cl.handleInputConsensusOutputRejected(input)
	case *inputConsensusTimeout:
		return cl.handleInputConsensusTimeout(input)
	case *inputSuspend:
		return cl.handleInputSuspend()
	}
	panic(fmt.Errorf("unexpected input %T: %+v", input, input))
}

// Implements the gpa.GPA interface.
func (cl *cmtLogImpl) Message(msg gpa.Message) gpa.OutMessages {
	msgNLI, ok := msg.(*msgNextLogIndex)
	if !ok {
		cl.log.Warnf("dropping unexpected message %T: %+v", msg, msg)
		return nil
	}
	return cl.handleMsgNextLogIndex(msgNLI)
}

// > UPON AliasOutput (AO) {Confirmed | Rejected} by L1:
// >   ...
func (cl *cmtLogImpl) handleInputAliasOutputConfirmed(input *inputAliasOutputConfirmed) gpa.OutMessages {
	if tipAO, tipUpdated := cl.varLocalView.AliasOutputConfirmed(input.aliasOutput); tipUpdated {
		if cl.suspended {
			cl.log.Infof("Committee resumed, tip replaced by L1 to %v", tipAO)
			cl.suspended = false
		}
		return cl.varLogIndex.L1ReplacedBaseAliasOutput(tipAO)
	}
	return nil
}

// >   ...
func (cl *cmtLogImpl) handleInputConsensusOutputConfirmed(input *inputConsensusOutputConfirmed) gpa.OutMessages {
	cl.varRunning.ConsensusOutputConfirmed(input.logIndex) // Not needed, in general.
	return nil
}

// >   ...
func (cl *cmtLogImpl) handleInputConsensusOutputRejected(input *inputConsensusOutputRejected) gpa.OutMessages {
	cl.varRunning.ConsensusOutputRejected(input.logIndex)
	if tipAO, tipUpdated := cl.varLocalView.AliasOutputRejected(input.aliasOutput); tipUpdated || cl.varRunning.IsLatest(input.logIndex) {
		return cl.varLogIndex.L1ReplacedBaseAliasOutput(tipAO)
	}
	return nil
}

// > ON ConsensusOutput/DONE (CD)
// >   ...
func (cl *cmtLogImpl) handleInputConsensusOutputDone(input *inputConsensusOutputDone) gpa.OutMessages {
	cl.varRunning.ConsensusOutputDone(input.logIndex)
	if tipAO, tipUpdated := cl.varLocalView.ConsensusOutputDone(input.logIndex, input.baseAliasOutputID, input.nextAliasOutput); tipUpdated {
		return cl.varLogIndex.ConsensusOutputReceived(input.logIndex, cons.Completed, tipAO)
	}
	return nil
}

// > ON ConsensusOutput/SKIP (CS)
// >   ...
func (cl *cmtLogImpl) handleInputConsensusOutputSkip(input *inputConsensusOutputSkip) gpa.OutMessages {
	latestCompleted := cl.varRunning.ConsensusOutputSkip(input.logIndex)
	tipAO := cl.varLocalView.Value()
	if latestCompleted && tipAO != nil {
		return cl.varLogIndex.ConsensusOutputReceived(input.logIndex, cons.Skipped, tipAO)
	}
	return nil
}

// > ON ConsensusTimeout (CT)
// >   ...
func (cl *cmtLogImpl) handleInputConsensusTimeout(input *inputConsensusTimeout) gpa.OutMessages {
	latestCompleted := cl.varRunning.ConsensusRecover(input.logIndex)
	tipAO := cl.varLocalView.Value()
	if latestCompleted && tipAO != nil {
		return cl.varLogIndex.ConsensusRecoverReceived(input.logIndex)
	}
	return nil
}

// > ON Suspend:
// >   ...
func (cl *cmtLogImpl) handleInputSuspend() gpa.OutMessages {
	cl.log.Infof("Committee suspended.")
	cl.suspended = true
	cl.output = nil
	return cl.tryProposeConsensus(nil)
}

// > ON Reception of ⟨NextLI, •⟩ message:
// >   ...
func (cl *cmtLogImpl) handleMsgNextLogIndex(msg *msgNextLogIndex) gpa.OutMessages {
	msgs := cl.varLogIndex.MsgNextLogIndexReceived(msg)
	return cl.tryProposeConsensus(msgs)
}

// Implements the gpa.GPA interface.
func (cl *cmtLogImpl) Output() gpa.Output {
	if cl.output == nil {
		return nil // Untyped nil!
	}
	return cl.output
}

// Implements the gpa.GPA interface.
func (cl *cmtLogImpl) StatusString() string {
	vliLI, _ := cl.varLogIndex.Value()
	return fmt.Sprintf("{cmtLogImpl, LogIndex=%v, output=%+v, %v, %v}", vliLI, cl.output, cl.varLocalView.StatusString(), cl.varLogIndex.StatusString())
}

// > PROCEDURE TryProposeConsensus:
// >     IF ∧ LocalView.BaseAO ≠ NIL
// >        ∧ LogIndex > ConsensusLI
// >        ∧ LogIndex ≥ MinLI // ⇒ LogIndex ≠ NIL
// >        ∧ ¬ Suspended
// >     THEN
// >         Persist LogIndex
// >         ConsensusLI <- LogIndex
// >         Propose LocalView.BaseAO for LogIndex
// >     ELSE
// >         Don't propose any consensus.
func (cl *cmtLogImpl) tryProposeConsensus(msgs gpa.OutMessages) gpa.OutMessages {
	logIndex, baseAO := cl.varLogIndex.Value()
	if logIndex == NilLogIndex() {
		// No log index decided yet.
		return msgs
	}
	//
	// Check, maybe it is already started.
	if cl.output != nil && cl.output.logIndex == logIndex {
		// Already started, keep it as is.
		return msgs
	}
	//
	// >     IF ∧ LocalView.BaseAO ≠ NIL
	// >        ∧ LogIndex > ConsensusLI
	// >        ∧ LogIndex ≥ MinLI // ⇒ LogIndex ≠ NIL
	// >        ∧ ¬ Suspended
	// TODO: previously was: baseAO := cl.varLocalView.GetBaseAliasOutput()
	if baseAO != nil && logIndex > cl.consensusLI && logIndex >= cl.minLI && !cl.suspended {
		// >     THEN
		// >         Persist LogIndex
		// >         ConsensusLI <- LogIndex
		// >         Propose LocalView.BaseAO for LogIndex
		//
		// Persist the log index to ensure we will not participate in the
		// same consensus after the restart.
		if err := cl.consensusStateRegistry.Set(cl.chainID, cl.cmtAddr, &State{LogIndex: logIndex}); err != nil {
			// Nothing to do, if we cannot persist this.
			panic(fmt.Errorf("cannot persist the cmtLog state: %w", err))
		}
		//
		// Start the consensus (ask the upper layer to start it).
		cl.consensusLI = logIndex
		cl.output = makeOutput(cl.consensusLI, baseAO)
		cl.varRunning.ConsensusProposed(logIndex)
	} else {
		// >     ELSE
		// >         Don't propose any consensus.
		cl.output = nil // Outdated, clear it away.
	}
	return msgs
}
