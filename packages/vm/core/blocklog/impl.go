package blocklog

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("blocklog.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}
