package suijsonrpc

import (
	"github.com/iotaledger/wasp/sui-go/sui"
)

type CoinType = string

type Balance struct {
	CoinType        CoinType            `json:"coinType"`
	CoinObjectCount uint64              `json:"coinObjectCount"`
	TotalBalance    uint64              `json:"totalBalance"`
	LockedBalance   map[EpochId]Uint128 `json:"lockedBalance"`
}

type MoveBalance struct {
	ID    *MoveUID
	Name  *sui.ResourceType
	Value *BigInt
}

type MoveUID struct {
	ID *sui.ObjectID `json:"id"`
}

type Allowance struct {
	CoinType        CoinType `json:"coinType"`
	CoinObjectCount uint64   `json:"coinObjectCount"`
}

type MoveAllowance struct {
	ID    *MoveUID
	Name  *sui.ResourceType
	Value *BigInt
}
