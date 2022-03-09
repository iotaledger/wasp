package panicutil

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/kv"
	"golang.org/x/xerrors"
	"runtime/debug"
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
				err = xerrors.Errorf("%s%v", s, err1)
			}
			log.Debugf("%s%v", s, err)
			log.Debugf(string(debug.Stack()))
		}()
		f()
	}()
	return err
}
