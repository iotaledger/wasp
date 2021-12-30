package dashboard

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

type RootInfo struct {
	ChainID *iscp.ChainID

	OwnerID          *iscp.AgentID
	OwnerIDDelegated *iscp.AgentID

	Description string
	Contracts   map[iscp.Hname]*root.ContractRecord

	FeeColor            []byte
	DefaultOwnerFee     uint64
	DefaultValidatorFee uint64
}

func (d *Dashboard) fetchRootInfo(chainID *iscp.ChainID) (ret RootInfo, err error) {
	info, err := d.wasp.CallView(chainID, governance.Contract.Name, governance.FuncGetChainInfo.Name, nil)
	if err != nil {
		return
	}

	ownerID, err := codec.DecodeAgentID(info.MustGet(governance.VarChainOwnerID))
	if err != nil {
		return
	}
	ret.OwnerID = ownerID
	if info.MustHas(governance.VarChainOwnerIDDelegated) {
		delegated, err := codec.DecodeAgentID(info.MustGet(governance.VarChainOwnerIDDelegated))
		if err != nil {
			return ret, err
		}
		ret.OwnerIDDelegated = delegated
	}
	ret.Description, err = codec.DecodeString(info.MustGet(governance.VarDescription), "")
	if err != nil {
		return
	}

	ret.FeeColor, ret.DefaultOwnerFee, ret.DefaultValidatorFee, err = governance.GetDefaultFeeInfo(info)
	if err != nil {
		return
	}

	recs, err := d.wasp.CallView(chainID, root.Contract.Name, root.FuncGetContractRecords.Name, nil)
	if err != nil {
		return
	}
	ret.Contracts, err = root.DecodeContractRegistry(collections.NewMapReadOnly(recs, root.StateVarContractRegistry))
	if err != nil {
		return
	}

	return ret, err
}
