package models

import (
	"github.com/iotaledger/wasp/packages/isc"
)

type TypeJSON = ObjectTypeJSON

type CoinJSON struct {
	CoinType TypeJSON `json:"coinType" swagger:"required"`
	Balance  string   `json:"balance" swagger:"required,desc(The balance (uint64 as string))"`
}

func CoinBalancesToJSON(b isc.CoinBalances) []CoinJSON {
	var coins []CoinJSON
	for t, v := range b.Iterate() {
		coins = append(coins, CoinJSON{
			CoinType: ToTypeJSON(t),
			Balance:  v.String(),
		})
	}
	return coins
}

// func (c *CoinBalances) UnmarshalJSON(b []byte) error {
// 	var coins []CoinJSON
// 	err := json.Unmarshal(b, &coins)
// 	if err != nil {
// 		return err
// 	}
// 	*c = NewCoinBalances()
// 	for _, cc := range coins {
// 		value := lo.Must(coin.ValueFromString(cc.Balance))
// 		c.Add(cc.CoinType.ToType(), value)
// 	}
// 	return nil
// }

// func (c CoinBalances) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(c.JSON())
// }
