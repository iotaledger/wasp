package main

import (
	"github.com/iotaledger/wasp/packages/isc"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
)

func OldChainIDToNewChainID(oldChainID old_isc.ChainID) isc.ChainID {
	return isc.ChainID(oldChainID)
}

func OldHnameToNewHname(oldHname old_isc.Hname) isc.Hname {
	return isc.Hname(oldHname)
}
