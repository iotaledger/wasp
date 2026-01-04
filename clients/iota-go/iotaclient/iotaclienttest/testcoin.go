package iotaclienttest

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/client"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iota-go/move"
)

func DeployCoinPackage(
	t require.TestingT,
	iotaClient client.IotaClient,
	signer iotasigner.Signer,
	bytecode move.PackageBytecode,
) (
	packageID *iotago.PackageID,
	treasuryCap *iotago.ObjectRef,
) {
	var modules [][]byte
	for _, m := range bytecode.Modules {
		modules = append(modules, m.Data())
	}
	ptb := iotago.NewProgrammableTransactionBuilder()
	argContract := ptb.Command(iotago.Command{Publish: &iotago.ProgrammablePublish{
		Modules:      modules,
		Dependencies: bytecode.Dependencies,
	}})
	ptb.Command(iotago.Command{TransferObjects: &iotago.ProgrammableTransferObjects{
		Objects: []iotago.Argument{argContract},
		Address: ptb.MustPure(signer.Address()),
	}})
	pt := ptb.Finish()

	txnResponse, err := iotaClient.SignAndExecuteTxWithRetry(
		context.Background(),
		signer,
		pt,
		nil,
		iotaclient.DefaultGasBudget*2,
		iotaclient.DefaultGasPrice,
		&iotagraphql.IotaTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		})
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err = txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	treasuryCap, err = txnResponse.GetCreatedObjectByName("coin", "TreasuryCap")
	require.NoError(t, err)

	return
}

func MintCoins(
	t require.TestingT,
	iotaClient client.IotaClient,
	signer iotasigner.Signer,
	packageID *iotago.PackageID,
	moduleName iotago.Identifier,
	typeTag iotago.Identifier,
	treasuryCapObjectID *iotago.ObjectRef,
	mintAmount uint64,
) *iotago.ObjectRef {
	txnRes, err := iotaClient.MintToken(
		context.Background(),
		signer,
		packageID,
		moduleName,
		treasuryCapObjectID,
		mintAmount,
		&iotagraphql.IotaTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnRes.Effects.Data.IsSuccess())
	time.Sleep(1 * time.Second)

	coinRef, err := txnRes.GetCreatedObjectByName(moduleName, typeTag)
	require.NoError(t, err)

	return coinRef
}
