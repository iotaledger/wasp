package commiteeimpl

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"time"
)

func (c *committeeObj) testTrace(msg *committee.TestTraceMsg) {
	if msg.Trace[0] == c.ownIndex {
		whenSent := time.Unix(0, msg.InitTime)
		c.log.Info("+++ testing round trip %v initiator #%d", time.Since(whenSent), c.ownIndex)
		return
	}
	msg.Trace = append(msg.Trace, c.ownIndex)
	target := (c.ownIndex + 1) % c.Size()
	err := c.SendMsg(target, committee.MsgTestTrace, hashing.MustBytes(msg))
	if err != nil {
		c.log.Errorf("testTrace::SendMsg: %v", err)
	}
}

func (c *committeeObj) InitTestRound() {
	msg := &committee.TestTraceMsg{
		InitTime: time.Now().UnixNano(),
		Trace:    []uint16{c.ownIndex},
	}
	target := (c.ownIndex + 1) % c.Size()
	err := c.SendMsg(target, committee.MsgTestTrace, hashing.MustBytes(msg))
	if err != nil {
		c.log.Errorf("InitTestRound::SendMsg: %v", err)
	} else {
		c.log.Infof("InitTestRound started")
	}
}
