package journal

import (
	"encoding/binary"
	"errors"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/common"
)

const (
	dbKeyForConsensusJournalLogIndex byte = iota
	dbKeyForConsensusJournalLocalView
)

type ConsensusJournal struct {
	store kvstore.KVStore
}

func NewConsensusJournal(store kvstore.KVStore) *ConsensusJournal {
	return &ConsensusJournal{
		store: store,
	}
}

func dbKeyForConsensusJournal(id journal.ID, subKey byte) []byte {
	return common.MakeKey(subKey, id[:])
}

func (j *ConsensusJournal) LoadConsensusJournal(id journal.ID) (journal.LogIndex, journal.LocalView, error) {
	liBytes, err := j.store.Get(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLogIndex))
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return 0, nil, journal.ErrConsensusJournalNotFound
		}
		return 0, nil, err
	}
	li := journal.LogIndex(binary.BigEndian.Uint32(liBytes))

	lvBytes, err := j.store.Get(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLocalView))
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

func (j *ConsensusJournal) SaveConsensusJournalLogIndex(id journal.ID, logIndex journal.LogIndex) error {
	liBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(liBytes, logIndex.AsUint32())

	err := j.store.Set(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLogIndex), liBytes)
	if err != nil {
		return err
	}
	return nil
}

func (j *ConsensusJournal) SaveConsensusJournalLocalView(id journal.ID, localView journal.LocalView) error {
	lvBytes, err := localView.AsBytes()
	if err != nil {
		return xerrors.Errorf("cannot serialize localView: %w", err)
	}
	if err := j.store.Set(dbKeyForConsensusJournal(id, dbKeyForConsensusJournalLocalView), lvBytes); err != nil {
		return err
	}
	return nil
}
