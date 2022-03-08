// http://localhost:14265/api/plugins/indexer/v1/outputs
// http://localhost:14265/api/plugins/indexer/v1/outputs?address=atoi1qpszqzadsym6wpppd6z037dvlejmjuke7s24hm95s9fg9vpua7vluehe53e
// http://localhost:14265/api/v2/outputs/00000000000000000000000000000000000000000000000000000000000000000000
//

package nodeconn_test

import (
	"context"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	iotagox "github.com/iotaledger/iota.go/v3/x"
	"github.com/stretchr/testify/require"
)

// func setupSuite(tb testing.TB) func(tb testing.TB) {
// 	tb.Logf("==============> Setup")
// 	return func(tb testing.TB) {
// 		tb.Logf("==============> Teardown")
// 	}
// }

func TestApi(t *testing.T) {
	ctx := context.Background()
	nodeAPI := nodeclient.New("http://localhost:14265", iotago.ZeroRentParas, nodeclient.WithIndexer())
	healthy, err := nodeAPI.Health(ctx)
	require.NoError(t, err)
	require.True(t, healthy)

	res, err := nodeAPI.Indexer().Outputs(ctx, &nodeclient.OutputsQuery{
		AddressBech32: "atoi1qpszqzadsym6wpppd6z037dvlejmjuke7s24hm95s9fg9vpua7vluehe53e",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	t.Logf("Outputs...")
	for res.Next() {
		outs, err := res.Outputs()
		require.NoError(t, err)
		for i, o := range outs {
			t.Logf("Outputs: %v=%#v", res.Response.Items[i], o)
		}
	}
	t.Logf("Outputs... Done")
}

func TestEvents(t *testing.T) {
	ctx := context.Background()
	nodeEvt := iotagox.NewNodeEventAPIClient("ws://localhost:14266/mqtt")
	require.NoError(t, nodeEvt.Connect(ctx))
	ch := nodeEvt.ConfirmedMilestones()
	for i := 0; i < 10; i++ {
		ms := <-ch
		t.Logf("MS=%+v", ms)
	}
}
