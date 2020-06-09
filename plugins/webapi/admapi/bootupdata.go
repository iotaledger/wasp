package admapi

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/publisher"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
)

type BootupDataJsonable struct {
	Address        string   `json:"address"`
	CommitteeNodes []string `json:"committee_nodes"`
	AccessNodes    []string `json:"access_nodes"`
}

//----------------------------------------------------------
func HandlerPutSCData(c echo.Context) error {
	var req BootupDataJsonable
	var err error

	if err := c.Bind(&req); err != nil {
		return misc.OkJsonErr(c, err)
	}

	rec := registry.BootupData{}

	if rec.Address, err = address.FromBase58(req.Address); err != nil {
		return misc.OkJsonErr(c, err)
	}
	rec.CommitteeNodes = req.CommitteeNodes
	rec.AccessNodes = req.AccessNodes

	// TODO it is always overwritten!

	if err = registry.SaveBootupData(&rec, true); err != nil {
		return misc.OkJsonErr(c, err)
	}
	publisher.Publish("bootuprec", req.Address)

	log.Infof("Bootup record saved for addr = %s", rec.Address.String())

	//if bd, exists, err := registry.GetBootupData(&rec.Address); err != nil || !exists {
	//	log.Debugw("reading back",
	//		"sc addr", req.Address,
	//		"exists", exists,
	//		"error", err)
	//} else {
	//	log.Debugf("reading back: %+v", *bd)
	//}
	return misc.OkJsonErr(c, nil)
}

type GetBootupDataRequest struct {
	Address string `json:"address"`
}

type GetBootupDataResponse struct {
	BootupDataJsonable
	Exists bool   `json:"exists"`
	Error  string `json:"err"`
}

func HandlerGetSCData(c echo.Context) error {
	var req GetBootupDataRequest

	if err := c.Bind(&req); err != nil {
		return misc.OkJson(c, &GetBootupDataResponse{
			Error: err.Error(),
		})
	}
	addr, err := address.FromBase58(req.Address)
	if err != nil {
		return misc.OkJson(c, &GetBootupDataResponse{
			Error: err.Error(),
		})
	}

	bd, exists, err := registry.GetBootupData(&addr)
	if err != nil {
		return misc.OkJson(c, &GetBootupDataResponse{Error: err.Error()})
	}
	if !exists {
		return misc.OkJson(c, &GetBootupDataResponse{Exists: false})
	}
	return misc.OkJson(c, &GetBootupDataResponse{
		BootupDataJsonable: BootupDataJsonable{
			Address:        bd.Address.String(),
			CommitteeNodes: bd.CommitteeNodes,
			AccessNodes:    bd.AccessNodes,
		},
		Exists: exists,
	})
}

type GetScAddressesResponse struct {
	Addresses []string `json:"addresses"`
	Error     string   `json:"err"`
}

func HandlerGetSCList(c echo.Context) error {
	lst, err := registry.GetBootupRecords()
	if err != nil {
		return misc.OkJson(c, &GetScAddressesResponse{Error: err.Error()})
	}
	ret := make([]string, len(lst))
	for i := range ret {
		ret[i] = lst[i].Address.String()
	}
	return misc.OkJson(c, &GetScAddressesResponse{Addresses: ret})
}
