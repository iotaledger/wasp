// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package journal

import (
	"encoding/binary"
	"errors"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

////////////////////////////////////////////////////////////////////////////////

// ConsensusJournal instances are per ChainID ⨉ CommitteeAddress.
// This ID represents that.
type ID [iotago.Ed25519AddressBytesLength]byte

func MakeID(chainID isc.ChainID, committeeAddress iotago.Address) (*ID, error) {
	var id ID
	addressBytes, err := committeeAddress.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, xerrors.Errorf("cannot serialize address: %v", err)
	}
	for i := range chainID {
		id[i] = addressBytes[i] ^ chainID[i]
	}
	return &id, nil
}

////////////////////////////////////////////////////////////////////////////////

type LogIndex uint32

func (li *LogIndex) AsUint32() uint32 {
	return uint32(*li)
}

// For the ACS runner mostly. Can be removed after moving stuff to GPA.
func (li *LogIndex) AsUint64Key(id ID) uint64 {
	liByes := make([]byte, 4)
	binary.BigEndian.PutUint32(liByes, li.AsUint32())
	return util.MustUint64From8Bytes(hashing.HashData(id[:], liByes).Bytes()[:8])
}

// For the Nonce Instance mostly. Can be removed after moving stuff to GPA.
func (li *LogIndex) AsStringKey(id ID) string {
	liByes := make([]byte, 4)
	binary.BigEndian.PutUint32(liByes, li.AsUint32())
	return hashing.HashData(id[:], liByes).String()
}

////////////////////////////////////////////////////////////////////////////////

// ConsensusJournal tracks the consensus instances (the consensus log).
// Its main purpose is to record, which instances have already been done.
type ConsensusJournal interface {
	// GetID just returns the ID of the current Consensus Journal.
	GetID() ID
	//
	// GetLogIndex returns the next non-decided log index for this journal.
	GetLogIndex() LogIndex
	//
	// Returns a persistence-supported version of the local view.
	GetLocalView() LocalView
	//
	// Notify, when a particular log index has been completed (decided).
	ConsensusReached(logIndex LogIndex)
}

////////////////////////////////////////////////////////////////////////////////

var ErrConsensusJournalNotFound = errors.New("ErrConsensusJournalNotFound")

// RegistryProvider has to be provided for the ConsensusJournal and should
// implement the persistent store.
type Registry interface {
	LoadConsensusJournal(id ID) (LogIndex, LocalView, error) // Can return ErrConsensusJournalNotFound
	SaveConsensusJournalLogIndex(id ID, logIndex LogIndex) error
	SaveConsensusJournalLocalView(id ID, localView LocalView) error
}

////////////////////////////////////////////////////////////////////////////////

// consensusJournalImpl implements ConsensusJournal and LocalView.
// Here the local view is made persistent and a dimension of history added.
//
type consensusJournalImpl struct {
	id        ID
	chainID   isc.ChainID
	committee iotago.Address
	registry  Registry
	logIndex  LogIndex
	localView LocalView
	log       *logger.Logger
}

var (
	_ ConsensusJournal = &consensusJournalImpl{}
	_ LocalView        = &consensusJournalImpl{}
)

func LoadConsensusJournal(chainID isc.ChainID, committee iotago.Address, registry Registry, log *logger.Logger) (ConsensusJournal, error) {
	id, err := MakeID(chainID, committee)
	if err != nil {
		return nil, err
	}
	j := &consensusJournalImpl{
		id:       *id,
		chainID:  chainID,
		registry: registry,
		log:      log,
	}
	li, lv, err := registry.LoadConsensusJournal(j.id)
	if err == nil {
		j.logIndex = li
		j.localView = lv
		return j, nil
	}
	if errors.Is(err, ErrConsensusJournalNotFound) {
		j.localView = NewLocalView()
		j.logIndex = 0
		if err := registry.SaveConsensusJournalLogIndex(j.id, j.logIndex); err != nil {
			return nil, xerrors.Errorf("cannot save consensus journal: %w", err)
		}
		if err := registry.SaveConsensusJournalLocalView(j.id, j.localView); err != nil {
			return nil, xerrors.Errorf("cannot save consensus journal: %w", err)
		}
		return j, nil
	}
	return nil, xerrors.Errorf("cannot load consensus journal: %w", err)
}

// Implements the ConsensusJournal interface.
func (j *consensusJournalImpl) GetID() ID {
	return j.id
}

// Implements the ConsensusJournal interface.
func (j *consensusJournalImpl) GetLogIndex() LogIndex {
	return j.logIndex
}

// Implements the ConsensusJournal interface.
func (j *consensusJournalImpl) GetLocalView() LocalView {
	return j // Return this object, it acts as a wrapper.
}

// Implements the ConsensusJournal interface.
func (j *consensusJournalImpl) ConsensusReached(logIndex LogIndex) {
	if j.logIndex != logIndex {
		j.log.Warnf("Consensus reached for logIndex=%v but we have current logIndex=%v, ignoring", logIndex, j.logIndex)
		return
	}
	j.logIndex++
	if err := j.registry.SaveConsensusJournalLogIndex(j.id, j.logIndex); err != nil {
		panic(xerrors.Errorf("cannot store the log index: %w", err))
	}
}

// Implements the LocalView interface.
func (j *consensusJournalImpl) GetBaseAliasOutputID() *iotago.OutputID {
	return j.localView.GetBaseAliasOutputID()
}

// Implements the LocalView interface.
func (j *consensusJournalImpl) AliasOutputReceived(confirmed *isc.AliasOutputWithID) {
	j.localView.AliasOutputReceived(confirmed)
	if err := j.registry.SaveConsensusJournalLocalView(j.id, j.localView); err != nil {
		panic(xerrors.Errorf("cannot persist local view after AliasOutputReceived: %w", err))
	}
}

// Implements the LocalView interface.
func (j *consensusJournalImpl) AliasOutputRejected(rejected *isc.AliasOutputWithID) {
	j.localView.AliasOutputRejected(rejected)
	if err := j.registry.SaveConsensusJournalLocalView(j.id, j.localView); err != nil {
		panic(xerrors.Errorf("cannot persist local view after AliasOutputRejected: %w", err))
	}
}

// Implements the LocalView interface.
func (j *consensusJournalImpl) AliasOutputPublished(consumed, published *isc.AliasOutputWithID) {
	j.localView.AliasOutputPublished(consumed, published)
	if err := j.registry.SaveConsensusJournalLocalView(j.id, j.localView); err != nil {
		panic(xerrors.Errorf("cannot persist local view after AliasOutputPublished: %w", err))
	}
}

// Implements the LocalView interface.
func (j *consensusJournalImpl) AsBytes() ([]byte, error) {
	return j.localView.AsBytes()
}
