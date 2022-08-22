// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package journal_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

func TestConsensusJournal(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	chainID := *isc.RandomChainID()
	committeeAddress := isc.RandomChainID().AsAddress()

	registry := testchain.NewMockedConsensusJournalRegistry()
	j, err := journal.LoadConsensusJournal(chainID, committeeAddress, registry, 1, 0, log)
	require.NoError(t, err)
	lv := j.GetLocalView()
	require.NotNil(t, lv)
	require.Equal(t, j.GetLogIndex(), journal.LogIndex(0))
	j.ConsensusReached(j.GetLogIndex())
	require.Equal(t, j.GetLogIndex(), journal.LogIndex(1))
}

func TestConsensusJournalPeers(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	chainID := *isc.RandomChainID()
	committeeAddress := isc.RandomChainID().AsAddress()

	registry := testchain.NewMockedConsensusJournalRegistry()
	j, err := journal.LoadConsensusJournal(chainID, committeeAddress, registry, 4, 1, log)
	require.NoError(t, err)
	require.Equal(t, j.GetLogIndex(), journal.LogIndex(0))
	require.False(t, j.PeerLogIndexReceived(1, 7))
	require.True(t, j.PeerLogIndexReceived(2, 13))
	require.Equal(t, j.GetLogIndex(), journal.LogIndex(7))
	require.True(t, j.PeerLogIndexReceived(1, 13))
	require.Equal(t, j.GetLogIndex(), journal.LogIndex(13))
	require.False(t, j.PeerLogIndexReceived(3, 10))
	require.Equal(t, j.GetLogIndex(), journal.LogIndex(13))
}
