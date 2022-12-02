package chainutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

// CallView executes a view call on the latest block of the chain
func CallView(iscBlockIndex uint32, ch chain.ChainCore, contractHname, viewHname isc.Hname, params dict.Dict) (dict.Dict, error) {
	vctx := viewcontext.New(ch, iscBlockIndex)
	return vctx.CallViewExternal(contractHname, viewHname, params)
}
