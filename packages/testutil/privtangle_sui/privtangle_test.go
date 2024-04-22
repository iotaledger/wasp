package privtangle_sui

import (
	"context"
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/testutil/privtangle_sui/miniclient"
)

func TestStart(t *testing.T) {
	ctx := context.Background()

	pt := Start(ctx, "/tmp/sui_test", 5000, 1, func(format string, args ...interface{}) {
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

func TestClient(t *testing.T) {
	client := miniclient.NewMiniClient("http://localhost:9000")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	res, err := client.GetLatestSuiSystemState(ctx)

	require.NoError(t, err)
	require.Equal(t, res.Jsonrpc, "2.0")
}

func TestA(t *testing.T) {
	syscall.Kill(362590, syscall.SIGINT)
}
