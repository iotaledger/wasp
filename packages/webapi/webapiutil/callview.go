package webapiutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

func CallView(ch chain.ChainCore, contractHname, viewHname iscp.Hname, params dict.Dict) (dict.Dict, error) {
	var ret dict.Dict
	err := optimism.RetryOnStateInvalidated(func() error {
		vctx := viewcontext.NewFromChain(ch)
		var err error
		ret, err = vctx.CallView(contractHname, viewHname, params)
		return err
	})

	return ret, err
}
