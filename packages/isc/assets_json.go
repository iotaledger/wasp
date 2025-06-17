package isc

type AssetsJSON struct {
	// Coins is a set of coin balances
	Coins []CoinJSON `json:"coins" swagger:"required"`
	// Objects is a set of non-Coin object IDs (e.g. NFTs)
	Objects []IotaObjectJSON `json:"objects" swagger:"required"`
}

func AssetsToAssetsJSON(a *Assets) AssetsJSON {
	if a == nil {
		return AssetsJSON{}
	}

	return AssetsJSON{
		Coins:   a.Coins.JSON(),
		Objects: a.Objects.JSON(),
	}
}
