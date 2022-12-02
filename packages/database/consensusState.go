package database

import (
	"errors"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/common"
)

const (
	dbKeyPrefixForConsensusStateCmtLog byte = iota
)

type ConsensusState struct {
	store kvstore.KVStore
}

var _ cmtLog.Store = &ConsensusState{}

func NewConsensusState(store kvstore.KVStore) *ConsensusState {
	return &ConsensusState{
		store: store,
	}
}

func (cs *ConsensusState) LoadCmtLogState(committeeAddress iotago.Address) (*cmtLog.State, error) { // Can return cmtLog.ErrCmtLogStateNotFound.
	stBytes, err := cs.store.Get(dbKeyForConsensusState(dbKeyPrefixForConsensusStateCmtLog, committeeAddress))
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			return nil, cmtLog.ErrCmtLogStateNotFound
		}
		return nil, err
	}
	st := &cmtLog.State{}
	if err := st.UnmarshalBinary(stBytes); err != nil {
		return nil, err
	}
	return st, nil
}

func (cs *ConsensusState) SaveCmtLogState(committeeAddress iotago.Address, state *cmtLog.State) error {
	stBytes, err := state.MarshalBinary()
	if err != nil {
		return err
	}
	return cs.store.Set(dbKeyForConsensusState(dbKeyPrefixForConsensusStateCmtLog, committeeAddress), stBytes)
}

func dbKeyForConsensusState(prefix byte, committeeAddress iotago.Address) []byte {
	addrBytes, err := committeeAddress.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic(xerrors.Errorf("cannot serialize committeeAddress: %w", err))
	}
	return common.MakeKey(prefix, addrBytes)
}
