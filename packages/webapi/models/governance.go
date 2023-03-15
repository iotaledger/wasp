package models

import "github.com/iotaledger/wasp/packages/vm/gas"

type GovChainInfoResponse struct {
	ChainID        string         `json:"chainID" swagger:"desc(ChainID (Bech32-encoded).),required"`
	ChainOwnerID   string         `json:"chainOwnerId" swagger:"desc(The chain owner address (Bech32-encoded).),required"`
	GasFeePolicy   *gas.FeePolicy `json:"gasFeePolicy" swagger:"desc(The gas fee policy),required"`
	GasLimits      *gas.Limits    `json:"gasLimits" swagger:"desc(The gas limits),required"`
	CustomMetadata string         `json:"customMetadata" swagger:"desc((base64) Optional extra metadata that is appended to the L1 AliasOutput)"`
}

type GovAllowedStateControllerAddressesResponse struct {
	Addresses []string `json:"addresses" swagger:"desc(The allowed state controller addresses (Bech32-encoded))"`
}

type GovChainOwnerResponse struct {
	ChainOwner string `json:"chainOwner" swagger:"desc(The chain owner (Bech32-encoded))"`
}
