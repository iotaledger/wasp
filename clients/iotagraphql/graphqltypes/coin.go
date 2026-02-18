package graphqltypes

import (
	"errors"
	"math/big"
	"sort"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

// Coin is the GraphQL coin type with helper methods.
type Coin = CoinData

type Coins []Coin

func (c *CoinData) ObjectID() iotago.ObjectID {
	return iotago.ObjectID(c.Address)
}

func (c *CoinData) ObjectRef() (*iotago.ObjectRef, error) {
	digest, err := iotago.NewDigest(c.Digest)
	if err != nil {
		return nil, err
	}
	objectID := c.ObjectID()
	return &iotago.ObjectRef{
		ObjectID: &objectID,
		Version:  c.Version,
		Digest:   digest,
	}, nil
}

func (c *CoinData) CoinType() CoinType {
	return MustCoinTypeFromString(c.Contents.Type.Repr)
}

func (c *CoinData) IsIOTA() bool {
	return c.CoinType() == IotaCoinType
}

func (c *CoinData) Balance() uint64 {
	return c.CoinBalance.Uint64()
}

func (cs Coins) TotalBalance() *big.Int {
	total := new(big.Int)
	for _, coin := range cs {
		total = total.Add(total, new(big.Int).SetUint64(coin.Balance()))
	}
	return total
}

func (cs Coins) PickCoinNoLess(amount uint64) (Coin, bool) {
	for _, coin := range cs {
		if coin.Balance() >= amount {
			return coin, true
		}
	}
	return Coin{}, false
}

func (cs Coins) PickMultipleCoinsNoLess(amount uint64) (Coins, error) {
	if amount == 0 {
		return nil, nil
	}

	sum := uint64(0)
	var coins Coins
	for _, c := range cs {
		if sum >= amount {
			return coins, nil
		}
		bal := c.Balance()

		need := amount - sum
		coins = append(coins, c)
		if bal >= need {
			return coins, nil
		}
		sum += bal
	}
	return nil, errors.New("insufficient balance")
}

func (cs Coins) CoinRefs() ([]*iotago.ObjectRef, error) {
	coinRefs := make([]*iotago.ObjectRef, len(cs))
	for idx := range cs {
		ref, err := cs[idx].ObjectRef()
		if err != nil {
			return nil, err
		}
		coinRefs[idx] = ref
	}
	return coinRefs, nil
}

func (cs Coins) ObjectIDs() []iotago.ObjectID {
	coinIDs := make([]iotago.ObjectID, len(cs))
	for idx := range cs {
		coinIDs[idx] = cs[idx].ObjectID()
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
		if cs[i].Balance() < gasAmount {
			continue
		}

		if gasCoin == nil || gasCoin.Balance() > cs[i].Balance() {
			gasCoin = &cs[i]
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
					return sortedCoins[i].Balance() < sortedCoins[j].Balance()
				} else {
					return sortedCoins[i].Balance() >= sortedCoins[j].Balance()
				}
			},
		)
	}

	result := make(Coins, 0)
	total := new(big.Int)
	for _, coin := range sortedCoins {
		result = append(result, coin)
		total = new(big.Int).Add(total, new(big.Int).SetUint64(coin.Balance()))
		if total.Cmp(amount) >= 0 {
			return result, nil
		}
	}

	return nil, ErrCoinsNeedMoreObject
}
