package committeelog

import (
	"fmt"
	"maps"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type onLIInc = func(li LogIndex) gpa.OutMessages

// VarConsInsts implements the algorithm modeled in WaspChainCommitteeLogSUI.tla
type VarConsInsts struct {
	haveConsOut bool
	lis         map[LogIndex]*isc.StateAnchor
	minLI       LogIndex         // Do not participate in LI lower than this.
	maxLI       LogIndex         // Cleanup all LIs smaller than this - hist.
	lastLI      LogIndex         // Just to wait for lastAnchor, if needed but not provided.
	lastAnchor  *isc.StateAnchor // Last Anchor seen confirmed in L1.
	hist        uint32           // How many instances to keep running.
	persistCB   func(li LogIndex)
	outputCB    func(lis Output)
	delayed     []LogIndex
	// pendingAfterSkipLI holds the next log index to advance to after a
	// consensus instance terminates with a SKIP/⊥ decision. It is applied on
	// the next Tick, which is driven by consensusDelay in the chain node.
	pendingAfterSkipLI  LogIndex
	hasPendingAfterSkip bool
	log                 log.Logger
}

// NewVarConsInsts is a constructor.
func NewVarConsInsts(
	minLI LogIndex,
	persistCB func(li LogIndex),
	outputCB func(lis Output),
	log log.Logger,
) *VarConsInsts {
	vci := &VarConsInsts{
		haveConsOut: false,
		lis: map[LogIndex]*isc.StateAnchor{
			minLI: nil,
		},
		minLI:      minLI,
		maxLI:      minLI,
		lastLI:     NilLogIndex(),
		lastAnchor: nil,
		hist:       3,
		persistCB:  persistCB,
		outputCB:   outputCB,
		delayed:    make([]LogIndex, 3), // Will wait for 3 time ticks before considering SeenLI.
		log:        log,
	}
	vci.outputCB(maps.Clone(vci.lis))
	return vci
}

// ConsOutputDone - Consensus at LI produced a TX.
func (vci *VarConsInsts) ConsOutputDone(li LogIndex, producedAnchor *isc.StateAnchor, cb onLIInc) gpa.OutMessages {
	vci.haveConsOut = true
	return vci.trySet(li.Next(), producedAnchor, cb)
}

// ConsOutputSkip - Consensus at LI terminate with a SKIP/⊥ decision.
func (vci *VarConsInsts) ConsOutputSkip(li LogIndex, cb onLIInc) gpa.OutMessages {
	vci.haveConsOut = true
	if vci.lastAnchor == nil {
		vci.lastLI = li.Next() // Will be set in LatestL1Anchor.
		return nil
	}
	// Defer advancing to the next LI until the next Tick, which is driven by
	// consensusDelay. This way, consecutive consensus runs are spaced by at
	// least the configured delay instead of tight-looping on SKIP results.
	vci.pendingAfterSkipLI = li.Next()
	vci.hasPendingAfterSkip = true
	return nil
}

// ConsOutputTimeout - Consensus at LI indicated a timeout.
func (vci *VarConsInsts) ConsOutputTimeout(li LogIndex, cb onLIInc) gpa.OutMessages {
	return vci.trySet(li.Next(), nil, cb)
}

// LatestSeenLI - If we see consensus proposals from F+1 nodes at seenLI...
func (vci *VarConsInsts) LatestSeenLI(seenLI LogIndex, cb onLIInc) gpa.OutMessages {
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

// LatestL1Anchor - Here we get the latest L1 state.
func (vci *VarConsInsts) LatestL1Anchor(ao *isc.StateAnchor, cb onLIInc) gpa.OutMessages {
	vci.lastAnchor = ao
	return vci.trySet(vci.lastLI, ao, cb) // Finish ConsOutputSkipBase, if pending.
}

func (vci *VarConsInsts) Tick(cb onLIInc) gpa.OutMessages {
	n := len(vci.delayed)
	last := vci.delayed[n-1]
	for i := n - 1; i > 0; i-- {
		vci.delayed[i] = vci.delayed[i-1]
	}
	vci.delayed[0] = NilLogIndex()
	msgs := gpa.NoMessages()
	if !last.IsNil() {
		msgs.AddAll(vci.trySet(last, nil, cb))
	}
	// Apply any pending advancement scheduled after a SKIP decision. This
	// ensures that the next consensus attempt only starts after at least one
	// consensusDelay tick has passed.
	if vci.hasPendingAfterSkip && vci.lastAnchor != nil {
		msgs.AddAll(vci.trySet(vci.pendingAfterSkipLI, vci.lastAnchor, cb))
		vci.hasPendingAfterSkip = false
	}
	return msgs
}

func (vci *VarConsInsts) trySet(li LogIndex, ao *isc.StateAnchor, cb onLIInc) gpa.OutMessages {
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

func (vci *VarConsInsts) StatusString() string {
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
