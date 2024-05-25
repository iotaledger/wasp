package isc_test

import (
	"context"
	"fmt"
	"github.com/iotaledger/wasp/sui-go/sui_types"
	"strconv"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/sui-go/isc"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/conn"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/utils"

	"github.com/stretchr/testify/require"
)

const SEEDFORUSER = "0x1234"
const SEEDFORCHAIN = "0x5678"

func newClient(t *testing.T) *isc.Client {
	// NOTE: comment out the next line to run local tests against sui-test-validator
	t.Skip("only for localnet")
	return isc.NewIscClient(sui.NewSuiClient(conn.LocalnetEndpointUrl))
}

func TestStartNewChain(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	modules, err := utils.MoveBuild(utils.GetGitRoot() + "/sui-go/isc/contracts/isc/")
	require.NoError(t, err)

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

	iscPackageID := txnResponse.GetPublishedPackageID()
	t.Log("packageID: ", iscPackageID)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)
	t.Log("anchorObjID: ", anchorObjID)
}

func TestSendCoin(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	iscPackageID := isc.BuildAndDeployIscContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	tokenPackageID, _ := isc.BuildDeployMintTestcoin(t, client, signer)
	coinType := fmt.Sprintf("%s::testcoin::TESTCOIN", tokenPackageID.String())
	fmt.Printf("coin type: %s\n", coinType)

	// the signer should have only one coin object which belongs to testcoin type
	coins, err := client.GetCoins(context.Background(), signer.Address, &coinType, nil, 10)
	require.NoError(t, err)

	sendCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins)

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

	iscPackageID := isc.BuildAndDeployIscContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	tokenPackageID, _ := isc.BuildDeployMintTestcoin(t, client, signer)
	coinType := fmt.Sprintf("%s::testcoin::TESTCOIN", tokenPackageID.String())
	fmt.Printf("coin type: %s\n", coinType)

	// the signer should have only one coin object which belongs to testcoin type
	coins, err := client.GetCoins(context.Background(), signer.Address, &coinType, nil, 10)
	require.NoError(t, err)

	sendCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins)

	getObjectRes, err := client.GetObject(
		context.Background(),
		coins.Data[0].CoinObjectID,
		&models.SuiObjectDataOptions{ShowOwner: true},
	)
	require.NoError(t, err)
	require.Equal(t, anchorObjID.String(), getObjectRes.Data.Owner.ObjectOwnerInternal.AddressOwner.String())

	receiveCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins)

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

	iscPackageID := isc.BuildAndDeployIscContracts(t, client, chainSigner)

	eventCh := subscribeEvents(t, iscPackageID)

	anchorObjID := startChainAnchor(t, client, chainSigner, iscPackageID)

	tokenPackageID, _ := isc.BuildDeployMintTestcoin(t, client, signer)
	coinType := fmt.Sprintf("%s::testcoin::TESTCOIN", tokenPackageID.String())
	fmt.Printf("coin type: %s\n", coinType)

	// the signer should have only one coin object which belongs to testcoin type
	coins, err := client.GetCoins(context.Background(), signer.Address, &coinType, nil, 10)
	require.NoError(t, err)

	sendCoin(t, client, signer, iscPackageID, anchorObjID, coinType, coins)

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
	const coinPrefix = "0x2::coin::Coin<"
	const coinSuffix = ">"
	require.True(t, strings.HasPrefix(coin.Type, coinPrefix) && strings.HasSuffix(coin.Type, coinSuffix))
	balance, err := strconv.ParseUint(coin.Fields.(map[string]interface{})["balance"].(string), 10, 64)
	require.NoError(t, err)

	// NOTE: this is the data that ISC should use to append the tokens to the account of the sender
	receivedCoin := &ReceivedCoin{
		objectID: object.Data.ObjectID,
		sender:   sender,
		coinType: coin.Type[len(coinPrefix) : len(coin.Type)-len(coinSuffix)],
		balance:  balance,
	}

	receiveCoin(t, client, chainSigner, iscPackageID, anchorObjID, receivedCoin.coinType, coins)

	assets, err := client.GetAssets(context.Background(), iscPackageID, anchorObjID)
	require.NoError(t, err)
	require.Len(t, assets.Coins, 1)
	require.Equal(t, coinType, "0x"+assets.Coins[0].CoinType)
	require.Equal(t, balance, assets.Coins[0].Balance.Uint64())
}

