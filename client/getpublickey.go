package client

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/client/jsonable"
)

func GetPubKeyInfoRoute(address string) string {
	return "dkg/pub/" + address
}

type PubKeyInfo struct {
	Address      *jsonable.Address `json:"address"`
	N            uint16            `json:"n"`
	T            uint16            `json:"t"`
	Index        uint16            `json:"index"`
	PubKeys      []string          `json:"pub_keys"`       // base58
	PubKeyMaster string            `json:"pub_key_master"` // base58
}

// GetPublicKeyInfo retrieves public info about key with specific address from host
func (c *WaspClient) GetPublicKeyInfo(addr *address.Address) (*PubKeyInfo, error) {
	res := &PubKeyInfo{}
	if err := c.do(http.MethodGet, AdminRoutePrefix+"/"+GetPubKeyInfoRoute(addr.String()), nil, res); err != nil {
		return nil, err
	}
	return res, nil
}
