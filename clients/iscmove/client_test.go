package iscmove_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/conn"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

const (
	SEEDFORUSER  = "0x1234"
	SEEDFORCHAIN = "0x5678"
)

// func newClient(_ *testing.T) *iscmove.Client {
// 	// NOTE: comment out the next line to run local tests against sui-test-validator
// 	// t.Skip("only for localnet")
// 	return iscmove.NewClient(sui.NewSuiClient(conn.LocalnetEndpointUrl))
// }

func TestStartNewChain(t *testing.T) {
	suiClient, signer := sui.NewSuiClient(conn.LocalnetEndpointUrl).WithSignerAndFund(sui_signer.TEST_SEED, 0)
	client := iscmove.NewClient(suiClient)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	startNewChainRes, err := client.StartNewChain(
		context.Background(),
		signer,
		iscPackageID,
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
		nil,
	)
	require.NoError(t, err)
	require.True(t, startNewChainRes.Effects.Data.IsSuccess())

	anchorRef, _, err := sui.GetCreatedObjectRefAndType(startNewChainRes, "anchor", "Anchor")
	require.NoError(t, err)
	t.Log("anchorObjID: ", anchorRef.ObjectID)

	// resAnchor, err := client.GetObject(context.Background(), anchorObjID, &models.SuiObjectDataOptions{ShowContent: true})
	// require.NoError(t, err)

	// var anchorFieldsRaw iscmove.AnchorFieldsRaw
	// err = json.Unmarshal(resAnchor.Data.Content.Data.MoveObject.Fields, &anchorFieldsRaw)

	// resAssets, err := client.GetObject(context.Background(), sui_types.MustObjectIDFromHex(anchorFieldsRaw.ID.ID), &models.SuiObjectDataOptions{ShowContent: true})
	// assetsRef := sui_types.ObjectRef{
	// 	ObjectID: resAssets.Data.ObjectID,
	// 	Version:  resAssets.Data.Version.Uint64(),
	// 	Digest:   resAssets.Data.Digest,
	// }
}

type ReceivedRequest struct {
	objectID  *sui_types.SuiAddress
	sender    *sui_types.SuiAddress
	anchorID  *sui_types.ObjectID
	contract  string
	function  string
	args      [][]byte
	allowance []byte
}

