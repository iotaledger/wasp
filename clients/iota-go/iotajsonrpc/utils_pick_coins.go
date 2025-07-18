package iotajsonrpc

import (
	"math/big"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

const MAX_INPUT_COUNT_MERGE = 256 - 1 // TODO find reference in Iota monorepo repo

type PickedCoins struct {
	Coins        Coins
	TotalAmount  *big.Int
	TargetAmount *big.Int
}

func (p *PickedCoins) Count() int {
	return len(p.Coins)
}

func (p *PickedCoins) CoinIds() []*iotago.ObjectID {
	coinIDs := make([]*iotago.ObjectID, len(p.Coins))
	for idx, coin := range p.Coins {
		coinIDs[idx] = coin.CoinObjectID
	}
	return coinIDs
}

func (p *PickedCoins) CoinRefs() []*iotago.ObjectRef {
	coinRefs := make([]*iotago.ObjectRef, len(p.Coins))
	for idx, coin := range p.Coins {
		coinRefs[idx] = coin.Ref()
	}
	return coinRefs
}

// Select coins whose sum >= (targetAmount + gasBudget)
// The return coin number will be maxCoinNum <= coin_obj_num <= minCoinNum
// @param inputCoins queried page coin data
// @param targetAmount total amount of coins to be selected from inputCoins
// @param gasBudget the transaction gas budget
// @param maxCoinNum the max number of returned coins. Default (maxCoinNum <= 0) is `MAX_INPUT_COUNT_MERGE`
// @param minCoinNum the min number of returned coins. Default (minCoinNum <= 0) is 3
// @throw ErrNoCoinsFound If the count of input coins is 0.
// @throw ErrInsufficientBalance If the input coins are all that is left and the total amount is less than the target amount.
// @throw ErrNeedMergeCoin If there are many coins, but the total amount of coins limited is less than the target amount.
func PickupCoins(
	inputCoins *CoinPage,
	targetAmount *big.Int,
	gasBudget uint64,
	maxCoinNum int,
	minCoinNum int,
) (*PickedCoins, error) {
	coins := inputCoins.Data
	inputCount := len(coins)
	if inputCount <= 0 {
		return nil, ErrNoCoinsFound
	}
	if maxCoinNum <= 0 {
		maxCoinNum = MAX_INPUT_COUNT_MERGE
	}
	if minCoinNum <= 0 {
		minCoinNum = 3
	}
	if minCoinNum > maxCoinNum {
		minCoinNum = maxCoinNum
	}
	totalTarget := new(big.Int).Add(targetAmount, new(big.Int).SetUint64(gasBudget))

	total := big.NewInt(0)
	pickedCoins := []*Coin{}
	for i, coin := range coins {
		total = total.Add(total, new(big.Int).SetUint64(coin.Balance.Uint64()))
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
		if inputCoins.HasNextPage {
			return nil, ErrNeedMergeCoin
		}
		sub := new(big.Int).Sub(totalTarget, total)
		if sub.Uint64() > gasBudget {
			return nil, ErrInsufficientBalance
		}
	}
	return &PickedCoins{
		Coins:        pickedCoins,
		TargetAmount: targetAmount,
		TotalAmount:  total,
	}, nil
}

func PickupCoinsWithCointype(
	inputCoins *CoinPage,
	targetAmount *big.Int,
	cointype CoinType,
) (*PickedCoins, error) {
	coins := inputCoins.Data
	inputCount := len(coins)
	if inputCount <= 0 {
		return nil, ErrNoCoinsFound
	}

	total := big.NewInt(0)
	pickedCoins := []*Coin{}
	for _, coin := range coins {
		if coin.CoinType != cointype {
			continue
		}
		total = total.Add(total, new(big.Int).SetUint64(coin.Balance.Uint64()))
		pickedCoins = append(pickedCoins, coin)

		if total.Cmp(targetAmount) >= 0 {
			break
		}
	}
	if total.Cmp(targetAmount) < 0 {
		if inputCoins.HasNextPage {
			return nil, ErrNeedMergeCoin
		}
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
	filter func(*Coin) bool,
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
		total += coin.Balance.Uint64()
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

func PickupCoinWithFilter(coins Coins, targetAmount uint64, filter func(*Coin) bool) (*Coin, error) {
	coins, err := PickupCoinsWithFilter(coins, targetAmount, filter)
	if err != nil {
		return nil, err
	}

	return coins.PickCoinNoLess(targetAmount)
}
