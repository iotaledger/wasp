package l1starter_test

import (
	"context"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/testutil/testmisc"

	"github.com/iotaledger/wasp/packages/testutil/l1starter"

	"github.com/stretchr/testify/require"
)

func TestStart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	iotaNode, destruct := l1starter.StartNode(context.Background())
	defer destruct()

	client := iotaNode.L1Client()
	state, err := client.GetLatestIotaSystemState(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 0, state.PendingActiveValidatorsSize.Uint64())

	w, cancel := context.WithTimeout(context.Background(), testmisc.GetTimeout(2*time.Second))
	defer cancel()
	<-w.Done()
}
