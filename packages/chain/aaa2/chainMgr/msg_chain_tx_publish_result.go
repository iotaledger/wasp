package chainMgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
)

// That's a local message.
type msgChainTxPublishResult struct {
	gpa.BasicMessage
	committeeID CommitteeID
	txID        iotago.TransactionID
	confirmed   bool
}

var _ gpa.Message = &msgChainTxPublishResult{}

func NewMsgChainTxPublishResult(recipient gpa.NodeID, committeeID CommitteeID, txID iotago.TransactionID, confirmed bool) gpa.Message {
	return &msgChainTxPublishResult{
		BasicMessage: gpa.NewBasicMessage(recipient),
		committeeID:  committeeID,
		txID:         txID,
		confirmed:    confirmed,
	}
}

func (m *msgChainTxPublishResult) MarshalBinary() ([]byte, error) {
	panic("trying to marshal local message")
}
