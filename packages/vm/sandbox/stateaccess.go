package sandbox

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/variables"
)

func (s *stateWrapper) Get(name string) ([]byte, bool) {
	// FIXME: this is O(N) with N = amount of accumulated mutations
	// it could be improved by caching the latest mutation for evey key
	muts := s.stateUpdate.Mutations()
	for i := muts.Len() - 1; i >= 0; i-- {
		m := muts.At(i)
		if (*m).Key() == name {
			// The key-value pair has been modified during the current request
			// return the latest assigned value
			return (*m).Value()
		}
	}

	// The key-value pair has not been modified
	// Fetch its value from the virtual state
	return s.virtualState.Variables().Get(name)
}

func (s *stateWrapper) GetInt64(name string) (int64, bool, error) {
	v, ok := s.Get(name)
	if !ok {
		return 0, false, nil
	}
	if len(v) != 8 {
		return 0, false, fmt.Errorf("variable %s: %v is not an int64", name, v)
	}
	return int64(util.Uint64From8Bytes(v)), true, nil
}

func (s *stateWrapper) Del(name string) {
	s.stateUpdate.Mutations().Add(variables.NewMutationDel(name))
}

func (s *stateWrapper) Set(name string, value []byte) {
	s.stateUpdate.Mutations().Add(variables.NewMutationSet(name, value))
}

func (s *stateWrapper) SetInt64(name string, value int64) {
	s.stateUpdate.Mutations().Add(variables.NewMutationSet(name, util.Uint64To8Bytes(uint64(value))))
}

func (s *stateWrapper) SetString(name string, value string) {
	s.stateUpdate.Mutations().Add(variables.NewMutationSet(name, []byte(value)))
}

func (s *stateWrapper) SetAddressValue(name string, addr address.Address) {
	s.stateUpdate.Mutations().Add(variables.NewMutationSet(name, addr[:]))
}

func (s *stateWrapper) SetHashValue(name string, h *hashing.HashValue) {
	s.stateUpdate.Mutations().Add(variables.NewMutationSet(name, h[:]))
}
