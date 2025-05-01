// Package iscmoveclienttest provides testing utilities for the ISC move client.
package iscmoveclienttest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func NewSignerWithFunds(t *testing.T, seed []byte, index int) cryptolib.Signer {
	seed[0] += byte(index)
	kp := cryptolib.KeyPairFromSeed(cryptolib.Seed(seed))
	err := iotaclient.RequestFundsFromFaucet(context.Background(), kp.Address().AsIotaAddress(), l1starter.Instance().FaucetURL())
	require.NoError(t, err)
	return kp
}

func NewRandomSignerWithFunds(t *testing.T, index int) cryptolib.Signer {
	seed := cryptolib.NewSeed()
	return NewSignerWithFunds(t, seed[:], index)
}

func NewWebSocketClient(ctx context.Context, log log.Logger) (*iscmoveclient.Client, error) {
	if l1starter.IsLocalConfigured() {
		panic("Right now no WS support")
	}

	return iscmoveclient.NewWebsocketClient(
		ctx,
		iotaconn.AlphanetWebsocketEndpointURL,
		l1starter.Instance().FaucetURL(),
		l1starter.WaitUntilEffectsVisible,
		log,
	)
}

func NewHTTPClient() *iscmoveclient.Client {
	return iscmoveclient.NewHTTPClient(
		l1starter.Instance().APIURL(),
		l1starter.Instance().FaucetURL(),
		l1starter.WaitUntilEffectsVisible,
	)
}
