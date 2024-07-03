package iscmove

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Config struct {
	APIURL       string
	FaucetURL    string
	GraphURL     string
	WebsocketURL string
}

type Client struct {
	*suiclient.Client
	*SuiGraph

	config Config
}

func NewClient(config Config) *Client {
	return &Client{
		suiclient.New(config.APIURL),
		NewGraph(config.GraphURL),
		config,
	}
}

func (c *Client) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	var faucetURL string = c.config.FaucetURL
	if faucetURL == "" {
		switch c.config.APIURL {
		case suiconn.TestnetEndpointURL:
			faucetURL = suiconn.TestnetFaucetURL
		case suiconn.DevnetEndpointURL:
			faucetURL = suiconn.DevnetFaucetURL
		case suiconn.LocalnetEndpointURL:
			faucetURL = suiconn.LocalnetFaucetURL
		default:
			panic("unspecified FaucetURL")
		}
	}
	return suiclient.RequestFundsFromFaucet(address.AsSuiAddress(), faucetURL)
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
	packageID *sui.PackageID,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
	treasuryCap *suijsonrpc.SuiObjectResponse,
) (*Anchor, error) {
	ptb := sui.NewProgrammableTransactionBuilder()
	// the return object is an Anchor object

	arguments := []sui.Argument{}
	if treasuryCap != nil {
		ref := treasuryCap.Data.Ref()

		arguments = []sui.Argument{
			ptb.MustObj(
				sui.ObjectArg{
					ImmOrOwnedObject: &ref,
				},
			),
		}
	}

	arg1 := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "anchor",
				Function:      "start_new_chain",
				TypeArguments: []sui.TypeTag{},
				Arguments:     arguments,
			},
		},
	)

	ptb.Command(
		sui.Command{
			TransferObjects: &sui.ProgrammableTransferObjects{
				Objects: []sui.Argument{arg1},
				Address: ptb.MustPure(signer.Address()),
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

	tx := sui.NewProgrammable(
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

	return c.getAnchorFromSuiTransactionBlockResponse(ctx, txnResponse)
}

func (c *Client) getAnchorFromSuiTransactionBlockResponse(
	ctx context.Context,
	response *suijsonrpc.SuiTransactionBlockResponse,
) (*Anchor, error) {
	anchorObjRef, err := response.GetCreatedObjectInfo("anchor", "Anchor")
	if err != nil {
		return nil, err
	}

	getObjectResponse, err := c.GetObject(
		ctx, suiclient.GetObjectRequest{
			ObjectID: anchorObjRef.ObjectID,
			Options:  &suijsonrpc.SuiObjectDataOptions{ShowBcs: true},
		},
	)
	if err != nil {
		return nil, err
	}
	anchorBCS := getObjectResponse.Data.Bcs.Data.MoveObject.BcsBytes

	anchor := Anchor{}
	n, err := bcs.Unmarshal(anchorBCS, &anchor)
	if err != nil {
		return nil, err
	}
	if n != len(anchorBCS) {
		return nil, errors.New("cannot decode anchor: excess bytes")
	}
	return &anchor, nil
}

// SendCoin calls <packageID>::anchor::send_coin(), which sends the given coin to the
// anchor's address.
func (c *Client) SendCoin(
	ctx context.Context,
	signer cryptolib.Signer,
	anchorPackageID *sui.PackageID,
	anchorAddress *sui.ObjectID,
	coinType string,
	coinObject *sui.ObjectID,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	txnBytes, err := c.MoveCall(ctx, suiclient.MoveCallRequest{
		Signer:    signer.Address().AsSuiAddress(),
		PackageID: anchorPackageID,
		Module:    "anchor",
		Function:  "send_coin",
		TypeArgs:  []string{coinType},
		Arguments: []any{anchorAddress.String(), coinObject.String()},
		GasBudget: suijsonrpc.NewBigInt(gasBudget),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call send_coin() move call: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, cryptolib.SignerToSuiSigner(signer), txnBytes.TxBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}

// ReceiveCoin calls <packageID>::anchor::receive_coin(), which adds the coin to the anchor's assets.
func (c *Client) ReceiveCoin(
	ctx context.Context,
	signer cryptolib.Signer,
	anchorPackageID *sui.PackageID,
	anchorAddress *sui.ObjectID,
	coinType string,
	receivingCoinObject *sui.ObjectID,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	txnBytes, err := c.MoveCall(ctx, suiclient.MoveCallRequest{
		Signer:    signer.Address().AsSuiAddress(),
		PackageID: anchorPackageID,
		Module:    "anchor",
		Function:  "receive_coin",
		TypeArgs:  []string{coinType},
		Arguments: []any{anchorAddress.String(), receivingCoinObject.String()},
		GasBudget: suijsonrpc.NewBigInt(gasBudget),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call receive_coin() move call: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, cryptolib.SignerToSuiSigner(signer), txnBytes.TxBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}

// GetAssets fetches the assets stored in the anchor object.
func (c *Client) GetAssets(
	ctx context.Context,
	anchorPackageID *sui.PackageID,
	anchorAddress *sui.ObjectID,
) (*Assets, error) {
	// object 'Assets' is owned by the Anchor object, and an 'Assets' object doesn't have ID, because it is a
	// dynamic-field of Anchor object.
	resGetObject, err := c.GetObject(ctx, suiclient.GetObjectRequest{
		ObjectID: anchorAddress,
		Options: &suijsonrpc.SuiObjectDataOptions{
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
	resDynamicFields, err := c.GetDynamicFields(context.Background(), suiclient.GetDynamicFieldsRequest{
		ParentObjectID: sui.MustObjectIDFromHex(CoinsID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call GetDynamicFields(): %w", err)
	}

	var assets Assets
	assets.Coins = make([]*suijsonrpc.Coin, len(resDynamicFields.Data))
	for i, coin := range resDynamicFields.Data {
		assets.Coins[i] = &suijsonrpc.Coin{
			CoinType:     coin.Name.Value.(string),
			CoinObjectID: &coin.ObjectID,
			Digest:       &coin.Digest,
		}
	}

	for _, coin := range assets.Coins {
		res, err := c.GetObject(context.Background(), suiclient.GetObjectRequest{
			ObjectID: coin.CoinObjectID,
			Options: &suijsonrpc.SuiObjectDataOptions{
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
		coin.Balance = suijsonrpc.NewBigInt(bal)
	}
	return &assets, nil
}

// CreateRequest calls <packageID>::request::create_request() and transfers the created
// Request to the signer.
func (c *Client) CreateRequest(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID *sui.PackageID,
	anchorAddress *sui.ObjectID,
	iscContractName string,
	iscFunctionName string,
	args [][]byte,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	ptb := sui.NewProgrammableTransactionBuilder()

	// the return object is an Anchor object
	arg1 := ptb.Command(
		sui.Command{
			MoveCall: &sui.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "request",
				Function:      "create_request",
				TypeArguments: []sui.TypeTag{},
				Arguments: []sui.Argument{
					ptb.MustPure(iscContractName),
					ptb.MustPure(iscFunctionName),
					ptb.MustPure(args),
				},
			},
		},
	)

	ptb.Command(
		sui.Command{
			TransferObjects: &sui.ProgrammableTransferObjects{
				Objects: []sui.Argument{arg1},
				Address: ptb.MustPure(signer.Address()),
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

	tx := sui.NewProgrammable(
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

	return txnResponse, nil
}

// SendRequest calls <packageID>::anchor::send_request(), which sends the request to the anchor.
func (c *Client) SendRequest(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID *sui.PackageID,
	anchorAddress *sui.ObjectID,
	reqObjID *sui.ObjectID,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	txnBytes, err := c.MoveCall(ctx, suiclient.MoveCallRequest{
		Signer:    signer.Address().AsSuiAddress(),
		PackageID: packageID,
		Module:    "anchor",
		Function:  "send_request",
		TypeArgs:  []string{},
		Arguments: []any{anchorAddress.String(), reqObjID.String()},
		GasBudget: suijsonrpc.NewBigInt(gasBudget),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call send_request() move call: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, cryptolib.SignerToSuiSigner(signer), txnBytes.TxBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}

// ReceiveRequest calls <packageID>::anchor::receive_request(), which receives and consumes
// the request object.
func (c *Client) ReceiveRequest(
	ctx context.Context,
	signer cryptolib.Signer,
	packageID *sui.PackageID,
	anchorAddress *sui.ObjectID,
	reqObjID *sui.ObjectID,
	gasPayments []*sui.ObjectRef, // optional
	gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
	gasBudget uint64,
	execOptions *suijsonrpc.SuiTransactionBlockResponseOptions,
) (*suijsonrpc.SuiTransactionBlockResponse, error) {
	txnBytes, err := c.MoveCall(ctx, suiclient.MoveCallRequest{
		Signer:    signer.Address().AsSuiAddress(),
		PackageID: packageID,
		Module:    "anchor",
		Function:  "receive_request",
		TypeArgs:  []string{},
		Arguments: []any{anchorAddress.String(), reqObjID.String()},
		GasBudget: suijsonrpc.NewBigInt(gasBudget),
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
