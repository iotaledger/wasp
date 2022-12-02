package inccounter

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

type StateAccess struct {
	state kv.KVStoreReader
}

func NewStateAccess(store kv.KVStoreReader) *StateAccess {
	state := subrealm.NewReadOnly(store, kv.Key(Contract.Hname().Bytes()))
	return &StateAccess{state: state}
}

func (sa *StateAccess) GetMaintenanceStatus() int64 {
	return codec.MustDecodeInt64(sa.state.MustGet(VarCounter), 0)
}
