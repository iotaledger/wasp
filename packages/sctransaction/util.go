package sctransaction

import (
	"crypto/rand"
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/mr-tron/base58"
)

func RandomColor() (ret balance.Color) {
	if _, err := rand.Read(ret[:]); err != nil {
		panic(err)
	}
	return
}

func ColorFromString(cs string) (ret balance.Color, err error) {
	var bin []byte
	bin, err = base58.Decode(cs)
	if err != nil {
		return
	}
	ret, err = ColorFromBytes(bin)
	return
}

func ColorFromBytes(cb []byte) (ret balance.Color, err error) {
	if len(cb) != balance.ColorLength {
		err = errors.New("must be exactly 32 bytes for color")
		return
	}
	copy(ret[:], cb)
	return
}

// sums value of valances with particular color
func SumBalancesOfColor(balances []*balance.Balance, color *balance.Color) int64 {
	var ret int64
	for _, bal := range balances {
		if bal.Color() == *color {
			ret += bal.Value()
		}
	}
	return ret
}
