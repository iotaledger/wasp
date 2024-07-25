package interfaces

import (
	"errors"

	"github.com/iotaledger/wasp/packages/isc"
)

var (
	ErrCantDeleteLastUser = errors.New("you can't delete the last user")
)

func NewChainNotFoundError(chainID isc.ChainID) error {
	return ChainNotFoundError{
		ChainID: chainID,
	}
}

type ChainNotFoundError struct {
	ChainID isc.ChainID
}

func (e ChainNotFoundError) Error() string {
	errStr := "Chain not found"

	if !e.ChainID.Empty() {
		errStr = errStr + ": " + e.ChainID.String()
	}

	return errStr
}
