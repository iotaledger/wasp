package suiclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
)

func TestRequestFundsFromFaucet_Devnet(t *testing.T) {
	err := suiclient.RequestFundsFromFaucet(context.Background(), testAddress, suiconn.DevnetFaucetURL)
	require.NoError(t, err)
}

func TestRequestFundsFromFaucet_Testnet(t *testing.T) {
	err := suiclient.RequestFundsFromFaucet(context.Background(), testAddress, suiconn.TestnetFaucetURL)
	require.NoError(t, err)
}

func TestRequestFundsFromFaucet_Localnet(t *testing.T) {
	t.Skip("only run with local node is set up")
	err := suiclient.RequestFundsFromFaucet(context.Background(), testAddress, suiconn.LocalnetFaucetURL)
	require.NoError(t, err)
}
