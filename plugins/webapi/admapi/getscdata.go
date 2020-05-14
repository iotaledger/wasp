package admapi

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

type GetSCDataRequest struct {
	Address *address.Address `json:"address"`
}

type GetSCDataResponse struct {
	registry.SCMetaDataJsonable
	Exists bool   `json:"exists"`
	Error  string `json:"err"`
}

func HandlerGetSCData(c echo.Context) error {
	var req GetSCDataRequest

	if err := c.Bind(&req); err != nil {
		return misc.OkJson(c, &GetSCDataResponse{
			Error: err.Error(),
		})
	}
	scdata, exists, err := registry.GetSCData(req.Address)
	if err != nil {
		return misc.OkJson(c, &GetScListResponse{Error: err.Error()})
	}
	if !exists {
		return misc.OkJson(c, &GetSCDataResponse{Exists: exists})
	}
	return misc.OkJson(c, &GetSCDataResponse{SCMetaDataJsonable: *scdata.Jsonable(), Exists: exists})
}
