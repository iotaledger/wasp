package statemgr

import (
	"github.com/iotaledger/wasp/packages/state"
	"os"
	"path"
)

// TODO temporary. Must take parameters from something global

const (
	saveRawBlocks          = false
	saveRawBlocksDirectory = "blocks"
)

func (sm *stateManager) setRawBlocksOptions() {
	if !saveRawBlocks {
		return
	}
	dir := path.Join(saveRawBlocksDirectory, sm.chain.ID().String())
	if err := os.MkdirAll(dir, 0o777); err != nil {
		sm.log.Errorf("create dir: %v", err)
		sm.log.Warnf("raw blocks won't be stored")
		return
	}
	sm.solidState.WithOnBlockSave(state.SaveRawBlockClosure(dir, sm.log))
	sm.log.Infof("raw blocks will be saved to %s", dir)
}
