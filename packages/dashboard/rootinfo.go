package dashboard

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// TODO test dashboard
type RootInfo struct {
	ChainID iscp.ChainID

	OwnerID          *iscp.AgentID
	OwnerIDDelegated *iscp.AgentID

	Description string
	Contracts   map[iscp.Hname]*root.ContractRecord

	FeeColor            colored.Color
	DefaultOwnerFee     uint64
	DefaultValidatorFee uint64
}

func (d *Dashboard) fetchRootInfo(chainID *iscp.ChainID) (ret RootInfo, err error) {
	info, err := d.wasp.CallView(chainID, root.Contract.Name, governance.FuncGetChainInfo.Name, nil)
	if err != nil {
		return
	}

	ret.Contracts, err = root.DecodeContractRegistry(collections.NewMapReadOnly(info, root.VarContractRegistry))
	if err != nil {
		return
	}

	ownerID, _, err := codec.DecodeAgentID(info.MustGet(governance.VarChainOwnerID))
	if err != nil {
		return
	}
	ret.OwnerID = &ownerID
	delegated, ok, err := codec.DecodeAgentID(info.MustGet(governance.VarChainOwnerIDDelegated))
	if err != nil {
		return
	}
	if ok {
		ret.OwnerIDDelegated = &delegated
	}
	ret.Description, _, err = codec.DecodeString(info.MustGet(governance.VarDescription))
	if err != nil {
		return
	}

	ret.FeeColor, ret.DefaultOwnerFee, ret.DefaultValidatorFee, err = governance.GetDefaultFeeInfo(info)
	if err != nil {
		return
	}

	return ret, err
}
