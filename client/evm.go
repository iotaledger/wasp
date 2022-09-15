package client

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
)

func (c *WaspClient) RequestIDByEVMTransactionHash(chainID *isc.ChainID, txHash common.Hash) (isc.RequestID, error) {
	var res model.RequestID
	if err := c.do(http.MethodGet, routes.RequestIDByEVMTransactionHash(chainID.String(), txHash.String()), nil, &res); err != nil {
		return isc.RequestID{}, err
	}
	return res.RequestID(), nil
}