func TestSendReceiveRequest(t *testing.T) {
	suiClient, _ := sui.NewSuiClient(conn.LocalnetEndpointUrl).WithSignerAndFund(sui_signer.TEST_SEED, 0)
	client := iscmove.NewClient(suiClient)
	_, chainSigner := sui.NewSuiClient(conn.LocalnetEndpointUrl).WithSignerAndFund(sui_signer.TEST_SEED, 1)

	iscPackageID := buildAndDeployISCContracts(t, client, chainSigner)

	// eventCh := subscribeEvents(t, iscPackageID)

	anchorObjID := startChainAnchor(t, client, chainSigner, iscPackageID)
	resAnchor, err := client.GetObject(context.Background(), anchorObjID, &models.SuiObjectDataOptions{ShowContent: true})
	require.NoError(t, err)

	var anchorFieldsRaw iscmove.AnchorFieldsRaw
	err = json.Unmarshal(resAnchor.Data.Content.Data.MoveObject.Fields, &anchorFieldsRaw)
	require.NoError(t, err)
	fmt.Printf("anchorFieldsRaw: %v\n", anchorFieldsRaw)
	resAssets, err := client.GetDynamicFieldObject(context.Background(), sui_types.MustObjectIDFromHex(anchorFieldsRaw.ID.ID), nil)
	require.NoError(t, err)
	assetsRef := &sui_types.ObjectRef{
		ObjectID: resAssets.Data.ObjectID,
		Version:  resAssets.Data.Version.Uint64(),
		Digest:   resAssets.Data.Digest,
	}
	sendReqRes, err := client.CreateAndSendRequest(
		context.Background(),
		chainSigner,
		iscPackageID,
		anchorObjID,
		assetsRef,
		"isc_test_contract_name",
		"isc_test_func_name",
		[][]byte{[]byte("one"), []byte("two"), []byte("three")}, // func input
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	fmt.Println("sendReqRes.Effects.Data.V1.Status.Error: ", sendReqRes)
	require.True(t, sendReqRes.Effects.Data.IsSuccess())
	// require.NoError(t, err)
	// reqObjID, reqType, err := sui.GetCreatedObjectRefAndType(createReqRes, "request", "Request")
	// require.NoError(t, err)
	// getObjectRes, err := client.GetObject(context.Background(), reqObjID, &models.SuiObjectDataOptions{ShowOwner: true})
	// require.NoError(t, err)
	// require.Equal(t, signer.Address, getObjectRes.Data.Owner.AddressOwner)
	// sendRequest(t, client, signer, iscPackageID, anchorObjID, reqObjID)

	// getObjectRes, err = client.GetObject(context.Background(), reqObjID, &models.SuiObjectDataOptions{ShowOwner: true})
	// require.NoError(t, err)
	// require.Equal(t, anchorObjID, getObjectRes.Data.Owner.AddressOwner)
	// sender, object := receiveEvent(t, client, eventCh)
	// require.Equal(t, signer.Address, sender)
	// req := object.Data.Content.Data.MoveObject
	// require.True(t, strings.HasSuffix(req.Type, "::request::Request"))
	// require.Equal(t, reqType, req.Type)

	// type Fields struct {
	// 	Data struct {
	// 		Fields struct {
	// 			Contract string
	// 			Function string
	// 			Args     []interface{}
	// 		}
	// 	}
	// }
	// var fields Fields
	// err = json.Unmarshal(req.Fields, &fields)
	// require.NoError(t, err)
	// args := make([][]byte, len(fields.Data.Fields.Args))
	// for i, argField := range fields.Data.Fields.Args {
	// 	argFieldBytes := argField.([]interface{})
	// 	arg := make([]byte, len(argFieldBytes))
	// 	for j, argFieldByte := range argFieldBytes {
	// 		arg[j] = byte(argFieldByte.(float64))
	// 	}
	// 	args[i] = arg
	// }

	// // NOTE: this is the data that ISC should use as the request from the sender
	// receivedRequest := &ReceivedRequest{
	// 	objectID:  object.Data.ObjectID,
	// 	sender:    sender,
	// 	anchorID:  anchorObjID,
	// 	contract:  fields.Data.Fields.Contract,
	// 	function:  fields.Data.Fields.Function,
	// 	args:      args,
	// 	allowance: nil,
	// }
	// require.Equal(t, "isc_test_contract_name", receivedRequest.contract)
	// require.Equal(t, "isc_test_func_name", receivedRequest.function)
	// require.EqualValues(t, 3, len(receivedRequest.args))
	// require.Equal(t, "one", string(receivedRequest.args[0]))
	// require.Equal(t, "two", string(receivedRequest.args[1]))
	// require.Equal(t, "three", string(receivedRequest.args[2]))
	// require.Nil(t, receivedRequest.allowance)

	// receiveRequest(t, client, chainSigner, iscPackageID, anchorObjID, reqObjID)
}

// func newSignerWithFunds(t *testing.T, seed string) *sui_signer.Signer {
// 	signer := sui_signer.NewSigner(sui_types.MustSuiAddressFromHex(seed)[:], sui_signer.KeySchemeFlagIotaEd25519)
// 	err := sui.RequestFundFromFaucet(signer.Address, conn.LocalnetFaucetUrl)
// 	require.NoError(t, err)
// 	return signer
// }

func startChainAnchor(
	t *testing.T,
	client *iscmove.Client,
	signer cryptolib.Signer,
	iscPackageID *sui_types.PackageID,
) *sui_types.ObjectID {
	// start a new chain
	startNewChainRes, err := client.StartNewChain(
		context.Background(),
		signer,
		iscPackageID,
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
		nil,
	)
	require.NoError(t, err)

	// 	for _, change:= {
	// 		startNewChainRes
	// 	}
	// fmt.Println("")

	anchorRef, _, err := sui.GetCreatedObjectRefAndType(startNewChainRes, "anchor", "Anchor")
	require.NoError(t, err)
	return anchorRef.ObjectID
}

func subscribeEvents(t *testing.T, iscPackageID *sui_types.PackageID) chan models.SuiEvent {
	api := sui.NewSuiWebsocketClient(conn.LocalnetWebsocketEndpointUrl)
	eventCh := make(chan models.SuiEvent)
	err := api.SubscribeEvent(
		context.TODO(),
		&models.EventFilter{
			Package: iscPackageID,
		},
		eventCh,
	)
	require.NoError(t, err)
	return eventCh
}

// func receiveEvent(t *testing.T, client *iscmove.Client, eventCh chan models.SuiEvent) (
// 	*sui_types.SuiAddress,
// 	*models.SuiObjectResponse,
// ) {
// 	event := <-eventCh
// 	fmt.Println("event: ", event)
// 	close(eventCh)

// 	eventId := sui_types.MustSuiAddressFromHex(event.ParsedJson.(map[string]interface{})["id"].(string))

// 	object, err := client.GetObject(
// 		context.Background(),
// 		eventId,
// 		&models.SuiObjectDataOptions{ShowContent: true},
// 	)
// 	require.NoError(t, err)
// 	require.Equal(t, eventId, object.Data.ObjectID)
// 	fmt.Println("object: ", object.Data.Content.Data.MoveObject)
// 	return event.Sender, object
// }

// func createAndSendRequest(
// 	t *testing.T,
// 	client *iscmove.Client,
// 	signer *sui_signer.Signer,
// 	iscPackageID *sui_types.PackageID,
// 	anchorObjID *sui_types.ObjectID,
// 	reqObjID *sui_types.ObjectID,
// ) {
// 	sendReqRes, err := client.SendRequest(
// 		context.Background(),
// 		signer,
// 		iscPackageID,
// 		anchorObjID,
// 		reqObjID,
// 		nil,
// 		sui.DefaultGasPrice,
// 		sui.DefaultGasBudget,
// 		&models.SuiTransactionBlockResponseOptions{
// 			ShowEffects:       true,
// 			ShowObjectChanges: true,
// 		},
// 	)
// 	require.NoError(t, err)
// 	require.True(t, sendReqRes.Effects.Data.IsSuccess())
// }

// func receiveRequest(
// 	t *testing.T,
// 	client *iscmove.Client,
// 	signer *sui_signer.Signer,
// 	iscPackageID *sui_types.PackageID,
// 	anchorObjID *sui_types.ObjectID,
// 	reqObjID *sui_types.ObjectID,
// ) {
// 	receiveReqRes, err := client.ReceiveRequest(
// 		context.Background(),
// 		signer,
// 		iscPackageID,
// 		anchorObjID,
// 		reqObjID,
// 		nil,
// 		sui.DefaultGasPrice,
// 		sui.DefaultGasBudget,
// 		&models.SuiTransactionBlockResponseOptions{
// 			ShowEffects:       true,
// 			ShowObjectChanges: true,
// 		},
// 	)
// 	require.NoError(t, err)
// 	require.True(t, receiveReqRes.Effects.Data.IsSuccess())
// }
