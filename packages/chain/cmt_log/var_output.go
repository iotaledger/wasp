package cmt_log

import (
	"fmt"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/types"
)

type VarOutput interface {
	// Summary of the internal state.
	StatusString() string
	Value() *Output
	LogIndexAgreed(li LogIndex)
	TipAOChanged(ao *types.Anchor)
	HaveRejection()
	HaveMilestone()
	CanPropose()
	Suspended(suspended bool)
}

type varOutputImpl struct {
	candidateLI                LogIndex
	candidateAO                *types.Anchor
	canPropose                 bool
	milestonesToWait           int
	suspended                  bool
	outValue                   *Output
	persistUsed                func(li LogIndex)
	postponeRecoveryMilestones int
	log                        *logger.Logger
}

func NewVarOutput(persistUsed func(li LogIndex), postponeRecoveryMilestones int, log *logger.Logger) VarOutput {
	return &varOutputImpl{
		candidateLI:                NilLogIndex(),
		candidateAO:                nil,
		canPropose:                 true,
		milestonesToWait:           0,
		suspended:                  false,
		outValue:                   nil,
		persistUsed:                persistUsed,
		postponeRecoveryMilestones: postponeRecoveryMilestones,
		log:                        log,
	}
}

func (vo *varOutputImpl) StatusString() string {
	return fmt.Sprintf(
		"{varOutput: output=%v, candidate{li=%v, ao=%v}, canPropose=%v, suspended=%v}",
		vo.outValue, vo.candidateLI, vo.candidateAO, vo.canPropose, vo.suspended,
	)
}

func (vo *varOutputImpl) Value() *Output {
	if vo.outValue == nil || vo.suspended {
		return nil // Untyped nil.
	}
	return vo.outValue
}

func (vo *varOutputImpl) LogIndexAgreed(li LogIndex) {
	vo.candidateLI = li
	vo.tryOutput()
}

func (vo *varOutputImpl) TipAOChanged(ao *types.Anchor) {
	vo.candidateAO = ao
	vo.tryOutput()
}

// This works in hand with HaveMilestone. See the comment there.
func (vo *varOutputImpl) HaveRejection() {
	vo.milestonesToWait = vo.postponeRecoveryMilestones
	vo.tryOutput()
}

// We set the milestonesToWait on any rejection, but start to decrease it only
// after the rejection is resolved completely. This way we make a grace-delay
// to work around the L1 problems with reporting rejection prematurely.
func (vo *varOutputImpl) HaveMilestone() {
	if vo.milestonesToWait == 0 {
		return
	}
	if vo.candidateLI.IsNil() || vo.candidateAO == nil {
		return
	}
	vo.milestonesToWait--
	vo.tryOutput()
}

func (vo *varOutputImpl) CanPropose() {
	vo.canPropose = true
	vo.tryOutput()
}

func (vo *varOutputImpl) Suspended(suspended bool) {
	if vo.suspended && !suspended {
		vo.log.Infof("Committee resumed.")
	}
	if !vo.suspended && suspended {
		vo.log.Infof("Committee suspended.")
	}
	vo.suspended = suspended
}

func (vo *varOutputImpl) tryOutput() {
	if vo.candidateLI.IsNil() || vo.candidateAO == nil || !vo.canPropose {
		// Keep output unchanged.
		return
	}
	if vo.milestonesToWait > 0 {
		// Postponed, wait for several milestones after a rejection.
		vo.log.Debugf("TIP decision postponed, milestonesToWait=%v", vo.milestonesToWait)
		return
	}
	//
	// Output the new data.
	vo.persistUsed(vo.candidateLI)
	vo.outValue = makeOutput(vo.candidateLI, vo.candidateAO)
	vo.log.Infof("⊪ Output %v", vo.outValue)
	vo.canPropose = false
	vo.candidateLI = NilLogIndex()
}
