package admapi

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

type PutSCDataRequest struct {
	ScId          string             `json:"sc_id"` // base58
	OwnerPubkey   *hashing.HashValue `json:"owner_pubkey"`
	Description   string             `json:"description"`
	NodeLocations []*registry.PortAddr
}

//----------------------------------------------------------
func HandlerPutSCData(c echo.Context) error {
	var req registry.SCData

	if err := c.Bind(&req); err != nil {
		return misc.OkJsonErr(c, err)
	}
	ok, err := registry.ExistDKShareInRegistry(req.Address)
	if err != nil {
		return misc.OkJsonErr(c, err)
	}
	if !ok {
		return misc.OkJsonErr(c, fmt.Errorf("address %s is not in registry. Can't save SCData", req.Address.String()))
	}

	if err := registry.SaveSCData(&req); err != nil {
		log.Errorf("failed to save SC data: %v", err)
		return misc.OkJsonErr(c, err)
	}
	log.Infof("SC data saved: sc addr = %s descr = '%s'", req.Address.String(), req.Description)

	log.Debugf("+++++ saved %v", req)

	if scdBack, err := registry.GetSCData(req.Address); err != nil {
		log.Debugw("reading back",
			"sc addr", req.Address.String(),
			"error", err)
	} else {
		log.Debugw("reading back",
			"sc addr", req.Address.String(),
			"record", *scdBack)
	}
	return misc.OkJsonErr(c, nil)
}
