package iotaclienttest

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/clients/iota-go/move"
)

func DeployCoinPackage(
	t require.TestingT,
	client *iotaclient.Client,
	signer iotasigner.Signer,
	bytecode move.PackageBytecode,
) (
	packageID *iotago.PackageID,
	treasuryCap *iotago.ObjectRef,
) {
	txnBytes, err := client.Publish(
		context.Background(),
		iotaclient.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: bytecode.Modules,
			Dependencies:    bytecode.Dependencies,
			GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 10),
		},
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err = txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	treasuryCap, err = txnResponse.GetCreatedObjectInfo("coin", "TreasuryCap")
	require.NoError(t, err)

	return
}

func MintCoins(
	t require.TestingT,
	client *iotaclient.Client,
	signer iotasigner.Signer,
	packageID *iotago.PackageID,
	moduleName iotago.Identifier,
	typeTag iotago.Identifier,
	treasuryCapObjectID *iotago.ObjectID,
	mintAmount uint64,
) *iotago.ObjectRef {
	txnRes, err := client.MintToken(
		context.Background(),
		signer,
		packageID,
		moduleName,
		treasuryCapObjectID,
		mintAmount,
		&iotajsonrpc.IotaTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnRes.Effects.Data.IsSuccess())

	coinRef, err := txnRes.GetCreatedObjectInfo(moduleName, typeTag)
	require.NoError(t, err)

	return coinRef
}
