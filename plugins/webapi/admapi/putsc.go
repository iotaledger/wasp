package admapi

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

type SCMetaDataJsonable struct {
	Address       string   `json:"address"`       // base58
	Color         string   `json:"color"`         // base58
	OwnerAddress  string   `json:"owner_address"` // base58
	Description   string   `json:"description"`   // base58
	ProgramHash   string   `json:"program_hash"`  // base58
	NodeLocations []string `json:"node_locations"`
}

//----------------------------------------------------------
func HandlerPutSCData(c echo.Context) error {
	var req SCMetaDataJsonable

	if err := c.Bind(&req); err != nil {
		return misc.OkJsonErr(c, err)
	}

	scdata := registry.SCMetaData{
		Address:       address.Address{},
		Color:         balance.Color{},
		OwnerAddress:  address.Address{},
		Description:   "",
		ProgramHash:   hashing.HashValue{},
		NodeLocations: nil,
	}
	var err error
	if scdata.Address, err = address.FromBase58(req.Address); err != nil {
		return misc.OkJsonErr(c, err)
	}
	if scdata.Color, err = sctransaction.ColorFromString(req.Color); err != nil {
		return misc.OkJsonErr(c, err)
	}
	if scdata.OwnerAddress, err = address.FromBase58(req.OwnerAddress); err != nil {
		return misc.OkJsonErr(c, err)
	}
	scdata.Description = req.Description
	if h, err := hashing.HashValueFromString(req.ProgramHash); err != nil {
		return misc.OkJsonErr(c, err)
	} else {
		scdata.ProgramHash = *h
	}
	scdata.NodeLocations = req.NodeLocations

	ok, err := registry.ExistDKShareInRegistry(&scdata.Address)
	if err != nil {
		return misc.OkJsonErr(c, err)
	}
	if !ok {
		return misc.OkJsonErr(c, fmt.Errorf("address %s is not in registry", req.Address))
	}

	if err := registry.SaveSCData(&scdata); err != nil {
		log.Errorf("failed to save SC data: %v", err)
		return misc.OkJsonErr(c, err)
	}
	log.Infof("SC data saved: sc addr = %s descr = '%s'", req.Address, req.Description)

	log.Debugf("+++++ saved %v", req)

	if scdBack, err := registry.GetSCData(&scdata.Address); err != nil {
		log.Debugw("reading back",
			"sc addr", req.Address,
			"error", err)
	} else {
		log.Debugf("reading back: %+v", *scdBack)
	}
	return misc.OkJsonErr(c, nil)
}
