package iscmove

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Config struct {
	APIURL       string
	FaucetURL    string
	GraphURL     string
	WebsocketURL string
}

type Client struct {
	*sui.ImplSuiAPI
	*SuiGraph

	config Config
}

func NewClient(config Config) *Client {
	return &Client{
		sui.NewSuiClient(config.APIURL),
		NewGraph(config.GraphURL),
		config,
	}
}

func (c *Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	paramJSON := fmt.Sprintf(`{"FixedAmountRequest":{"recipient":"%v"}}`, address)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.FaucetURL, bytes.NewBuffer([]byte(paramJSON)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("post %v response code: %v", c.config.FaucetURL, res.Status)
	}
	defer res.Body.Close()

	resByte, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	//
	var response struct {
		Task  string `json:"task,omitempty"`
		Error string `json:"error,omitempty"`
	}
	err = json.Unmarshal(resByte, &response)
	if err != nil {
		return err
	}
	if strings.TrimSpace(response.Error) != "" {
		return errors.New(response.Error)
	}

	return nil
}

func (c *Client) Health(ctx context.Context) error {
	_, err := c.GetLatestSuiSystemState(ctx)
	return err
}

// StartNewChain calls <packageID>::anchor::start_new_chain(), and then transfers the created
// Anchor to the signer.
func (c *Client) StartNewChain(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID *sui_types.PackageID,
	gasPayments []*sui_types.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *models.SuiTransactionBlockResponseOptions,
	treasuryCap *models.SuiObjectResponse,
) (*Anchor, error) {
	ptb := sui_types.NewProgrammableTransactionBuilder()
	// the return object is an Anchor object

	arguments := []sui_types.Argument{}
	if treasuryCap != nil {
		ref := treasuryCap.Data.Ref()

		arguments = []sui_types.Argument{
			ptb.MustObj(
				sui_types.ObjectArg{
					ImmOrOwnedObject: &ref,
				},
			),
		}
	}

	arg1 := ptb.Command(
		sui_types.Command{
			MoveCall: &sui_types.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "anchor",
				Function:      "start_new_chain",
				TypeArguments: []sui_types.TypeTag{},
				Arguments:     arguments,
			},
		},
	)

	ptb.Command(
		sui_types.Command{
			TransferObjects: &sui_types.ProgrammableTransferObjects{
				Objects: []sui_types.Argument{arg1},
				Address: ptb.MustPure(signer.Address),
			},
		},
	)
	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address().AsSuiAddress(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui_types.NewProgrammable(
		signer.Address().AsSuiAddress(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)
	txnBytes, err := bcs.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, cryptolib.SignerToSuiSigner(signer), txnBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return GetAnchorFromSuiTransactionBlockResponse(ctx, c, txnResponse)
}

// GetAssets fetches the assets stored in the anchor object.
func (c *Client) GetAssets(
	ctx context.Context,
	anchorPackageID *sui_types.PackageID,
	anchorAddress *sui_types.ObjectID,
) (*Assets, error) {
	// object 'Assets' is owned by the Anchor object, and an 'Assets' object doesn't have ID, because it is a
	// dynamic-field of Anchor object.
	resGetObject, err := c.GetObject(ctx, &models.GetObjectRequest{
		ObjectID: anchorAddress,
		Options: &models.SuiObjectDataOptions{
			ShowContent: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call GetObject(): %w", err)
	}
	type ResGetObjectFields struct {
		Assets json.RawMessage
	}
	var resGetObjectFields ResGetObjectFields
	err = json.Unmarshal(resGetObject.Data.Content.Data.MoveObject.Fields, &resGetObjectFields)
	if err != nil {
		return nil, fmt.Errorf("failed to get Assets fields %w", err)
	}
	var normalizedAssets NormalizedAssets
	err = json.Unmarshal(resGetObjectFields.Assets, &normalizedAssets)
	if err != nil {
		return nil, fmt.Errorf("failed to cast to 'NormalizedAssets' type: %w", err)
	}

	CoinsID := normalizedAssets.Fields.Coins.Fields.ID.ID
	resDynamicFields, err := c.GetDynamicFields(context.Background(), &models.GetDynamicFieldsRequest{
		ParentObjectID: sui_types.MustObjectIDFromHex(CoinsID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call GetDynamicFields(): %w", err)
	}

	var assets Assets
	assets.Coins = make([]*models.Coin, len(resDynamicFields.Data))
	for i, coin := range resDynamicFields.Data {
		assets.Coins[i] = &models.Coin{
			CoinType:     coin.Name.Value.(string),
			CoinObjectID: &coin.ObjectID,
			Digest:       &coin.Digest,
		}
	}

	for _, coin := range assets.Coins {
		res, err := c.GetObject(context.Background(), &models.GetObjectRequest{
			ObjectID: coin.CoinObjectID,
			Options: &models.SuiObjectDataOptions{
				ShowContent: true,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to call GetObject(): %w", err)
		}
		type ResFields struct {
			Value string
		}
		var resFields ResFields
		err = json.Unmarshal(res.Data.Content.Data.MoveObject.Fields, &resFields)
		if err != nil {
			panic(err)
		}
		bal, _ := strconv.ParseUint(resFields.Value, 10, 64)
		coin.Balance = models.NewBigInt(bal)
	}
	return &assets, nil
}

// CreateAndSendRequest calls <packageID>::request::create_and_send_request() and transfers the created
// Request to the signer.
func (c *Client) CreateAndSendRequest(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID *sui_types.PackageID,
	anchorAddress *sui_types.ObjectID,
	assetsBagRef *sui_types.ObjectRef,
	iscContractName string,
	iscFunctionName string,
	args [][]byte,
	gasPayments []*sui_types.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *models.SuiTransactionBlockResponseOptions,
) (*models.SuiTransactionBlockResponse, error) {
	ptb := sui_types.NewProgrammableTransactionBuilder()

	// the return object is an Anchor object
	ptb.Command(
		sui_types.Command{
			MoveCall: &sui_types.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "request",
				Function:      "create_and_send_request",
				TypeArguments: []sui_types.TypeTag{},
				Arguments: []sui_types.Argument{
					ptb.MustPure(anchorAddress),
					ptb.MustObj(sui_types.ObjectArg{ImmOrOwnedObject: assetsBagRef}),
					ptb.MustPure(iscContractName),
					ptb.MustPure(iscFunctionName),
					ptb.MustPure(args),
				},
			},
		},
	)

	pt := ptb.Finish()

	if len(gasPayments) == 0 {
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address().AsSuiAddress(), gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui_types.NewProgrammable(
		signer.Address().AsSuiAddress(),
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)
	// txnBytes, err := bcs.Marshal(tx)
	// if err != nil {
	// 	return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	// }

	// txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes, execOptions)
	// if err != nil {
	// 	return nil, fmt.Errorf("can't execute the transaction: %w", err)
	// }

	txnBytes, err := bcs.Marshal(tx.V1.Kind)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}

	txnResponse, err := c.DevInspectTransactionBlock(ctx, signer.Address, txnBytes, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}
	fmt.Println("txnResponse: ", txnResponse.Effects.Data.V1)
	return nil, nil
}

// ReceiveRequest calls <packageID>::anchor::receive_request(), which receives and consumes
// the request object.
func (c *Client) ReceiveRequest(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID *sui_types.PackageID,
	anchorAddress *sui_types.ObjectID,
	reqObjID *sui_types.ObjectID,
	gasPayments []*sui_types.ObjectRef, // optional
	gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
	gasBudget uint64,
	execOptions *models.SuiTransactionBlockResponseOptions,
) (*models.SuiTransactionBlockResponse, error) {
	txnBytes, err := c.MoveCall(ctx, &models.MoveCallRequest{
		Signer:    signer.Address().AsSuiAddress(),
		PackageID: packageID,
		Module:    "anchor",
		Function:  "receive_request",
		TypeArgs:  []string{},
		Arguments: []any{anchorAddress.String(), reqObjID.String()},
		GasBudget: models.NewBigInt(gasBudget),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call receive_request() move call: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, cryptolib.SignerToSuiSigner(signer), txnBytes.TxBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}
