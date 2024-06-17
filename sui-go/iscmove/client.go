package iscmove

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/fardream/go-bcs/bcs"

	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

// Client provides convenient methods to interact with the `isc` Move contracts.
type Client struct {
	*sui.ImplSuiAPI
}

func NewClient(api *sui.ImplSuiAPI) *Client {
	return &Client{
		api,
	}
}

// StartNewChain calls <packageID>::anchor::start_new_chain(), and then transfers the created
// Anchor to the signer.
func (c *Client) StartNewChain(
	ctx context.Context,
	signer *sui_signer.Signer,
	packageID *sui_types.PackageID,
	gasPayments []*sui_types.ObjectRef, // optional
	gasPrice uint64,
	gasBudget uint64,
	execOptions *models.SuiTransactionBlockResponseOptions,
	treasuryCapID *sui_types.ObjectID,
) (*models.SuiTransactionBlockResponse, error) {
	txObj, err := c.GetObject(ctx, treasuryCapID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Treasury object: %w", err)
	}

	ref := txObj.Data.Ref()

	fmt.Print(txObj)

	ptb := sui_types.NewProgrammableTransactionBuilder()
	// the return object is an Anchor object

	f, _ := ptb.Obj(
		sui_types.ObjectArg{
			ImmOrOwnedObject: &ref,
		},
	)

	arg1 := ptb.Command(
		sui_types.Command{

			MoveCall: &sui_types.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "anchor",
				Function:      "start_new_chain",
				TypeArguments: []sui_types.TypeTag{},
				Arguments:     []sui_types.Argument{f},
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
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address, gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui_types.NewProgrammable(
		signer.Address,
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)
	txnBytes, err := bcs.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}

// SendCoin calls <packageID>::anchor::send_coin(), which sends the given coin to the
// anchor's address.
func (c *Client) SendCoin(
	ctx context.Context,
	signer *sui_signer.Signer,
	anchorPackageID *sui_types.PackageID,
	anchorAddress *sui_types.ObjectID,
	coinType string,
	coinObject *sui_types.ObjectID,
	gasPayments []*sui_types.ObjectRef, // optional
	gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
	gasBudget uint64,
	execOptions *models.SuiTransactionBlockResponseOptions,
) (*models.SuiTransactionBlockResponse, error) {
	txnBytes, err := c.MoveCall(
		ctx,
		signer.Address,
		anchorPackageID,
		"anchor",
		"send_coin",
		[]string{coinType},
		[]any{anchorAddress.String(), coinObject.String()},
		nil,
		models.NewSafeSuiBigInt(gasBudget),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call send_coin() move call: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes.TxBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}

// ReceiveCoin calls <packageID>::anchor::receive_coin(), which adds the coin to the anchor's assets.
func (c *Client) ReceiveCoin(
	ctx context.Context,
	signer *sui_signer.Signer,
	anchorPackageID *sui_types.PackageID,
	anchorAddress *sui_types.ObjectID,
	coinType string,
	receivingCoinObject *sui_types.ObjectID,
	gasPayments []*sui_types.ObjectRef, // optional
	gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
	gasBudget uint64,
	execOptions *models.SuiTransactionBlockResponseOptions,
) (*models.SuiTransactionBlockResponse, error) {
	txnBytes, err := c.MoveCall(
		ctx,
		signer.Address,
		anchorPackageID,
		"anchor",
		"receive_coin",
		[]string{coinType},
		[]any{anchorAddress.String(), receivingCoinObject.String()},
		nil,
		models.NewSafeSuiBigInt(gasBudget),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call receive_coin() move call: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes.TxBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}

// GetAssets fetches the assets stored in the anchor object.
func (c *Client) GetAssets(
	ctx context.Context,
	anchorPackageID *sui_types.PackageID,
	anchorAddress *sui_types.ObjectID,
) (*Assets, error) {
	// object 'Assets' is owned by the Anchor object, and an 'Assets' object doesn't have ID, because it is a
	// dynamic-field of Anchor object.
	resGetObject, err := c.GetObject(
		context.Background(),
		anchorAddress,
		&models.SuiObjectDataOptions{
			ShowContent: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call GetObject(): %w", err)
	}

	b, err := json.Marshal(resGetObject.Data.Content.Data.MoveObject.Fields.(map[string]interface{})["assets"])
	if err != nil {
		return nil, fmt.Errorf("failed to access 'assets' fields: %w", err)
	}
	var normalizedAssets NormalizedAssets
	err = json.Unmarshal(b, &normalizedAssets)
	if err != nil {
		return nil, fmt.Errorf("failed to cast to 'NormalizedAssets' type: %w", err)
	}

	CoinsID := normalizedAssets.Fields.Coins.Fields.ID.ID
	resDynamicFields, err := c.GetDynamicFields(
		context.Background(),
		sui_types.MustObjectIDFromHex(CoinsID),
		nil,
		nil,
	)
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
		res, err := c.GetObject(
			context.Background(),
			coin.CoinObjectID,
			&models.SuiObjectDataOptions{
				ShowContent: true,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to call GetObject(): %w", err)
		}
		fieldsMap := res.Data.Content.Data.MoveObject.Fields.((map[string]interface{}))
		bal, _ := strconv.ParseUint(fieldsMap["value"].(string), 10, 64)
		coin.Balance = models.NewSafeSuiBigInt(bal)
	}
	return &assets, nil
}

// CreateRequest calls <packageID>::request::create_request() and transfers the created
// Request to the signer.
func (c *Client) CreateRequest(
	ctx context.Context,
	signer *sui_signer.Signer,
	packageID *sui_types.PackageID,
	anchorAddress *sui_types.ObjectID,
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
	arg1 := ptb.Command(
		sui_types.Command{
			MoveCall: &sui_types.ProgrammableMoveCall{
				Package:       packageID,
				Module:        "request",
				Function:      "create_request",
				TypeArguments: []sui_types.TypeTag{},
				Arguments: []sui_types.Argument{
					ptb.MustPure(iscContractName),
					ptb.MustPure(iscFunctionName),
					ptb.MustPure(args),
				},
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
		coins, err := c.GetCoinObjsForTargetAmount(ctx, signer.Address, gasBudget)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GasPayment object: %w", err)
		}
		gasPayments = coins.CoinRefs()
	}

	tx := sui_types.NewProgrammable(
		signer.Address,
		pt,
		gasPayments,
		gasBudget,
		gasPrice,
	)
	txnBytes, err := bcs.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}

// SendRequest calls <packageID>::anchor::send_request(), which sends the request to the anchor.
func (c *Client) SendRequest(
	ctx context.Context,
	signer *sui_signer.Signer,
	packageID *sui_types.PackageID,
	anchorAddress *sui_types.ObjectID,
	reqObjID *sui_types.ObjectID,
	gasPayments []*sui_types.ObjectRef, // optional
	gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
	gasBudget uint64,
	execOptions *models.SuiTransactionBlockResponseOptions,
) (*models.SuiTransactionBlockResponse, error) {
	txnBytes, err := c.MoveCall(
		ctx,
		signer.Address,
		packageID,
		"anchor",
		"send_request",
		[]string{},
		[]any{anchorAddress.String(), reqObjID.String()},
		nil,
		models.NewSafeSuiBigInt(gasBudget),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call send_request() move call: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes.TxBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}

// ReceiveRequest calls <packageID>::anchor::receive_request(), which receives and consumes
// the request object.
func (c *Client) ReceiveRequest(
	ctx context.Context,
	signer *sui_signer.Signer,
	packageID *sui_types.PackageID,
	anchorAddress *sui_types.ObjectID,
	reqObjID *sui_types.ObjectID,
	gasPayments []*sui_types.ObjectRef, // optional
	gasPrice uint64, // TODO use gasPrice when we change MoveCall API to PTB version
	gasBudget uint64,
	execOptions *models.SuiTransactionBlockResponseOptions,
) (*models.SuiTransactionBlockResponse, error) {
	txnBytes, err := c.MoveCall(
		ctx,
		signer.Address,
		packageID,
		"anchor",
		"receive_request",
		[]string{},
		[]any{anchorAddress.String(), reqObjID.String()},
		nil,
		models.NewSafeSuiBigInt(gasBudget),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call receive_request() move call: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes.TxBytes, execOptions)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}
