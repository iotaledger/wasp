package logger

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/iotaledger/hive.go/logger"
)

func initGoEthLogger(waspLogger *logger.Logger) {
	log.Root().SetHandler(log.FuncHandler(func(r *log.Record) error {
		switch r.Lvl {
		case log.LvlCrit, log.LvlError:
			waspLogger.Errorf("[%s] %s", r.Lvl.AlignedString(), r.Msg)
		case log.LvlTrace, log.LvlDebug:
			waspLogger.Debugf("[%s] %s", r.Lvl.AlignedString(), r.Msg)
		default:
			waspLogger.Infof("[%s] %s", r.Lvl.AlignedString(), r.Msg)
		}
		return nil
	}))
}
