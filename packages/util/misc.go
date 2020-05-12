package util

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/pkg/errors"
	"io"
)

func Short(s string) string {
	if len(s) <= 6 {
		return s
	}
	return s[:6] + ".."
}

func ReadAddress(r io.Reader, addr *address.Address) error {
	n, err := r.Read(addr[:])
	if err != nil {
		return err
	}
	if n != address.Length {
		return errors.New("error while reading address")
	}
	return nil
}

func ReadColor(r io.Reader, color *balance.Color) error {
	n, err := r.Read(color[:])
	if err != nil {
		return err
	}
	if n != balance.ColorLength {
		return errors.New("error while reading color code")
	}
	return nil
}
