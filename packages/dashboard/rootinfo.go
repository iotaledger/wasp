package dashboard

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

type RootInfo struct {
	ChainID iscp.ChainID

	OwnerID          *iscp.AgentID
	OwnerIDDelegated *iscp.AgentID

	Description string
	Contracts   map[iscp.Hname]*root.ContractRecord

	FeeColor            ledgerstate.Color
	DefaultOwnerFee     uint64
	DefaultValidatorFee uint64
}

func (d *Dashboard) fetchRootInfo(ch chain.ChainCore) (ret RootInfo, err error) {
	info, err := d.wasp.CallView(ch, root.Interface.Hname(), root.FuncGetChainInfo, nil)
	if err != nil {
		err = fmt.Errorf("root view call failed: %v", err)
		return
	}

	ret.Contracts, err = root.DecodeContractRegistry(collections.NewMapReadOnly(info, root.VarContractRegistry))
	if err != nil {
		err = fmt.Errorf("DecodeContractRegistry() failed: %v", err)
		return
	}

	ownerID, _, err := codec.DecodeAgentID(info.MustGet(root.VarChainOwnerID))
	if err != nil {
		return
	}
	ret.OwnerID = &ownerID
	delegated, ok, err := codec.DecodeAgentID(info.MustGet(root.VarChainOwnerIDDelegated))
	if err != nil {
		return
	}
	if ok {
		ret.OwnerIDDelegated = &delegated
	}
	ret.Description, _, err = codec.DecodeString(info.MustGet(root.VarDescription))
	if err != nil {
		return
	}

	ret.FeeColor, ret.DefaultOwnerFee, ret.DefaultValidatorFee, err = root.GetDefaultFeeInfo(info)
	if err != nil {
		return
	}

	return ret, err
}
