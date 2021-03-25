package mocknode

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/client"
	"github.com/stretchr/testify/require"
)

func TestMockNode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mocknode test in short mode")
	}

	m := Start(":5000", ":8080")
	defer m.Stop()

	time.Sleep(1 * time.Second)

	_, addr := m.Ledger.NewKeyPairByIndex(2)

	cl := client.NewGoShimmerAPI("http://127.0.0.1:8080")
	_, err := cl.SendFaucetRequest(addr.Base58())
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	r, err := cl.GetUnspentOutputs([]string{addr.Base58()})

	require.NoError(t, err)
	require.Equal(t, 1, len(r.UnspentOutputs))
	require.Equal(t, addr.Base58(), r.UnspentOutputs[0].Address)
	require.Equal(t, 1, len(r.UnspentOutputs[0].OutputIDs))
	require.Equal(t, 1, len(r.UnspentOutputs[0].OutputIDs[0].Balances))
	require.EqualValues(t, 1337, r.UnspentOutputs[0].OutputIDs[0].Balances[0].Value)
}
