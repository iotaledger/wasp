package nodeconn

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/tangle"
	"github.com/iotaledger/goshimmer/packages/waspconn"
)

func (n *NodeConn) msgDataToEvent(data []byte) {
	msg, err := waspconn.DecodeMsg(data, true)
	if err != nil {
		n.log.Errorf("wrong message from node: %v", err)
		return
	}

	switch msg := msg.(type) {
	case *waspconn.WaspMsgChunk:
		finalData, err := n.msgChopper.IncomingChunk(msg.Data, tangle.MaxMessageSize, waspconn.ChunkMessageHeaderSize)
		if err != nil {
			n.log.Errorf("receiving message chunk: %v", err)
			return
		}
		if finalData != nil {
			n.msgDataToEvent(finalData)
		}

	case *waspconn.WaspPingMsg:
		roundtrip := time.Since(msg.Timestamp)
		n.log.Infof("PING %d response from node. Roundtrip %v", msg.Id, roundtrip)

	default:
		n.EventMessageReceived.Trigger(msg)
	}
}
