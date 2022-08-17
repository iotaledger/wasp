// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testchain

import (
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
)

//
// Mock for the journal.RegistryProvider.
//
type mockedConsensusJournalRegistryImpl struct {
	li map[journal.ID]journal.LogIndex
	lv map[journal.ID][]byte
}

var _ journal.Registry = &mockedConsensusJournalRegistryImpl{}

func NewMockedConsensusJournalRegistry() journal.Registry {
	return &mockedConsensusJournalRegistryImpl{
		li: map[journal.ID]journal.LogIndex{},
		lv: map[journal.ID][]byte{},
	}
}

func (m *mockedConsensusJournalRegistryImpl) SaveConsensusJournalLogIndex(id journal.ID, logIndex journal.LogIndex) error {
	m.li[id] = logIndex
	return nil
}

func (m *mockedConsensusJournalRegistryImpl) SaveConsensusJournalLocalView(id journal.ID, localView journal.LocalView) error {
	var err error
	m.lv[id], err = localView.AsBytes()
	return err
}

func (m *mockedConsensusJournalRegistryImpl) LoadConsensusJournal(id journal.ID) (journal.LogIndex, journal.LocalView, error) {
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
