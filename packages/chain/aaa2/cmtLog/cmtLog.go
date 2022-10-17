// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package cmtLog is responsible for producing a log (a journal) of a chain for a particular committee.
// The main tasks for this module are:
//   - Initiate and manage the consensus instances.
//   - Handle startup and recovery scenarios.
//
// The algorithm at a high level:
//
// TODO: Something is missing here. Maybe LogIndex + WantToLogIndex has to be used to differentiate the current and wanted?
//
// > ON Startup:
// >     OutLI <- TRY restoring the last started LogIndex ELSE 0
// >     MinLI <- OurLI + 1
// >     send NextLI with the MinLI as a proposed log index.
// > UPON Reception of N-F NextLI with LI > OurLI:
// >	 OurLI <- LI
// >     IF not suspended AND have BaseAO THEN
// >         Start consensus for LI with the same BaseAO.
// > UPON Reception of ConsensusOutput:
// >     update LocalView
// >     store the output.			// TODO: What component does that? The upper layer probably.
// >     publish to node conn. 		// TODO: What component does that? The upper layer probably.
// >     Increase OurLI
// >     IF not suspended THEN
// >         Start next Consensus round // TODO: if not started.
// > UPON Reception of ConsensusTimeout:
// >     send NextLI with the next log index.
// > UPON Reception of Suspend:
// >     mark instance as suspended.
// > UPON Reception of AOConfirmed:
// >     update the LocalView accordingly.
// >     IF LocalView was reset THEN
// >         mark instance as not suspended
// >         IF consensus is active THEN
// >             Increase OutLI.
// >         Start next consensus round.
// > UPON Reception of AORejected:
// >     update the LocalView accordingly.
// >     IF localView.BaseAO = NIL THEN
// >         IF consensus running THEN
// >             OutLI <- OurLI + 1
// >             cancel consensus.
// >     ELSE IF localView.baseAO != currentConsensus.baseAO THEN
// >         OutLI <- OurLI + 1 	// consensus is running, implied by the above condition.
// >         IF not suspended THEN
// >             Start next Consensus round
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
// Messages exchanged:
//   - Consensus: Start -> Done | Timeout.
//   - AliasOutput: Confirmed | Rejected -> {}.
//   - NextLogIndex (private, between cmtLog instances).
//   - Suspend.
//   - TimerTick // TODO: Do we need it?
package cmtLog

