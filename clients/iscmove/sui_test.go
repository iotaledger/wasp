package iscmove

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

	"github.com/iotaledger/wasp/clients/iscmove/mock_contract"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suisigner"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/serialization"
	"github.com/iotaledger/wasp/sui-go/suiconn"
)

type testSetup struct {
	iscClient *Client
	signer    cryptolib.Signer
	packageID sui.PackageID
	chain     *Anchor
}

func setupAndDeploy(t *testing.T) testSetup {
	client := NewClient(
		Config{
			APIURL:       suiconn.LocalnetEndpointURL,
			FaucetURL:    suiconn.LocalnetFaucetURL,
			WebsocketURL: suiconn.LocalnetWebsocketEndpointURL,
		},
	)

	kp := cryptolib.KeyPairFromSeed(cryptolib.SubSeed(suisigner.TestSeed, 0))

	iscBytecode := mock_contract.MockISCContract()

	fmt.Printf("%s", kp.Address().String())
	txnBytes, err := client.Publish(context.Background(), suiclient.PublishRequest{
		Sender:          kp.Address().AsSuiAddress(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       suijsonrpc.NewBigInt(uint64(100000000)),
	})
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), cryptolib.SignerToSuiSigner(kp), txnBytes.TxBytes, &suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	createdCap, _ := lo.Find(
		txnResponse.ObjectChanges, func(item serialization.TagJson[suijsonrpc.ObjectChange]) bool {
			if item.Data.Created != nil && strings.Contains(item.Data.Created.ObjectType, "TreasuryCap") {
				return true
			}

			return false
		},
	)

	capObj, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: &createdCap.Data.Created.ObjectID,
	})
	require.NoError(t, err)
	startNewChainRes, err := client.StartNewChain(
		context.Background(),
		kp,
		packageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
		capObj,
	)
	require.NoError(t, err)
	t.Logf("StartNewChain response: %#v\n", startNewChainRes)

	return testSetup{
		signer:    kp,
		chain:     startNewChainRes,
		iscClient: client,
		packageID: *packageID,
	}
}

func GetAnchor(t *testing.T, setup testSetup) Anchor {
	anchor, err := setup.iscClient.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: &setup.chain.ID,
		Options: &suijsonrpc.SuiObjectDataOptions{
			ShowType:    true,
			ShowContent: true,
			ShowBcs:     true,
			ShowDisplay: true,
		},
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

	graph := NewGraph("http://localhost:9001")
	ret, err := graph.GetAssetBag(context.Background(), anchor.Assets.Value.ID)

	require.NoError(t, err)
	require.NotNil(t, ret)
}
