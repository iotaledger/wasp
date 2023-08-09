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

func (voi *varOutputImpl) StatusString() string {
	return fmt.Sprintf(
		"{varOutput: output=%v, candidate{li=%v, ao=%v}, canPropose=%v, suspended=%v}",
		voi.outValue, voi.candidateLI, voi.candidateAO, voi.canPropose, voi.suspended,
	)
}

func (voi *varOutputImpl) Value() *Output {
	if voi.outValue == nil || voi.suspended {
		return nil // Untyped nil.
	}
	return voi.outValue
}

func (voi *varOutputImpl) LogIndexAgreed(li LogIndex) {
	voi.candidateLI = li
	voi.tryOutput()
}

func (voi *varOutputImpl) TipAOChanged(ao *isc.AliasOutputWithID) {
	voi.candidateAO = ao
	voi.tryOutput()
}

func (voi *varOutputImpl) CanPropose() {
	voi.canPropose = true
	voi.tryOutput()
}

func (voi *varOutputImpl) Suspended(suspended bool) {
	if voi.suspended && !suspended {
		voi.log.Infof("Committee resumed.")
	}
	if !voi.suspended && suspended {
		voi.log.Infof("Committee suspended.")
	}
	voi.suspended = suspended
}

func (voi *varOutputImpl) tryOutput() {
	if voi.candidateLI.IsNil() || voi.candidateAO == nil || !voi.canPropose {
		// Keep output unchanged.
		return
	}
	//
	// Output the new data.
	voi.persistUsed(voi.candidateLI)
	voi.outValue = makeOutput(voi.candidateLI, voi.candidateAO)
	voi.log.Infof("⊪ Output %p", voi.outValue)
	voi.canPropose = false
	voi.candidateLI = NilLogIndex()
}
