package client

import (
	"encoding/hex"
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

// StateGet fetches the raw value associated with the given key in the chain state
func (c *WaspClient) StateGet(chainID *iscp.ChainID, key string) ([]byte, error) {
	var res []byte
	if err := c.do(http.MethodGet, routes.StateGet(chainID.Hex(), hex.EncodeToString([]byte(key))), nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}
