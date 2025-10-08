// Package committeelog is responsible for producing a log of chain's block decisions.
// for a particular committee. The main functions:
//
//   - Propose to start a consensus instance at a specific LI.
//   - Propose the input for a consensus instance (Anchor).
//
// The main idea:
//
//   - In the normal operation, don't consider L1 state at all.
//     Use the anchor ref returned by L1 as a result the previously posted TX.
//
//   - On boot a node proposes ⊥ as a base Anchor.
//     It uses ConsStarted Quorum counter to find that?
//
//   - If consensus decides BaseAnchor=⊥, the nodes will propose latest known L1 Anchor,
//     as reported by their clients.
//
//   - A lagging node will join LI-1 instance with ⊥ and then will input its
//     output to the round LI. This assumes the latest round is LI.
package committeelog

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/metrics"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
	"github.com/iotaledger/wasp/v2/packages/util/byzquorum"
)

// CommitteeLog is the public interface for this algorithm.
type CommitteeLog interface {
	AsGPA() gpa.GPA
}

type State struct {
	LogIndex LogIndex
}

// ConsensusStateRegistry is the interface used to store and recover the existing persistent state.
// To be implemented by the registry.
type ConsensusStateRegistry interface {
	Get(chainID isc.ChainID, committeeAddress *cryptolib.Address) (*State, error) // Can return ErrCommitteeLogStateNotFound.
	Set(chainID isc.ChainID, committeeAddress *cryptolib.Address, state *State) error
}

var ErrCommitteeLogStateNotFound = errors.New("errCmtLogStateNotFound")

// Output is a set of log indexes for which we should run the consensus with
// the values indicated here. Nil means ⊥ here. The output might change Anchor to ⊥
// for a particular LI, but not opposite. The updated value should only be used if
// the previous value was not yet proposed to the consensus (unlikely), otherwise
// the update can be ignored.
type Output = map[LogIndex]*isc.StateAnchor

// Protocol implementation.
type committeeLogImpl struct {
	chainID                isc.ChainID            // Chain, for which this log is maintained by this committee.
	committeeAddr          *cryptolib.Address     // Address of the committee running this chain.
	consensusStateRegistry ConsensusStateRegistry // Persistent storage.
	varLogIndex            VarLogIndex            // Calculates the current log index.
	varLocalView           VarLocalView           // Tracks the pending alias outputs.
	varConsInsts           VarConsInsts           // The main algorithm.
	suspended              bool                   // Is this committee currently suspended?
	output                 Output                 // The current output.
	first                  bool                   // A workaround to senf the first nextLI messages.
	asGPA                  gpa.GPA                // This object, just with all the needed wrappers.
	log                    log.Logger
}

var _ gpa.GPA = &committeeLogImpl{}

