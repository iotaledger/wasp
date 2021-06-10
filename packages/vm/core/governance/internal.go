package governance

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func MustIsAllowedCommitteeAddress(state kv.KVStoreReader, addr ledgerstate.Address) bool {
	amap := collections.NewMapReadOnly(state, StateVarAllowedCommitteeAddresses)
	return amap.MustHasAt(addr.Bytes())
}
