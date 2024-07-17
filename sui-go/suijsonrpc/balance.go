package suijsonrpc

import (
	"github.com/iotaledger/wasp/sui-go/sui"
)

type CoinType = string

type Balance struct {
	CoinType        CoinType          `json:"coinType"`
	CoinObjectCount uint64            `json:"coinObjectCount"`
	TotalBalance    *BigInt           `json:"totalBalance"`
	LockedBalance   map[BigInt]BigInt `json:"lockedBalance"` // FIXME the type may not be wrong
}

type MoveBalance struct {
	ID    *MoveUID
	Name  *sui.ResourceType
	Value *BigInt
}

type MoveUID struct {
	ID *sui.ObjectID `json:"id"`
}