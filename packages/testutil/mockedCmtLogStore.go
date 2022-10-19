package testutil

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
)

type mockedCmtLogStore struct {
	data map[string]*cmtLog.State
}

var _ cmtLog.Store = &mockedCmtLogStore{}

func NewMockedCmtLogStore() cmtLog.Store {
	return &mockedCmtLogStore{data: map[string]*cmtLog.State{}}
}

func (s *mockedCmtLogStore) LoadCmtLogState(cmtAddr iotago.Address) (*cmtLog.State, error) {
	if store, ok := s.data[cmtAddr.Key()]; ok {
		return store, nil
	}
	return nil, cmtLog.ErrCmtLogStateNotFound
}

func (s *mockedCmtLogStore) SaveCmtLogState(cmtAddr iotago.Address, state *cmtLog.State) error {
	s.data[cmtAddr.Key()] = state
	return nil
}
