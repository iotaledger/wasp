package controllerutils

import (
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
	"github.com/iotaledger/wasp/packages/webapi/params"
)

// both used by the prometheus metrics middleware
const (
	EchoContextKeyChainID   = "chainid"
	EchoContextKeyOperation = "operation"
)

func ChainIDFromParams(c echo.Context, cs interfaces.ChainService) (isc.ChainID, error) {
	chainID, err := params.DecodeChainID(c)
	if err != nil {
		return chainID, err
	}

	if !cs.HasChain(chainID) {
		return chainID, apierrors.ChainNotFoundError(chainID.String())
	}
	// set chainID to be used by the prometheus metrics
	c.Set(EchoContextKeyChainID, chainID)
	return chainID, nil
}

// sets the label of the operation (endpoint being called) to be used by the prometheus metrics middleware
func SetOperation(c echo.Context, op string) {
	c.Set(EchoContextKeyOperation, op)
}
