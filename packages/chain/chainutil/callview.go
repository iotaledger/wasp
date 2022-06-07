package chainutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

func CallView(ch chain.ChainCore, contractHname, viewHname iscp.Hname, params dict.Dict) (dict.Dict, error) {
	var ret dict.Dict
	err := optimism.RetryOnStateInvalidated(func() (err error) {
		vctx := viewcontext.New(ch)
		ret, err = vctx.CallViewExternal(contractHname, viewHname, params)
		return err
	})

	return ret, err
}
