package webapi_validation

import (
	"log"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/tools/stardust-migration/webapi-validation/base"
	"github.com/stretchr/testify/require"
)

type GovernanceValidation struct {
	base.ValidationContext
	client base.GovernanceClientWrapper
}

func NewGovernanceValidation(validationContext base.ValidationContext) GovernanceValidation {
	return GovernanceValidation{
		ValidationContext: validationContext,
		client:            base.GovernanceClientWrapper{ValidationContext: validationContext},
	}
}

func (a *GovernanceValidation) ValidateGovernance(stateIndex uint32) {
	log.Printf("Validating governance for state index %d", stateIndex)
	sRes, rRes := a.client.GovernanceGetChainInfo(stateIndex)

	_, addr, err := iotago.ParseBech32(sRes.ChainID)
	require.NoError(base.T, err)
	oldChainID := addr.String()

	_, oldOwnerAddr, err := iotago.ParseBech32(sRes.ChainOwnerId)
	require.NoError(base.T, err)
	oldOwnerChainID := oldOwnerAddr.String()

	_ = oldChainID
	_ = oldOwnerChainID
	// require.Equal(base.T, oldChainID, rRes.ChainID)
	// require.Equal(base.T, oldOwnerChainID, rRes.ChainAdmin)
	require.Equal(base.T, sRes.GasFeePolicy.EvmGasRatio.A, rRes.GasFeePolicy.EvmGasRatio.A)
	require.Equal(base.T, sRes.GasFeePolicy.EvmGasRatio.B, rRes.GasFeePolicy.EvmGasRatio.B)
	require.Equal(base.T, sRes.GasFeePolicy.GasPerToken.A, rRes.GasFeePolicy.GasPerToken.A)
	require.Equal(base.T, sRes.GasFeePolicy.GasPerToken.B, rRes.GasFeePolicy.GasPerToken.B)
	require.Equal(base.T, sRes.GasFeePolicy.ValidatorFeeShare, rRes.GasFeePolicy.ValidatorFeeShare)

	require.Equal(base.T, sRes.GasLimits.MaxGasPerBlock, rRes.GasLimits.MaxGasPerBlock)
	require.Equal(base.T, sRes.GasLimits.MaxGasPerRequest, rRes.GasLimits.MaxGasPerRequest)
	require.Equal(base.T, sRes.GasLimits.MinGasPerRequest, rRes.GasLimits.MinGasPerRequest)
	require.Equal(base.T, sRes.GasLimits.MaxGasExternalViewCall, rRes.GasLimits.MaxGasExternalViewCall)

	require.Equal(base.T, sRes.Metadata.Description, rRes.Metadata.Description)
	require.Equal(base.T, sRes.Metadata.EvmJsonRpcURL, rRes.Metadata.EvmJsonRpcURL)
	require.Equal(base.T, sRes.Metadata.EvmWebSocketURL, rRes.Metadata.EvmWebSocketURL)
	require.Equal(base.T, sRes.Metadata.Name, rRes.Metadata.Name)
	require.Equal(base.T, sRes.Metadata.Website, rRes.Metadata.Website)

	require.Equal(base.T, sRes.PublicURL, rRes.PublicURL)
	log.Printf("Completed governance validation for state index %d", stateIndex)
}