import (
	"errors"
	"fmt"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

// Public interface for this algorithm.
type CmtLog interface {
	AsGPA() gpa.GPA
}

type State struct {
	LogIndex journal.LogIndex
}

// Interface used to store and recover the existing persistent state.
// To be implemented by the registry.
type Store interface {
	LoadCmtLogState(committeeAddress iotago.Address) (*State, error) // Can return ErrCmtLogStateNotFound.
	SaveCmtLogState(committeeAddress iotago.Address, state *State) error
}

var ErrCmtLogStateNotFound = errors.New("errErrCmtLogStateNotFound")

// Output for this protocol indicates, which
type Output struct {
	logIndex          journal.LogIndex
	baseAliasOutputID iotago.OutputID
}

func makeOutput(logIndex journal.LogIndex, baseAliasOutputID iotago.OutputID) *Output {
	return &Output{logIndex: logIndex, baseAliasOutputID: baseAliasOutputID}
}

func (o *Output) GetLogIndex() journal.LogIndex {
	return o.logIndex
}

func (o *Output) GetBaseAliasOutputID() iotago.OutputID {
	return o.baseAliasOutputID
}

// Protocol implementation.
type cmtLogImpl struct {
	chainID     isc.ChainID                     // Chain, for which this log is maintained by this committee.
	cmtAddr     iotago.Address                  // Address of the committee running this chain.
	dkShare     tcrypto.DKShare                 // Committee that runs the chain ans maintains the log. // TODO: Should be not used here.
	nodeIDs     []gpa.NodeID                    // All the peers in this committee.
	me          gpa.NodeID                      // ID of this node.
	n           int                             // TODO: Get it somehow.
	f           int                             // TODO: Get it somehow.
	store       Store                           // Persistent storage.
	minLogIndex journal.LogIndex                // Lowest log index this instance is allowed to participate.
	logIndex    journal.LogIndex                // Latest log index we are working on. // TODO: remove and use LocalView instead?
	localView   journal.LocalView               // TODO: Consider...
	sentNextLI  journal.LogIndex                // LogIndex for which the MsgNextLogIndex was sent.
	consensusLI journal.LogIndex                // Latest LogIndex for which consensus was been started.
	output      *Output                         // The current request for a consensus.
	suspended   bool                            // Is this committee suspended?
	maxPeerLIs  map[gpa.NodeID]journal.LogIndex // Latest peer indexes received from peers.
	asGPA       gpa.GPA                         // This object, just with all the needed wrappers.
	log         *logger.Logger
}

var _ gpa.GPA = &cmtLogImpl{}

// Construct new node instance for this protocol.
//
// > ON Startup:
// >     OutLI <- TRY restoring the last started LogIndex ELSE 0
// >     MinLI <- OurLI + 1
// >     ...
func New(
	me gpa.NodeID,
	chainID isc.ChainID,
	dkShare tcrypto.DKShare,
	store Store,
	nodeIDFromPubKey func(pubKey *cryptolib.PublicKey) gpa.NodeID,
	log *logger.Logger,
) (CmtLog, error) {
	cmtAddr := dkShare.GetSharedPublic().AsEd25519Address()
	//
	// Load the last LogIndex we were working on.
	var logIndex journal.LogIndex
	state, err := store.LoadCmtLogState(cmtAddr)
	if err == nil {
		// Don't participate in the last stored LI, because maybe we have already sent some messages.
		logIndex = state.LogIndex
	}
	if errors.Is(err, ErrCmtLogStateNotFound) {
		logIndex = journal.NilLogIndex()
	} else if err != nil {
		return nil, xerrors.Errorf("cannot load cmtLogState for %v: %w", cmtAddr, err)
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
	f := (n - 1) / 3
	cl := &cmtLogImpl{
		chainID:     chainID,
		cmtAddr:     cmtAddr,
		dkShare:     dkShare,
		nodeIDs:     nodeIDs,
		me:          me,
		n:           n,
		f:           f,
		store:       store,
		minLogIndex: logIndex.Next(),
		logIndex:    logIndex,
		localView:   journal.NewLocalView(),
		sentNextLI:  journal.NilLogIndex(),
		consensusLI: journal.NilLogIndex(),
		suspended:   false,
		maxPeerLIs:  map[gpa.NodeID]journal.LogIndex{},
		log:         log,
	}
	cl.asGPA = gpa.NewOwnHandler(me, cl)
	return cl, nil
}

// Implements the CmtLog interface.
func (cl *cmtLogImpl) AsGPA() gpa.GPA {
	return cl.asGPA
}

// Implements the gpa.GPA interface.
//
// > ON Startup:
// >     ...
// >     send NextLI with the MinLI as a proposed log index.
func (cl *cmtLogImpl) Input(input gpa.Input) gpa.OutMessages {
	cl.log.Debugf("Input: %v", input)
	if input != nil {
		panic(xerrors.Errorf("cmtLogImpl: expecting nil input, got: %v", input))
	}
	return cl.maybeSendNextLogIndex(cl.minLogIndex)
}

// Implements the gpa.GPA interface.
func (cl *cmtLogImpl) Message(msg gpa.Message) gpa.OutMessages {
	switch msgT := msg.(type) {
	case *msgAliasOutputConfirmed:
		return cl.handleMsgAliasOutputConfirmed(msgT)
	case *msgAliasOutputRejected:
		return cl.handleMsgAliasOutputRejected(msgT)
	case *msgConsensusOutput:
		return cl.handleMsgConsensusOutput(msgT)
	case *msgConsensusTimeout:
		return cl.handleMsgConsensusTimeout(msgT)
	case *msgSuspend:
		return cl.handleMsgSuspend()
	case *msgNextLogIndex:
		return cl.handleMsgNextLogIndex(msgT)
	}
	cl.log.Warnf("dropping unexpected message: %+v", msg)
	return nil
}

// > UPON Reception of AOConfirmed:
// >     update the LocalView accordingly.
// >     IF LocalView was reset THEN
// >         mark instance as not suspended
// >         Increase OutLI.          // TODO: Do we need this???
// >         Start next consensus round.
func (cl *cmtLogImpl) handleMsgAliasOutputConfirmed(msg *msgAliasOutputConfirmed) gpa.OutMessages {
	cl.log.Debugf("handleMsgAliasOutputConfirmed, msg=%+v", msg)
	cl.localView.AliasOutputConfirmed(msg.aliasOutput)
	cl.maybeStartConsensus() // TODO: Not always.
	return nil
}

func (cl *cmtLogImpl) handleMsgAliasOutputRejected(msg *msgAliasOutputRejected) gpa.OutMessages {
	cl.log.Debugf("handleMsgAliasOutputRejected, msg=%+v", msg)
	cl.localView.AliasOutputRejected(msg.aliasOutput)
	cl.maybeStartConsensus() // TODO: Not always.
	return nil
}

// > UPON Reception of ConsensusOutput:
// >     update LocalView
// >     store the output.			// TODO: What component does that? The upper layer probably.
// >     publish to node conn. 		// TODO: What component does that? The upper layer probably.
// >     Increase OurLI
// >     IF not suspended THEN Start next Consensus round.
func (cl *cmtLogImpl) handleMsgConsensusOutput(msg *msgConsensusOutput) gpa.OutMessages {
	if msg.logIndex < cl.consensusLI {
		cl.log.Warnf("Dropping outdated consensus response: %+v", msg)
		return nil
	}
	//
	// If we receive a consensus from the future LogIndex, maybe other nodes have already
	// agreed on it and we have to catch up. Doing this by adjusting our current LogIndex.
	if msg.logIndex > cl.consensusLI {
		cl.log.Warnf("received consensus for log index we haven't asked yet: %+v", msg)
		cl.consensusLI = msg.logIndex
		if cl.logIndex < msg.logIndex {
			cl.logIndex = msg.logIndex
		}
	}
	//
	// Now start a consensus for the next entry.
	// This should succeed, unless the instance is suspended
	// or we don't have baseOA (e.g. there was a reject and we are re-syncing).
	cl.localView.AliasOutputPublished(msg.baseAliasOutputID, msg.nextAliasOutput)
	cl.logIndex = cl.logIndex.Next()
	cl.maybeStartConsensus()
	return nil
}

// > UPON Reception of ConsensusTimeout:
// >     send NextLI with the next log index.
//
// NOTE: Consensus has not finished in long time, try recover by voting
// for the next LogIndex. This actually breaks the asynchronous assumption.
func (cl *cmtLogImpl) handleMsgConsensusTimeout(msg *msgConsensusTimeout) gpa.OutMessages {
	if msg.logIndex < cl.consensusLI {
		cl.log.Warnf("Dropping outdated consensus timeout: %+v", msg)
		return nil
	}
	if msg.logIndex > cl.consensusLI {
		cl.log.Warnf("received consensus timeout for log index we haven't asked yet: %+v", msg)
		cl.consensusLI = msg.logIndex
		if cl.logIndex < msg.logIndex {
			cl.logIndex = msg.logIndex
		}
	}
	return cl.maybeSendNextLogIndex(cl.logIndex.Next())
}

// > UPON Reception of Suspend:
// >     mark instance as suspended.
//
// NOTE: A suspended cmtLog will not start the next consensus round
// after the current consensus will complete. Suspend will be sent
// by the chain, if an alias output with different state controller
// is received.
func (cl *cmtLogImpl) handleMsgSuspend() gpa.OutMessages {
	cl.log.Debugf("handleMsgSuspend")
	cl.suspended = true
	return nil
}

// > UPON Reception of N-F NextLI with LI > OurLI (and having AO):
// >	 OurLI <- LI + 1
// >     IF not suspended THEN Start consensus for LI+1 with the same AliasOutput.
func (cl *cmtLogImpl) handleMsgNextLogIndex(msg *msgNextLogIndex) gpa.OutMessages {
	cl.log.Debugf("handleMsgNextLogIndex, msg=%+v", msg)
	msgs := gpa.NoMessages()
	sender := msg.Sender()
	//
	// Validate and record the vote.
	if !cl.knownNodeID(sender) {
		cl.log.Warnf("MsgNextLogIndex from unknown sender: %+v", msg)
		return nil
	}
	var prevPeerLogIndex journal.LogIndex
	var found bool
	if prevPeerLogIndex, found = cl.maxPeerLIs[sender]; !found {
		prevPeerLogIndex = journal.NilLogIndex()
	}
	if prevPeerLogIndex.AsUint32() >= msg.nextLogIndex.AsUint32() {
		return nil
	}
	cl.maxPeerLIs[sender] = msg.nextLogIndex
	//
	// Support log indexes, if there are F+1 votes for that log index.
	supportLogIndex := cl.votedFor(cl.f + 1)
	if supportLogIndex > cl.logIndex {
		msgs.AddAll(cl.maybeSendNextLogIndex(supportLogIndex))
	}
	//
	// Proceed to the next log index, if needed.
	newLogIndex := cl.votedFor(cl.n - cl.f)
	if newLogIndex > cl.logIndex {
		cl.logIndex = newLogIndex
		cl.maybeStartConsensus()
	}
	return msgs
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
	return fmt.Sprintf("{cmtLogImpl, LogIndex=%v, output=%+v}", cl.logIndex, cl.output)
}

func (cl *cmtLogImpl) knownNodeID(nodeID gpa.NodeID) bool {
	for i := range cl.nodeIDs {
		if cl.nodeIDs[i] == nodeID {
			return true
		}
	}
	return false
}

func (cl *cmtLogImpl) maybeSendNextLogIndex(logIndex journal.LogIndex) gpa.OutMessages {
	if logIndex < cl.logIndex {
		return nil
	}
	if cl.sentNextLI.AsUint32() >= logIndex.AsUint32() {
		return nil
	}
	cl.sentNextLI = logIndex
	msgs := gpa.NoMessages()
	for i := range cl.nodeIDs {
		msgs.Add(newMsgNextLogIndex(cl.nodeIDs[i], logIndex))
	}
	return msgs
}

func (cl *cmtLogImpl) maybeStartConsensus() {
	//
	// Check, maybe it is already started.
	if cl.output != nil && cl.output.logIndex == cl.logIndex {
		// Already started, keep it as is.
		return
	}
	cl.output = nil // Outdated, clear it away.
	//
	// No current consensus, check, maybe we can start the next one.
	if cl.logIndex < cl.minLogIndex {
		return // Don't participate in the LI < minimal.
	}
	if cl.logIndex.IsNil() {
		return // The current LogIndex is not defined yet.
	}
	if cl.consensusLI.AsUint32() >= cl.logIndex.AsUint32() {
		return // We have already started consensus for latter LogIndex.
	}
	if cl.suspended {
		return
	}
	baseOut := cl.localView.GetBaseAliasOutputID()
	if baseOut == nil {
		return // Have no AO to use.
	}
	//
	// Persist the log index to ensure we will not participate in the
	// same consensus after the restart.
	if err := cl.store.SaveCmtLogState(cl.cmtAddr, &State{cl.logIndex}); err != nil {
		// Nothing to do, if we cannot persist this.
		panic(xerrors.Errorf("cannot persist the cmtLog state: %w", err))
	}
	//
	// Start the consensus (ask the upper layer to start it).
	cl.consensusLI = cl.logIndex
	cl.output = makeOutput(cl.consensusLI, *baseOut)
}

// Find highest LogIndex for which N-F nodes have voted.
// Returns 0, if not found.
func (cl *cmtLogImpl) votedFor(quorum int) journal.LogIndex {
	counts := map[journal.LogIndex]int{}
	for _, li := range cl.maxPeerLIs {
		counts[li]++
	}
	max := journal.NilLogIndex()
	for li, c := range counts {
		if c >= quorum && li.AsUint32() > max.AsUint32() {
			max = li
		}
	}
	return max
}
