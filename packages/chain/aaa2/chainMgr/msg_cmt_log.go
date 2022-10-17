package chainMgr

import (
	"bytes"

	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

// gpa.Wrapper is not applicable here, because here the addressing
// is by CommitteeID, not by integer index.
type msgCmtLog struct {
	committeeID CommitteeID
	wrapped     gpa.Message
}

var _ gpa.Message = &msgCmtLog{}

func NewMsgCmtLog(committeeID CommitteeID, wrapped gpa.Message) gpa.Message {
	return &msgCmtLog{
		committeeID: committeeID,
		wrapped:     wrapped,
	}
}

func (msg *msgCmtLog) Recipient() gpa.NodeID {
	return msg.wrapped.Recipient()
}

func (msg *msgCmtLog) SetSender(sender gpa.NodeID) {
	msg.wrapped.SetSender(sender)
}

func (msg *msgCmtLog) MarshalBinary() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	if err := util.WriteString16(w, msg.committeeID); err != nil {
		return nil, err
	}
	bin, err := msg.wrapped.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if err := util.WriteBytes16(w, bin); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (msg *msgCmtLog) UnmarshalBinary(data []byte, cmtLogs map[CommitteeID]gpa.GPA) error {
	var err error
	r := bytes.NewReader(data)
	msg.committeeID, err = util.ReadString16(r)
	if err != nil {
		return err
	}
	cmtLog, ok := cmtLogs[msg.committeeID]
	if !ok {
		return xerrors.Errorf("cannot find cmtLog: %v", msg.committeeID)
	}
	wrappedMsgData, err := util.ReadBytes16(r)
	if err != nil {
		return err
	}
	msg.wrapped, err = cmtLog.UnmarshalMessage(wrappedMsgData)
	if err != nil {
		return err
	}
	return nil
}
