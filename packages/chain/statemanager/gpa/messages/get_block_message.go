package messages

import (
	"github.com/iotaledger/wasp/v2/packages/state"
)

type GetBlockMessage struct {
	commitment *state.L1Commitment `bcs:"export"`
}

func NewGetBlockMessage(commitment *state.L1Commitment) GetBlockMessage {
	return GetBlockMessage{
		commitment: commitment,
	}
}

func (msg *GetBlockMessage) GetL1Commitment() *state.L1Commitment {
	return msg.commitment
}
