package cmt_log

import (
	"fmt"
	"maps"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type onLIInc = func(li LogIndex) gpa.OutMessages

type VarConsInsts interface {
	ConsOutputDone(li LogIndex, producedAO *isc.StateAnchor, cb onLIInc) gpa.OutMessages
	ConsOutputSkip(li LogIndex, cb onLIInc) gpa.OutMessages
	ConsTimeout(li LogIndex, cb onLIInc) gpa.OutMessages
	LatestSeenLI(seenLI LogIndex, cb onLIInc) gpa.OutMessages
	LatestL1AO(ao *isc.StateAnchor, cb onLIInc) gpa.OutMessages
	StatusString() string
}

// consInsts implements the algorithm modelled in WaspChainCmtLogSUI.tla
type varConsInstsImpl struct {
	lis       map[LogIndex]*isc.StateAnchor
	minLI     LogIndex         // Do not participate in LI lower than this.
	maxLI     LogIndex         // Cleanup all LIs smaller than this - hist.
	lastLI    LogIndex         // Just to wait for lastAO, if needed but not provided.
	lastAO    *isc.StateAnchor // Last AO seen confirmed in L1.
	hist      uint32           // How many instances to keep running.
	persistCB func(li LogIndex)
	outputCB  func(lis Output)
	log       *logger.Logger
}

var _ VarConsInsts = &varConsInstsImpl{}

// Constructor.
func NewVarConsInsts(
	minLI LogIndex,
	persistCB func(li LogIndex),
	outputCB func(lis Output),
	log *logger.Logger,
) VarConsInsts {
	return &varConsInstsImpl{
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
		log:       log,
	}
}

// Consensus at LI produced a TX.
func (vci *varConsInstsImpl) ConsOutputDone(li LogIndex, producedAO *isc.StateAnchor, cb onLIInc) gpa.OutMessages {
	return vci.trySet(li.Next(), producedAO, cb)
}

// Consensus at LI terminate with a SKIP/⊥ decision.
func (vci *varConsInstsImpl) ConsOutputSkip(li LogIndex, cb onLIInc) gpa.OutMessages {
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
	vci.maxLI = MaxLogIndex(vci.maxLI, seenLI)
	return vci.trySet(seenLI.Prev(), nil, cb)
}

// Here we get the latest L1 state.
func (vci *varConsInstsImpl) LatestL1AO(ao *isc.StateAnchor, cb onLIInc) gpa.OutMessages {
	vci.lastAO = ao
	return vci.trySet(vci.lastLI, ao, cb) // Finish ConsOutputSkipBase, if pending.
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
		vci.minLI = MaxLogIndex(vci.minLI, LogIndex(vci.maxLI.AsUint32()-vci.hist))
		msgs.AddAll(cb(li))
	}
	//
	// Cleanup old instances.
	for i := range vci.lis {
		if i < vci.minLI {
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
			buf = buf + fmt.Sprintf(" LI#%d=…", li)
		} else if ao == nil {
			buf = buf + fmt.Sprintf(" LI#%d=⊥", li)
		} else {
			buf = buf + fmt.Sprintf(" LI#%d=%s", li, ao.Anchor().String())
		}
	}
	return fmt.Sprintf("{varConsInsts: minLI=%v,%s}", vci.minLI, buf)
}
