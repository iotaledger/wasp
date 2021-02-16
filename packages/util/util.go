package util

import (
	"errors"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/mr-tron/base58"
)

func ColorFromString(cs string) (ret balance.Color, err error) {
	if cs == "IOTA" {
		ret = balance.ColorIOTA
		return
	}
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
