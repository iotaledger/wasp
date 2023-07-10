package evmimpl

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
)

type StateAccess struct {
	state kv.KVStoreReader
}

func NewStateAccess(store kv.KVStoreReader) *StateAccess {
	contractState := subrealm.NewReadOnly(store, kv.Key(evm.Contract.Hname().Bytes()))
	return &StateAccess{state: contractState}
}

func (sa *StateAccess) Nonce(addr common.Address) uint64 {
	return Nonce(sa.state, addr)
}
