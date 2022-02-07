package util

import (
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"golang.org/x/xerrors"
)

func CatchPanicReturnError(fun func(), catchErrors ...*errors.Error) *errors.Error {
	var err *errors.Error
	func() {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			if err1, ok := r.(error); ok {
				for _, targetError := range catchErrors {
					if xerrors.Is(err1, targetError) {
						err = targetError
						return
					}
				}
			}
			panic(r)
		}()
		fun()
	}()
	return err
}
