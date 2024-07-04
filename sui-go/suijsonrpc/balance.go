package suijsonrpc

type CoinType string

type Balance struct {
	CoinType        CoinType          `json:"coinType"`
	CoinObjectCount uint64            `json:"coinObjectCount"`
	TotalBalance    *BigInt           `json:"totalBalance"`
	LockedBalance   map[BigInt]BigInt `json:"lockedBalance"` // FIXME the type may not be wrong
}
