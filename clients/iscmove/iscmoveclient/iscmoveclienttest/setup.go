// Package iscmoveclienttest provides testing utilities for the ISC move client.
package iscmoveclienttest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

const ExpectedCoinCount = 5

func NewSignerWithFunds(t *testing.T, seed []byte, index int) cryptolib.Signer {
	seedCopy := make([]byte, len(seed))
	copy(seedCopy, seed)
	seedCopy[0] += byte(index)
	kp := cryptolib.KeyPairFromSeed(cryptolib.Seed(seedCopy))
	client := l1starter.Instance().L1Client()
	addr := kp.Address().AsIotaAddress()

	err := client.RequestFundsFromFaucet(context.Background(), addr)
	require.NoError(t, err)

	iotatest.EnsureCoinCount(t, cryptolib.SignerToIotaSigner(kp), client, ExpectedCoinCount)
	return kp
}

func NewRandomSignerWithFunds(t *testing.T, index int) cryptolib.Signer {
	seed := cryptolib.NewSeed()
	return NewSignerWithFunds(t, seed[:], index)
}

func NewWebSocketClient(ctx context.Context) (*iscmoveclient.Client, error) {
	if l1starter.IsLocalConfigured() { //nolint:contextcheck
		panic("Right now no WS support")
	}

	return iscmoveclient.NewWebsocketClient(
		ctx,
		iotaconn.AlphanetWebsocketEndpointURL,
		l1starter.Instance().FaucetURL(),
		l1starter.WaitUntilEffectsVisible,
	)
}

func NewClient() *iscmoveclient.Client {
	if l1starter.IsSimulatorConfigured() {
		return iscmoveclient.NewClient(l1starter.Instance().L1Client().GetIotaClient())
	}
	return iscmoveclient.NewClient(
		iotagraphql.NewGraphQLClientWithWaitParams(
			l1starter.Instance().APIURL(),
			l1starter.Instance().FaucetURL(),
			l1starter.WaitUntilEffectsVisible,
		),
	)
}

func NewAlphanetClient() *iscmoveclient.Client {
	return iscmoveclient.NewClient(
		iotagraphql.NewGraphQLClientWithWaitParams(
			iotaconn.AlphanetEndpointURL,
			iotaconn.AlphanetFaucetURL,
			l1starter.WaitUntilEffectsVisible,
		),
	)
}

func NewAlphanetSignerWithFunds(t *testing.T, seed []byte, index int) cryptolib.Signer {
	return newSignerWithFunds(t, seed, index, iotaconn.AlphanetEndpointURL, iotaconn.AlphanetFaucetURL)
}

func newSignerWithFunds(t *testing.T, seed []byte, index int, apiURL, faucetURL string) cryptolib.Signer {
	seedCopy := make([]byte, len(seed))
	copy(seedCopy, seed)
	seedCopy[0] += byte(index)
	kp := cryptolib.KeyPairFromSeed(cryptolib.Seed(seedCopy))
	addr := kp.Address().AsIotaAddress()
	l1Client := clients.NewL1ClientFromIotaClient(iotagraphql.NewGraphQLClient(apiURL, faucetURL))
	err := l1Client.RequestFundsFromFaucet(context.Background(), addr)
	require.NoError(t, err)

	iotatest.EnsureCoinCount(t, cryptolib.SignerToIotaSigner(kp), l1Client, ExpectedCoinCount)
	return kp
}
