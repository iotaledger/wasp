package dashboard

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/iotaledger/wasp/plugins/database"
)

func callView(chain chain.Chain, hname coretypes.Hname, fname string, params dict.Dict) (dict.Dict, error) {
	vctx, err := viewcontext.NewFromDB(database.GetInstance(), *chain.ID(), chain.Processors())
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to create context: %v", err))
	}

	ret, err := vctx.CallView(hname, coretypes.Hn(fname), params)
	if err != nil {
		return nil, fmt.Errorf("root view call failed: %v", err)
	}
	return ret, nil
}
