package iotajsonrpc

import (
	"encoding/json"
	"errors"
	"math/big"
	"sort"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type Coin struct {
	CoinType     CoinType             `json:"coinType"`
	CoinObjectID *iotago.ObjectID     `json:"coinObjectID"`
	Version      *BigInt              `json:"version"`
	Digest       *iotago.ObjectDigest `json:"digest"`
	Balance      *BigInt              `json:"balance"`

	LockedUntilEpoch    *BigInt                  `json:"lockedUntilEpoch,omitempty"`
	PreviousTransaction iotago.TransactionDigest `json:"previousTransaction"`
}

type CoinPage = Page[*Coin, iotago.ObjectID]

func (c *Coin) Ref() *iotago.ObjectRef {
	return &iotago.ObjectRef{
		Digest:   c.Digest,
		Version:  c.Version.Uint64(),
		ObjectID: c.CoinObjectID,
	}
}

func (c *Coin) String() string {
	if c == nil {
		panic("coin is nil")
	}
	b, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (c *Coin) IsIOTA() bool {
	return MustCoinTypeFromString(c.CoinType.String()) == IotaCoinType
}

type CoinFields struct {
	Balance *BigInt
	ID      struct {
		ID *iotago.ObjectID
	}
}

type Coins []*Coin

func (cs Coins) TotalBalance() *big.Int {
	total := new(big.Int)
	for _, coin := range cs {
		total = total.Add(total, new(big.Int).SetUint64(coin.Balance.Uint64()))
	}
	return total
}

func (cs Coins) PickCoinNoLess(amount uint64) (*Coin, error) {
	for i, coin := range cs {
		if coin.Balance.Uint64() >= amount {
			cs = append(cs[:i], cs[i+1:]...)
			return coin, nil
		}
	}
	if len(cs) <= 3 {
		return nil, errors.New("insufficient balance")
	}
	return nil, errors.New("no coin is enough to cover the gas")
}

func (cs Coins) CoinRefs() []*iotago.ObjectRef {
	coinRefs := make([]*iotago.ObjectRef, len(cs))
	for idx, coin := range cs {
		coinRefs[idx] = coin.Ref()
	}
	return coinRefs
}

func (cs Coins) ObjectIDs() []*iotago.ObjectID {
	coinIDs := make([]*iotago.ObjectID, len(cs))
	for idx, coin := range cs {
		coinIDs[idx] = coin.CoinObjectID
	}
	return coinIDs
}

func (cs Coins) ObjectIDVals() []iotago.ObjectID {
	coinIDs := make([]iotago.ObjectID, len(cs))
	for idx, coin := range cs {
		coinIDs[idx] = *coin.CoinObjectID
	}
	return coinIDs
}

const (
	PickMethodSmaller = iota // pick smaller coins to match amount
	PickMethodBigger         // pick bigger coins to match amount
	PickMethodByOrder        // pick coins by coins order to match amount
)

// PickIOTACoinsWithGas pick coins, which sum >= amount, and pick a gas coin >= gasAmount which not in coins
// if not satisfied amount/gasAmount, an ErrCoinsNotMatchRequest/ErrCoinsNeedMoreObject error will return
// if gasAmount == 0, a nil gasCoin will return
// pickMethod, see PickMethodSmaller|PickMethodBigger|PickMethodByOrder
func (cs Coins) PickIOTACoinsWithGas(amount *big.Int, gasAmount uint64, pickMethod int) (Coins, *Coin, error) {
	if gasAmount == 0 {
		res, err := cs.PickCoins(amount, pickMethod)
		return res, nil, err
	}

	if amount.Cmp(new(big.Int)) == 0 && gasAmount == 0 {
		return make(Coins, 0), nil, nil
	} else if len(cs) == 0 {
		return cs, nil, ErrCoinsNeedMoreObject
	}

	// find smallest to match gasAmount
	var gasCoin *Coin
	var selectIndex int
	for i := range cs {
		if cs[i].Balance.Uint64() < gasAmount {
			continue
		}

		if gasCoin == nil || gasCoin.Balance.Uint64() > cs[i].Balance.Uint64() {
			gasCoin = cs[i]
			selectIndex = i
		}
	}
	if gasCoin == nil {
		return nil, nil, ErrCoinsNotMatchRequest
	}

	lastCoins := make(Coins, 0, len(cs)-1)
	lastCoins = append(lastCoins, cs[0:selectIndex]...)
	lastCoins = append(lastCoins, cs[selectIndex+1:]...)
	pickCoins, err := lastCoins.PickCoins(amount, pickMethod)
	return pickCoins, gasCoin, err
}

// PickCoins pick coins, which sum >= amount,
// pickMethod, see PickMethodSmaller|PickMethodBigger|PickMethodByOrder
// if not satisfied amount, an ErrCoinsNeedMoreObject error will return
func (cs Coins) PickCoins(amount *big.Int, pickMethod int) (Coins, error) {
	var sortedCoins Coins
	if pickMethod == PickMethodByOrder {
		sortedCoins = cs
	} else {
		sortedCoins = make(Coins, len(cs))
		copy(sortedCoins, cs)
		sort.Slice(
			sortedCoins, func(i, j int) bool {
				if pickMethod == PickMethodSmaller {
					return sortedCoins[i].Balance.Uint64() < sortedCoins[j].Balance.Uint64()
				} else {
					return sortedCoins[i].Balance.Uint64() >= sortedCoins[j].Balance.Uint64()
				}
			},
		)
	}

	result := make(Coins, 0)
	total := new(big.Int)
	for _, coin := range sortedCoins {
		result = append(result, coin)
		total = new(big.Int).Add(total, new(big.Int).SetUint64(coin.Balance.Uint64()))
		if total.Cmp(amount) >= 0 {
			return result, nil
		}
	}

	return nil, ErrCoinsNeedMoreObject
}
