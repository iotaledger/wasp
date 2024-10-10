package iotajsonrpc

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
)

type CoinType = string

type CoinValue uint64

func (t CoinValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%d", t))
}

func (t *CoinValue) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*t = CoinValue(v)
	return nil
}

type Balance struct {
	CoinType        CoinType            `json:"coinType"`
	CoinObjectCount uint64              `json:"coinObjectCount"`
	TotalBalance    CoinValue           `json:"totalBalance"`
	LockedBalance   map[EpochId]Uint128 `json:"lockedBalance"`
}

type MoveBalance struct {
	ID    *MoveUID
	Name  *iotago.ResourceType
	Value *BigInt
}

type MoveUID struct {
	ID *iotago.ObjectID `json:"id"`
}

type Allowance struct {
	CoinType        CoinType `json:"coinType"`
	CoinObjectCount uint64   `json:"coinObjectCount"`
}

type MoveAllowance struct {
	ID    *MoveUID
	Name  *iotago.ResourceType
	Value *BigInt
}
