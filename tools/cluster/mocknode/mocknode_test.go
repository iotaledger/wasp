package mocknode

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMockNode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mocknode test in short mode")
	}

	m := Start(":5001", ":8081") // use different ports so this doesn't conflict with cluster tests
	defer m.Stop()

	time.Sleep(1 * time.Second)

	_, addr := m.Ledger.NewKeyPairByIndex(2)

	cl := client.NewGoShimmerAPI("http://127.0.0.1:8081")
	_, err := cl.SendFaucetRequest(addr.Base58(), 0)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	r, err := cl.GetAddressUnspentOutputs(addr.Base58())
	require.NoError(t, err)

	require.Equal(t, 1, len(r.Outputs))
	out, err := r.Outputs[0].ToLedgerstateOutput()
	require.NoError(t, err)
	require.Equal(t, addr.Base58(), out.Address().Base58())
	require.Equal(t, 1, out.Balances().Size())
	b, _ := out.Balances().Get(ledgerstate.ColorIOTA)
	require.EqualValues(t, utxodb.RequestFundsAmount, b)
}
