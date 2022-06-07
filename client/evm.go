package client

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func (c *WaspClient) EVMRequestIDByTransactionHash(chainID *iscp.ChainID, txHash common.Hash) (iscp.RequestID, error) {
	var res model.RequestID
	if err := c.do(http.MethodGet, routes.EVMRequestIDByTransactionHash(chainID.String(), txHash.String()), nil, &res); err != nil {
		return iscp.RequestID{}, err
	}
	return res.RequestID(), nil
}
