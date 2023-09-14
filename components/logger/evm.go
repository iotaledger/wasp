package logger

import (
	"github.com/ethereum/go-ethereum/log"

	"github.com/iotaledger/hive.go/logger"
)

var format = log.TerminalFormat(false)

func initGoEthLogger(waspLogger *logger.Logger) {
	log.Root().SetHandler(log.FuncHandler(func(r *log.Record) error {
		s := string(format.Format(r))
		switch r.Lvl {
		case log.LvlCrit, log.LvlError:
			waspLogger.Error(s)
		case log.LvlTrace, log.LvlDebug:
			waspLogger.Debug(s)
		default:
			waspLogger.Info(s)
		}
		return nil
	}))
}
