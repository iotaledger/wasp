package testmisc

import (
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
)

func RandChainID() *isc.ChainID {
	ret := isc.ChainIDFromAliasID(tpkg.RandAliasAddress().AliasID())
	return &ret
}
