package chainutil

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

// CallView executes a view call on the latest block of the chain
func CallView(
	anchor *isc.StateAnchor,
	chainState state.State,
	processors *processors.Cache,
	log *logger.Logger,
	msg isc.Message,
) (isc.CallArguments, error) {
	vctx, err := viewcontext.New(anchor, chainState, processors, log, false)
	if err != nil {
		return nil, err
	}
	return vctx.CallViewExternal(msg)
}
