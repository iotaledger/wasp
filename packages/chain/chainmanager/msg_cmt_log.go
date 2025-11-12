package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type msgNextLogIndex struct {
	CommitteeAddr      cryptolib.Address
	MsgForCommitteeLog committeelog.MsgNextLogIndex
}

func NewMsgNextLogIndex(committeeAddr cryptolib.Address, msg committeelog.MsgNextLogIndex) *msgNextLogIndex {
	return &msgNextLogIndex{
		CommitteeAddr:      committeeAddr,
		MsgForCommitteeLog: msg,
	}
}

func (msg *msgNextLogIndex) MsgType() gpa.MessageType {
	return msgTypeMsgNextLogIndex
}

func (msg *msgNextLogIndex) String() string {
	return fmt.Sprintf("{chainMgr.msgNextLogIndex, committeeAddr=%v, MsgNextLogIndex=%+v}", msg.CommitteeAddr.String(), msg.MsgForCommitteeLog)
}
