package chain

import (
	"strconv"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/publisher"
)

// LogStateTransition also used in testing
func LogStateTransition(blockIndex uint32, outputID string, rootCommitment trie.VCommitment, reqids []isc.RequestID, log *logger.Logger) {
	if blockIndex > 0 {
		log.Infof("STATE TRANSITION TO #%d. Requests: %d, chain output: %s", blockIndex, len(reqids), outputID)
		log.Debugf("STATE TRANSITION. Root commitment: %s", rootCommitment)
	} else {
		log.Infof("ORIGIN STATE SAVED. State output id: %s", outputID)
		log.Debugf("ORIGIN STATE SAVED. Root commitment: %s", rootCommitment)
	}
}

// LogGovernanceTransition
func LogGovernanceTransition(blockIndex uint32, outputID string, rootCommitment trie.VCommitment, log *logger.Logger) {
	log.Infof("GOVERNANCE TRANSITION. State index #%d, anchor output: %s", blockIndex, outputID)
	log.Debugf("GOVERNANCE TRANSITION. Root commitment: %s", rootCommitment)
}

func PublishRequestsSettled(chainID *isc.ChainID, stateIndex uint32, reqids []isc.RequestID) {
	for _, reqid := range reqids {
		publisher.Publish("request_out",
			chainID.String(),
			reqid.String(),
			strconv.Itoa(int(stateIndex)),
			strconv.Itoa(len(reqids)),
		)
	}
}

func PublishStateTransition(chainID *isc.ChainID, stateOutput *isc.AliasOutputWithID, reqIDsLength int) {
	stateHash, _ := hashing.HashValueFromBytes(stateOutput.GetStateMetadata())

	publisher.Publish("state",
		chainID.String(),
		strconv.Itoa(int(stateOutput.GetStateIndex())),
		strconv.Itoa(reqIDsLength),
		isc.OID(stateOutput.ID()),
		stateHash.String(),
	)
}

func PublishGovernanceTransition(stateOutput *isc.AliasOutputWithID) {
	stateHash, _ := hashing.HashValueFromBytes(stateOutput.GetStateMetadata())
	chainID := isc.ChainIDFromAliasID(stateOutput.GetAliasID())

	publisher.Publish("rotate",
		chainID.String(),
		strconv.Itoa(int(stateOutput.GetStateIndex())),
		isc.OID(stateOutput.ID()),
		stateHash.String(),
	)
}
