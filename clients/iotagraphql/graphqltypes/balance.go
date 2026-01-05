package graphqltypes

import (
	"encoding/json"
	"fmt"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

// this type "CoinType" is used only in iota-go and iscmoveclient
type CoinType string

func CoinTypeFromString(s string) (CoinType, error) {
	coinType, err := iotago.NewResourceType(s)
	if err != nil {
		return "", err
	}
	return CoinType(coinType.String()), nil
}

func MustCoinTypeFromString(s string) CoinType {
	coinType, err := CoinTypeFromString(s)
	if err != nil {
		panic(err)
	}
	return coinType
}

func (t CoinType) String() string {
	return string(t)
}

func (t *CoinType) MarshalBCS(e *bcs.Encoder) error {
	rt, err := iotago.NewResourceType(t.String())
	if err != nil {
		return err
	}
	// use ShortString to save space
	e.WriteString(rt.ShortString())
	return nil
}

func (t *CoinType) UnmarshalBCS(d *bcs.Decoder) error {
	var err error
	s := bcs.Decode[string](d)
	*t, err = CoinTypeFromString(s)
	return err
}

func (t CoinType) TypeTag() iotago.TypeTag {
	coinTypeTag, err := iotago.TypeTagFromString(t.String())
	if err != nil {
		panic(err)
	}
	return *coinTypeTag
}

func (t CoinType) MarshalJSON() ([]byte, error) {
	coinType, err := iotago.NewResourceType(t.String())
	if err != nil {
		return []byte{}, err
	}

	return json.Marshal(coinType.String())
}

type CoinValue uint64

func (t CoinValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%d", t))
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

func (balance *Balance) String() string {
	b, err := json.Marshal(balance)
	if err != nil {
		panic(err)
	}
	return string(b)
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
