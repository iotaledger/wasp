package smGPA

import (
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/state"
)

type consensusStateProposalBlockRequest struct {
	consensusStateProposal *smInputs.ConsensusStateProposal
	done                   bool
	lastBlockHash          state.BlockHash
}

var _ blockRequest = &consensusStateProposalBlockRequest{}

func newConsensusStateProposalBlockRequest(input *smInputs.ConsensusStateProposal) (blockRequest, error) {
	stateCommitment, err := state.L1CommitmentFromAliasOutput(input.GetAliasOutputWithID().GetAliasOutput())
	if err != nil {
		return nil, err
	}
	return &consensusStateProposalBlockRequest{
		consensusStateProposal: input,
		done:                   false,
		lastBlockHash:          stateCommitment.BlockHash,
	}, nil
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

func (cspbrT *consensusStateProposalBlockRequest) blockAvailable(block state.Block) {}

func (cspbrT *consensusStateProposalBlockRequest) getPriority() uint32 {
	return topPriority
}

func (cspbrT *consensusStateProposalBlockRequest) markCompleted(createStateFun) {
	if cspbrT.isValid() {
		cspbrT.done = true
		cspbrT.consensusStateProposal.Respond()
	}
}
