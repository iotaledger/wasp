package chain

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"strconv"
)

// LogStateTransition also used in testing
func LogStateTransition(msg *StateTransitionEventData, log *logger.Logger) {
	reqids := blocklog.GetRequestIDsForLastBlock(msg.VirtualState)
	if msg.ChainOutput.GetStateIndex() > 0 {
		log.Infof("STATE TRANSITION TO #%d. Chain output: %s, block size: %d",
			msg.VirtualState.BlockIndex(), coretypes.OID(msg.ChainOutput.ID()), len(reqids))
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

func PublishStateTransition(newState state.VirtualState, stateOutput *ledgerstate.AliasOutput) []coretypes.RequestID {
	reqids := blocklog.GetRequestIDsForLastBlock(newState)
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
