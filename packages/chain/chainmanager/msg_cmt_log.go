package chainmanager

import (
	"fmt"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

// gpa.Wrapper is not applicable here, because here the addressing
// is by CommitteeID, not by integer index.
type msgCommitteeLog struct {
	committeeAddr cryptolib.Address
	wrapped       gpa.Message
}

var _ gpa.Message = new(msgCommitteeLog)

func NewMsgCommitteeLog(committeeAddr cryptolib.Address, wrapped gpa.Message) gpa.Message {
	return &msgCommitteeLog{
		committeeAddr: committeeAddr,
		wrapped:       wrapped,
	}
}

func (msg *msgCommitteeLog) MsgType() gpa.MessageType {
	return msgTypeCommitteeLog
}

func (msg *msgCommitteeLog) String() string {
	return fmt.Sprintf("{chainMgr.msgCommitteeLog, committeeAddr=%v, wrapped=%+v}", msg.committeeAddr.String(), msg.wrapped)
}

func (msg *msgCommitteeLog) Recipient() gpa.NodeID {
	return msg.wrapped.Recipient()
}

func (msg *msgCommitteeLog) SetSender(sender gpa.NodeID) {
	msg.wrapped.SetSender(sender)
}

func (msg *msgCommitteeLog) MarshalBCS(e *bcs.Encoder) error {
	wrappedBytes, err := gpa.MarshalMessage(msg.wrapped)
	if err != nil {
		return fmt.Errorf("marshaling wrapped message: %w", err)
	}

	e.Encode(msg.committeeAddr)
	e.Encode(wrappedBytes)

	return nil
}

func (msg *msgCommitteeLog) UnmarshalBCS(d *bcs.Decoder) error {
	d.Decode(&msg.committeeAddr)
	wrappedBytes := bcs.Decode[[]byte](d)

	var err error
	msg.wrapped, err = committeelog.UnmarshalMessage(wrappedBytes)
	if err != nil {
		return fmt.Errorf("unmarshaling wrapped message: %w", err)
	}

	return nil
}
