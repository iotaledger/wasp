package statemgr

import (
	"os"
	"path"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
)

func (sm *stateManager) setRawBlocksOptions() {
	if !parameters.GetBool(parameters.RawBlocksEnabled) {
		return
	}
	dir := path.Join(parameters.GetString(parameters.RawBlocksDir), sm.chain.ID().String())
	if err := os.MkdirAll(dir, 0o777); err != nil {
		sm.log.Errorf("create dir: %v", err)
		sm.log.Warnf("raw blocks won't be stored")
		return
	}
	sm.solidState.WithOnBlockSave(state.SaveRawBlockClosure(dir, sm.log))
	sm.log.Infof("raw blocks will be saved to %s", dir)
}
