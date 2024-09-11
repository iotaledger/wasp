package iscmoveclient_test

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/sui-go/contracts"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func TestMain(m *testing.M) {
	flag.Parse()
	stv := l1starter.Start(context.Background(), l1starter.DefaultConfig)
	defer stv.Stop()
	m.Run()
}

func buildDeployMintTestcoin(t *testing.T, client *iscmoveclient.Client, signer cryptolib.Signer) (*sui.ObjectRef, *sui.ObjectInfo) {
	testcoinBytecode := contracts.Testcoin()
	suiSigner := cryptolib.SignerToSuiSigner(signer)

	txnBytes, err := client.Publish(context.Background(), suiclient.PublishRequest{
		Sender:          signer.Address().AsSuiAddress(),
		CompiledModules: testcoinBytecode.Modules,
		Dependencies:    testcoinBytecode.Dependencies,
		GasBudget:       suijsonrpc.NewBigInt(suiclient.DefaultGasBudget * 10),
	})
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		suiSigner,
		txnBytes.TxBytes,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)
	getObjectRes, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: packageID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowType: true},
	})
	require.NoError(t, err)
	packageRef := getObjectRes.Data.Ref()

	treasuryCapRef, err := txnResponse.GetCreatedObjectInfo("coin", "TreasuryCap")
	require.NoError(t, err)

	mintAmount := uint64(1000000)
	txnRes, err := client.MintToken(
		context.Background(),
		suiSigner,
		packageID,
		"testcoin",
		treasuryCapRef.ObjectID,
		mintAmount,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnRes.Effects.Data.IsSuccess())

	coinRef, err := txnRes.GetCreatedObjectInfo("testcoin", "TESTCOIN")
	require.NoError(t, err)
	coinInfo := sui.NewObjectInfo(coinRef, &sui.ResourceType{Address: packageID, Module: "testcoin", ObjectName: "TESTCOIN"})

	return &packageRef, coinInfo
}
