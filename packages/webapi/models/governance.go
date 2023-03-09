package models

import "github.com/iotaledger/wasp/packages/util"

type GasFeePolicy struct {
	GasPerToken       util.Ratio32 `json:"gasPerToken" swagger:"desc(The gas per token ratio (A/B) (gas/token)),required"`
	ValidatorFeeShare uint8        `json:"validatorFeeShare" swagger:"desc(The validator fee share.),required"`
	EVMGasRatio       util.Ratio32 `json:"evmGasRatio" swagger:"desc(The EVM gas ratio (ISC gas = EVM gas * A/B)),required"`
}

type GovChainInfoResponse struct {
	ChainID        string       `json:"chainID" swagger:"desc(ChainID (Bech32-encoded).),required"`
	ChainOwnerID   string       `json:"chainOwnerId" swagger:"desc(The chain owner address (Bech32-encoded).),required"`
	GasFeePolicy   GasFeePolicy `json:"gasFeePolicy" swagger:"desc(The gas fee policy),required"`
	CustomMetadata string       `json:"customMetadata" swagger:"desc((base64) Optional extra metadata that is appended to the L1 AliasOutput)"`
}

type GovAllowedStateControllerAddressesResponse struct {
	Addresses []string `json:"addresses" swagger:"desc(The allowed state controller addresses (Bech32-encoded))"`
}

type GovChainOwnerResponse struct {
	ChainOwner string `json:"chainOwner" swagger:"desc(The chain owner (Bech32-encoded))"`
}
