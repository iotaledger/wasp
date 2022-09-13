// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Implementation for the journal.Registry interface.

package registry

import (
	"encoding/binary"
	"errors"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"golang.org/x/xerrors"
)

const (
	dbKeyForConsensusJournalLogIndex byte = iota
	dbKeyForConsensusJournalLocalView
)

func (r *Impl) LoadConsensusJournal(id journal.ID) (journal.LogIndex, journal.LocalView, error) {
	liBytes, err := r.store.Get(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLogIndex))
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return 0, nil, journal.ErrConsensusJournalNotFound
		}
		return 0, nil, err
	}
	li := journal.LogIndex(binary.BigEndian.Uint32(liBytes))

	lvBytes, err := r.store.Get(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLocalView))
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return 0, nil, journal.ErrConsensusJournalNotFound
		}
		return 0, nil, err
	}
	lv, err := journal.NewLocalViewFromBytes(lvBytes)
	if err != nil {
		return 0, nil, xerrors.Errorf("cannot deserialize LocalView")
	}

	return li, lv, nil
}

func (r *Impl) SaveConsensusJournalLogIndex(id journal.ID, logIndex journal.LogIndex) error {
	liBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(liBytes, logIndex.AsUint32())

	err := r.store.Set(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLogIndex), liBytes)
	if err != nil {
		return err
	}
	return nil
}

func (r *Impl) SaveConsensusJournalLocalView(id journal.ID, localView journal.LocalView) error {
	lvBytes, err := localView.AsBytes()
	if err != nil {
		return xerrors.Errorf("cannot serialize localView: %w", err)
	}
	if err := r.store.Set(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLocalView), lvBytes); err != nil {
		return err
	}
	return nil
}

func dbKeyForConsensusJournal(id journal.ID, subKey byte) []byte {
	return dbkeys.MakeKey(dbkeys.ObjectTypeConsensusJournal, []byte{subKey}, id[:])
}
