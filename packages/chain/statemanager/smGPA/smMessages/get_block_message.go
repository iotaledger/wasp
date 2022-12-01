package smMessages

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

type GetBlockMessage struct {
	gpa.BasicMessage
	blockHash  state.BlockHash
	blockIndex uint32 // TODO: temporary field. Remove it after DB is refactored.
}

var _ gpa.Message = &GetBlockMessage{}

// TODO: `blockIndex` is a temporary parameter. Remove it after DB is refactored.
func NewGetBlockMessage(blockIndex uint32, blockHash state.BlockHash, to gpa.NodeID) *GetBlockMessage {
	return &GetBlockMessage{
		BasicMessage: gpa.NewBasicMessage(to),
		blockHash:    blockHash,
		blockIndex:   blockIndex,
	}
}

func NewEmptyGetBlockMessage() *GetBlockMessage { // `UnmarshalBinary` must be called afterwards
	return NewGetBlockMessage(0, state.BlockHash{}, "UNKNOWN")
}

func (gbmT *GetBlockMessage) MarshalBinary() (data []byte, err error) {
	result := append([]byte{MsgTypeGetBlockMessage}, util.Uint32To4Bytes(gbmT.blockIndex)...)
	return append(result, gbmT.blockHash[:]...), nil
}

func (gbmT *GetBlockMessage) UnmarshalBinary(data []byte) error {
	if data[0] != MsgTypeGetBlockMessage {
		return fmt.Errorf("Error creating get block message from bytes: wrong message type %v", data[0])
	}
	// TODO: temporary code. Remove it after DB is refactored.
	if len(data) < 5 {
		return fmt.Errorf("Error creating get block message from bytes: wrong size %v, expecting 5 or more", len(data))
	}
	var err error
	gbmT.blockIndex, err = util.Uint32From4Bytes(data[1:5])
	if err != nil {
		return err
	}
	// End of temporary code
	actualData := data[5:] //data[1:]
	if len(actualData) != state.BlockHashSize {
		return fmt.Errorf("Error creating get block message from bytes: wrong size %v, expecting %v", len(actualData), state.BlockHashSize)
	}
	copy(gbmT.blockHash[:], actualData)
	return nil
}

func (gbmT *GetBlockMessage) GetBlockHash() state.BlockHash {
	return gbmT.blockHash
}

func (gbmT *GetBlockMessage) GetBlockIndex() uint32 { // TODO: temporary function. Remove it after DB is refactored.
	return gbmT.blockIndex
}
