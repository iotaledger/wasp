package iotaclienttest

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iota-go/move"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
)

func DeployCoinPackage(
	t require.TestingT,
	iotaClient iotagraphql.IotaClient,
	signer iotasigner.Signer,
	packageBytecode move.PackageBytecode,
) (
	packageID *iotago.PackageID,
	treasuryCap *iotago.ObjectRef,
) {
	if th, ok := t.(interface{ Helper() }); ok {
		th.Helper()
	}

	txnBytes, err := iotaClient.Publish(
		context.Background(),
		iotagraphql.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: packageBytecode.Modules,
			Dependencies:    packageBytecode.Dependencies,
			GasBudget:       iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget * 5),
		},
	)
	require.NoError(t, err)

	txnResponse, err := iotaClient.SignAndExecuteTransaction(
		context.Background(),
		&iotagraphql.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      signer,
			Options: &iotagraphql.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, txnResponse)
	require.NotNil(t, txnResponse.Effects)
	require.True(t, txnResponse.Effects.Data.IsSuccess(), txnResponse.Effects.Data.V1.Status.Error)

	packageID, err = txnResponse.GetPublishedPackageID()
	require.NoError(t, err)
	require.NotNil(t, packageID)

	treasuryCap, err = txnResponse.GetCreatedObjectByName("coin", "TreasuryCap")
	require.NoError(t, err)
	require.NotNil(t, treasuryCap)

	return packageID, treasuryCap
}

func MintCoins(
	t require.TestingT,
	iotaClient iotagraphql.IotaClient,
	signer iotasigner.Signer,
	packageID *iotago.PackageID,
	moduleName iotago.Identifier,
	typeTag iotago.Identifier,
	treasuryCapObject *iotago.ObjectRef,
	mintAmount uint64,
) (coinRef *iotago.ObjectRef) {
	if th, ok := t.(interface{ Helper() }); ok {
		th.Helper()
	}

	resp, err := iotaClient.MintToken(
		context.Background(),
		signer,
		packageID,
		string(moduleName),
		treasuryCapObject,
		mintAmount,
		&iotagraphql.IotaTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Effects)
	require.True(t, resp.Effects.Data.IsSuccess(), resp.Effects.Data.V1.Status.Error)

	coinRef, err = resp.GetCreatedCoinByType(string(moduleName), string(typeTag))
	require.NoError(t, err)
	require.NotNil(t, coinRef)

	return coinRef
}
