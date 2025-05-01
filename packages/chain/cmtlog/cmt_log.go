// Package cmtlog is responsible for producing a log of chain's block decisions.
// for a particular committee. The main functions:
//
//   - Propose to start a consensus instance at a specific LI.
//   - Propose the input for a consensus instance (AO).
//
// The main idea:
//
//   - In the normal operation, don't consider L1 state at all.
//     Use the anchor ref returned by L1 as a result the previously posted TX.
//
//   - On boot a node proposes ⊥ as a base AO.
//     It uses ConsStarted Quorum counter to find that?
//
//   - If consensus decides BaseAO=⊥, the nodes will propose latest known L1 AO,
//     as reported by their clients.
//
//   - A lagging node will join LI-1 instance with ⊥ and then will input its
//     output to the round LI. This assumes the latest round is LI.
package cmtlog

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util/byzquorum"
)

// CmtLog is the public interface for this algorithm.
type CmtLog interface {
	AsGPA() gpa.GPA
}

type State struct {
	LogIndex LogIndex
}

// ConsensusStateRegistry is the interface used to store and recover the existing persistent state.
// To be implemented by the registry.
type ConsensusStateRegistry interface {
	Get(chainID isc.ChainID, committeeAddress *cryptolib.Address) (*State, error) // Can return ErrCmtLogStateNotFound.
	Set(chainID isc.ChainID, committeeAddress *cryptolib.Address, state *State) error
}

var ErrCmtLogStateNotFound = errors.New("errCmtLogStateNotFound")

// Output is a set of log indexes for which we should run the consensus with
// the values indicated here. Nil means ⊥ here. The output might change AO to ⊥
// for a particular LI, but not opposite. The updated value should only be used if
// the previous value was not yet proposed to the consensus (unlikely), otherwise
// the update can be ignored.
type Output = map[LogIndex]*isc.StateAnchor

// Protocol implementation.
type cmtLogImpl struct {
	chainID                isc.ChainID            // Chain, for which this log is maintained by this committee.
	cmtAddr                *cryptolib.Address     // Address of the committee running this chain.
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

var _ gpa.GPA = &cmtLogImpl{}

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
	deriveAOByQuorum bool,
	pipeliningLimit int,
	cclMetrics *metrics.ChainCmtLogMetrics,
	log log.Logger,
) (CmtLog, error) {
	cmtAddr := dkShare.GetSharedPublic().AsAddress()
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
	if f > byzquorum.MaxF(n) {
		log.LogPanicf("invalid f=%v for n=%v", f, n)
	}
	//
	// Log important info.
	log.LogInfof("Committee: N=%v, F=%v, address=%v, address=%v", n, f, cmtAddr.String(), cmtAddr.String())
	for i := range nodePKs {
		log.LogInfof("Committee node[%v]=%v", i, nodePKs[i])
	}
	//
	// Create it.
	cl := &cmtLogImpl{
		chainID:                chainID,
		cmtAddr:                cmtAddr,
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
		if err := consensusStateRegistry.Set(chainID, cmtAddr, &State{LogIndex: li}); err != nil {
			// Nothing to do, if we cannot persist this.
			panic(fmt.Errorf("cannot persist the cmtLog state: %w", err))
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
		return cl.varConsInsts.LatestL1AO(ao, cl.varLogIndex.ConsensusStarted)
	}, log.NewChildLogger("VLV"))
	cl.asGPA = gpa.NewOwnHandler(me, cl)
	return cl, nil
}

// Implements the CmtLog interface.
func (cl *cmtLogImpl) AsGPA() gpa.GPA {
	return cl.asGPA
}

// Implements the gpa.GPA interface.
func (cl *cmtLogImpl) Input(input gpa.Input) gpa.OutMessages {
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
func (cl *cmtLogImpl) Message(msg gpa.Message) gpa.OutMessages {
	msgNLI, ok := msg.(*MsgNextLogIndex)
	if !ok {
		cl.log.LogWarnf("dropping unexpected message %T: %+v", msg, msg)
		return nil
	}
	return cl.handleMsgNextLogIndex(msgNLI)
}

// The latest anchor object's version confirmed at the L1.
func (cl *cmtLogImpl) handleInputAnchorConfirmed(input *inputAnchorConfirmed) gpa.OutMessages {
	cl.suspended = false
	return cl.varLocalView.AnchorObjectConfirmed(input.anchor)
}

// Consensus completed with a decision to SKIP/⊥.
func (cl *cmtLogImpl) handleInputConsensusOutputSkip(input *inputConsensusOutputSkip) gpa.OutMessages {
	return cl.varConsInsts.ConsOutputSkip(input.logIndex, cl.varLogIndex.ConsensusStarted)
}

// Consensus has decided, produced a TX and it is now confirmed by L1.
func (cl *cmtLogImpl) handleInputConsensusOutputConfirmed(input *inputConsensusOutputConfirmed) gpa.OutMessages {
	return cl.varConsInsts.ConsOutputDone(input.logIndex, input.nextAnchorObject, cl.varLogIndex.ConsensusStarted)
}

// Consensus has decided, produced a TX but it was rejected by L1.
func (cl *cmtLogImpl) handleInputConsensusOutputRejected(input *inputConsensusOutputRejected) gpa.OutMessages {
	return cl.varConsInsts.ConsOutputSkip(input.logIndex, cl.varLogIndex.ConsensusStarted) // This will cause proposal of our latest L1 AO.
}

// Consensus tries to decide for too long. Maybe quorum assumption has been violated.
func (cl *cmtLogImpl) handleInputConsensusTimeout(input *inputConsensusTimeout) gpa.OutMessages {
	return cl.varConsInsts.ConsTimeout(input.logIndex, cl.varLogIndex.ConsensusStarted)
}

func (cl *cmtLogImpl) handleInputCanPropose() gpa.OutMessages {
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

func (cl *cmtLogImpl) handleInputSuspend() {
	cl.suspended = true
}

// > ON Reception of ⟨NextLI, •⟩ message:
// >   ...
func (cl *cmtLogImpl) handleMsgNextLogIndex(msg *MsgNextLogIndex) gpa.OutMessages {
	return cl.varLogIndex.MsgNextLogIndexReceived(msg)
}

// Implements the gpa.GPA interface.
func (cl *cmtLogImpl) Output() gpa.Output {
	out := cl.output
	if out == nil || cl.suspended {
		return nil // Untyped nil.
	}
	return out
}

// Implements the gpa.GPA interface.
func (cl *cmtLogImpl) StatusString() string {
	return fmt.Sprintf(
		"{cmtLogImpl, %v, %v, %v}",
		cl.varConsInsts.StatusString(),
		cl.varLocalView.StatusString(),
		cl.varLogIndex.StatusString(),
	)
}
