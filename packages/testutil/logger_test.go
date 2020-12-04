package testutil

import (
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestLoggerBasic(t *testing.T) {
	l := WithLevel(NewLogger(t), zapcore.DebugLevel)
	require.NotNil(t, l)
	l.Info("testing the logger 1")
	l.Debug("testing debug 1")
	l = WithLevel(l, zapcore.InfoLevel)
	l.Debug("should not appear")
}
