package l1starter_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestStart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	stv := l1starter.Start(ctx, l1starter.DefaultConfig)
	defer stv.Stop()

	client := stv.Client()
	state, err := client.GetLatestIotaSystemState(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 0, state.PendingActiveValidatorsSize.Uint64())

	w, _ := context.WithTimeout(context.Background(), time.Second*2)
	<-w.Done()
}
