package cmtLog

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/generics/shrinkingmap"
	"github.com/iotaledger/wasp/packages/isc"
)

type VarRunning interface {
	IsLatest(li LogIndex) bool
	ConsensusProposed(li LogIndex, proposedAO *isc.AliasOutputWithID)
	ConsensusOutput(li LogIndex, proposedAO *isc.AliasOutputWithID) bool // Returns, if the event was outdated.
	Inconsistent()
	StatusString() string
}

type varRunning struct {
	latest    LogIndex
	consInsts *shrinkingmap.ShrinkingMap[LogIndex, *isc.AliasOutputWithID]
}

func NewVarRunning() VarRunning {
	return &varRunning{
		latest:    NilLogIndex(),
		consInsts: shrinkingmap.New[LogIndex, *isc.AliasOutputWithID](),
	}
}

func (vr *varRunning) IsLatest(li LogIndex) bool {
	return li == vr.latest
}

func (vr *varRunning) ConsensusProposed(li LogIndex, proposedAO *isc.AliasOutputWithID) {
	if vr.consInsts.Has(li) {
		panic(fmt.Errorf("duplicate consensus instance started, li=%v, proposedAO=%v", li, proposedAO))
	}
	if vr.latest >= li {
		panic(fmt.Errorf("trying to start outdated consensus instance, li=%v, proposedAO=%v, latest=%v", li, proposedAO, vr.latest))
	}
	vr.latest = li
	vr.consInsts.Set(li, proposedAO)
}

// If consensus is done, it is not running anymore. Also, the instances
// started before it are not needed anymore, thus outdated.
func (vr *varRunning) ConsensusOutput(li LogIndex, proposedAO *isc.AliasOutputWithID) bool {
	has := vr.consInsts.Has(li)
	vr.consInsts.ForEachKey(func(runningLI LogIndex) bool {
		if runningLI <= li {
			vr.consInsts.Delete(runningLI)
		}
		return true
	})
	return has
}

// If CmtLog becomes inconsistent, we will ignore all the running consensus
// instances and will only track the confirmed outputs.
func (vr *varRunning) Inconsistent() {
	vr.consInsts.ForEachKey(func(runningLI LogIndex) bool {
		vr.consInsts.Delete(runningLI)
		return true
	})
}

func (vr *varRunning) StatusString() string {
	return fmt.Sprintf("{running, count=%v, latest=%v}", vr.consInsts.Size(), vr.latest)
}
