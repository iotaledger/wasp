package models

import "github.com/iotaledger/wasp/packages/util"

type GasFeePolicy struct {
	GasPerToken       util.Ratio32 `json:"gasPerToken" swagger:"desc(The gas per token ratio (fee = gas * A/B)),required"`
	ValidatorFeeShare uint8        `json:"validatorFeeShare" swagger:"desc(The validator fee share.),required"`
	EVMGasRatio       util.Ratio32 `json:"evmGasRatio" swagger:"desc(The EVM gas ratio (ISC gas = EVM gas * A/B)),required"`
}

type GovChainInfoResponse struct {
	ChainID         string       `json:"chainID" swagger:"desc(ChainID (Bech32-encoded).),required"`
	ChainOwnerID    string       `json:"chainOwnerId" swagger:"desc(The chain owner address (Bech32-encoded).),required"`
	Description     string       `json:"description" swagger:"desc(The description of the chain.),required"`
	GasFeePolicy    GasFeePolicy `json:"gasFeePolicy" swagger:"desc(The gas fee policy),required"`
	MaxBlobSize     uint32       `json:"maxBlobSize" swagger:"desc(The maximum contract blob size.),required"`
	MaxEventSize    uint16       `json:"maxEventSize" swagger:"desc(The maximum event size.),required"`                      // TODO: Clarify
	MaxEventsPerReq uint16       `json:"maxEventsPerReq" swagger:"desc(The maximum amount of events per request.),required"` // TODO: Clarify
}

type GovAllowedStateControllerAddressesResponse struct {
	Addresses []string `json:"addresses" swagger:"desc(The allowed state controller addresses (Bech32-encoded))"`
}

type GovChainOwnerResponse struct {
	ChainOwner string `json:"chainOwner" swagger:"desc(The chain owner (Bech32-encoded))"`
}
