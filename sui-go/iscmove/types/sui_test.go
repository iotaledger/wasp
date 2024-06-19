package types

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/fardream/go-bcs/bcs"
	"github.com/kr/pretty"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/sui-go/iscmove"
	"github.com/iotaledger/wasp/sui-go/iscmove/types/mock_contract"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/conn"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
	"github.com/iotaledger/wasp/sui-go/sui_types/serialization"
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

	iscBytecode := mock_contract.MockISCContract()

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

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	cap, _ := lo.Find(txnResponse.ObjectChanges, func(item serialization.TagJson[models.ObjectChange]) bool {
		if item.Data.Created != nil && strings.Contains(item.Data.Created.ObjectType, "TreasuryCap") {
			return true
		}

		return false
	})

	capObj, err := client.GetObject(context.Background(), &cap.Data.Created.ObjectID, nil)
	require.NoError(t, err)
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
		capObj,
	)
	require.NoError(t, err)
	require.True(t, startNewChainRes.Effects.Data.IsSuccess())
	t.Logf("StartNewChain response: %#v\n", startNewChainRes)

	return testSetup{
		suiClient: suiClient,
		signer:    signer,
		chain:     startNewChainRes,
		iscClient: client,
		packageID: *packageID,
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

	return decodedAnchor
}

func TestMinimalClient(t *testing.T) {
	setup := setupAndDeploy(t)

	anchor := GetAnchor(t, setup)
	t.Log(anchor)

	graph := iscmove.NewGraph(setup.suiClient, "http://localhost:9001")
	ret, err := graph.GetAssetBag(context.Background(), anchor.Assets.Value.ID)

	require.NoError(t, err)
	require.NotNil(t, ret)
}
