package smGPA

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/state"
)

type consensusStateProposalBlockRequest struct {
	consensusStateProposal *smInputs.ConsensusStateProposal
	done                   bool
	lastBlockHash          state.BlockHash
	log                    *logger.Logger
	id                     blockRequestID
}

var _ blockRequest = &consensusStateProposalBlockRequest{}

func newConsensusStateProposalBlockRequest(input *smInputs.ConsensusStateProposal, log *logger.Logger, id blockRequestID) (blockRequest, error) {
	stateCommitment, err := state.L1CommitmentFromAliasOutput(input.GetAliasOutputWithID().GetAliasOutput())
	if err != nil {
		return nil, err
	}
	result := &consensusStateProposalBlockRequest{
		consensusStateProposal: input,
		done:                   false,
		lastBlockHash:          stateCommitment.BlockHash,
		id:                     id,
	}
	result.log = log.Named(fmt.Sprintf("r%s-%v", result.getType(), id))
	result.log.Debugf("Consensus state proposal block request id %v for block %s is created", result.id, result.lastBlockHash)
	return result, nil
}

func (cspbrT *consensusStateProposalBlockRequest) getLastBlockHash() state.BlockHash {
	return cspbrT.lastBlockHash
}

func (cspbrT *consensusStateProposalBlockRequest) getLastBlockIndex() uint32 { // TODO: temporary function. Remove it after DB refactoring.
	return cspbrT.consensusStateProposal.GetAliasOutputWithID().GetStateIndex()
}

func (cspbrT *consensusStateProposalBlockRequest) isValid() bool {
	return !cspbrT.done && cspbrT.consensusStateProposal.IsValid()
}

func (cspbrT *consensusStateProposalBlockRequest) blockAvailable(block state.Block) {
	cspbrT.log.Debugf("Consensus state proposal block request received block %s, ignoring it", block.GetHash())
}

func (cspbrT *consensusStateProposalBlockRequest) getPriority() uint64 {
	return topPriority
}

func (cspbrT *consensusStateProposalBlockRequest) markCompleted(createStateFun) {
	if cspbrT.isValid() {
		cspbrT.log.Debugf("Consensus state proposal block request is valid, marking it completed and responding to consensus")
		cspbrT.done = true
		cspbrT.consensusStateProposal.Respond()
		cspbrT.log.Debugf("Responded")
	} else {
		cspbrT.log.Debugf("Consensus state proposal block request is not valid, ignoring mark completed call")
	}
}

func (cspbrT *consensusStateProposalBlockRequest) getType() string {
	return "csp"
}

func (cspbrT *consensusStateProposalBlockRequest) getID() blockRequestID {
	return cspbrT.id
}
