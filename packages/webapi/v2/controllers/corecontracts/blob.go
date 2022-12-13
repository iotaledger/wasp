package corecontracts

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/packages/hashing"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/labstack/echo/v4"
)

type Blob struct {
	Hash string
	Size uint32
}

type BlobListResponse struct {
	Blobs []Blob
}

func (c *Controller) listBlobs(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	blobList, err := c.blob.ListBlobs(chainID)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	blobListResponse := &BlobListResponse{
		Blobs: make([]Blob, 0, len(blobList)),
	}

	for k, v := range blobList {
		blobListResponse.Blobs = append(blobListResponse.Blobs, Blob{
			Hash: k.Hex(),
			Size: v,
		})
	}

	return e.JSON(http.StatusOK, blobListResponse)
}

type BlobValueResponse struct {
	ValueData string
}

func (c *Controller) getBlobValue(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	blobHash, err := hashing.HashValueFromHex(e.Param("blobHash"))
	if err != nil {
		return apierrors.InvalidPropertyError("blobHash", err)
	}

	fieldKey := e.Param("fieldKey")

	blobValueBytes, err := c.blob.GetBlobValue(chainID, blobHash, fieldKey)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	blobValueResponse := &BlobValueResponse{
		ValueData: hexutil.Encode(blobValueBytes),
	}

	return e.JSON(http.StatusOK, blobValueResponse)
}

type BlobInfoResponse struct {
	Fields map[string]uint32
}

func (c *Controller) getBlobInfo(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	blobHash, err := hashing.HashValueFromHex(e.Param("blobHash"))
	if err != nil {
		return apierrors.InvalidPropertyError("blobHash", err)
	}

	blobInfo, ok, err := c.blob.GetBlobInfo(chainID, blobHash)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	if !ok {
		return e.NoContent(http.StatusNotFound)
	}

	blobInfoResponse := &BlobInfoResponse{
		Fields: map[string]uint32{},
	}

	for k, v := range blobInfo {
		blobInfoResponse.Fields[k] = v
	}

	return e.JSON(http.StatusOK, blobInfoResponse)
}
