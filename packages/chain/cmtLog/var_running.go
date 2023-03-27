package cmtLog

import (
	"fmt"

	"github.com/iotaledger/hive.go/logger"
)

// All functions returns true, if the latest LI become completed.
type VarRunning interface {
	IsLatest(li LogIndex) bool
	ConsensusProposed(li LogIndex)
	ConsensusRecover(li LogIndex) bool
	ConsensusOutputDone(li LogIndex) bool
	ConsensusOutputSkip(li LogIndex) bool
	ConsensusOutputConfirmed(li LogIndex) bool
	ConsensusOutputRejected(li LogIndex) bool
	StatusString() string
}

type varRunning struct {
	latestLI  LogIndex
	completed bool
	log       *logger.Logger
}

func NewVarRunning(log *logger.Logger) VarRunning {
	return &varRunning{
		latestLI:  NilLogIndex(),
		completed: true,
		log:       log,
	}
}

func (vr *varRunning) IsLatest(li LogIndex) bool                 { return li == vr.latestLI }
func (vr *varRunning) ConsensusProposed(li LogIndex)             { vr.markLogIndex(li, false) }
func (vr *varRunning) ConsensusRecover(li LogIndex) bool         { return vr.markLogIndex(li, true) }
func (vr *varRunning) ConsensusOutputDone(li LogIndex) bool      { return vr.markLogIndex(li, true) }
func (vr *varRunning) ConsensusOutputSkip(li LogIndex) bool      { return vr.markLogIndex(li, true) }
func (vr *varRunning) ConsensusOutputConfirmed(li LogIndex) bool { return vr.markLogIndex(li, true) }
func (vr *varRunning) ConsensusOutputRejected(li LogIndex) bool  { return vr.markLogIndex(li, true) }

func (vr *varRunning) markLogIndex(li LogIndex, completed bool) bool {
	if li < vr.latestLI {
		vr.log.Debugf("⫢ li=%v/%v outdated, latest=%v/%v", li, completed, vr.latestLI, vr.completed)
		return false
	}
	if li == vr.latestLI {
		if !vr.completed && completed {
			vr.log.Debugf("⫢ marking li=%v as completed", li)
			vr.completed = true
			return true
		}
		vr.log.Debugf("⫢ keeping li=%v as completed=%v", li, vr.completed)
		return false
	}
	//
	// li > vr.latestLI
	vr.latestLI = li
	vr.completed = completed
	vr.log.Debugf("⫢ new latest li=%v, completed=%v", li, completed)
	return vr.completed
}

func (vr *varRunning) StatusString() string {
	return fmt.Sprintf("{running, latestLI=%v, completed=%v}", vr.latestLI, vr.completed)
}
