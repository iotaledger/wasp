package evmlogger

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/log"

	hiveLog "github.com/iotaledger/hive.go/logger"
)

func Init(hiveLogger *hiveLog.Logger) {
	log.SetDefault(log.NewLogger(&hiveLogHandler{hiveLogger}))
}

type hiveLogHandler struct{ *hiveLog.Logger }

// Enabled implements slog.Handler.
func (*hiveLogHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// Handle implements slog.Handler.
func (h *hiveLogHandler) Handle(ctx context.Context, r slog.Record) error {
	switch {
	case r.Level >= slog.LevelError:
		h.Logger.Error(r.Message)
	case r.Level <= slog.LevelDebug:
		h.Logger.Debug(r.Message)
	case r.Level == slog.LevelWarn:
		h.Logger.Warn(r.Message)
	default:
		h.Logger.Info(r.Message)
	}
	return nil
}

// WithAttrs implements slog.Handler.
func (h *hiveLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// TODO: unimplemented in hive logger?
	return h
}

// WithGroup implements slog.Handler.
func (h *hiveLogHandler) WithGroup(name string) slog.Handler {
	// TODO: unimplemented in hive logger?
	return h
}
