package iscmoveclient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

var testSeed = []byte{50, 230, 119, 9, 86, 155, 106, 30, 245, 81, 234, 122, 116, 90, 172, 148, 59, 33, 88, 252, 134, 42, 231, 198, 208, 141, 209, 116, 78, 21, 216, 24}

func newSignerWithFunds(t *testing.T, seed []byte, index int) cryptolib.Signer {
	seed[0] = seed[0] + byte(index)
	kp := cryptolib.KeyPairFromSeed(cryptolib.Seed(seed))
	err := iotaclient.RequestFundsFromFaucet(context.TODO(), kp.Address().AsSuiAddress(), iotaconn.LocalnetFaucetURL)
	require.NoError(t, err)
	return kp
}

func newLocalnetClient() *iscmoveclient.Client {
	return iscmoveclient.NewHTTPClient(
		iotaconn.LocalnetEndpointURL,
		iotaconn.LocalnetFaucetURL,
	)
}

func TestKeys(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()
	iscBytecode := contracts.ISC()

	txnBytes, err := client.Publish(context.Background(), iotaclient.PublishRequest{
		Sender:          cryptolibSigner.Address().AsSuiAddress(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 10),
	})
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(cryptolibSigner),
		txnBytes.TxBytes,
		&iotajsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	fmt.Println(txnResponse)
}
