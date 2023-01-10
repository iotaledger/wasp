package chainutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

// CallView executes a view call on a given block of the chain (nil blockIndex for the latest block)
func CallView(ch chain.ChainCore, b *viewcontext.BlockIndexOrTrieRoot, contractHname, viewHname isc.Hname, params dict.Dict) (dict.Dict, error) {
	vctx, err := viewcontext.New(ch, b)
	if err != nil {
		return nil, err
	}
	return vctx.CallViewExternal(contractHname, viewHname, params)
}
