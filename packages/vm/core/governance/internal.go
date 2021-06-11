package governance

import (
	"github.com/iotaledger/wasp/packages/kv"
)

func IsBlockMarkedFake(state kv.KVStoreReader) bool {
	return state.MustHas(StateVarFakeBlockMarker)
}
