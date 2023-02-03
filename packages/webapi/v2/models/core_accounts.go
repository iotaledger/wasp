package models

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

type AccountsResponse struct {
	AccountIDs []string `json:"accountIds" swagger:"required"`
}

type AccountListResponse struct {
	Accounts []string `json:"accounts" swagger:"required"`
}

type NativeToken struct {
	ID     string `json:"id" swagger:"required"`
	Amount string `json:"amount" swagger:"required"`
}

type AssetsResponse struct {
	BaseTokens   uint64         `json:"baseTokens" swagger:"required"`
	NativeTokens []*NativeToken `json:"nativeTokens" swagger:"required"`
}

func MapNativeToken(token *iotago.NativeToken) *NativeToken {
	return &NativeToken{
		ID:     token.ID.ToHex(),
		Amount: token.Amount.String(),
	}
}

func MapNativeTokens(tokens iotago.NativeTokens) []*NativeToken {
	nativeTokens := make([]*NativeToken, len(tokens))

	for k, v := range tokens {
		nativeTokens[k] = MapNativeToken(v)
	}

	return nativeTokens
}

type AccountNFTsResponse struct {
	NFTIDs []string `json:"nftIds" swagger:"required"`
}

type AccountNonceResponse struct {
	Nonce uint64 `json:"nonce" swagger:"required"`
}

type NFTDataResponse struct {
	ID       string `json:"id" swagger:"required"`
	Issuer   string `json:"issuer" swagger:"required"`
	Metadata string `json:"metadata" swagger:"required"`
	Owner    string `json:"owner" swagger:"required"`
}

func MapNFTDataResponse(nft *isc.NFT) *NFTDataResponse {
	if nft == nil {
		return nil
	}

	return &NFTDataResponse{
		ID:       nft.ID.ToHex(),
		Issuer:   nft.Issuer.String(),
		Metadata: iotago.EncodeHex(nft.Metadata),
		Owner:    nft.Owner.String(),
	}
}

type NativeTokenIDRegistryResponse struct {
	NativeTokenRegistryIDs []string `json:"nativeTokenRegistryIds" swagger:"required"`
}

type FoundryOutputResponse struct {
	FoundryID string         `json:"foundryId" swagger:"required"`
	Assets    AssetsResponse `json:"assets" swagger:"required"`
}
