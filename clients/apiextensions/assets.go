package apiextensions

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
)

func NewAssetsFromAPIResponse(assetsResponse *apiclient.AssetsResponse) (*isc.Assets, error) {
	assets := isc.NewEmptyAssets()

	baseTokens, err := iotago.DecodeUint64(assetsResponse.BaseTokens)
	if err != nil {
		return nil, err
	}

	assets.BaseTokens = baseTokens

	for _, nativeToken := range assetsResponse.NativeTokens {
		nativeTokenIDHex, err := iotago.DecodeHex(nativeToken.Id)
		if err != nil {
			return nil, err
		}

		nativeTokenID, err := isc.NativeTokenIDFromBytes(nativeTokenIDHex)
		if err != nil {
			return nil, err
		}

		amount, err := iotago.DecodeUint256(nativeToken.Amount)
		if err != nil {
			return nil, err
		}

		assets.NativeTokens = append(assets.NativeTokens, &iotago.NativeToken{
			ID:     nativeTokenID,
			Amount: amount,
		})
	}

	return assets, err
}
