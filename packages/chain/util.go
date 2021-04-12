package chain

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
)

// LogStateTransition also used in testing
func LogStateTransition(msg *StateTransitionEventData, log *logger.Logger) {
	if msg.ChainOutput.GetStateIndex() > 0 {
		log.Infof("STATE TRANSITION TO #%d. Chain output: %s, block size: %d",
			msg.VariableState.BlockIndex(), coretypes.OID(msg.ChainOutput.ID()), len(msg.RequestIDs))
		log.Debugf("STATE TRANSITION. State hash: %s, block essence: %s",
			msg.VariableState.Hash().String(), msg.BlockEssenceHash.String())
	} else {
		log.Infof("ORIGIN STATE SAVED. State output id: %s", coretypes.OID(msg.ChainOutput.ID()))
		log.Debugf("ORIGIN STATE SAVED. state hash: %s, block essence: %s",
			msg.VariableState.Hash().String(), msg.BlockEssenceHash.String())
	}
}

func LogSyncedEvent(outputID ledgerstate.OutputID, blockIndex uint32, log *logger.Logger) {
	log.Infof("EVENT: state was synced to block index #%d, approving output: %s", blockIndex, coretypes.OID(outputID))
}
