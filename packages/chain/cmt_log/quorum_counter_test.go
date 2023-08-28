package cmt_log_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestQuorumCounter(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 7
	f := 2
	nodeIDs := gpa.MakeTestNodeIDs(n)
	lin := cmt_log.NilLogIndex()
	li7 := cmt_log.LogIndex(7)
	li8 := cmt_log.LogIndex(8)

	qc := cmt_log.NewQuorumCounter(cmt_log.MsgNextLogIndexCauseRecover, nodeIDs, log)

	require.Equal(t, lin, qc.EnoughVotes(f+1))

	makeVote := func(from gpa.NodeID, li cmt_log.LogIndex) *cmt_log.MsgNextLogIndex {
		vote := cmt_log.NewMsgNextLogIndex(nodeIDs[0], li, cmt_log.MsgNextLogIndexCauseRecover, false)
		vote.SetSender(from)
		return vote
	}

	qc.VoteReceived(makeVote(nodeIDs[0], li7))
	qc.VoteReceived(makeVote(nodeIDs[1], li7))
	qc.VoteReceived(makeVote(nodeIDs[2], li8))
	qc.VoteReceived(makeVote(nodeIDs[3], li8))
	qc.VoteReceived(makeVote(nodeIDs[4], li8))

	require.Equal(t, li8, qc.EnoughVotes(f+1))
	require.Equal(t, li7, qc.EnoughVotes(n-f))

	require.True(t, qc.HaveVoteFrom(nodeIDs[4]))
	require.False(t, qc.HaveVoteFrom(nodeIDs[5]))
}
