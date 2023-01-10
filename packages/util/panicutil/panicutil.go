package panicutil

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/kv"
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
					if errors.Is(err1, targetError) {
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

func CatchAllButDBError(f func(), log *logger.Logger, prefix ...string) (err error) {
	s := ""
	if len(prefix) > 0 {
		s = prefix[0]
	}
	func() {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			switch err1 := r.(type) {
			case *kv.DBError:
				log.Panicf("DB error: %v", err1)
			case error:
				err = err1
			default:
				err = fmt.Errorf("%s%v", s, err1)
			}
			log.Debugf("%s%v", s, err)
			log.Debugf(string(debug.Stack()))
		}()
		f()
	}()
	return err
}

func CatchPanic(f func()) (err error) {
	func() {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			var ok bool
			if err, ok = r.(error); !ok {
				err = fmt.Errorf("%v", r)
			}
			if os.Getenv("DEBUG") != "" {
				fmt.Println(string(debug.Stack()))
			}
		}()
		f()
	}()
	return err
}

func CatchAllExcept(f func(), exceptErrors ...error) (err error) {
	func() {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			if os.Getenv("DEBUG") != "" {
				fmt.Println(string(debug.Stack()))
			}
			if recoveredError, ok := r.(error); ok {
				for _, e := range exceptErrors {
					if errors.Is(recoveredError, e) {
						panic(err)
					}
				}
				if err, ok := r.(*kv.DBError); ok {
					panic(err)
				}
				err = recoveredError
			} else {
				err = fmt.Errorf("%v", r)
			}
		}()
		f()
	}()
	return err
}
