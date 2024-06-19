package iscmove_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/sui-go/iscmove"
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

func newClient(_ *testing.T) *iscmove.Client {
	// NOTE: comment out the next line to run local tests against sui-test-validator
	// t.Skip("only for localnet")
	return iscmove.NewClient(sui.NewSuiClient(conn.LocalnetEndpointUrl))
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
	coins, err := client.GetCoins(context.Background(), signer.Address, &coinType, nil, 10)
	require.NoError(t, err)
	require.Len(t, coins, 1)

	sendCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins.Data[0].CoinObjectID)

	getObjectRes, err := client.GetObject(
		context.Background(),
		coins.Data[0].CoinObjectID,
		&models.SuiObjectDataOptions{ShowOwner: true},
	)
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
	coins, err := client.GetCoins(context.Background(), signer.Address, &coinType, nil, 10)
	require.NoError(t, err)

	sendCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins.Data[0].CoinObjectID)

	getObjectRes, err := client.GetObject(
		context.Background(),
		coins.Data[0].CoinObjectID,
		&models.SuiObjectDataOptions{ShowOwner: true},
	)
	require.NoError(t, err)
	require.Equal(t, anchorObjID.String(), getObjectRes.Data.Owner.ObjectOwnerInternal.AddressOwner.String())

	receiveCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins.Data[0].CoinObjectID)

	assets, err := client.GetAssets(context.Background(), iscPackageID, anchorObjID)
	require.NoError(t, err)
	require.Len(t, assets.Coins, 1)
}

type ReceivedCoin struct {
	objectID *sui_types.SuiAddress
	sender   *sui_types.SuiAddress
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
	coins, err := client.GetCoins(context.Background(), signer.Address, &coinType, nil, 10)
	require.NoError(t, err)

	sendCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins.Data[0].CoinObjectID)

	getObjectRes, err := client.GetObject(
		context.Background(),
		coins.Data[0].CoinObjectID,
		&models.SuiObjectDataOptions{ShowOwner: true},
	)
	require.NoError(t, err)
	require.Equal(t, anchorObjID.String(), getObjectRes.Data.Owner.ObjectOwnerInternal.AddressOwner.String())

	sender, object := receiveEvent(t, client, eventCh)
	require.Equal(t, signer.Address, sender)

	coin := object.Data.Content.Data.MoveObject
	resource, err := models.NewResourceType(coin.Type)
	require.NoError(t, err)
	require.Equal(t, "0x2", resource.Address.ShortString())
	require.Equal(t, "coin", resource.ModuleName)
	require.Equal(t, "Coin", resource.FuncName)
	balance, err := strconv.ParseUint(coin.Fields["balance"].(string), 10, 64)
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
	require.Equal(t, balance, assets.Coins[0].Balance.Uint64())
}

func TestCreateRequest(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	createReqRes, err := createRequest(t, client, signer, iscPackageID, anchorObjID)
	require.NoError(t, err)

	_, _, err = sui.GetCreatedObjectIdAndType(createReqRes, "request", "Request")
	require.NoError(t, err)
}

func TestSendRequest(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	iscPackageID := buildAndDeployISCContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	createReqRes, err := createRequest(t, client, signer, iscPackageID, anchorObjID)
	require.NoError(t, err)

	reqObjID, _, err := sui.GetCreatedObjectIdAndType(createReqRes, "request", "Request")
	require.NoError(t, err)
	getObjectRes, err := client.GetObject(context.Background(), reqObjID, &models.SuiObjectDataOptions{ShowOwner: true})
	require.NoError(t, err)
	require.Equal(t, signer.Address, getObjectRes.Data.Owner.AddressOwner)

	sendRequest(t, client, signer, iscPackageID, anchorObjID, reqObjID)

	getObjectRes, err = client.GetObject(context.Background(), reqObjID, &models.SuiObjectDataOptions{ShowOwner: true})
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

	reqObjID, _, err := sui.GetCreatedObjectIdAndType(createReqRes, "request", "Request")
	require.NoError(t, err)
	getObjectRes, err := client.GetObject(context.Background(), reqObjID, &models.SuiObjectDataOptions{ShowOwner: true})
	require.NoError(t, err)
	require.Equal(t, signer.Address, getObjectRes.Data.Owner.AddressOwner)

	sendRequest(t, client, signer, iscPackageID, anchorObjID, reqObjID)

	getObjectRes, err = client.GetObject(context.Background(), reqObjID, &models.SuiObjectDataOptions{ShowOwner: true})
	require.NoError(t, err)
	require.Equal(t, anchorObjID, getObjectRes.Data.Owner.AddressOwner)

	receiveRequest(t, client, signer, iscPackageID, anchorObjID, reqObjID)

	getObjectRes, err = client.GetObject(context.Background(), reqObjID, &models.SuiObjectDataOptions{ShowOwner: true})
	require.NoError(t, err)
	require.NotNil(t, getObjectRes.Error.Data.Deleted)
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
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)
	chainSigner := newSignerWithFunds(t, SEEDFORCHAIN)

	iscPackageID := buildAndDeployISCContracts(t, client, chainSigner)

	eventCh := subscribeEvents(t, iscPackageID)

	anchorObjID := startChainAnchor(t, client, chainSigner, iscPackageID)

	createReqRes, err := createRequest(t, client, signer, iscPackageID, anchorObjID)
	require.NoError(t, err)

	reqObjID, reqType, err := sui.GetCreatedObjectIdAndType(createReqRes, "request", "Request")
	require.NoError(t, err)
	getObjectRes, err := client.GetObject(context.Background(), reqObjID, &models.SuiObjectDataOptions{ShowOwner: true})
	require.NoError(t, err)
	require.Equal(t, signer.Address, getObjectRes.Data.Owner.AddressOwner)

	sendRequest(t, client, signer, iscPackageID, anchorObjID, reqObjID)

	getObjectRes, err = client.GetObject(context.Background(), reqObjID, &models.SuiObjectDataOptions{ShowOwner: true})
	require.NoError(t, err)
	require.Equal(t, anchorObjID, getObjectRes.Data.Owner.AddressOwner)

	sender, object := receiveEvent(t, client, eventCh)
	require.Equal(t, signer.Address, sender)

	req := object.Data.Content.Data.MoveObject
	require.True(t, strings.HasSuffix(req.Type, "::request::Request"))
	require.Equal(t, reqType, req.Type)

	reqData := req.Fields["data"].(map[string]interface{})
	reqFields := reqData["fields"].(map[string]interface{})
	argFields := reqFields["args"].([]interface{})
	args := make([][]byte, len(argFields))
	for i, argField := range argFields {
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
		contract:  reqFields["contract"].(string),
		function:  reqFields["function"].(string),
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

	receiveRequest(t, client, chainSigner, iscPackageID, anchorObjID, reqObjID)
}

