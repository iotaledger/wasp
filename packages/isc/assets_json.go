package isc

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
)

type AssetsJSON struct {
	// Coins is a set of coin balances
	Coins []CoinJSON `json:"coins" swagger:"required"`
	// Objects is a set of non-Coin object IDs (e.g. NFTs)
	Objects []iotago.ObjectID `json:"objects" swagger:"required"`
}

func AssetsToAssetsJSON(a *Assets) AssetsJSON {
	if a == nil {
		return AssetsJSON{}
	}
	coins := a.Coins.JSON()
	objs := a.Objects.Sorted()

	return AssetsJSON{
		Coins:   coins,
		Objects: objs,
	}
}
