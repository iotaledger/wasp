package graphqltypes

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type Coin struct {
	CoinType            string
	CoinObjectID        *iotago.ObjectID
	Version             uint64
	Digest              *iotago.ObjectDigest
	Balance             uint64
	LockedUntilEpoch    *uint64
	PreviousTransaction iotago.TransactionDigest
}

func (c *Coin) Ref() *iotago.ObjectRef {
	return &iotago.ObjectRef{
		Digest:   c.Digest,
		Version:  c.Version,
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
	return c.CoinType == "0x2::iota::IOTA"
}

type Coins []*Coin

func (cs Coins) TotalBalance() *big.Int {
	total := new(big.Int)
	for _, coin := range cs {
		total = total.Add(total, new(big.Int).SetUint64(coin.Balance))
	}
	return total
}

// PickCoinNoLess picks a single coin with balance >= amount
func (cs Coins) PickCoinNoLess(amount uint64) (*Coin, error) {
	for i, coin := range cs {
		if coin.Balance >= amount {
			cs = append(cs[:i], cs[i+1:]...)
			return coin, nil
		}
	}
	if len(cs) <= 3 {
		return nil, errors.New("insufficient balance")
	}
	return nil, errors.New("no coin is enough to cover the gas")
}

// PickMultipleCoinsNoLess picks multiple coins with total balance >= amount
func (cs Coins) PickMultipleCoinsNoLess(amount uint64) ([]*Coin, error) {
	if amount == 0 {
		return nil, nil
	}

	sum := uint64(0)
	var coins []*Coin
	for _, c := range cs {
		if sum >= amount {
			return coins, nil
		}
		bal := c.Balance

		need := amount - sum
		coins = append(coins, c)
		if bal >= need {
			return coins, nil
		}
		sum += bal
	}
	return nil, errors.New("insufficient balance")
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

const (
	PickMethodSmaller = iota // pick smaller coins to match amount
	PickMethodBigger         // pick bigger coins to match amount
	PickMethodByOrder        // pick coins by coins order to match amount
)

var (
	ErrCoinsNotMatchRequest = fmt.Errorf("coins not match request")
	ErrCoinsNeedMoreObject  = fmt.Errorf("need more coins")
)

// PickIOTACoinsWithGas picks coins >= amount and a gas coin >= gasAmount
func (cs Coins) PickIOTACoinsWithGas(amount *big.Int, gasAmount uint64, pickMethod int) (Coins, *Coin, error) {
	if gasAmount == 0 {
		res, err := cs.PickCoins(amount, pickMethod)
		return res, nil, err
	}

	if amount.Cmp(big.NewInt(0)) == 0 && gasAmount == 0 {
		return make(Coins, 0), nil, nil
	} else if len(cs) == 0 {
		return cs, nil, ErrCoinsNeedMoreObject
	}

	// find smallest to match gasAmount
	var gasCoin *Coin
	var selectIndex int
	for i := range cs {
		if cs[i].Balance < gasAmount {
			continue
		}

		if gasCoin == nil || gasCoin.Balance > cs[i].Balance {
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

// PickCoins picks coins with total >= amount using the specified method
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
					return sortedCoins[i].Balance < sortedCoins[j].Balance
				} else {
					return sortedCoins[i].Balance >= sortedCoins[j].Balance
				}
			},
		)
	}

	result := make(Coins, 0)
	total := new(big.Int)
	for _, coin := range sortedCoins {
		result = append(result, coin)
		total = new(big.Int).Add(total, new(big.Int).SetUint64(coin.Balance))
		if total.Cmp(amount) >= 0 {
			return result, nil
		}
	}

	return nil, ErrCoinsNeedMoreObject
}

type Balance struct {
	CoinType        string
	CoinObjectCount uint64
	TotalBalance    uint64
	LockedBalance   map[uint64]uint64
}

func (balance *Balance) String() string {
	b, err := json.Marshal(balance)
	if err != nil {
		panic(err)
	}
	return string(b)
}

type CoinMetadata struct {
	Name        string
	Symbol      string
	Decimals    uint8
	Description string
	IconUrl     string
	Id          *iotago.ObjectID
}

type Supply struct {
	Value uint64
}
