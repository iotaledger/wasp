package isc

type AssetsJSON struct {
	// Coins is a set of coin balances
	Coins []CoinJSON `json:"coins" swagger:"required"`
	// Objects is a set of non-Coin object IDs (e.g. NFTs)
	Objects []IotaObject `json:"objects" swagger:"required"`
}

func AssetsToAssetsJSON(a *Assets) AssetsJSON {
	if a == nil {
		return AssetsJSON{
			Coins:   []CoinJSON{},
			Objects: []IotaObject{},
		}
	}

	coins := a.Coins.JSON()
	objs := a.Objects.JSON()

	return AssetsJSON{
		Coins:   coins,
		Objects: objs,
	}
}
