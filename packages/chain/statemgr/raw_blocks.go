package statemgr

import (
	"os"
	"path"

	"github.com/iotaledger/wasp/packages/state"
)

func (sm *stateManager) setRawBlocksOptions() {
	// parameters are not loaded in the context of unit tests
	if sm.unitTests || !sm.rawBlocksEnabled {
		return
	}

	dir := path.Join(sm.rawBlocksDir, sm.chain.ID().String())
	if err := os.MkdirAll(dir, 0o777); err != nil {
		sm.log.Errorf("create dir: %v", err)
		sm.log.Warnf("raw blocks won't be stored")
		return
	}
	sm.solidState.WithOnBlockSave(state.SaveRawBlockClosure(dir, sm.log))
	sm.log.Infof("raw blocks will be saved to %s", dir)
}
