package util

import (
	"golang.org/x/xerrors"
)

func CatchPanicReturnError(fun func(), catchErrors ...error) error {
	var err error
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
