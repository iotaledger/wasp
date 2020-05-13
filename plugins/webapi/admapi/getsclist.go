package admapi

import (
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

type GetScListResponse struct {
	SCDataList []*registry.SCData `json:"sc_data_list"`
	Error      string             `json:"err"`
}

func HandlerGetSCList(c echo.Context) error {
	sclist, err := registry.GetSCDataList()
	if err != nil {
		return misc.OkJson(c, &GetScListResponse{Error: err.Error()})
	}
	return misc.OkJson(c, &GetScListResponse{SCDataList: sclist})
}
