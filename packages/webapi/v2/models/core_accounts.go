package models

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

type AccountsResponse struct {
	AccountIDs []string `json:"accountIds"`
}

type AccountListResponse struct {
	Accounts []string `json:"accounts"`
}

type NativeToken struct {
	ID     string `json:"id"`
	Amount string `json:"amount"`
}

type AssetsResponse struct {
	BaseTokens uint64         `json:"baseTokens"`
	Tokens     []*NativeToken `json:"nativeTokens"`
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
	NFTIDs []string `json:"nftIds"`
}

type AccountNonceResponse struct {
	Nonce uint64 `json:"nonce"`
}

type NFTDataResponse struct {
	ID       string `json:"id"`
	Issuer   string `json:"issuer"`
	Metadata string `json:"metadata"`
	Owner    string `json:"owner"`
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
	NativeTokenRegistryIDs []string `json:"nativeTokenRegistryIds"`
}

type FoundryOutputResponse struct {
	FoundryID string         `json:"foundryId"`
	Assets    AssetsResponse `json:"assets"`
}
