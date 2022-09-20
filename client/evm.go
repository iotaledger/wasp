package client

import (
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
	"net/http"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/isc"
)

func (c *WaspClient) RequestIDByEVMTransactionHash(chainID isc.ChainID, txHash common.Hash) (isc.RequestID, error) {
	var res model.RequestID
	if err := c.do(http.MethodGet, routes.RequestIDByEVMTransactionHash(chainID.String(), txHash.String()), nil, &res); err != nil {
		return isc.RequestID{}, err
	}
	return res.RequestID(), nil
}
