package iscmoveclient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient/iscmoveclienttest"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

type PTBTestWrapperRequest struct {
	Client      *iscmoveclient.Client
	Signer      cryptolib.Signer
	PackageID   iotago.PackageID
	GasPayments []*iotago.ObjectRef // optional
	GasPrice    uint64
	GasBudget   uint64
}

func PTBTestWrapper(
	req *PTBTestWrapperRequest,
	f func(ptb *iotago.ProgrammableTransactionBuilder) *iotago.ProgrammableTransactionBuilder,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	return req.Client.SignAndExecutePTB(
		context.Background(),
		req.Signer,
		f(ptb).Finish(),
		req.GasPayments,
		req.GasPrice,
		req.GasBudget,
	)
}

func TestKeys(t *testing.T) {
	cryptolibSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	client := iscmoveclienttest.NewHTTPClient()
	iscBytecode := contracts.ISC()

	txnBytes, err := client.Publish(context.Background(), iotaclient.PublishRequest{
		Sender:          cryptolibSigner.Address().AsIotaAddress(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 10),
	})
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      cryptolib.SignerToIotaSigner(cryptolibSigner),
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	require.NoError(t, err)
	fmt.Println(txnResponse)
}

type CompiledMoveModules struct {
	Modules      []*iotago.Base64Data `json:"modules"`
	Dependencies []*iotago.Address    `json:"dependencies"`
	Digest       []int                `json:"digest"`
}

func TestBuildISCContract(t *testing.T) {
	var err error
	cmd := exec.Command("iotago", "move", "build", "--dump-bytecode-as-base64")
	// TODO skip to fetch latest deps if there is no internet
	// cmd := exec.Command("iotago", "move", "build", "--dump-bytecode-as-base64", "--skip-fetch-latest-git-deps")
	cmd.Dir = "clients/iota-go/contracts/isc/Move.toml"

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	require.NoError(t, err)

	var modules CompiledMoveModules
	err = json.Unmarshal(stdout.Bytes(), &modules)
	require.NoError(t, err)

	client := iscmoveclient.NewClient(iotaclient.NewHTTP("https://api.iota-rebased-alphanet.iota.cafe", iotaclient.WaitForEffectsDisabled), "https://faucet.iota-rebased-alphanet.iota.cafe/gas")
	cryptolibSigner := iscmoveclienttest.NewSignerWithFunds(t, testcommon.TestSeed, 0)
	signer := cryptolib.SignerToIotaSigner(cryptolibSigner)

	txnBytes, err := client.Publish(
		context.Background(),
		iotaclient.PublishRequest{
			Sender:          cryptolibSigner.Address().AsIotaAddress(),
			CompiledModules: modules.Modules,
			Dependencies:    modules.Dependencies,
			GasBudget:       iotajsonrpc.NewBigInt(10 * iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(),
		&iotaclient.SignAndExecuteTransactionRequest{
			Signer:      signer,
			TxDataBytes: txnBytes.TxBytes,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	if err != nil || !txnResponse.Effects.Data.IsSuccess() {
		panic(err)
	}
	packageId, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: signer.Address().AsIotaAddress()})
	require.NoError(t, err)

	anchor1, err := client.StartNewChain(
		context.Background(),
		&iscmoveclient.StartNewChainRequest{
			Signer:            cryptolibSigner,
			ChainOwnerAddress: cryptolibSigner.Address(),
			PackageID:         *packageId,
			StateMetadata:     []byte{1, 2, 3, 4},
			InitCoinRef:       getCoinsRes.Data[1].Ref(),
			GasPrice:          iotaclient.DefaultGasPrice,
			GasBudget:         iotaclient.DefaultGasBudget,
		},
	)

	txnResponse, err = newAssetsBag(client, cryptolibSigner)
	require.NoError(t, err)
	sentAssetsBagRef, err := txnResponse.GetCreatedObjectInfo(iscmove.AssetsBagModuleName, iscmove.AssetsBagObjectName)
	require.NoError(t, err)

	getCoinsRes, err = client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: cryptolibSigner.Address().AsIotaAddress()})
	require.NoError(t, err)

	_, err = assetsBagPlaceCoinAmountWithGasCoin(
		client,
		cryptolibSigner,
		sentAssetsBagRef,
		iotajsonrpc.IotaCoinType,
		10,
	)
	require.NoError(t, err)

	sentAssetsBagRef, err = client.UpdateObjectRef(context.Background(), sentAssetsBagRef)
	require.NoError(t, err)

}
