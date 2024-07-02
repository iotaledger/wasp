package iscmove_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

const (
	SEEDFORUSER  = "0x1234"
	SEEDFORCHAIN = "0x5678"
)

func newClient(_ *testing.T) *iscmove.Client {
	// NOTE: comment out the next line to run local tests against sui-test-validator
	// t.Skip("only for localnet")
	return iscmove.NewClient(
		iscmove.Config{
			APIURL:       suiconn.LocalnetEndpointURL,
			FaucetURL:    suiconn.LocalnetFaucetURL,
			WebsocketURL: suiconn.LocalnetWebsocketEndpointURL,
		},
	)
}

func TestStartNewChain(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)
	t.Log("anchorObjID: ", anchorObjID)
}

func TestSendCoin(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	tokenPackageID, _ := buildDeployMintTestcoin(t, client, signer)
	coinType := fmt.Sprintf("%s::testcoin::TESTCOIN", tokenPackageID.String())
	fmt.Printf("coin type: %s\n", coinType)

	// the signer should have only one coin object which belongs to testcoin type
	coins, err := client.GetCoins(context.Background(), suiclient.GetCoinsRequest{
		Owner:    signer.Address().AsSuiAddress(),
		CoinType: &coinType,
		Limit:    10,
	})
	require.NoError(t, err)
	require.Len(t, coins, 1)

	sendCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins.Data[0].CoinObjectID)

	getObjectRes, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: coins.Data[0].CoinObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, anchorObjID.String(), getObjectRes.Data.Owner.ObjectOwnerInternal.AddressOwner.String())
}

func TestReceiveCoin(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	tokenPackageID, _ := buildDeployMintTestcoin(t, client, signer)
	coinType := fmt.Sprintf("%s::testcoin::TESTCOIN", tokenPackageID.String())
	fmt.Printf("coin type: %s\n", coinType)

	// the signer should have only one coin object which belongs to testcoin type
	coins, err := client.GetCoins(context.Background(), suiclient.GetCoinsRequest{
		Owner:    signer.Address().AsSuiAddress(),
		CoinType: &coinType,
		Limit:    10,
	})
	require.NoError(t, err)

	sendCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins.Data[0].CoinObjectID)

	getObjectRes, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: coins.Data[0].CoinObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, anchorObjID.String(), getObjectRes.Data.Owner.ObjectOwnerInternal.AddressOwner.String())

	receiveCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins.Data[0].CoinObjectID)

	assets, err := client.GetAssets(context.Background(), iscPackageID, anchorObjID)
	require.NoError(t, err)
	require.Len(t, assets.Coins, 1)
}

type ReceivedCoin struct {
	objectID *sui.Address
	sender   *sui.Address
	coinType string
	balance  uint64
}

func TestSendReceiveCoin(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)
	chainSigner := newSignerWithFunds(t, SEEDFORCHAIN)

	iscPackageID := buildAndDeployISCContracts(t, client, chainSigner)

	eventCh := subscribeEvents(t, iscPackageID)

	anchorObjID := startChainAnchor(t, client, chainSigner, iscPackageID)

	tokenPackageID, _ := buildDeployMintTestcoin(t, client, signer)
	coinType := fmt.Sprintf("%s::testcoin::TESTCOIN", tokenPackageID.String())
	fmt.Printf("coin type: %s\n", coinType)

	// the signer should have only one coin object which belongs to testcoin type
	coins, err := client.GetCoins(context.Background(), suiclient.GetCoinsRequest{
		Owner:    signer.Address().AsSuiAddress(),
		CoinType: &coinType,
		Limit:    10,
	})
	require.NoError(t, err)

	sendCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins.Data[0].CoinObjectID)

	getObjectRes, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: coins.Data[0].CoinObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, anchorObjID.String(), getObjectRes.Data.Owner.ObjectOwnerInternal.AddressOwner.String())

	sender, object := receiveEvent(t, client, eventCh)
	require.Equal(t, signer.Address(), sender)

	coin := object.Data.Content.Data.MoveObject
	resource, err := sui.NewResourceType(coin.Type)
	require.NoError(t, err)
	require.Equal(t, "0x2", resource.Address.ShortString())
	require.Equal(t, "coin", resource.Module)
	require.Equal(t, "Coin", resource.ObjectName)
	var fields suijsonrpc.CoinFields
	err = json.Unmarshal(coin.Fields, &fields)
	require.NoError(t, err)

	// NOTE: this is the data that ISC should use to append the tokens to the account of the sender
	// receivedCoin := &ReceivedCoin{
	// 	objectID: object.Data.ObjectID,
	// 	sender:   sender,
	// 	coinType: resource.SubType.String(),
	// 	balance:  balance,
	// }

	receiveCoin(
		t,
		client,
		chainSigner,
		iscPackageID,
		anchorObjID,
		resource.SubType.String(),
		coins.Data[0].CoinObjectID,
	)

	assets, err := client.GetAssets(context.Background(), iscPackageID, anchorObjID)
	require.NoError(t, err)
	require.Len(t, assets.Coins, 1)
	require.Equal(t, coinType, "0x"+assets.Coins[0].CoinType)
	require.Equal(t, fields.Balance.Uint64(), assets.Coins[0].Balance.Uint64())
}

