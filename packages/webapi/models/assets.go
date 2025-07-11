package models

import "github.com/iotaledger/wasp/packages/isc"

type AssetsJSON struct {
	// Coins is a set of coin balances
	Coins []CoinJSON `json:"coins" swagger:"required"`
	// Objects is a set of non-Coin object IDs (e.g. NFTs)
	Objects []IotaObjectJSON `json:"objects" swagger:"required"`
}

func AssetsToAssetsJSON(a *isc.Assets) AssetsJSON {
	if a == nil {
		return AssetsJSON{}
	}

	return AssetsJSON{
		Coins:   ToCoinBalancesJSON(a.Coins),
		Objects: ToIotaObjectsJSON(&a.Objects),
	}
}

func ToIotaObjectsJSON(o *isc.ObjectSet) []IotaObjectJSON {
	objs := make([]IotaObjectJSON, 0, o.Size())
	for obj := range o.Iterate() {
		objs = append(objs, ToIotaObjectJSON(obj))
	}
	return objs
}
