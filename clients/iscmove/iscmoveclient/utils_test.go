package iscmoveclient_test

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestMain(m *testing.M) {
	flag.Parse()
	stv := l1starter.Start(context.Background(), l1starter.DefaultConfig)
	defer stv.Stop()
	m.Run()
}

func buildDeployMintTestcoin(t *testing.T, client *iscmoveclient.Client, signer cryptolib.Signer) (*iotago.ObjectRef, *iotago.ObjectInfo) {
	testcoinBytecode := contracts.Testcoin()
	iotaSigner := cryptolib.SignerToIotaSigner(signer)

	txnBytes, err := client.Publish(context.Background(), iotaclient.PublishRequest{
		Sender:          signer.Address().AsIotaAddress(),
		CompiledModules: testcoinBytecode.Modules,
		Dependencies:    testcoinBytecode.Dependencies,
		GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 10),
	})
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		iotaSigner,
		txnBytes.TxBytes,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)
	getObjectRes, err := client.GetObject(context.Background(), iotaclient.GetObjectRequest{
		ObjectID: packageID,
		Options:  &iotajsonrpc.IotaObjectDataOptions{ShowType: true},
	})
	require.NoError(t, err)
	packageRef := getObjectRes.Data.Ref()

	treasuryCapRef, err := txnResponse.GetCreatedObjectInfo("coin", "TreasuryCap")
	require.NoError(t, err)

	mintAmount := uint64(1000000)
	txnRes, err := client.MintToken(
		context.Background(),
		iotaSigner,
		packageID,
		"testcoin",
		treasuryCapRef.ObjectID,
		mintAmount,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnRes.Effects.Data.IsSuccess())

	coinRef, err := txnRes.GetCreatedObjectInfo("testcoin", "TESTCOIN")
	require.NoError(t, err)
	coinInfo := iotago.NewObjectInfo(coinRef, &iotago.ResourceType{Address: packageID, Module: "testcoin", ObjectName: "TESTCOIN"})

	return &packageRef, coinInfo
}
