package admapi

import (
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

type GetScListResponse struct {
	SCDataList []*SCMetaDataJsonable `json:"sc_data_list"`
	Error      string                `json:"err"`
}

func HandlerGetSCList(c echo.Context) error {
	sclist, err := registry.GetSCDataList()
	if err != nil {
		return misc.OkJson(c, &GetScListResponse{Error: err.Error()})
	}
	retSclist := make([]*SCMetaDataJsonable, len(sclist))
	for i, scd := range sclist {
		retSclist[i] = toJsonable(scd)
	}
	return misc.OkJson(c, &GetScListResponse{SCDataList: retSclist})
}

func toJsonable(scdata *registry.SCMetaData) *SCMetaDataJsonable {
	return &SCMetaDataJsonable{
		Address:       scdata.Address.String(),
		Color:         scdata.Color.String(),
		OwnerAddress:  scdata.OwnerAddress.String(),
		Description:   scdata.Description,
		ProgramHash:   scdata.ProgramHash.String(),
		NodeLocations: scdata.NodeLocations,
	}
}