func TestCreateRequest(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	createReqRes, err := createRequest(t, client, signer, iscPackageID, anchorObjID)
	require.NoError(t, err)

	_, err = createReqRes.GetCreatedObjectInfo("request", "Request")
	require.NoError(t, err)
}

func TestSendRequest(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	createReqRes, err := createRequest(t, client, signer, iscPackageID, anchorObjID)
	require.NoError(t, err)

	reqObjRef, err := createReqRes.GetCreatedObjectInfo("request", "Request")
	require.NoError(t, err)
	getObjectRes, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: reqObjRef.ObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, signer.Address(), getObjectRes.Data.Owner.AddressOwner)

	sendRequest(t, client, signer, iscPackageID, anchorObjID, reqObjRef.ObjectID)

	getObjectRes, err = client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: reqObjRef.ObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, anchorObjID, getObjectRes.Data.Owner.AddressOwner)
}

func TestReceiveRequest(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	createReqRes, err := createRequest(t, client, signer, iscPackageID, anchorObjID)
	require.NoError(t, err)

	reqObjRef, err := createReqRes.GetCreatedObjectInfo("request", "Request")
	require.NoError(t, err)
	getObjectRes, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: reqObjRef.ObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, signer.Address(), getObjectRes.Data.Owner.AddressOwner)

	sendRequest(t, client, signer, iscPackageID, anchorObjID, reqObjRef.ObjectID)

	getObjectRes, err = client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: reqObjRef.ObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, anchorObjID, getObjectRes.Data.Owner.AddressOwner)

	receiveRequest(t, client, signer, iscPackageID, anchorObjID, reqObjRef.ObjectID)

	getObjectRes, err = client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: reqObjRef.ObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.NotNil(t, getObjectRes.Error.Data.Deleted)
}

type ReceivedRequest struct {
	objectID  *sui.Address
	sender    *sui.Address
	anchorID  *sui.ObjectID
	contract  string
	function  string
	args      [][]byte
	allowance []byte
}

func TestSendReceiveRequest(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)
	chainSigner := newSignerWithFunds(t, SEEDFORCHAIN)

	iscPackageID := buildAndDeployISCContracts(t, client, chainSigner)

	eventCh := subscribeEvents(t, iscPackageID)

	anchorObjID := startChainAnchor(t, client, chainSigner, iscPackageID)

	createReqRes, err := createRequest(t, client, signer, iscPackageID, anchorObjID)
	require.NoError(t, err)
	reqObjRef, err := createReqRes.GetCreatedObjectInfo("request", "Request")
	require.NoError(t, err)
	getObjectRes, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: reqObjRef.ObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, signer.Address(), getObjectRes.Data.Owner.AddressOwner)

	sendRequest(t, client, signer, iscPackageID, anchorObjID, reqObjRef.ObjectID)

	getObjectRes, err = client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: reqObjRef.ObjectID,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowOwner: true},
	})
	require.NoError(t, err)
	require.Equal(t, anchorObjID, getObjectRes.Data.Owner.AddressOwner)
	sender, object := receiveEvent(t, client, eventCh)
	require.Equal(t, signer.Address(), sender)

	req := object.Data.Content.Data.MoveObject
	require.True(t, strings.HasSuffix(req.Type, "::request::Request"))

	type Fields struct {
		Data struct {
			Fields struct {
				Contract string
				Function string
				Args     []interface{}
			}
		}
	}
	var fields Fields
	err = json.Unmarshal(req.Fields, &fields)
	require.NoError(t, err)
	args := make([][]byte, len(fields.Data.Fields.Args))
	for i, argField := range fields.Data.Fields.Args {
		argFieldBytes := argField.([]interface{})
		arg := make([]byte, len(argFieldBytes))
		for j, argFieldByte := range argFieldBytes {
			arg[j] = byte(argFieldByte.(float64))
		}
		args[i] = arg
	}

	// NOTE: this is the data that ISC should use as the request from the sender
	receivedRequest := &ReceivedRequest{
		objectID:  object.Data.ObjectID,
		sender:    sender,
		anchorID:  anchorObjID,
		contract:  fields.Data.Fields.Contract,
		function:  fields.Data.Fields.Function,
		args:      args,
		allowance: nil,
	}
	require.Equal(t, "isc_test_contract_name", receivedRequest.contract)
	require.Equal(t, "isc_test_func_name", receivedRequest.function)
	require.EqualValues(t, 3, len(receivedRequest.args))
	require.Equal(t, "one", string(receivedRequest.args[0]))
	require.Equal(t, "two", string(receivedRequest.args[1]))
	require.Equal(t, "three", string(receivedRequest.args[2]))
	require.Nil(t, receivedRequest.allowance)

	receiveRequest(t, client, chainSigner, iscPackageID, anchorObjID, reqObjRef.ObjectID)
}

