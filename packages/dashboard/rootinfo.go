package dashboard

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

type RootInfo struct {
	OwnerID     coretypes.AgentID
	Description string
	Contracts   map[coretypes.Hname]*root.ContractRecord
}

func fetchRootInfo(chain chain.Chain) (ret RootInfo, err error) {
	info, err := callView(chain, root.Interface.Hname(), root.FuncGetChainInfo, nil)
	if err != nil {
		err = fmt.Errorf("root view call failed: %v", err)
		return
	}

	ret.Contracts, err = root.DecodeContractRegistry(datatypes.NewMustMap(info, root.VarContractRegistry))
	if err != nil {
		err = fmt.Errorf("DecodeContractRegistry() failed: %v", err)
		return
	}

	ret.OwnerID, _, _ = codec.DecodeAgentID(info.MustGet(root.VarChainOwnerID))
	ret.Description, _, _ = codec.DecodeString(info.MustGet(root.VarDescription))
	return
}