// New constructs a new node instance for this protocol.
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
	deriveAnchorByQuorum bool,
	pipeliningLimit int,
	cclMetrics *metrics.ChainCommitteeLogMetrics,
	log log.Logger,
) (CommitteeLog, error) {
	committeeAddr := dkShare.GetSharedPublic().AsAddress()
	//
	// Load the last LogIndex we were working on.
	var prevLI LogIndex
	state, err := consensusStateRegistry.Get(chainID, committeeAddr)
	if err != nil {
		if !errors.Is(err, ErrCommitteeLogStateNotFound) {
			return nil, fmt.Errorf("cannot load committeeLogState for %v: %w", committeeAddr, err)
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
	if f > byzquorum.MaxF(n) {
		log.LogPanicf("invalid f=%v for n=%v", f, n)
	}
	//
	// Log important info.
	log.LogInfof("Committee: N=%v, F=%v, address=%v, address=%v", n, f, committeeAddr.String(), committeeAddr.String())
	for i := range nodePKs {
		log.LogInfof("Committee node[%v]=%v", i, nodePKs[i])
	}
	//
	// Create it.
	cl := &committeeLogImpl{
		chainID:                chainID,
		committeeAddr:          committeeAddr,
		consensusStateRegistry: consensusStateRegistry,
		varLogIndex:            nil, // Set bellow.
		varLocalView:           nil, // Set bellow.
		varConsInsts:           nil, // Set bellow.
		suspended:              true,
		output:                 nil,
		first:                  true,
		log:                    log,
	}
	persistLIFunc := func(li LogIndex) {
		if err := consensusStateRegistry.Set(chainID, committeeAddr, &State{LogIndex: li}); err != nil {
			// Nothing to do, if we cannot persist this.
			panic(fmt.Errorf("cannot persist the committeeLog state: %w", err))
		}
	}
	cl.varConsInsts = NewVarConsInsts(prevLI.Next(), persistLIFunc, func(out Output) {
		log.LogDebugf("VarConsInsts: Output received, %v", out)
		cl.output = out
	}, log.NewChildLogger("VCI"))
	cl.varLogIndex = NewVarLogIndex(nodeIDs, n, f, prevLI, func(li LogIndex) gpa.OutMessages {
		log.LogDebugf("VarLogIndex: Output received, %v", li)
		return cl.varConsInsts.LatestSeenLI(li, cl.varLogIndex.ConsensusStarted)
	}, cclMetrics, log.NewChildLogger("VLI"))
	cl.varLocalView = NewVarLocalView(pipeliningLimit, func(ao *isc.StateAnchor) gpa.OutMessages {
		log.LogDebugf("VarLocalView: Output received, %v", ao)
		return cl.varConsInsts.LatestL1Anchor(ao, cl.varLogIndex.ConsensusStarted)
	}, log.NewChildLogger("VLV"))
	cl.asGPA = gpa.NewOwnHandler(me, cl)
	return cl, nil
}

// Implements the CommitteeLog interface.
func (cl *committeeLogImpl) AsGPA() gpa.GPA {
	return cl.asGPA
}

// Implements the gpa.GPA interface.
func (cl *committeeLogImpl) Input(input gpa.Input) gpa.OutMessages {
	switch input.(type) {
	case *inputCanPropose:
		break // Don't log, its periodic.
	default:
		cl.log.LogDebugf("Input %T: %+v", input, input)
	}
	switch input := input.(type) {
	case *inputAnchorConfirmed:
		return cl.handleInputAnchorConfirmed(input)
	case *inputConsensusOutputSkip:
		return cl.handleInputConsensusOutputSkip(input)
	case *inputConsensusOutputConfirmed:
		return cl.handleInputConsensusOutputConfirmed(input)
	case *inputConsensusOutputRejected:
		return cl.handleInputConsensusOutputRejected(input)
	case *inputConsensusTimeout:
		return cl.handleInputConsensusTimeout(input)
	case *inputCanPropose:
		return cl.handleInputCanPropose()
	case *inputSuspend:
		cl.handleInputSuspend()
		return nil
	}
	panic(fmt.Errorf("unexpected input %T: %+v", input, input))
}

// Implements the gpa.GPA interface.
func (cl *committeeLogImpl) Message(msg gpa.Message) gpa.OutMessages {
	msgNLI, ok := msg.(*MsgNextLogIndex)
	if !ok {
		cl.log.LogWarnf("dropping unexpected message %T: %+v", msg, msg)
		return nil
	}
	return cl.handleMsgNextLogIndex(msgNLI)
}

// The latest anchor object's version confirmed at the L1.
func (cl *committeeLogImpl) handleInputAnchorConfirmed(input *inputAnchorConfirmed) gpa.OutMessages {
	cl.suspended = false
	return cl.varLocalView.AnchorConfirmed(input.anchor)
}

// Consensus completed with a decision to SKIP/⊥.
func (cl *committeeLogImpl) handleInputConsensusOutputSkip(input *inputConsensusOutputSkip) gpa.OutMessages {
	return cl.varConsInsts.ConsOutputSkip(input.logIndex, cl.varLogIndex.ConsensusStarted)
}

// Consensus has decided, produced a TX and it is now confirmed by L1.
func (cl *committeeLogImpl) handleInputConsensusOutputConfirmed(input *inputConsensusOutputConfirmed) gpa.OutMessages {
	return cl.varConsInsts.ConsOutputDone(input.logIndex, input.nextAnchor, cl.varLogIndex.ConsensusStarted)
}

// Consensus has decided, produced a TX but it was rejected by L1.
func (cl *committeeLogImpl) handleInputConsensusOutputRejected(input *inputConsensusOutputRejected) gpa.OutMessages {
	return cl.varConsInsts.ConsOutputSkip(input.logIndex, cl.varLogIndex.ConsensusStarted) // This will cause proposal of our latest L1 Anchor.
}

// Consensus tries to decide for too long. Maybe quorum assumption has been violated.
func (cl *committeeLogImpl) handleInputConsensusTimeout(input *inputConsensusTimeout) gpa.OutMessages {
	return cl.varConsInsts.ConsTimeout(input.logIndex, cl.varLogIndex.ConsensusStarted)
}

func (cl *committeeLogImpl) handleInputCanPropose() gpa.OutMessages {
	msgs := gpa.NoMessages()
	msgs.AddAll(cl.varConsInsts.Tick(cl.varLogIndex.ConsensusStarted))

	if cl.first && cl.output != nil && len(cl.output) > 0 {
		// This is a workaround for sending initial NextLI messages on boot.
		cl.first = false
		for li := range cl.output {
			cl.log.LogDebugf("Sending initial NextLI messages for LI=%v", li)
			msgs.AddAll(cl.varLogIndex.ConsensusStarted(li))
		}
		return msgs
	}
	return msgs
}

func (cl *committeeLogImpl) handleInputSuspend() {
	cl.suspended = true
}

// > ON Reception of ⟨NextLI, •⟩ message:
// >   ...
func (cl *committeeLogImpl) handleMsgNextLogIndex(msg *MsgNextLogIndex) gpa.OutMessages {
	return cl.varLogIndex.MsgNextLogIndexReceived(msg)
}

// Implements the gpa.GPA interface.
func (cl *committeeLogImpl) Output() gpa.Output {
	out := cl.output
	if out == nil || cl.suspended {
		return nil // Untyped nil.
	}
	return out
}

// Implements the gpa.GPA interface.
func (cl *committeeLogImpl) StatusString() string {
	return fmt.Sprintf(
		"{committeeLogImpl, %v, %v, %v}",
		cl.varConsInsts.StatusString(),
		cl.varLocalView.StatusString(),
		cl.varLogIndex.StatusString(),
	)
}