func newSignerWithFunds(t *testing.T, seed string) cryptolib.Signer {
	kp := cryptolib.KeyPairFromSeed(cryptolib.SubSeed(sui.MustAddressFromHex(seed)[:], 0))
	err := suiclient.RequestFundsFromFaucet(kp.Address().AsSuiAddress(), suiconn.LocalnetFaucetURL)
	require.NoError(t, err)
	return kp
}

func startChainAnchor(
	t *testing.T,
	client *iscmove.Client,
	signer cryptolib.Signer,
	iscPackageID *sui.PackageID,
) *sui.ObjectID {
	// start a new chain
	startNewChainRes, err := client.StartNewChain(
		context.Background(),
		signer,
		iscPackageID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
		nil,
	)
	require.NoError(t, err)

	return &startNewChainRes.ID
}

func subscribeEvents(t *testing.T, iscPackageID *sui.PackageID) chan suijsonrpc.SuiEvent {
	api := suiclient.NewWebsocket(suiconn.LocalnetWebsocketEndpointURL)
	eventCh := make(chan suijsonrpc.SuiEvent)
	err := api.SubscribeEvent(
		context.TODO(),
		&suijsonrpc.EventFilter{
			Package: iscPackageID,
		},
		eventCh,
	)
	require.NoError(t, err)
	return eventCh
}

func receiveEvent(t *testing.T, client *iscmove.Client, eventCh chan suijsonrpc.SuiEvent) (
	*sui.Address,
	*suijsonrpc.SuiObjectResponse,
) {
	event := <-eventCh
	fmt.Println("event: ", event)
	close(eventCh)

	eventId := sui.MustAddressFromHex(event.ParsedJson.(map[string]interface{})["id"].(string))

	object, err := client.GetObject(context.Background(), suiclient.GetObjectRequest{
		ObjectID: eventId,
		Options:  &suijsonrpc.SuiObjectDataOptions{ShowContent: true},
	})
	require.NoError(t, err)
	require.Equal(t, eventId, object.Data.ObjectID)
	fmt.Println("object: ", object.Data.Content.Data.MoveObject)
	return event.Sender, object
}

func sendCoin(
	t *testing.T,
	client *iscmove.Client,
	signer cryptolib.Signer,
	iscPackageID *sui.PackageID,
	anchorObjID *sui.ObjectID,
	coinType string,
	coin *sui.ObjectID,
) {
	sendCoinRes, err := client.SendCoin(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
		coinType,
		coin,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, sendCoinRes.Effects.Data.IsSuccess())
}

func receiveCoin(
	t *testing.T,
	client *iscmove.Client,
	signer cryptolib.Signer,
	iscPackageID *sui.PackageID,
	anchorObjID *sui.ObjectID,
	coinType string,
	receivingCoinObject *sui.ObjectID,
) {
	receiveCoinRes, err := client.ReceiveCoin(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
		coinType,
		receivingCoinObject,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, receiveCoinRes.Effects.Data.IsSuccess())
}

func createRequest(
	t *testing.T,
	client *iscmove.Client,
	signer cryptolib.Signer,
	iscPackageID *sui.PackageID,
	anchorObjID *sui.ObjectID,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	createReqRes, err := client.CreateRequest(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
		"isc_test_contract_name",
		"isc_test_func_name",
		[][]byte{[]byte("one"), []byte("two"), []byte("three")}, // func input
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, createReqRes.Effects.Data.IsSuccess())
	return createReqRes, err
}

func sendRequest(
	t *testing.T,
	client *iscmove.Client,
	signer cryptolib.Signer,
	iscPackageID *sui.PackageID,
	anchorObjID *sui.ObjectID,
	reqObjID *sui.ObjectID,
) {
	sendReqRes, err := client.SendRequest(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
		reqObjID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, sendReqRes.Effects.Data.IsSuccess())
}

func receiveRequest(
	t *testing.T,
	client *iscmove.Client,
	signer cryptolib.Signer,
	iscPackageID *sui.PackageID,
	anchorObjID *sui.ObjectID,
	reqObjID *sui.ObjectID,
) {
	receiveReqRes, err := client.ReceiveRequest(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
		reqObjID,
		nil,
		suiclient.DefaultGasPrice,
		suiclient.DefaultGasBudget,
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, receiveReqRes.Effects.Data.IsSuccess())
}