func TestCreateRequest(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	iscPackageID := isc.BuildAndDeployIscContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	createReqRes, err := createRequest(t, client, signer, iscPackageID, anchorObjID)

	_, _, err = sui.GetCreatedObjectIdAndType(createReqRes, "request", "Request")
	require.NoError(t, err)
}

func TestSendRequest(t *testing.T) {
	client := newClient(t)
	signer := newSignerWithFunds(t, SEEDFORUSER)

	iscPackageID := isc.BuildAndDeployIscContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	createReqRes, err := createRequest(t, client, signer, iscPackageID, anchorObjID)

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

	iscPackageID := isc.BuildAndDeployIscContracts(t, client, signer)

	anchorObjID := startChainAnchor(t, client, signer, iscPackageID)

	createReqRes, err := createRequest(t, client, signer, iscPackageID, anchorObjID)

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

	iscPackageID := isc.BuildAndDeployIscContracts(t, client, chainSigner)

	eventCh := subscribeEvents(t, iscPackageID)

	anchorObjID := startChainAnchor(t, client, chainSigner, iscPackageID)

	createReqRes, err := createRequest(t, client, signer, iscPackageID, anchorObjID)

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

	reqData := req.Fields.(map[string]interface{})["data"].(map[string]interface{})
	reqFields := reqData["fields"].(map[string]interface{})
	args := [][]byte{}

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
	require.Empty(t, receivedRequest.args)
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
	client *isc.Client,
	signer *sui_signer.Signer,
	iscPackageID *sui_types.PackageID,
) *sui_types.ObjectID {
	// start a new chain
	startNewChainRes, err := client.StartNewChain(
		context.Background(),
		signer,
		iscPackageID,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
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

func receiveEvent(t *testing.T, client *isc.Client, eventCh chan models.SuiEvent) (
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
	return &event.Sender, object
}

func sendCoin(
	t *testing.T,
	client *isc.Client,
	signer *sui_signer.Signer,
	iscPackageID *sui_types.PackageID,
	anchorObjID *sui_types.ObjectID,
	coinType string,
	coins *models.CoinPage,
) {
	sendCoinRes, err := client.SendCoin(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
		coinType,
		coins.Data[0].CoinObjectID,
		sui.DefaultGasBudget, &models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, sendCoinRes.Effects.Data.IsSuccess())
}

func receiveCoin(
	t *testing.T,
	client *isc.Client,
	signer *sui_signer.Signer,
	iscPackageID *sui_types.PackageID,
	anchorObjID *sui_types.ObjectID,
	coinType string,
	coins *models.CoinPage,
) {
	receiveCoinRes, err := client.ReceiveCoin(
		context.Background(),
		signer,
		iscPackageID,
		anchorObjID,
		coinType,
		coins.Data[0].CoinObjectID,
		sui.DefaultGasBudget, &models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, receiveCoinRes.Effects.Data.IsSuccess())
}

func createRequest(
	t *testing.T,
	client *isc.Client,
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
		[][]byte{}, // func input
		sui.DefaultGasBudget, &models.SuiTransactionBlockResponseOptions{
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
	client *isc.Client,
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
		sui.DefaultGasBudget, &models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, sendReqRes.Effects.Data.IsSuccess())
}

func receiveRequest(
	t *testing.T,
	client *isc.Client,
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
		sui.DefaultGasBudget, &models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, receiveReqRes.Effects.Data.IsSuccess())
}
