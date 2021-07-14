package webapiutil

import (
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

const retryOnStateInvalidatedRetry = 100 * time.Millisecond //nolint:gofumpt
const retryOnStateInvalidatedTimeout = 5 * time.Minute

func CallView(ch chain.ChainCore, contractHname, viewHname coretypes.Hname, params dict.Dict) (dict.Dict, error) {
	vctx := viewcontext.NewFromChain(ch)
	var ret dict.Dict
	err := optimism.RetryOnStateInvalidated(func() error {
		var err error
		ret, err = vctx.CallView(contractHname, viewHname, params)
		return err
	}, retryOnStateInvalidatedRetry, time.Now().Add(retryOnStateInvalidatedTimeout))
	return ret, err
}
