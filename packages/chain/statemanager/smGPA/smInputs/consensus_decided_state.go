package smInputs

import (
	"context"

	"github.com/iotaledger/wasp/packages/chain/aaa2/cons/gr"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/state"
)

type ConsensusDecidedState struct {
	context         context.Context
	blockIndex      uint32 // TODO: temporary field. Remove it after DB refactoring.
	stateCommitment *state.L1Commitment
	resultCh        chan<- *consGR.StateMgrDecidedState
}

var _ gpa.Input = &ConsensusDecidedState{}

func NewConsensusDecidedState(ctx context.Context, aliasOutput *isc.AliasOutputWithID) (*ConsensusDecidedState, <-chan *consGR.StateMgrDecidedState) {
	sc, err := state.L1CommitmentFromAliasOutput(aliasOutput.GetAliasOutput())
	if err != nil {
		panic("Cannot make L1 commitment from alias output")
	}
	resultChannel := make(chan *consGR.StateMgrDecidedState, 1)
	return &ConsensusDecidedState{
		context:         ctx,
		blockIndex:      aliasOutput.GetStateIndex(),
		stateCommitment: sc,
		resultCh:        resultChannel,
	}, resultChannel
}

func (cdsT *ConsensusDecidedState) GetBlockIndex() uint32 { // TODO: temporary function. Remove it after DB refactoring.
	return cdsT.blockIndex
}

func (cdsT *ConsensusDecidedState) GetStateCommitment() *state.L1Commitment {
	return cdsT.stateCommitment
}

func (cdsT *ConsensusDecidedState) IsValid() bool {
	return cdsT.context.Err() == nil
}

func (cdsT *ConsensusDecidedState) Respond(virtualStateAccess state.VirtualStateAccess) {
	if cdsT.IsValid() && !cdsT.IsResultChClosed() {
		cdsT.resultCh <- &consGR.StateMgrDecidedState{
			StateBaseline:      coreutil.NewChainStateSync().SetSolidIndex(virtualStateAccess.BlockIndex()).GetSolidIndexBaseline(), // TODO - move it to Respond method parameters?
			VirtualStateAccess: virtualStateAccess,
		}
		cdsT.closeResultCh()
	}
}

func (cdsT *ConsensusDecidedState) IsResultChClosed() bool {
	return cdsT.resultCh == nil
}

func (cdsT *ConsensusDecidedState) closeResultCh() {
	close(cdsT.resultCh)
	cdsT.resultCh = nil
}
