// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package journal_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

func TestConsensusJournal(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Sync()

	chainID := *isc.RandomChainID()
	committeeAddress := isc.RandomChainID().AsAddress()

	registry := newMockedRegistryImpl()
	j, err := journal.LoadConsensusJournal(chainID, committeeAddress, registry, log)
	require.NoError(t, err)
	lv := j.GetLocalView()
	require.NotNil(t, lv)
	require.Equal(t, j.GetLogIndex(), journal.LogIndex(0))
	j.ConsensusReached(j.GetLogIndex())
	require.Equal(t, j.GetLogIndex(), journal.LogIndex(1))
}

//
// Mock for the journal.RegistryProvider.
//
type mockedRegistryImpl struct {
	li map[journal.ID]journal.LogIndex
	lv map[journal.ID][]byte
}

func newMockedRegistryImpl() *mockedRegistryImpl {
	return &mockedRegistryImpl{
		li: map[journal.ID]journal.LogIndex{},
		lv: map[journal.ID][]byte{},
	}
}

func (m *mockedRegistryImpl) SaveConsensusJournalLogIndex(id journal.ID, logIndex journal.LogIndex) error {
	m.li[id] = logIndex
	return nil
}

func (m *mockedRegistryImpl) SaveConsensusJournalLocalView(id journal.ID, localView journal.LocalView) error {
	var err error
	m.lv[id], err = localView.AsBytes()
	return err
}

func (m *mockedRegistryImpl) LoadConsensusJournal(id journal.ID) (journal.LogIndex, journal.LocalView, error) {
	lvBytes, found := m.lv[id]
	if !found {
		return 0, nil, journal.ErrConsensusJournalNotFound
	}
	lv, err := journal.NewLocalViewFromBytes(lvBytes)
	if err != nil {
		return 0, nil, err
	}
	li, found := m.li[id]
	if !found {
		return 0, nil, journal.ErrConsensusJournalNotFound
	}
	return li, lv, nil
}
