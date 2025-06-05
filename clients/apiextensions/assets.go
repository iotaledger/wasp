package apiextensions

import (
	"strconv"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
)

// AssetsFromAPIResponse TODO: Handle Coins (other than base tokens) and Objects
func AssetsFromAPIResponse(assetsResponse *apiclient.AssetsResponse) (*isc.Assets, error) {
	baseTokens, err := strconv.ParseUint(assetsResponse.BaseTokens, 10, 64)
	if err != nil {
		return nil, err
	}

	assets := isc.NewAssets(coin.Value(baseTokens))
	/*
		for _, nativeToken := range assetsResponse.NativeTokens {
			nativeTokenIDHex, err2 := cryptolib.DecodeHex(nativeToken.Id)
			if err2 != nil {
				return nil, err2
			}

			nativeTokenID, err2 := isc.NativeTokenIDFromBytes(nativeTokenIDHex)
			if err2 != nil {
				return nil, err2
			}

			amount, err2 := iotago.DecodeUint256(nativeToken.Amount)
			if err2 != nil {
				return nil, err2
			}

			assets.NativeTokens = append(assets.NativeTokens, &iotago.NativeToken{
				ID:     nativeTokenID,
				Amount: amount,
			})
		}*/

	return assets, err
}
