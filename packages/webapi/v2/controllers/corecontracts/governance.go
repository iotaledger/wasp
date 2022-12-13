package corecontracts

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/vm/core/governance"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/apierrors"
	"github.com/labstack/echo/v4"
)

type gasFeePolicy struct {
	GasFeeTokenID     string `swagger:"desc(The gas fee token id. Empty if base token.)"`
	GasPerToken       uint64 `swagger:"desc(The amount of gas per token.)"`
	ValidatorFeeShare uint8  `swagger:"desc(The validator fee share.)"`
}

type ChainInfoResponse struct {
	ChainID         string       `swagger:"desc(ChainID (Bech32-encoded).)"`
	ChainOwnerID    string       `swagger:"desc(The chain owner address (Bech32-encoded).)"`
	Description     string       `swagger:"desc(The description of the chain.)"`
	GasFeePolicy    gasFeePolicy `json:"GasFeePolicy"`
	MaxBlobSize     uint32       `swagger:"desc(The maximum contract blob size.)"`
	MaxEventSize    uint16       `swagger:"desc(The maximum event size.)"`                   // TODO: Clarify
	MaxEventsPerReq uint16       `swagger:"desc(The maximum amount of events per request.)"` // TODO: Clarify
}

func MapChainInfoResponse(chainInfo *governance.ChainInfo) ChainInfoResponse {
	gasFeeTokenID := ""

	if chainInfo.GasFeePolicy.GasFeeTokenID != nil {
		gasFeeTokenID = chainInfo.GasFeePolicy.GasFeeTokenID.String()
	}

	chainInfoResponse := ChainInfoResponse{
		ChainID:      chainInfo.ChainID.String(),
		ChainOwnerID: chainInfo.ChainOwnerID.String(),
		Description:  chainInfo.Description,
		GasFeePolicy: gasFeePolicy{
			GasFeeTokenID:     gasFeeTokenID,
			GasPerToken:       chainInfo.GasFeePolicy.GasPerToken,
			ValidatorFeeShare: chainInfo.GasFeePolicy.ValidatorFeeShare,
		},
		MaxBlobSize:     chainInfo.MaxBlobSize,
		MaxEventSize:    chainInfo.MaxEventSize,
		MaxEventsPerReq: chainInfo.MaxEventsPerReq,
	}

	return chainInfoResponse
}

func (c *Controller) getChainInfo(e echo.Context) error {
	chainID, err := isc.ChainIDFromString(e.Param("chainID"))
	if err != nil {
		return apierrors.InvalidPropertyError("chainID", err)
	}

	chainInfo, err := c.governance.GetChainInfo(chainID)

	if err != nil {
		return apierrors.ContractExecutionError(err)
	}

	chainInfoResponse := MapChainInfoResponse(chainInfo)

	return e.JSON(http.StatusOK, chainInfoResponse)
}
