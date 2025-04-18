package apiextensions

import (
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
)

// TODO: Handle Objects
func AssetsFromAPIResponse(assetsResponse *apiclient.AssetsResponse) (*isc.Assets, error) {
	assets := isc.NewAssets(0)
	for _, resCoin := range assetsResponse.Coins {
		bal, err := coin.ValueFromString(resCoin.Balance)
		if err != nil {
			return nil, err
		}
		assets.AddCoin(coin.MustTypeFromString(resCoin.CoinType), bal)
	}

	return assets, nil
}
