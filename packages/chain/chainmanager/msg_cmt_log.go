package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

// gpa.Wrapper is not applicable here, because here the addressing
// is by CommitteeID, not by integer index.
type msgCmtLog struct {
	committeeAddr cryptolib.Address `bcs:""`
	wrapped       gpa.Message       `bcs:"not_enum,bytearr"`
}

var _ gpa.Message = new(msgCmtLog)

func NewMsgCmtLog(committeeAddr cryptolib.Address, wrapped gpa.Message) gpa.Message {
	return &msgCmtLog{
		committeeAddr: committeeAddr,
		wrapped:       wrapped,
	}
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

func (msg *msgCmtLog) UnmarshalBCS(d *bcs.Decoder) error {
	d.Decode(&msg.committeeAddr)
	var wrapped []byte
	d.Decode(&wrapped)

	if d.Err() != nil {
		return d.Err()
	}

	var err error
	msg.wrapped, err = cmt_log.UnmarshalMessage(wrapped)

	return err
}
