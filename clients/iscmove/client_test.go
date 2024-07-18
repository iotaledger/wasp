package iscmove_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/contracts"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

var testSeed = []byte{50, 230, 119, 9, 86, 155, 106, 30, 245, 81, 234, 122, 116, 90, 172, 148, 59, 33, 88, 252, 134, 42, 231, 198, 208, 141, 209, 116, 78, 21, 216, 24}

func newSignerWithFunds(t *testing.T, seed []byte, index int) cryptolib.Signer {
	seed[0] = seed[0] + byte(index)
	kp := cryptolib.KeyPairFromSeed(cryptolib.Seed(seed))
	err := suiclient.RequestFundsFromFaucet(context.TODO(), kp.Address().AsSuiAddress(), suiconn.LocalnetFaucetURL)
	require.NoError(t, err)
	return kp
}

func newLocalnetClient() *iscmove.Client {
	return iscmove.NewHTTPClient(
		suiconn.LocalnetEndpointURL,
		"", // graphURL
		suiconn.LocalnetFaucetURL,
	)
}

func TestKeys(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()
	iscBytecode := contracts.ISC()

	txnBytes, err := client.Publish(context.Background(), suiclient.PublishRequest{
		Sender:          cryptolibSigner.Address().AsSuiAddress(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       suijsonrpc.NewBigInt(suiclient.DefaultGasBudget * 10),
	})
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		cryptolib.SignerToSuiSigner(cryptolibSigner),
		txnBytes.TxBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	fmt.Println(txnResponse)
}
