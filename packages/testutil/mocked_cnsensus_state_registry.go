package testutil

import (
	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type mockedConsensusStateRegistry struct {
	data map[cryptolib.AddressKey]*committeelog.State
}

func (s *mockedConsensusStateRegistry) MarshalJSON() ([]byte, error) {
	panic("not used in tests")
}

func (s *mockedConsensusStateRegistry) UnmarshalJSON(bytes []byte) error {
	panic("not used in tests")
}

var _ committeelog.ConsensusStateRegistry = &mockedConsensusStateRegistry{}

func NewConsensusStateRegistry() committeelog.ConsensusStateRegistry {
	return &mockedConsensusStateRegistry{data: map[cryptolib.AddressKey]*committeelog.State{}}
}

func (s *mockedConsensusStateRegistry) Get(chainID isc.ChainID, cmtAddr *cryptolib.Address) (*committeelog.State, error) {
	if store, ok := s.data[cmtAddr.Key()]; ok {
		return store, nil
	}
	return nil, committeelog.ErrCommitteeLogStateNotFound
}

func (s *mockedConsensusStateRegistry) Set(chainID isc.ChainID, cmtAddr *cryptolib.Address, state *committeelog.State) error {
	s.data[cmtAddr.Key()] = state
	return nil
}
