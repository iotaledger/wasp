package transaction

import "golang.org/x/xerrors"

var (
	ErrNotEnoughIotas               = xerrors.New("not enough iotas")
	ErrNotEnoughIotasForDustDeposit = xerrors.New("not enough iotas for dust deposit")
	ErrNotEnoughNativeTokens        = xerrors.New("not enough native tokens")
)
