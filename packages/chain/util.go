package chain

import (
	"strconv"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
)

// LogStateTransition also used in testing
func LogStateTransition(msg *ChainTransitionEventData, reqids []iscp.RequestID, log *logger.Logger) {
	if msg.ChainOutput.GetStateIndex() > 0 {
		log.Infof("STATE TRANSITION TO #%d. Requests: %d, chain output: %s",
			msg.VirtualState.BlockIndex(), len(reqids), iscp.OID(msg.ChainOutput.ID()))
		log.Debugf("STATE TRANSITION. Root commitment: %s", state.RootCommitment(msg.VirtualState.TrieNodeStore()))
	} else {
		log.Infof("ORIGIN STATE SAVED. State output id: %s", iscp.OID(msg.ChainOutput.ID()))
		log.Debugf("ORIGIN STATE SAVED. Root commitment: %s", state.RootCommitment(msg.VirtualState.TrieNodeStore()))
	}
}

// LogGovernanceTransition
func LogGovernanceTransition(msg *ChainTransitionEventData, log *logger.Logger) {
	log.Infof("GOVERNANCE TRANSITION. State index #%d, anchor output: %s",
		msg.VirtualState.BlockIndex(), iscp.OID(msg.ChainOutput.ID()))
	log.Debugf("GOVERNANCE TRANSITION. Root commitment: %s", state.RootCommitment(msg.VirtualState.TrieNodeStore()))
}

func PublishRequestsSettled(chainID *iscp.ChainID, stateIndex uint32, reqids []iscp.RequestID) {
	for _, reqid := range reqids {
		publisher.Publish("request_out",
			chainID.String(),
			reqid.String(),
			strconv.Itoa(int(stateIndex)),
			strconv.Itoa(len(reqids)),
		)
	}
}

func PublishStateTransition(chainID *iscp.ChainID, stateOutput *iscp.AliasOutputWithID, reqIDsLength int) {
	stateHash, _ := hashing.HashValueFromBytes(stateOutput.GetStateMetadata())

	publisher.Publish("state",
		chainID.String(),
		strconv.Itoa(int(stateOutput.GetStateIndex())),
		strconv.Itoa(reqIDsLength),
		iscp.OID(stateOutput.ID()),
		stateHash.String(),
	)
}

func PublishGovernanceTransition(stateOutput *iscp.AliasOutputWithID) {
	stateHash, _ := hashing.HashValueFromBytes(stateOutput.GetStateMetadata())
	chainID := iscp.ChainIDFromAliasID(stateOutput.GetAliasID())

	publisher.Publish("rotate",
		chainID.String(),
		strconv.Itoa(int(stateOutput.GetStateIndex())),
		iscp.OID(stateOutput.ID()),
		stateHash.String(),
	)
}
