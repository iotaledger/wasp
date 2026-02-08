// Package iscmoveclienttest provides testing utilities for the ISC move client.
package iscmoveclienttest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func NewSignerWithFunds(t *testing.T, seed []byte, index int) cryptolib.Signer {
	return newSignerWithFunds(t, seed, index, l1starter.Instance().FaucetURL(), l1starter.Instance().APIURL())
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

func NewHTTPClient() *iscmoveclient.Client {
	return iscmoveclient.NewHTTPClient(
		l1starter.Instance().APIURL(),
		l1starter.Instance().FaucetURL(),
		l1starter.WaitUntilEffectsVisible,
	)
}

func NewAlphanetHTTPClient() *iscmoveclient.Client {
	return iscmoveclient.NewHTTPClient(
		iotaconn.AlphanetEndpointURL,
		iotaconn.AlphanetFaucetURL,
		l1starter.WaitUntilEffectsVisible,
	)
}

func NewAlphanetSignerWithFunds(t *testing.T, seed []byte, index int) cryptolib.Signer {
	return newSignerWithFunds(t, seed, index, iotaconn.AlphanetFaucetURL, iotaconn.AlphanetEndpointURL)
}

func newSignerWithFunds(t *testing.T, seed []byte, index int, faucetURL, apiURL string) cryptolib.Signer {
	seed[0] += byte(index)
	kp := cryptolib.KeyPairFromSeed(cryptolib.Seed(seed))
	err := iotagraphql.RequestFundsFromFaucetAndWait(context.Background(), kp.Address().AsIotaAddress(), faucetURL, apiURL)
	require.NoError(t, err)
	return kp
}
