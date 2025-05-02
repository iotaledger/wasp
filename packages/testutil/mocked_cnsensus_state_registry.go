package testutil

import (
	"github.com/iotaledger/wasp/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

type mockedConsensusStateRegistry struct {
	data map[cryptolib.AddressKey]*cmtlog.State
}

func (s *mockedConsensusStateRegistry) MarshalJSON() ([]byte, error) {
	panic("not used in tests")
}

func (s *mockedConsensusStateRegistry) UnmarshalJSON(bytes []byte) error {
	panic("not used in tests")
}

var _ cmtlog.ConsensusStateRegistry = &mockedConsensusStateRegistry{}

func NewConsensusStateRegistry() cmtlog.ConsensusStateRegistry {
	return &mockedConsensusStateRegistry{data: map[cryptolib.AddressKey]*cmtlog.State{}}
}

func (s *mockedConsensusStateRegistry) Get(chainID isc.ChainID, cmtAddr *cryptolib.Address) (*cmtlog.State, error) {
	if store, ok := s.data[cmtAddr.Key()]; ok {
		return store, nil
	}
	return nil, cmtlog.ErrCmtLogStateNotFound
}

func (s *mockedConsensusStateRegistry) Set(chainID isc.ChainID, cmtAddr *cryptolib.Address, state *cmtlog.State) error {
	s.data[cmtAddr.Key()] = state
	return nil
}
