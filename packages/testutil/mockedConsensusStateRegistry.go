package testutil

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/isc"
)

type mockedConsensusStateRegistry struct {
	data map[string]*cmtLog.State
}

func (s *mockedConsensusStateRegistry) MarshalJSON() ([]byte, error) {
	panic("not used in tests")
}

func (s *mockedConsensusStateRegistry) UnmarshalJSON(bytes []byte) error {
	panic("not used in tests")
}

var _ cmtLog.ConsensusStateRegistry = &mockedConsensusStateRegistry{}

func NewConsensusStateRegistry() cmtLog.ConsensusStateRegistry {
	return &mockedConsensusStateRegistry{data: map[string]*cmtLog.State{}}
}

func (s *mockedConsensusStateRegistry) Get(chainID isc.ChainID, cmtAddr iotago.Address) (*cmtLog.State, error) {
	if store, ok := s.data[cmtAddr.Key()]; ok {
		return store, nil
	}
	return nil, cmtLog.ErrCmtLogStateNotFound
}

func (s *mockedConsensusStateRegistry) Set(chainID isc.ChainID, cmtAddr iotago.Address, state *cmtLog.State) error {
	s.data[cmtAddr.Key()] = state
	return nil
}
