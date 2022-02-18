package testmisc

import (
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/iscp"
)

func RandChainID() *iscp.ChainID {
	ret := iscp.ChainIDFromAliasID(tpkg.RandAliasAddress().AliasID())
	return &ret
}
