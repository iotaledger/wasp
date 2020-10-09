package client

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client/jsonable"
	"github.com/iotaledger/wasp/packages/registry"
)

const (
	PutBootupDataRoute     = "bootup"
	GetBootupDataListRoute = "bootup"
)

func GetBootupDataRoute(address string) string {
	return "bootup/" + address
}

// PutBootupData calls node to write BootupData record
func (c *WaspClient) PutBootupData(bd *registry.BootupData) error {
	return c.do(http.MethodPost, AdminRoutePrefix+"/"+PutBootupDataRoute, jsonable.NewBootupData(bd), nil)
}

// GetBootupData calls node to get BootupData record by address
func (c *WaspClient) GetBootupData(addr *address.Address) (*registry.BootupData, error) {
	res := jsonable.BootupData{}
	if err := c.do(http.MethodGet, AdminRoutePrefix+"/"+GetBootupDataRoute(addr.String()), nil, res); err != nil {
		return nil, err
	}
	return res.BootupData(), nil
}

// gets list of all SCs from the node
func (c *WaspClient) GetBootupDataList() ([]*registry.BootupData, error) {
	var res []*jsonable.BootupData
	if err := c.do(http.MethodGet, AdminRoutePrefix+"/"+GetBootupDataListRoute, nil, res); err != nil {
		return nil, err
	}
	list := make([]*registry.BootupData, len(res))
	for i, bd := range res {
		list[i] = bd.BootupData()
	}
	return list, nil
}
