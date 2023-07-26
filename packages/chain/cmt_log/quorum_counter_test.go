package cmt_log_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestQuorumCounter(t *testing.T) {
	log := testlogger.NewLogger(t)
	n := 7
	f := 2
	nodeIDs := gpa.MakeTestNodeIDs(n)
	ao1 := isc.RandomAliasOutputWithID()
	ao2 := isc.RandomAliasOutputWithID()
	lin := cmt_log.NilLogIndex()
	li7 := cmt_log.LogIndex(7)
	li8 := cmt_log.LogIndex(8)

	qc := cmt_log.NewQuorumCounter(cmt_log.MsgNextLogIndexCauseRecover, nodeIDs, log)
	var li cmt_log.LogIndex
	var ao *isc.AliasOutputWithID

	li, ao = qc.EnoughVotes(f+1, true)
	require.Equal(t, lin, li)
	require.Nil(t, ao)
	li, ao = qc.EnoughVotes(f+1, false)
	require.Equal(t, lin, li)
	require.Nil(t, ao)

	makeVote := func(from gpa.NodeID, li cmt_log.LogIndex, ao *isc.AliasOutputWithID) *cmt_log.MsgNextLogIndex {
		vote := cmt_log.NewMsgNextLogIndex(nodeIDs[0], li, ao, cmt_log.MsgNextLogIndexCauseRecover, false)
		vote.SetSender(from)
		return vote
	}

	qc.VoteReceived(makeVote(nodeIDs[0], li7, ao1))
	qc.VoteReceived(makeVote(nodeIDs[1], li7, ao1))
	qc.VoteReceived(makeVote(nodeIDs[2], li8, ao1))
	qc.VoteReceived(makeVote(nodeIDs[3], li8, ao2))
	qc.VoteReceived(makeVote(nodeIDs[4], li8, ao2))

	li, ao = qc.EnoughVotes(f+1, false)
	require.Equal(t, li8, li)
	require.Nil(t, ao) // Don't return AO i the case of false.

	li, ao = qc.EnoughVotes(f+1, true)
	require.Equal(t, li7, li)

	li, ao = qc.EnoughVotes(n-f, false)
	require.Equal(t, li7, li)
	require.Nil(t, ao) // Don't return AO i the case of false.

	li, ao = qc.EnoughVotes(n-f, true)
	require.Equal(t, lin, li)

	require.True(t, qc.HaveVoteFrom(nodeIDs[4]))
	require.False(t, qc.HaveVoteFrom(nodeIDs[5]))
}
