package chainmanager

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// This message is used to inform access nodes on new blocks
// produced so that they can update their active state faster.
type msgBlockProduced struct {
	gpa.BasicMessage
	tx    *iotago.Transaction
	block state.Block
}

var _ gpa.Message = &msgCmtLog{}

func NewMsgBlockProduced(recipient gpa.NodeID, tx *iotago.Transaction, block state.Block) gpa.Message {
	return &msgBlockProduced{
		BasicMessage: gpa.NewBasicMessage(recipient),
		tx:           tx,
		block:        block,
	}
}

func (msg *msgBlockProduced) String() string {
	txID, err := msg.tx.ID()
	if err != nil {
		panic(fmt.Errorf("cannot extract TX ID: %w", err))
	}
	return fmt.Sprintf(
		"{chainMgr.msgBlockProduced, stateIndex=%v, l1Commitment=%v, tx.ID=%v}",
		msg.block.StateIndex(), msg.block.L1Commitment(), txID.ToHex(),
	)
}

func (msg *msgBlockProduced) MarshalBinary() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	if err := rwutil.WriteByte(w, msgTypeBlockProduced); err != nil {
		return nil, fmt.Errorf("cannot serialize msgType: %w", err)
	}
	//
	// TX
	txBytes, err := msg.tx.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot serialize tx: %w", err)
	}
	if err := rwutil.WriteBytes(w, txBytes); err != nil {
		return nil, fmt.Errorf("cannot write tx bytes: %w", err)
	}
	//
	// Block
	if err := rwutil.WriteBytes(w, msg.block.Bytes()); err != nil {
		return nil, fmt.Errorf("cannot serialize block: %w", err)
	}
	return w.Bytes(), nil
}

func (msg *msgBlockProduced) UnmarshalBinary(data []byte) error {
	var err error
	r := bytes.NewReader(data)
	//
	// MsgType
	msgType, err := rwutil.ReadByte(r)
	if err != nil {
		return fmt.Errorf("cannot read msgType byte: %w", err)
	}
	if msgType != msgTypeBlockProduced {
		return fmt.Errorf("unexpected msgType: %v", msgType)
	}
	//
	// TX
	txBytes, err := rwutil.ReadBytes(r)
	if err != nil {
		return fmt.Errorf("cannot read tx bytes: %w", err)
	}
	tx := &iotago.Transaction{}
	_, err = tx.Deserialize(txBytes, serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return fmt.Errorf("cannot deserialize tx: %w", err)
	}
	msg.tx = tx
	//
	// Block
	blockBytes, err := rwutil.ReadBytes(r)
	if err != nil {
		return fmt.Errorf("cannot read block bytes: %w", err)
	}
	block, err := state.BlockFromBytes(blockBytes)
	if err != nil {
		return fmt.Errorf("cannot deserialize block: %w", err)
	}
	msg.block = block
	return nil
}
