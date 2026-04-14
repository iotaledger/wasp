package graphqltypes

import (
	"math/big"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

const MaxInputCountMerge = 256 - 1 // TODO find reference in Iota monorepo repo

type PickedCoins struct {
	Coins        Coins
	TotalAmount  *big.Int
	TargetAmount *big.Int
}

func (p *PickedCoins) Count() int {
	return len(p.Coins)
}

func (p *PickedCoins) CoinIds() []iotago.ObjectID {
	return p.Coins.ObjectIDs()
}

func (p *PickedCoins) CoinRefs() ([]*iotago.ObjectRef, error) {
	return p.Coins.CoinRefs()
}

// PickupCoins selects coins whose sum >= (targetAmount + gasBudget).
// The return coin number will be maxCoinNum <= coin_obj_num <= minCoinNum.
// Parameters:
//   - coins: coin data to select from
//   - hasNextPage: whether more coins are available beyond this set
//   - targetAmount: total amount of coins to be selected
//   - gasBudget: the transaction gas budget
//   - maxCoinNum: the max number of returned coins. Default (maxCoinNum <= 0) is MaxInputCountMerge
//   - minCoinNum: the min number of returned coins. Default (minCoinNum <= 0) is 3
//
// Returns ErrNoCoinsFound if the count of input coins is 0.
// Returns ErrInsufficientBalance if the input coins are all that is left and the total amount is less than the target amount.
// Returns ErrNeedMergeCoin if there are many coins, but the total amount of coins limited is less than the target amount.
func PickupCoins(
	coins Coins,
	targetAmount *big.Int,
	gasBudget uint64,
	maxCoinNum int,
	minCoinNum int,
) (*PickedCoins, error) {
	if len(coins) == 0 {
		return nil, ErrNoCoinsFound
	}
	if maxCoinNum <= 0 {
		maxCoinNum = MaxInputCountMerge
	}
	if minCoinNum <= 0 {
		minCoinNum = 3
	}
	if minCoinNum > maxCoinNum {
		minCoinNum = maxCoinNum
	}
	totalTarget := new(big.Int).Add(targetAmount, new(big.Int).SetUint64(gasBudget))

	total := big.NewInt(0)
	pickedCoins := Coins{}
	for i, coin := range coins {
		total = total.Add(total, new(big.Int).SetUint64(coin.Balance()))
		pickedCoins = append(pickedCoins, coin)
		if i+1 > maxCoinNum {
			return nil, ErrNeedMergeCoin
		}
		if i+1 < minCoinNum {
			continue
		}
		if total.Cmp(totalTarget) >= 0 {
			break
		}
	}
	if total.Cmp(totalTarget) < 0 {
		return nil, ErrInsufficientBalance
	}
	return &PickedCoins{
		Coins:        pickedCoins,
		TargetAmount: targetAmount,
		TotalAmount:  total,
	}, nil
}

func PickupCoinsWithCointype(
	coins Coins,
	targetAmount *big.Int,
	cointype CoinType,
) (*PickedCoins, error) {
	if len(coins) == 0 {
		return nil, ErrNoCoinsFound
	}

	total := big.NewInt(0)
	pickedCoins := Coins{}
	for _, coin := range coins {
		if coin.CoinType() != cointype {
			continue
		}
		total = total.Add(total, new(big.Int).SetUint64(coin.Balance()))
		pickedCoins = append(pickedCoins, coin)

		if total.Cmp(targetAmount) >= 0 {
			break
		}
	}
	if total.Cmp(targetAmount) < 0 {
		return nil, ErrInsufficientBalance
	}
	return &PickedCoins{
		Coins:        pickedCoins,
		TargetAmount: targetAmount,
		TotalAmount:  total,
	}, nil
}

func PickupCoinsSimple(coins Coins, targetAmount uint64) (Coins, error) {
	return PickupCoinsWithFilter(coins, targetAmount, nil)
}

func PickupCoinsWithFilter(
	coins Coins,
	targetAmount uint64,
	filter func(Coin) bool,
) (Coins, error) {
	if len(coins) == 0 {
		return nil, ErrNoCoinsFound
	}
	total := uint64(0)
	pickedCoins := Coins{}
	for _, coin := range coins {
		if filter != nil && !filter(coin) {
			continue
		}
		total += coin.Balance()
		pickedCoins = append(pickedCoins, coin)
		if total >= targetAmount {
			break
		}
	}
	if total < targetAmount {
		return nil, ErrInsufficientBalance
	}
	return pickedCoins, nil
}

func PickupCoinWithFilter(coins Coins, targetAmount uint64, filter func(Coin) bool) (Coin, bool, error) {
	coins, err := PickupCoinsWithFilter(coins, targetAmount, filter)
	if err != nil {
		return Coin{}, false, err
	}
	coin, ok := coins.PickCoinNoLess(targetAmount)
	return coin, ok, nil
}
