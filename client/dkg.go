package client

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
)

const DKGImportRoute = "dkg/import"

func DKGExportRoute(address string) string {
	return "dkg/export/" + address
}

type DKShare struct {
	Blob []byte `json:"blob"`
}

func (c *WaspClient) ExportDKShare(addr *address.Address) ([]byte, error) {
	res := &DKShare{}
	if err := c.do(http.MethodGet, AdminRoutePrefix+"/"+DKGExportRoute(addr.String()), nil, res); err != nil {
		return nil, err
	}
	return res.Blob, nil
}

func (c *WaspClient) ImportDKShare(blob []byte) error {
	return c.do(http.MethodPost, AdminRoutePrefix+"/"+DKGImportRoute, &DKShare{Blob: blob}, nil)
}
