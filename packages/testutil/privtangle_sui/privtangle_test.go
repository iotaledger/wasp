package privtangle_sui

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStart(t *testing.T) {
	ctx := context.Background()

	pt := Start(ctx, "/tmp/sui_test", 5000, func(format string, args ...interface{}) {
		fmt.Printf(format+"\n", args...)
	})

	client := pt.nodeClient(0)
	state, err := client.GetLatestSuiSystemState(ctx)
	require.NoError(t, err)
	require.Equal(t, state.Jsonrpc, "2.0")
	require.Equal(t, state.Result.PendingActiveValidatorsSize, "0")

	w, _ := context.WithTimeout(context.Background(), time.Second*2)
	<-w.Done()

	pt.Stop()

}
