package chainmanager

import (
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// gpa.Wrapper is not applicable here, because here the addressing
// is by CommitteeID, not by integer index.
type msgCmtLog struct {
	committeeAddr iotago.Ed25519Address
	wrapped       gpa.Message
}

var _ gpa.Message = new(msgCmtLog)

func NewMsgCmtLog(committeeAddr iotago.Ed25519Address, wrapped gpa.Message) gpa.Message {
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

func (msg *msgCmtLog) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	msgTypeCmtLog.ReadAndVerify(rr)
	rr.ReadN(msg.committeeAddr[:])
	msg.wrapped = rwutil.ReadFromFunc(rr, cmt_log.UnmarshalMessage)
	return rr.Err
}

func (msg *msgCmtLog) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	msgTypeCmtLog.Write(ww)
	ww.WriteN(msg.committeeAddr[:])
	ww.WriteBytes(rwutil.WriteToBytes(msg.wrapped))
	return ww.Err
}
