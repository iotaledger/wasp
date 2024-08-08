package iscmoveclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func TestAllowanceNewAndDestroyEmpty(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	txnResponse, err := client.AllowanceNew(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	allowanceRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AllowanceModuleName, iscmove.AllowanceObjectName)
	require.NoError(t, err)

	allowanceDestroyRes, err := client.AllowanceDestroy(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		allowanceRef,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	_, err = allowanceDestroyRes.GetCreatedObjectInfo(iscmove.AllowanceModuleName, iscmove.AllowanceObjectName)
	require.Error(t, err, "not found")
}

func TestAllowanceAddItems(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	txnResponse, err := client.AllowanceNew(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	allowanceMainRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AllowanceModuleName, iscmove.AllowanceObjectName)
	require.NoError(t, err)

	_, coinInfo := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		suiclient.GetObjectRequest{
			ObjectID: coinInfo.Ref.ObjectID,
			Options:  &suijsonrpc.SuiObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := sui.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := suijsonrpc.CoinType(coinResource.SubType.String())
	_, err = client.AllowanceWithCoinBalance(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		allowanceMainRef,
		100,
		testCointype,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
}

func TestGetAllowanceFromAllowanceID(t *testing.T) {
	cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	client := newLocalnetClient()

	iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	txnResponse, err := client.AllowanceNew(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	allowanceMainRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AllowanceModuleName, iscmove.AllowanceObjectName)
	require.NoError(t, err)

	_, coinInfo := buildDeployMintTestcoin(t, client, cryptolibSigner)
	getCoinRef, err := client.GetObject(
		context.Background(),
		suiclient.GetObjectRequest{
			ObjectID: coinInfo.Ref.ObjectID,
			Options:  &suijsonrpc.SuiObjectDataOptions{ShowType: true},
		},
	)
	require.NoError(t, err)

	coinResource, err := sui.NewResourceType(*getCoinRef.Data.Type)
	require.NoError(t, err)
	testCointype := suijsonrpc.CoinType(coinResource.SubType.String())
	_, err = client.AllowanceWithCoinBalance(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		allowanceMainRef,
		100,
		testCointype,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	allowanceGetObjectRes, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: allowanceMainRef.ObjectID,
	})
	require.NoError(t, err)

	allowanceMainRefTmp := allowanceGetObjectRes.Data.Ref()

	_, err = client.AllowanceWithCoinBalance(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		&allowanceMainRefTmp,
		111,
		suijsonrpc.SuiCoinType,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)

	allowance, err := client.GetAllowance(context.Background(), allowanceMainRef.ObjectID)
	require.NoError(t, err)
	for i := range allowance.CoinTypes {
		if allowance.CoinTypes[i] == suijsonrpc.SuiCoinType {
			require.Equal(t, 111, allowance.CoinAmounts[i])
		}
		if allowance.CoinTypes[i] == testCointype {
			require.Equal(t, 100, allowance.CoinAmounts[i])
		}
	}
}

func TestGetAllowanceFromRequestID(t *testing.T) {
	// cryptolibSigner := newSignerWithFunds(t, testSeed, 0)
	// client := newLocalnetClient()

	// iscPackageID := buildAndDeployISCContracts(t, client, cryptolibSigner)

	// anchor := startNewChain(t, client, cryptolibSigner, iscPackageID)

	// _, testcoinInfo := buildDeployMintTestcoin(t, client, cryptolibSigner)
	// getCoinRef, err := client.GetObject(
	// 	context.Background(),
	// 	suiclient.GetObjectRequest{
	// 		ObjectID: testcoinInfo.Ref.ObjectID,
	// 		Options:  &suijsonrpc.SuiObjectDataOptions{ShowType: true},
	// 	},
	// )
	// require.NoError(t, err)

	// coinResource, err := sui.NewResourceType(*getCoinRef.Data.Type)
	// require.NoError(t, err)
	// testCointype := suijsonrpc.CoinType(coinResource.SubType.String())

	// txnResponse, err := client.AllowanceNew(
	// 	context.Background(),
	// 	cryptolibSigner,
	// 	iscPackageID,
	// 	nil,
	// 	suiclient.DefaultGasPrice,
	// 	suiclient.DefaultGasBudget,
	// 	false,
	// )
	// require.NoError(t, err)
	// allowanceRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AllowanceModuleName, iscmove.AllowanceObjectName)
	// require.NoError(t, err)

	// _, err = client.AllowanceWithCoinBalance(
	// 	context.Background(),
	// 	cryptolibSigner,
	// 	iscPackageID,
	// 	allowanceRef,
	// 	testcoinInfo.Ref,
	// 	testCointype,
	// 	nil,
	// 	suiclient.DefaultGasPrice,
	// 	suiclient.DefaultGasBudget,
	// 	false,
	// )
	// require.NoError(t, err)

	// allowanceGetObjectRes, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{ObjectID: allowanceRef.ObjectID})
	// require.NoError(t, err)
	// tmpAllowanceRef := allowanceGetObjectRes.Data.Ref()
	// createAndSendRequestRes, err := client.CreateAndSendRequest(
	// 	context.Background(),
	// 	cryptolibSigner,
	// 	iscPackageID,
	// 	anchor.ObjectID,
	// 	&tmpAllowanceRef,
	// 	uint32(isc.Hn("test_isc_contract")),
	// 	uint32(isc.Hn("test_isc_func")),
	// 	[][]byte{[]byte("one"), []byte("two"), []byte("three")},
	// 	nil,
	// 	suiclient.DefaultGasPrice,
	// 	suiclient.DefaultGasBudget,
	// 	false,
	// )
	// require.NoError(t, err)

	// reqRef, err := createAndSendRequestRes.GetCreatedObjectInfo(iscmove.RequestModuleName, iscmove.RequestObjectName)
	// require.NoError(t, err)

	// req, err := client.GetRequestFromObjectID(context.Background(), reqRef.ObjectID)
	// require.NoError(t, err)

	// allowance, err := client.GetAllowanceWithBalances(context.Background(), &req.Allowance.Value.ID)
	// require.NoError(t, err)
	// require.Equal(t, uint64(1), allowance.Size)
	// bal, ok := allowance.Balances[testCointype]
	// require.True(t, ok)
	// require.Equal(t, testCointype, bal.CoinType)
	// require.Equal(t, uint64(1000000), bal.TotalBalance.Uint64())
}

func createEmptyAllowance(t *testing.T, client *iscmoveclient.Client, cryptolibSigner cryptolib.Signer, iscPackageID sui.PackageID) *sui.ObjectRef {
	txnResponse, err := client.AllowanceNew(
		context.Background(),
		cryptolibSigner,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		false,
	)
	require.NoError(t, err)
	allowanceRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AllowanceModuleName, iscmove.AllowanceObjectName)
	require.NoError(t, err)
	return allowanceRef
}
