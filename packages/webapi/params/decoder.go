package params

import (
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/apierrors"
)

func DecodeChainID(e echo.Context) (isc.ChainID, error) {
	chainID, err := isc.ChainIDFromString(e.Param(ParamChainID))
	if err != nil {
		return isc.ChainID{}, apierrors.InvalidPropertyError(ParamChainID, err)
	}
	return chainID, nil
}

func DecodePublicKey(e echo.Context) (*cryptolib.PublicKey, error) {
	publicKey, err := cryptolib.NewPublicKeyFromString(e.Param(ParamPublicKey))
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
	hname, err := isc.HnameFromHexString(e.Param(key))
	if err != nil {
		return 0, apierrors.InvalidPropertyError(key, err)
	}

	return hname, nil
}

func DecodeAgentID(e echo.Context) (isc.AgentID, error) {
	agentIDDecoded, err := url.QueryUnescape(e.Param(ParamAgentID))
	if err != nil {
		return nil, apierrors.InvalidPropertyError(ParamAgentID, err)
	}

	agentID, err := isc.NewAgentIDFromString(agentIDDecoded)
	if err != nil {
		return nil, apierrors.InvalidPropertyError(ParamAgentID, err)
	}

	return agentID, nil
}

func DecodeNFTID(e echo.Context) (*iotago.NFTID, error) {
	nftIDBytes, err := iotago.DecodeHex(e.Param(ParamNFTID))
	if err != nil {
		return nil, apierrors.InvalidPropertyError(ParamNFTID, err)
	}

	if len(nftIDBytes) != iotago.NFTIDLength {
		return nil, apierrors.InvalidPropertyError(ParamNFTID, err)
	}

	var nftID iotago.NFTID
	copy(nftID[:], nftIDBytes)

	if err != nil {
		return nil, apierrors.InvalidPropertyError(ParamNFTID, err)
	}

	return &nftID, nil
}

func DecodeBlobHash(e echo.Context) (*hashing.HashValue, error) {
	blobHash, err := hashing.HashValueFromHex(e.Param(ParamBlobHash))
	if err != nil {
		return nil, apierrors.InvalidPropertyError(ParamBlobHash, err)
	}

	return &blobHash, nil
}

// DecodeUInt decodes params to Uint64. If a lower Uint is expected it can be casted with uintX(returnValue) but validate the result
func DecodeUInt(e echo.Context, key string) (uint64, error) {
	value, err := strconv.ParseUint(e.Param(key), 10, 64)
	if err != nil {
		return 0, apierrors.InvalidPropertyError(key, err)
	}
	return value, nil
}
