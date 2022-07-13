package client

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func (c *WaspClient) RequestIDByEVMTransactionHash(chainID *iscp.ChainID, txHash common.Hash) (iscp.RequestID, error) {
	var res model.RequestID
	if err := c.do(http.MethodGet, routes.RequestIDByEVMTransactionHash(chainID.String(), txHash.String()), nil, &res); err != nil {
		return iscp.RequestID{}, err
	}
	return res.RequestID(), nil
}
