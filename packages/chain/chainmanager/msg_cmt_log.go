package chainmanager

import (
	"fmt"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

// gpa.Wrapper is not applicable here, because here the addressing
// is by CommitteeID, not by integer index.
type msgCmtLog struct {
	committeeAddr cryptolib.Address
	wrapped       gpa.Message
}

var _ gpa.Message = new(msgCmtLog)

func NewMsgCmtLog(committeeAddr cryptolib.Address, wrapped gpa.Message) gpa.Message {
	return &msgCmtLog{
		committeeAddr: committeeAddr,
		wrapped:       wrapped,
	}
}

func (msg *msgCmtLog) MsgType() gpa.MessageType {
	return msgTypeCmtLog
}

func (msg *msgCmtLog) String() string {
	return fmt.Sprintf("{chainMgr.msgCmtLog, committeeAddr=%v, wrapped=%+v}", msg.committeeAddr.String(), msg.wrapped)
}

func (msg *msgCmtLog) Recipient() gpa.NodeID {
	return msg.wrapped.Recipient()
}

func (msg *msgCmtLog) SetSender(sender gpa.NodeID) {
	msg.wrapped.SetSender(sender)
}

func (msg *msgCmtLog) MarshalBCS(e *bcs.Encoder) error {
	wrappedBytes, err := gpa.MarshalMessage(msg.wrapped)
	if err != nil {
		return fmt.Errorf("marshaling wrapped message: %w", err)
	}

	e.Encode(msg.committeeAddr)
	e.Encode(wrappedBytes)

	return nil
}

func (msg *msgCmtLog) UnmarshalBCS(d *bcs.Decoder) error {
	d.Decode(&msg.committeeAddr)
	wrappedBytes := bcs.Decode[[]byte](d)

	var err error
	msg.wrapped, err = cmtlog.UnmarshalMessage(wrappedBytes)
	if err != nil {
		return fmt.Errorf("unmarshaling wrapped message: %w", err)
	}

	return nil
}
