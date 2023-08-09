package cmt_log

import (
	"fmt"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc"
)

type VarOutput interface {
	// Summary of the internal state.
	StatusString() string
	Value() *Output
	LogIndexAgreed(li LogIndex)
	TipAOChanged(ao *isc.AliasOutputWithID)
	CanPropose()
	Suspended(suspended bool)
}

type varOutputImpl struct {
	candidateLI LogIndex
	candidateAO *isc.AliasOutputWithID
	canPropose  bool
	suspended   bool
	outValue    *Output
	persistUsed func(li LogIndex)
	log         *logger.Logger
}

func NewVarOutput(persistUsed func(li LogIndex), log *logger.Logger) VarOutput {
	return &varOutputImpl{
		candidateLI: NilLogIndex(),
		candidateAO: nil,
		canPropose:  true,
		suspended:   false,
		outValue:    nil,
		persistUsed: persistUsed,
		log:         log,
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

func (vo *varOutputImpl) TipAOChanged(ao *isc.AliasOutputWithID) {
	vo.candidateAO = ao
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
	//
	// Output the new data.
	vo.persistUsed(vo.candidateLI)
	vo.outValue = makeOutput(vo.candidateLI, vo.candidateAO)
	vo.log.Infof("⊪ Output %v", vo.outValue)
	vo.canPropose = false
	vo.candidateLI = NilLogIndex()
}
