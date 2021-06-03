package chain

import (
	"strconv"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
)

// LogStateTransition also used in testing
func LogStateTransition(msg *StateTransitionEventData, log *logger.Logger) {
	if msg.ChainOutput.GetStateIndex() > 0 {
		log.Infof("STATE TRANSITION TO #%d. Chain output: %s, block size: %d",
			msg.VirtualState.BlockIndex(), coretypes.OID(msg.ChainOutput.ID()), len(msg.RequestIDs))
		log.Debugf("STATE TRANSITION. State hash: %s",
			msg.VirtualState.Hash().String())
	} else {
		log.Infof("ORIGIN STATE SAVED. State output id: %s", coretypes.OID(msg.ChainOutput.ID()))
		log.Debugf("ORIGIN STATE SAVED. state hash: %s",
			msg.VirtualState.Hash().String())
	}
}

func LogSyncedEvent(outputID ledgerstate.OutputID, blockIndex uint32, log *logger.Logger) {
	log.Infof("EVENT: state was synced to block index #%d, approving output: %s", blockIndex, coretypes.OID(outputID))
}

func PublishStateTransition(newState state.VirtualState, stateOutput *ledgerstate.AliasOutput, reqids []coretypes.RequestID) []coretypes.RequestID {
	chainID := coretypes.NewChainID(stateOutput.GetAliasAddress())

	publisher.Publish("state",
		chainID.String(),
		strconv.Itoa(int(newState.BlockIndex())),
		strconv.Itoa(len(reqids)),
		coretypes.OID(stateOutput.ID()),
		newState.Hash().String(),
	)
	for _, reqid := range reqids {
		publisher.Publish("request_out",
			chainID.String(),
			reqid.String(),
			strconv.Itoa(int(newState.BlockIndex())),
			strconv.Itoa(len(reqids)),
		)
	}
	return reqids
}
