package webapiutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

func CallView(ch chain.ChainCore, contractHname, viewHname coretypes.Hname, params dict.Dict) (dict.Dict, error) {
	vctx := viewcontext.NewFromChain(ch)
	var ret dict.Dict
	var err error
	err = optimism.RepeatOnceIfUnlucky(func() error {
		ret, err = vctx.CallView(contractHname, viewHname, params)
		return err
	})
	return ret, err
}
