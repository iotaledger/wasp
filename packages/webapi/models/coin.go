package models

import (
	"github.com/iotaledger/wasp/packages/isc"
)

type TypeJSON = ObjectTypeJSON

type CoinJSON struct {
	CoinType TypeJSON `json:"coinType" swagger:"required"`
	Balance  string   `json:"balance" swagger:"required,desc(The balance (uint64 as string))"`
}

func ToCoinBalancesJSON(b isc.CoinBalances) []CoinJSON {
	var coins []CoinJSON
	for t, v := range b.Iterate() {
		coins = append(coins, CoinJSON{
			CoinType: ToTypeJSON(t),
			Balance:  v.String(),
		})
	}
	return coins
}
