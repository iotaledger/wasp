package sui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/fardream/go-bcs/bcs"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/sui-go/isc"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/conn"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
	"github.com/iotaledger/wasp/sui-go/utils"
)

type testSetup struct {
	suiClient *sui.ImplSuiAPI
	iscClient *isc.Client
	signer    *sui_signer.Signer
	packageID sui_types.PackageID
	chain     *models.SuiTransactionBlockResponse
}

func setupAndDeploy(t *testing.T) testSetup {
	suiClient, signer := sui.NewTestSuiClientWithSignerAndFund(conn.LocalnetEndpointUrl, sui_signer.TEST_MNEMONIC)
	client := isc.NewIscClient(suiClient)

	printCoinsForAddress(t, suiClient, *signer.Address)

	modules, err := utils.MoveBuild(utils.GetGitRoot() + "/contracts/move/sources")
	require.NoError(t, err)

	fmt.Printf("%s", signer.Address.String())

	txnBytes, err := client.Publish(
		context.Background(),
		signer.Address,
		modules.Modules,
		modules.Dependencies,
		nil,
		models.NewSafeSuiBigInt(uint64(100000000)),
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	t.Log("Before StartNewChain")
	printCoinsForAddress(t, suiClient, *signer.Address)

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	t.Log("packageID: ", packageID)

	startNewChainRes, err := client.StartNewChain(
		context.Background(),
		signer,
		packageID,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, startNewChainRes.Effects.Data.IsSuccess())
	t.Logf("StartNewChain response: %#v\n", startNewChainRes)

	printCoinsForAddress(t, suiClient, *signer.Address)

	return testSetup{
		suiClient: suiClient,
		signer:    signer,
		chain:     startNewChainRes,
		iscClient: client,
		packageID: *packageID,
	}
}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}

func printCoinsForAddress(t *testing.T, suiClient *sui.ImplSuiAPI, address sui_types.SuiAddress) {
	coins, err := suiClient.GetSuiCoinsOwnedByAddress(context.Background(), &address)
	require.NoError(t, err)

	t.Logf("Coins for address: %v", address.String())
	for _, v := range coins {
		t.Logf("COIN -> %v: %v (%v)", v.CoinObjectID, v.Balance, v.CoinType)
	}
}

func printGasCoinsForAddress(t *testing.T, suiClient *sui.ImplSuiAPI, address sui_types.SuiAddress) {
	coins, err := suiClient.GetCoinObjsForTargetAmount(context.Background(), &address, 10000)
	require.NoError(t, err)

	t.Logf("Gas for address: %v", address.String())
	for _, v := range coins {
		t.Logf("GAS -> %v: %v sui", v.CoinObjectID, v.Balance)
	}
}

func GetAnchor(t *testing.T, setup testSetup) Anchor {
	anchor, err := setup.suiClient.GetObject(context.Background(), &setup.chain.ObjectChanges[1].Data.Created.ObjectID, &models.SuiObjectDataOptions{
		ShowType:                true,
		ShowContent:             true,
		ShowBcs:                 true,
		ShowOwner:               true,
		ShowPreviousTransaction: true,
		ShowStorageRebate:       true,
		ShowDisplay:             true,
	})
	require.NoError(t, err)

	decodedAnchor := Anchor{}
	_, err = bcs.Unmarshal(anchor.Data.Bcs.Data.MoveObject.BcsBytes.Data(), &decodedAnchor)
	require.NoError(t, err)

	return decodedAnchor
}

func TestMinimalClient(t *testing.T) {
	setup := setupAndDeploy(t)

	suiUserClient, userSigner := sui.NewTestSuiClientWithSignerAndFund(conn.LocalnetEndpointUrl, sui_signer.TEST_CLIENT_MNEMONIC)
	iscUserClient := isc.NewIscClient(suiUserClient)

	printCoinsForAddress(t, setup.suiClient, *setup.signer.Address)
	printCoinsForAddress(t, suiUserClient, *userSigner.Address)

	printGasCoinsForAddress(t, setup.suiClient, *setup.signer.Address)
	printGasCoinsForAddress(t, suiUserClient, *userSigner.Address)

	anchor := GetAnchor(t, setup)
	t.Log(anchor)

	coins, err := setup.suiClient.GetSuiCoinsOwnedByAddress(context.Background(), userSigner.Address)
	require.NoError(t, err)

	_, err = iscUserClient.SendCoin(context.Background(), userSigner, &setup.packageID, &anchor.ID, coins[0].CoinType, coins[0].CoinObjectID, coins[1].CoinObjectID, sui.DefaultGasBudget, &models.SuiTransactionBlockResponseOptions{})
	require.NoError(t, err)

	contractCoins, err := setup.suiClient.GetSuiCoinsOwnedByAddress(context.Background(), setup.signer.Address)
	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	_, err = setup.iscClient.ReceiveCoin(context.Background(), setup.signer, &setup.packageID, &anchor.ID, coins[0].CoinType, coins[0].CoinObjectID, contractCoins[0].CoinObjectID, sui.DefaultGasBudget, &models.SuiTransactionBlockResponseOptions{})
	require.NoError(t, err)

	printCoinsForAddress(t, setup.suiClient, *setup.signer.Address)

	anchor = GetAnchor(t, setup)
	t.Log(anchor)
}
