package evmlogger

import (
	"strings"

	"github.com/ethereum/go-ethereum/log"

	"github.com/iotaledger/hive.go/logger"
)

var format = log.TerminalFormat(false)

func Init(waspLogger *logger.Logger) {
	log.Root().SetHandler(log.FuncHandler(func(r *log.Record) error {
		s := strings.TrimRight(string(format.Format(r)), "\n")
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
