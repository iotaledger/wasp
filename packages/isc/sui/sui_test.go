package sui

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/fardream/go-bcs/bcs"
	"github.com/kr/pretty"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/sui-go/contracts"
	"github.com/iotaledger/wasp/sui-go/iscmove"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/conn"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
	"github.com/iotaledger/wasp/sui-go/sui_types/serialization"
	"github.com/iotaledger/wasp/sui-go/utils"
)

type testSetup struct {
	suiClient *sui.ImplSuiAPI
	iscClient *iscmove.Client
	signer    *sui_signer.Signer
	packageID sui_types.PackageID
	chain     *models.SuiTransactionBlockResponse
}

func setupAndDeploy(t *testing.T) testSetup {
	suiClient, signer := sui.NewSuiClient(conn.LocalnetEndpointUrl).WithSignerAndFund(sui_signer.TEST_SEED, 0)
	client := iscmove.NewClient(suiClient)

	printCoinsForAddress(t, suiClient, *signer.Address)

	iscBytecode := contracts.ISC()

	fmt.Printf("%s", signer.Address.String())

	txnBytes, err := client.Publish(
		context.Background(),
		signer.Address,
		iscBytecode.Modules,
		iscBytecode.Dependencies,
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

	printCoinsForAddress(t, suiClient, *signer.Address)

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	t.Logf("PackageID: %s", packageID.String())

	cap, _ := lo.Find(txnResponse.ObjectChanges, func(item serialization.TagJson[models.ObjectChange]) bool {
		if item.Data.Created != nil && strings.Contains(item.Data.Created.ObjectType, "TreasuryCap") {
			return true
		}

		return false
	})

	startNewChainRes, err := client.StartNewChain(
		context.Background(),
		signer,
		packageID,
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
		&cap.Data.Created.ObjectID,
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
	cap, _ := lo.Find(setup.chain.ObjectChanges, func(item serialization.TagJson[models.ObjectChange]) bool {
		if item.Data.Created != nil && strings.Contains(item.Data.Created.ObjectType, "Anchor") {
			return true
		}

		return false
	})

	anchor, err := setup.suiClient.GetObject(context.Background(), &cap.Data.Created.ObjectID, &models.SuiObjectDataOptions{
		ShowType:    true,
		ShowContent: true,
		ShowBcs:     true,
		ShowDisplay: true,
	})
	require.NoError(t, err)

	t.Logf("%# v\n", pretty.Formatter(anchor.Data.Content.Data))

	decodedAnchor := Anchor{}
	_, err = bcs.Unmarshal(anchor.Data.Bcs.Data.MoveObject.BcsBytes.Data(), &decodedAnchor)

	fmt.Printf("BCS Data Anchor: %v", hex.EncodeToString(anchor.Data.Bcs.Data.MoveObject.BcsBytes.Data()))
	t.Logf("%# v\n", pretty.Formatter(decodedAnchor))
	require.NoError(t, err)

	return decodedAnchor
}

func TestAnchorDeserialization(t *testing.T) {
	anchorBCSDataHex := "9722e2a90361273cac7b5c652b7e65a356ca53088e2da187cda8ec2732739a9da35a5f09f2a93b817c99afe64455babfb90e92774baa30db6e271691ddcf9d5e010f985f3bde360af527ff5caffd4103124f02f5a16ceb86d55bc7eb1369e32b8d020000000000000000"
	anchorBCSBytes, err := hex.DecodeString(anchorBCSDataHex)
	require.NoError(t, err)

	decodedAnchor := Anchor{}
	_, err = bcs.Unmarshal(anchorBCSBytes, &decodedAnchor)
	t.Logf("%# v\n", pretty.Formatter(decodedAnchor))
	require.NoError(t, err) // This will fail
}

func TestMinimalClient(t *testing.T) {
	setup := setupAndDeploy(t)

	suiUserClient, userSigner := sui.NewSuiClient(conn.LocalnetEndpointUrl).WithSignerAndFund(sui_signer.TEST_SEED, 1)
	iscUserClient := iscmove.NewClient(suiUserClient)

	printCoinsForAddress(t, setup.suiClient, *setup.signer.Address)
	printCoinsForAddress(t, suiUserClient, *userSigner.Address)

	printGasCoinsForAddress(t, setup.suiClient, *setup.signer.Address)
	printGasCoinsForAddress(t, suiUserClient, *userSigner.Address)

	anchor := GetAnchor(t, setup)
	t.Log(anchor)

	coins, err := setup.suiClient.GetSuiCoinsOwnedByAddress(context.Background(), userSigner.Address)
	require.NoError(t, err)

	_, err = iscUserClient.SendCoin(context.Background(), userSigner, &setup.packageID, &anchor.ID, coins[0].CoinType, coins[0].CoinObjectID, nil, sui.DefaultGasPrice, sui.DefaultGasBudget, &models.SuiTransactionBlockResponseOptions{})
	require.NoError(t, err)

	time.Sleep(3 * time.Second)

	_, err = setup.iscClient.ReceiveCoin(context.Background(), setup.signer, &setup.packageID, &anchor.ID, coins[0].CoinType, coins[0].CoinObjectID, nil, sui.DefaultGasPrice, sui.DefaultGasBudget, &models.SuiTransactionBlockResponseOptions{})
	require.NoError(t, err)

	printCoinsForAddress(t, setup.suiClient, *setup.signer.Address)

	anchor = GetAnchor(t, setup)
	t.Log(anchor)
}
