package admapi

import (
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

type GetScListResponse struct {
	SCDataList []*registry.SCMetaDataJsonable `json:"sc_data_list"`
	Error      string                         `json:"err"`
}

func HandlerGetSCList(c echo.Context) error {
	sclist, err := registry.GetSCDataList()
	if err != nil {
		return misc.OkJson(c, &GetScListResponse{Error: err.Error()})
	}
	retSclist := make([]*registry.SCMetaDataJsonable, len(sclist))
	for i, scd := range sclist {
		retSclist[i] = scd.Jsonable()
	}
	return misc.OkJson(c, &GetScListResponse{SCDataList: retSclist})
}
