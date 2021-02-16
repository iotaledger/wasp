// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainimpl

import (
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/prometheus/common/log"
)

func (c *chainObj) testTrace(msg *chain.TestTraceMsg) {
	log.Debugf("++++ received TestTraceMsg from #%d", msg.SenderIndex)

	if len(msg.Sequence) != int(c.Size()) || !util.ValidPermutation(msg.Sequence) {
		c.log.Panicf("wrong permutation %+v received from #%d", msg.Sequence, msg.SenderIndex)
	}

	msg.NumHops++
	if msg.InitPeerIndex == c.ownIndex {
		whenSent := time.Unix(0, msg.InitTime)
		c.log.Infof("TEST PASSED with %d hops in %v", msg.NumHops, time.Since(whenSent))
		return
	}
	seqIndex := c.mustFindIndexOf(c.ownIndex, msg.Sequence)
	targetSeqIndex := (seqIndex + 1) % c.Size()

	sentToSeqIdx, err := c.SendMsgInSequence(chain.MsgTestTrace, util.MustBytes(msg), targetSeqIndex, msg.Sequence)
	if err != nil {
		c.log.Errorf("testTrace: %v", err)
	} else {
		if targetSeqIndex != sentToSeqIdx {
			c.log.Infof("testTrace: #%d -> #%d -> #%d %+v",
				c.ownIndex, msg.Sequence[targetSeqIndex], msg.Sequence[sentToSeqIdx], msg.Sequence)
		}
	}
}

func (c *chainObj) InitTestRound() {
	seed := hashing.RandomHash(nil)
	msg := &chain.TestTraceMsg{
		InitTime:      time.Now().UnixNano(),
		InitPeerIndex: c.ownIndex,
		Sequence:      util.NewPermutation16(c.NumPeers(), seed[:]).GetArray(),
	}
	// found own seqIndex in permutation
	seqIndex := c.mustFindIndexOf(c.ownIndex, msg.Sequence)
	targetSeqIndex := (seqIndex + 1) % c.Size()

	sentToSeqIdx, err := c.SendMsgInSequence(chain.MsgTestTrace, util.MustBytes(msg), targetSeqIndex, msg.Sequence)
	if err != nil {
		c.log.Errorf("TEST FAILED: initial send returned an error: %v", err)
	} else {
		c.log.Infof("InitTestRound started: #%d -> #%d -> #%d %+v",
			c.ownIndex, msg.Sequence[targetSeqIndex], msg.Sequence[sentToSeqIdx], msg.Sequence)
	}
}

func (c *chainObj) mustFindIndexOf(val uint16, sequence []uint16) uint16 {
	for i, s := range sequence {
		if s == val {
			return uint16(i)
		}
	}
	c.log.Panicf("mustFindIndexOf: search for %d: wrong value or permutation %+v", val, sequence)
	return 0
}
