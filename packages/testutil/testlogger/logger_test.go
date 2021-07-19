package testlogger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestLoggerBasic(t *testing.T) {
	l := WithLevel(NewLogger(t), zapcore.DebugLevel, true)
	require.NotNil(t, l)
	l.Info("testing the logger 1")
	l.Debug("testing debug 1")
	l = WithLevel(l, zapcore.InfoLevel, false)
	l.Debug("should not appear")
}
