package params

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/webapi/apierrors"
)

func DecodeChainID(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param(ParamChainID))
	if err != nil {
		return isc.ChainID{}, apierrors.InvalidPropertyError(ParamChainID, err)
	}
	return chainID, nil
}

func DecodePublicKey(e echo.Context) (*cryptolib.PublicKey, error) {
	publicKey, err := cryptolib.PublicKeyFromString(e.Param(ParamPublicKey))
	if err != nil {
		return nil, apierrors.InvalidPropertyError(ParamPublicKey, err)
	}
	return publicKey, nil
}

func DecodeRequestID(e echo.Context) (isc.RequestID, error) {
	requestID, err := isc.RequestIDFromString(e.Param(ParamRequestID))
	if err != nil {
		return isc.RequestID{}, apierrors.InvalidPropertyError(ParamRequestID, err)
	}

	return requestID, nil
}

func DecodeHNameFromHNameHexString(e echo.Context, key string) (isc.Hname, error) {
	hname, err := isc.HnameFromString(e.Param(key))
	if err != nil {
		return 0, apierrors.InvalidPropertyError(key, err)
	}

	return hname, nil
}

func DecodeAgentID(e echo.Context) (isc.AgentID, error) {
	agentID, err := isc.AgentIDFromString(e.Param(ParamAgentID))
	if err != nil {
		return nil, apierrors.InvalidPropertyError(ParamAgentID, err)
	}

	return agentID, nil
}

// DecodeUInt decodes params to Uint64. If a lower Uint is expected it can be casted with uintX(returnValue) but validate the result
func DecodeUInt(e echo.Context, key string) (uint64, error) {
	value, err := strconv.ParseUint(e.Param(key), 10, 64)
	if err != nil {
		return 0, apierrors.InvalidPropertyError(key, err)
	}
	return value, nil
}
