package shutdown_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/shutdown"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

func TestShutdownCoordinator(t *testing.T) {
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	sc := shutdown.NewCoordinator("test", log)
	require.True(t, sc.CheckNestedDone())

	sc1 := sc.Nested("1")
	require.False(t, sc.CheckNestedDone())
	sc1.Done()
	require.True(t, sc.CheckNestedDone())
	sc.WaitNested()
}
