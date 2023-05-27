package corecontracts

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/webapi/common"
)

func GetAllowedStateControllerAddresses(ch chain.Chain) ([]iotago.Address, error) {
	ret, err := common.CallView(ch, governance.Contract.Hname(), governance.ViewGetAllowedStateControllerAddresses.Hname(), nil)
	if err != nil {
		return nil, err
	}

	stateControllerAddressesCollection := collections.NewArray16ReadOnly(ret, governance.ParamAllowedStateControllerAddresses)
	stateControllerAddressesCount := stateControllerAddressesCollection.Len()

	stateControllerAddresses := make([]iotago.Address, 0)
	for i := uint16(0); i < stateControllerAddressesCount; i++ {
		addressBytes := stateControllerAddressesCollection.GetAt(i)
		address, err := codec.DecodeAddress(addressBytes)
		if err != nil {
			return nil, err
		}

		stateControllerAddresses = append(stateControllerAddresses, address)
	}

	return stateControllerAddresses, nil
}

func GetChainOwner(ch chain.Chain) (isc.AgentID, error) {
	ret, err := common.CallView(ch, governance.Contract.Hname(), governance.ViewGetChainOwner.Hname(), nil)
	if err != nil {
		return nil, err
	}

	ownerBytes := ret.Get(governance.ParamChainOwner)
	ownerID, err := isc.AgentIDFromBytes(ownerBytes)
	if err != nil {
		return nil, err
	}

	return ownerID, nil
}

func GetChainInfo(ch chain.Chain) (*isc.ChainInfo, error) {
	ret, err := common.CallView(ch, governance.Contract.Hname(), governance.ViewGetChainInfo.Hname(), nil)
	if err != nil {
		return nil, err
	}

	var chainInfo *isc.ChainInfo

	if chainInfo, err = governance.GetChainInfo(ret, ch.ID()); err != nil {
		return nil, err
	}

	return chainInfo, nil
}
