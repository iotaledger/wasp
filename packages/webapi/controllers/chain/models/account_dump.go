package models

import "github.com/iotaledger/wasp/packages/isc"

type AccountBalance struct {
	Coins   isc.CoinBalances
	Objects []isc.IotaObject
}

type DumpAccountsResponse struct {
	StateIndex uint32                    `json:"state_index"`
	Accounts   map[string]AccountBalance `json:"accounts"`
}
