package apiextensions

import (
	"strconv"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
)

// TODO: Discuss API structures once we have time for it.
/**
	Should we also just return a map of coins similar to isc.Assets, or should we make it "simpler" and keep the current structure of
     * BaseTokens
     * Coins []Coin
*/
func AssetsFromAPIResponse(assetsResponse *apiclient.AssetsResponse) (*isc.Assets, error) {
	baseTokens, err := strconv.ParseUint(assetsResponse.BaseTokens, 10, 64)
	if err != nil {
		return nil, err
	}

	assets := isc.NewAssets(coin.Value(baseTokens))
	/*
		for _, nativeToken := range assetsResponse.NativeTokens {
			nativeTokenIDHex, err2 := iotago.DecodeHex(nativeToken.Id)
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
