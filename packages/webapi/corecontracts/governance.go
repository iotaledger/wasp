package corecontracts

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
)

type Governance struct {
	vmService interfaces.VMService
}

func NewGovernance(vmService interfaces.VMService) *Governance {
	return &Governance{
		vmService: vmService,
	}
}

func (g *Governance) GetAllowedStateControllerAddresses(chainID isc.ChainID) ([]iotago.Address, error) {
	ret, err := g.vmService.CallViewByChainID(chainID, governance.Contract.Hname(), governance.ViewGetAllowedStateControllerAddresses.Hname(), nil)
	if err != nil {
		return nil, err
	}

	stateControllerAddressesCollection := collections.NewArray16ReadOnly(ret, governance.ParamAllowedStateControllerAddresses)
	stateControllerAddressesCount, err := stateControllerAddressesCollection.Len()
	if err != nil {
		return nil, err
	}

	stateControllerAddresses := make([]iotago.Address, 0)
	for i := uint16(0); i < stateControllerAddressesCount; i++ {
		addressBytes, err := stateControllerAddressesCollection.GetAt(i)
		if err != nil {
			return nil, err
		}

		address, err := codec.DecodeAddress(addressBytes)
		if err != nil {
			return nil, err
		}

		stateControllerAddresses = append(stateControllerAddresses, address)
	}

	return stateControllerAddresses, nil
}

func (g *Governance) GetChainOwner(chainID isc.ChainID) (isc.AgentID, error) {
	ret, err := g.vmService.CallViewByChainID(chainID, governance.Contract.Hname(), governance.ViewGetChainOwner.Hname(), nil)
	if err != nil {
		return nil, err
	}

	ownerBytes, err := ret.Get(governance.ParamChainOwner)
	if err != nil {
		return nil, err
	}

	ownerID, err := isc.AgentIDFromBytes(ownerBytes)
	if err != nil {
		return nil, err
	}

	return ownerID, nil
}

func (g *Governance) GetChainInfo(chainID isc.ChainID) (*isc.ChainInfo, error) {
	ret, err := g.vmService.CallViewByChainID(chainID, governance.Contract.Hname(), governance.ViewGetChainInfo.Hname(), nil)
	if err != nil {
		return nil, err
	}

	var chainInfo *isc.ChainInfo

	if chainInfo, err = governance.GetChainInfo(ret, chainID); err != nil {
		return nil, err
	}

	return chainInfo, nil
}
