package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
)

func (e *ChainEnv) testDumpAccounts(t *testing.T) {
	// create 10 accounts with funds on-chain
	accs := make([]string, 0, 100)
	for i := 0; i < 5; i++ {
		// 5 L1 accounts
		keyPair, addr, err := e.Clu.NewKeyPairWithFunds()
		require.NoError(t, err)
		e.DepositFunds(10*isc.Million, keyPair)
		accs = append(accs, addr.String())
	}

	for i := 0; i < 5; i++ {
		// 5 EVM accounts
		_, evmAddr := solo.NewEthereumAccount()
		keyPair, _, err := e.Clu.NewKeyPairWithFunds()
		require.NoError(t, err)
		evmAgentID := isc.NewEthereumAddressAgentID(evmAddr)
		e.TransferFundsTo(isc.NewAssets(iotaclient.DefaultGasBudget-1*isc.Million), keyPair, evmAgentID)
		accs = append(accs, evmAgentID.String())
	}

	client, _ := e.NewRandomChainClient()
	resp, err := client.WaspClient.ChainsAPI.DumpAccounts(
		context.Background(),
	).Execute()
	require.NoError(t, err)
	require.Equal(t, 202, resp.StatusCode)
	time.Sleep(1 * time.Second) // wait for the file to be produced

	path := filepath.Join(e.Clu.NodeDataPath(0), "waspdb", "account_dumps", e.Chain.ChainID.String())
	entries, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	contents, err := os.ReadFile(filepath.Join(path, entries[0].Name()))
	require.NoError(t, err)
	// assert all accounts are present in the dump
	for _, acc := range accs {
		require.Contains(t, string(contents), acc)
	}
}
