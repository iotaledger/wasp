package iotajsonrpc

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
)

// this type "CoinType" is used only in iota-go and iscmoveclient
type CoinType string

func CoinTypeFromString(s string) CoinType {
	return CoinType(s)
}

func (t CoinType) String() string {
	return string(t)
}

func (t CoinType) TypeTag() iotago.TypeTag {
	coinTypeTag, err := iotago.TypeTagFromString(t.String())
	if err != nil {
		panic(err)
	}
	return *coinTypeTag
}

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

func (t CoinValue) Uint64() uint64 {
	return uint64(t)
}

type Balance struct {
	CoinType        CoinType            `json:"coinType"`
	CoinObjectCount *BigInt             `json:"coinObjectCount"`
	TotalBalance    *BigInt             `json:"totalBalance"`
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
