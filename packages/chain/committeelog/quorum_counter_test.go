package committeelog_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

func TestQuorumCounter(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 7
	f := 2
	nodeIDs := gpa.MakeTestNodeIDs(n)
	lin := committeelog.NilLogIndex()
	li7 := committeelog.LogIndex(7)
	li8 := committeelog.LogIndex(8)

	qc := committeelog.NewQuorumCounter(committeelog.MsgNextLogIndexCauseStarted, nodeIDs, log)

	require.Equal(t, lin, qc.EnoughVotes(f+1))

	makeVote := func(from gpa.NodeID, li committeelog.LogIndex) *gpa.TypedMessageIn[*committeelog.MsgNextLogIndex] {
		vote := committeelog.NewMsgNextLogIndex(li, committeelog.MsgNextLogIndexCauseStarted, false)
		return &gpa.TypedMessageIn[*committeelog.MsgNextLogIndex]{
			Sender:  from,
			Payload: vote,
		}
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
