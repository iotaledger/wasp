package testutil

import (
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

type mockedConsensusStateRegistry struct {
	data map[cryptolib.AddressKey]*cmt_log.State
}

func (s *mockedConsensusStateRegistry) MarshalJSON() ([]byte, error) {
	panic("not used in tests")
}

func (s *mockedConsensusStateRegistry) UnmarshalJSON(bytes []byte) error {
	panic("not used in tests")
}

var _ cmt_log.ConsensusStateRegistry = &mockedConsensusStateRegistry{}

func NewConsensusStateRegistry() cmt_log.ConsensusStateRegistry {
	return &mockedConsensusStateRegistry{data: map[cryptolib.AddressKey]*cmt_log.State{}}
}

func (s *mockedConsensusStateRegistry) Get(chainID isc.ChainID, cmtAddr *cryptolib.Address) (*cmt_log.State, error) {
	if store, ok := s.data[cmtAddr.Key()]; ok {
		return store, nil
	}
	return nil, cmt_log.ErrCmtLogStateNotFound
}

func (s *mockedConsensusStateRegistry) Set(chainID isc.ChainID, cmtAddr *cryptolib.Address, state *cmt_log.State) error {
	s.data[cmtAddr.Key()] = state
	return nil
}
