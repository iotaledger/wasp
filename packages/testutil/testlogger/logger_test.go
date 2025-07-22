package testlogger_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

func TestLoggerBasic(t *testing.T) {
	l := testlogger.WithLevel(testlogger.NewLogger(t), log.LevelDebug, true)
	require.NotNil(t, l)
	l.LogInfo("testing the logger 1")
	l.LogDebug("testing debug 1")
	l = testlogger.WithLevel(l, log.LevelInfo, false)
	l.LogDebug("should not appear")
}
