package models

import "github.com/iotaledger/wasp/packages/isc"

type AccountsResponse struct {
	AccountIDs []string `json:"accountIds" swagger:"required"`
}

type AccountListResponse struct {
	Accounts []string `json:"accounts" swagger:"required"`
}

type FungibleTokensResponse struct {
	BaseTokens   string                 `json:"baseTokens" swagger:"required,desc(The base tokens (uint64 as string))"`
	NativeTokens isc.NativeTokenMapJSON `json:"nativeTokens" swagger:"required"`
}

type AccountNFTsResponse struct {
	NFTIDs []string `json:"nftIds" swagger:"required"`
}

type AccountFoundriesResponse struct {
	FoundrySerialNumbers []uint32 `json:"foundrySerialNumbers" swagger:"required"`
}

type AccountNonceResponse struct {
	Nonce string `json:"nonce" swagger:"required,desc(The nonce (uint64 as string))"`
}

type NativeTokenIDRegistryResponse struct {
	NativeTokenRegistryIDs []string `json:"nativeTokenRegistryIds" swagger:"required"`
}

type FoundryOutputResponse struct {
	FoundryID     string `json:"foundryId" swagger:"required"`
	BaseTokens    string `json:"baseTokens" swagger:"required,desc(The base tokens (uint64 as string))"`
	NativeTokenID string `json:"nativeTokenId" swagger:"required,desc(The native token id)"`
}
