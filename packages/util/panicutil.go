package util

import (
	errorlib "errors"
	"github.com/iotaledger/wasp/packages/vm/vmerrors"
)

func CatchPanicReturnError(fun func(), catchErrors ...error) error {
	var err error
	func() {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			if err1, ok := r.(*vmerrors.Error); ok {
				for _, targetError := range catchErrors {
					if err1.Error() == targetError.Error() {
						err = targetError
						return
					}
				}
			}
			if err1, ok := r.(error); ok {
				for _, targetError := range catchErrors {
					if errorlib.Is(err1, targetError) {
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