func newSignerWithFunds(t *testing.T, seed string) *sui_signer.Signer {
	signer := sui_signer.NewSigner(sui_types.MustSuiAddressFromHex(seed)[:], sui_signer.KeySchemeFlagIotaEd25519)
	err := sui.RequestFundFromFaucet(signer.Address, conn.LocalnetFaucetUrl)
	require.NoError(t, err)
	return signer
}

func startChainAnchor(
	t *testing.T,
	client *iscmove.Client,
	signer *sui_signer.Signer,
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
	require.True(t, startNewChainRes.Effects.Data.IsSuccess())

	anchorObjID, _, err := sui.GetCreatedObjectIdAndType(startNewChainRes, "anchor", "Anchor")
	require.NoError(t, err)
	return anchorObjID
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

func receiveEvent(t *testing.T, client *iscmove.Client, eventCh chan models.SuiEvent) (
	*sui_types.SuiAddress,
	*models.SuiObjectResponse,
) {
	event := <-eventCh
	fmt.Println("event: ", event)
	close(eventCh)

	eventId := sui_types.MustSuiAddressFromHex(event.ParsedJson.(map[string]interface{})["id"].(string))

	object, err := client.GetObject(
		context.Background(),
		eventId,
		&models.SuiObjectDataOptions{ShowContent: true},
	)
	require.NoError(t, err)
	require.Equal(t, eventId, object.Data.ObjectID)
	fmt.Println("object: ", object.Data.Content.Data.MoveObject)
	return event.Sender, object
}

func sendCoin(
	t *testing.T,
	client *iscmove.Client,
	signer *sui_signer.Signer,
	iscPackageID *sui_types.PackageID,
	anchorObjID *sui_types.ObjectID,
	coinType string,
	coin *sui_types.ObjectID,
) {
	sendCoinRes, err := client.SendCoin(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
		coinType,
		coin,
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
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
	signer *sui_signer.Signer,
	iscPackageID *sui_types.PackageID,
	anchorObjID *sui_types.ObjectID,
	coinType string,
	receivingCoinObject *sui_types.ObjectID,
) {
	receiveCoinRes, err := client.ReceiveCoin(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
		coinType,
		receivingCoinObject,
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
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
	signer *sui_signer.Signer,
	iscPackageID *sui_types.PackageID,
	anchorObjID *sui_types.ObjectID,
) (*models.SuiTransactionBlockResponse, error) {
	createReqRes, err := client.CreateRequest(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
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
	require.True(t, createReqRes.Effects.Data.IsSuccess())
	return createReqRes, err
}

func sendRequest(
	t *testing.T,
	client *iscmove.Client,
	signer *sui_signer.Signer,
	iscPackageID *sui_types.PackageID,
	anchorObjID *sui_types.ObjectID,
	reqObjID *sui_types.ObjectID,
) {
	sendReqRes, err := client.SendRequest(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
		reqObjID,
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
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
	signer *sui_signer.Signer,
	iscPackageID *sui_types.PackageID,
	anchorObjID *sui_types.ObjectID,
	reqObjID *sui_types.ObjectID,
) {
	receiveReqRes, err := client.ReceiveRequest(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
		reqObjID,
		nil,
		sui.DefaultGasPrice,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, receiveReqRes.Effects.Data.IsSuccess())
}
