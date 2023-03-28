package chainMgr

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util"
)

// gpa.Wrapper is not applicable here, because here the addressing
// is by CommitteeID, not by integer index.
type msgCmtLog struct {
	committeeAddr iotago.Ed25519Address
	wrapped       gpa.Message
}

var _ gpa.Message = &msgCmtLog{}

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

func (msg *msgCmtLog) MarshalBinary() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	committeeAddrBytes, err := msg.committeeAddr.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}
	if err2 := util.WriteBytes16(w, committeeAddrBytes); err2 != nil {
		return nil, err2
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

func (msg *msgCmtLog) UnmarshalBinary(data []byte) error {
	var err error
	r := bytes.NewReader(data)
	committeeAddrBytes, err := util.ReadBytes16(r)
	if err != nil {
		return err
	}
	_, err = msg.committeeAddr.Deserialize(committeeAddrBytes, serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return err
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
