package cmtlog

import (
	"fmt"
	"maps"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type onLIInc = func(li LogIndex) gpa.OutMessages

type VarConsInsts interface {
	ConsOutputDone(li LogIndex, producedAO *isc.StateAnchor, cb onLIInc) gpa.OutMessages
	ConsOutputSkip(li LogIndex, cb onLIInc) gpa.OutMessages
	ConsTimeout(li LogIndex, cb onLIInc) gpa.OutMessages
	LatestSeenLI(seenLI LogIndex, cb onLIInc) gpa.OutMessages
	LatestL1AO(ao *isc.StateAnchor, cb onLIInc) gpa.OutMessages
	Tick(cb onLIInc) gpa.OutMessages
	StatusString() string
}

// consInsts implements the algorithm modeled in WaspChainCmtLogSUI.tla
type varConsInstsImpl struct {
	haveConsOut bool
	lis         map[LogIndex]*isc.StateAnchor
	minLI       LogIndex         // Do not participate in LI lower than this.
	maxLI       LogIndex         // Cleanup all LIs smaller than this - hist.
	lastLI      LogIndex         // Just to wait for lastAO, if needed but not provided.
	lastAO      *isc.StateAnchor // Last AO seen confirmed in L1.
	hist        uint32           // How many instances to keep running.
	persistCB   func(li LogIndex)
	outputCB    func(lis Output)
	delayed     []LogIndex
	log         log.Logger
}

var _ VarConsInsts = &varConsInstsImpl{}

// NewVarConsInsts is a constructor.
func NewVarConsInsts(
	minLI LogIndex,
	persistCB func(li LogIndex),
	outputCB func(lis Output),
	log log.Logger,
) VarConsInsts {
	vci := &varConsInstsImpl{
		haveConsOut: false,
		lis: map[LogIndex]*isc.StateAnchor{
			minLI: nil,
		},
		minLI:     minLI,
		maxLI:     minLI,
		lastLI:    NilLogIndex(),
		lastAO:    nil,
		hist:      3,
		persistCB: persistCB,
		outputCB:  outputCB,
		delayed:   make([]LogIndex, 3), // Will wait for 3 time ticks before considering SeenLI.
		log:       log,
	}
	vci.outputCB(maps.Clone(vci.lis))
	return vci
}

// Consensus at LI produced a TX.
func (vci *varConsInstsImpl) ConsOutputDone(li LogIndex, producedAO *isc.StateAnchor, cb onLIInc) gpa.OutMessages {
	vci.haveConsOut = true
	return vci.trySet(li.Next(), producedAO, cb)
}

// Consensus at LI terminate with a SKIP/⊥ decision.
func (vci *varConsInstsImpl) ConsOutputSkip(li LogIndex, cb onLIInc) gpa.OutMessages {
	vci.haveConsOut = true
	if vci.lastAO == nil {
		vci.lastLI = li.Next() // Will be set in LatestL1AO.
		return nil
	}
	return vci.trySet(li.Next(), vci.lastAO, cb)
}

// Consensus at LI indicated a timeout.
func (vci *varConsInstsImpl) ConsTimeout(li LogIndex, cb onLIInc) gpa.OutMessages {
	return vci.trySet(li.Next(), nil, cb)
}

// If we see consensus proposals from F+1 nodes at seenLI...
func (vci *varConsInstsImpl) LatestSeenLI(seenLI LogIndex, cb onLIInc) gpa.OutMessages {
	msgs := gpa.NoMessages()
	msgs.AddAll(vci.trySet(seenLI.Prev(), nil, cb))
	if !vci.haveConsOut {
		// Still don't have the initial round succeeded, thus keep proposing the NIL.
		// A race condition is possible between receiving the next LI from the VarLogIndex,
		// and receiving the consensus output. While the actual convergence at runtime
		// happens anyway, we delay reaction to the VarLogIndex output to some delay to make
		// the test-cases more deterministic.
		vci.delayed[0] = MaxLogIndex(vci.delayed[0], seenLI)
	}
	return msgs
}

// Here we get the latest L1 state.
func (vci *varConsInstsImpl) LatestL1AO(ao *isc.StateAnchor, cb onLIInc) gpa.OutMessages {
	vci.lastAO = ao
	return vci.trySet(vci.lastLI, ao, cb) // Finish ConsOutputSkipBase, if pending.
}

func (vci *varConsInstsImpl) Tick(cb onLIInc) gpa.OutMessages {
	n := len(vci.delayed)
	last := vci.delayed[n-1]
	for i := n - 1; i > 0; i-- {
		vci.delayed[i] = vci.delayed[i-1]
	}
	vci.delayed[0] = NilLogIndex()
	if last.IsNil() {
		return nil
	}
	return vci.trySet(last, nil, cb)
}

func (vci *varConsInstsImpl) trySet(li LogIndex, ao *isc.StateAnchor, cb onLIInc) gpa.OutMessages {
	//
	// Is it outdated?
	if li < vci.minLI {
		return nil
	}
	//
	// Is it already proposed?
	if _, ok := vci.lis[li]; ok {
		return nil
	}
	//
	// Propose it.
	vci.lis[li] = ao
	//
	// Track the max.
	msgs := gpa.NoMessages()
	if li > vci.maxLI {
		vci.persistCB(li)
		vci.maxLI = li
		vci.minLI = MaxLogIndex(vci.minLI, vci.maxLI.Sub(vci.hist))
		msgs.AddAll(cb(li))
	}
	//
	// Cleanup old instances.
	for i := range vci.lis {
		if i < vci.minLI {
			vci.log.LogDebugf("Cleaning up LI=%v, minLI=%v, maxLI=%v", i, vci.minLI, vci.maxLI)
			delete(vci.lis, i)
			continue
		}
	}
	//
	// Set all non-last positions to ⊥, if not set yet.
	for li := vci.minLI; li < vci.maxLI; li = li.Next() {
		if _, ok := vci.lis[li]; !ok {
			vci.lis[li] = nil
		}
	}
	//
	// Notify updated state.
	vci.outputCB(maps.Clone(vci.lis))
	return msgs
}

func (vci *varConsInstsImpl) StatusString() string {
	buf := ""
	for li := vci.minLI; li <= vci.maxLI; li = li.Next() {
		ao, ok := vci.lis[li]
		if !ok {
			buf += fmt.Sprintf(" LI#%d=…", li)
		} else if ao == nil {
			buf += fmt.Sprintf(" LI#%d=⊥", li)
		} else {
			buf += fmt.Sprintf(" LI#%d=%s", li, ao.Anchor().String())
		}
	}
	return fmt.Sprintf("{varConsInsts: minLI=%v,%s}", vci.minLI, buf)
}
