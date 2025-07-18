package iotaclienttest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestRequestFundsFromFaucet_Devnet(t *testing.T) {
	t.Skip("Disable Faucet request until its stable on L1")

	err := iotaclient.RequestFundsFromFaucet(
		context.Background(), iotago.MustAddressFromHex(testcommon.TestAddress),
		iotaconn.DevnetFaucetURL,
	)
	require.NoError(t, err)
}

func TestRequestFundsFromFaucet_Testnet(t *testing.T) {
	t.Skip("Disable Faucet request until its stable on L1")

	err := iotaclient.RequestFundsFromFaucet(
		context.Background(),
		iotago.MustAddressFromHex(testcommon.TestAddress),
		iotaconn.TestnetFaucetURL,
	)
	require.NoError(t, err)
}

func TestRequestFundsFromFaucet_Localnet(t *testing.T) {
	if !l1starter.Instance().IsLocal() {
		t.Skip("only run with local node is set up")
	}

	err := iotaclient.RequestFundsFromFaucet(
		context.Background(),
		iotago.MustAddressFromHex(testcommon.TestAddress),
		l1starter.Instance().FaucetURL(),
	)
	require.NoError(t, err)
}
