package chainutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

func CallView(ch chain.ChainCore, contractHname, viewHname isc.Hname, params dict.Dict) (dict.Dict, error) {
	vctx := viewcontext.New(ch, ch.LatestBlockIndex())
	return vctx.CallViewExternal(contractHname, viewHname, params)
}

func CallViewAtBlockIndex(ch chain.ChainCore, blockIndex uint32, contractHname, viewHname isc.Hname, params dict.Dict) (dict.Dict, error) {
	vctx := viewcontext.New(ch, blockIndex)
	return vctx.CallViewExternal(contractHname, viewHname, params)
}
