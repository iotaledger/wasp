package params

import (
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
)

func DecodeChainID(e echo.Context) (isc.ChainID, error) {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return isc.ChainID{}, apierrors.InvalidPropertyError("chainID", err)
	}
	return chainID, nil
}

func DecodeRequestID(e echo.Context) (*isc.RequestID, error) {
	requestID, err := isc.RequestIDFromString(e.Param("requestID"))
	if err != nil {
		return nil, apierrors.InvalidPropertyError("requestID", err)
	}

	return &requestID, nil
}

func DecodeHNameFromHNameString(e echo.Context, key string) (isc.Hname, error) {
	hname, err := isc.HnameFromString(e.Param(key))
	if err != nil {
		return 0, apierrors.InvalidPropertyError(key, err)
	}

	return hname, nil
}

func DecodeHNameFromNameString(e echo.Context, key string) isc.Hname {
	hname := isc.Hn(e.Param(key))

	return hname
}

func DecodeAgentID(e echo.Context) (isc.AgentID, error) {
	agentIDDecoded, err := url.QueryUnescape(e.Param("agentID"))
	if err != nil {
		return nil, apierrors.InvalidPropertyError("agentID", err)
	}

	agentID, err := isc.NewAgentIDFromString(agentIDDecoded)
	if err != nil {
		return nil, apierrors.InvalidPropertyError("agentID", err)
	}

	return agentID, nil
}

func DecodeNFTID(e echo.Context) (*iotago.NFTID, error) {
	nftIDBytes, err := iotago.DecodeHex(e.Param("nftID"))
	if err != nil {
		return nil, apierrors.InvalidPropertyError("nftID", err)
	}

	if len(nftIDBytes) != iotago.NFTIDLength {
		return nil, apierrors.InvalidPropertyError("nftID", err)
	}

	var nftID iotago.NFTID
	copy(nftID[:], nftIDBytes)

	if err != nil {
		return nil, apierrors.InvalidPropertyError("nftID", err)
	}

	return &nftID, nil
}

func DecodeBlobHash(e echo.Context) (*hashing.HashValue, error) {
	blobHash, err := hashing.HashValueFromHex(e.Param("blobHash"))
	if err != nil {
		return nil, apierrors.InvalidPropertyError("blobHash", err)
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
