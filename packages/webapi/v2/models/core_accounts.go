package models

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

type AccountsResponse struct {
	AccountIDs []string
}

type AccountListResponse struct {
	Accounts []string
}

type NativeToken struct {
	ID     string
	Amount string
}

type AssetsResponse struct {
	BaseTokens uint64
	Tokens     []*NativeToken
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
	NFTIDs []string
}

type AccountNonceResponse struct {
	Nonce uint64
}

type NFTDataResponse struct {
	ID       string
	Issuer   string
	Metadata string
	Owner    string
}

func MapNFTDataResponse(nft *isc.NFT) *NFTDataResponse {
	if nft == nil {
		return nil
	}

	return &NFTDataResponse{
		ID:       nft.ID.ToHex(),
		Issuer:   nft.Issuer.String(),
		Metadata: hexutil.Encode(nft.Metadata),
		Owner:    nft.Owner.String(),
	}
}

type NativeTokenIDRegistryResponse struct {
	NativeTokenRegistryIDs []string
}

type FoundryOutputResponse struct {
	FoundryID string
	Token     AssetsResponse
}
