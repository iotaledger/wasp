package iotaclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
)

func TestRequestFundsFromFaucet_Devnet(t *testing.T) {
	err := iotaclient.RequestFundsFromFaucet(context.Background(), testAddress, iotaconn.AlphanetFaucetURL)
	require.NoError(t, err)
}

func TestRequestFundsFromFaucet_Testnet(t *testing.T) {
	err := iotaclient.RequestFundsFromFaucet(context.Background(), testAddress, iotaconn.AlphanetFaucetURL)
	require.NoError(t, err)
}

func TestRequestFundsFromFaucet_Localnet(t *testing.T) {
	t.Skip("only run with local node is set up")
	err := iotaclient.RequestFundsFromFaucet(context.Background(), testAddress, iotaconn.LocalnetFaucetURL)
	require.NoError(t, err)
}
