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
	"github.com/iotaledger/wasp/sui-go/suisigner"
)

func newSignerWithFunds(t *testing.T, seed []byte, index int) cryptolib.Signer {
	seed[0] = seed[0] + byte(index)
	kp := cryptolib.KeyPairFromSeed(cryptolib.Seed(seed))
	err := suiclient.RequestFundsFromFaucet(context.TODO(), kp.Address().AsSuiAddress(), suiconn.LocalnetFaucetURL)
	require.NoError(t, err)
	return kp
}

func TestKeys(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, suisigner.TestSeed, 0)
	client := iscmove.NewClient(iscmove.Config{
		APIURL: suiconn.LocalnetEndpointURL,
	})
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
