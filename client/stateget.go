package client

import (
	"encoding/hex"
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// StateGet fetches the raw value associated with the given key in the chain state
func (c *WaspClient) StateGet(chainID coretypes.ChainID, key string) ([]byte, error) {
	var res []byte
	if err := c.do(http.MethodGet, routes.StateGet(chainID.Base58(), hex.EncodeToString([]byte(key))), nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}
