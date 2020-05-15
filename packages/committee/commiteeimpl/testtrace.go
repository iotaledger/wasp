package commiteeimpl

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/prometheus/common/log"
	"time"
)

func (c *committeeObj) testTrace(msg *committee.TestTraceMsg) {
	log.Debug("++++ received TestTraceMsg from #%d", msg.SenderIndex)

	if msg.InitPeerIndex == c.ownIndex {
		whenSent := time.Unix(0, msg.InitTime)
		c.log.Infof("+++ testing round trip %v initiator #%d", time.Since(whenSent), c.ownIndex)
		return
	}
	seqIndex := findIndexOf(c.ownIndex, msg.Sequence)
	targetSeqIndex := msg.Sequence[(seqIndex+1)%c.Size()]

	sentToSeqIdx, err := c.SendMsgInSequence(committee.MsgTestTrace, hashing.MustBytes(msg), targetSeqIndex, msg.Sequence)
	if err != nil {
		c.log.Errorf("testTrace: %v", err)
	} else {
		if targetSeqIndex != sentToSeqIdx {
			c.log.Infof("testTrace: #%d -> #%d -> #%d %+v",
				c.ownIndex, msg.Sequence[targetSeqIndex], msg.Sequence[sentToSeqIdx], msg.Sequence)
		}
	}
}

func (c *committeeObj) InitTestRound() {
	var b [2]byte

	msg := &committee.TestTraceMsg{
		InitTime:      time.Now().UnixNano(),
		InitPeerIndex: c.ownIndex,
		Sequence:      util.GetPermutation(c.Size(), b[:]),
	}
	// found own seqIndex in permutation
	seqIndex := findIndexOf(c.ownIndex, msg.Sequence)
	targetSeqIndex := msg.Sequence[(seqIndex+1)%c.Size()]

	sentToSeqIdx, err := c.SendMsgInSequence(committee.MsgTestTrace, hashing.MustBytes(msg), targetSeqIndex, msg.Sequence)
	if err != nil {
		c.log.Errorf("InitTestRound: %v", err)
	} else {
		c.log.Infof("InitTestRound started: #%d -> #%d -> #%d %+v", c.ownIndex, targetSeqIndex, sentToSeqIdx, msg.Sequence)
	}
}

func findIndexOf(val uint16, sequence []uint16) uint16 {
	for i, s := range sequence {
		if s == val {
			return uint16(i)
		}
	}
	panic("wrong permutation")
}
