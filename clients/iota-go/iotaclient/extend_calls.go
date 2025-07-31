package iotaclient

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
)

func (c *Client) GetCoinObjsForTargetAmount(
	ctx context.Context,
	address *iotago.Address,
	targetAmount uint64,
	gasAmount uint64,
) (iotajsonrpc.Coins, error) {
	coins, err := c.GetCoins(
		ctx, GetCoinsRequest{
			Owner: address,
			Limit: 200,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call GetCoins(): %w", err)
	}
	pickedCoins, err := iotajsonrpc.PickupCoins(coins, new(big.Int).SetUint64(targetAmount), gasAmount, 0, 25)
	if err != nil {
		return nil, err
	}
	return pickedCoins.Coins, nil
}

type SignAndExecuteTransactionRequest struct {
	TxDataBytes iotago.Base64Data
	Signer      iotasigner.Signer
	Options     *iotajsonrpc.IotaTransactionBlockResponseOptions // optional
}

func isResponseComplete(
	res *iotajsonrpc.IotaTransactionBlockResponse,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) bool {
	// In Rebased, it can happen that Effects are available before ObjectChanges are.
	// This function checks if ShowEffects/ShowObjectChanges are enabled, and validates the state of the response.

	// If object changes were requested, we need both object changes and effects (if effects were also requested)
	if options.ShowObjectChanges {
		if res.ObjectChanges == nil {
			return false
		}
		// Need to check effects too if they were requested
		if options.ShowEffects && res.Effects == nil {
			return false
		}
		return true
	}

	// If only effects were requested, we just need to wait for effects
	if options.ShowEffects {
		return res.Effects != nil
	}

	// If neither effects nor object changes were requested, response is complete
	return true
}

func (c *Client) SignAndExecuteTransaction(
	ctx context.Context,
	req *SignAndExecuteTransactionRequest,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	// FIXME we need to support other intent
	signature, err := req.Signer.SignTransactionBlock(req.TxDataBytes, iotasigner.DefaultIntent())
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction block: %w", err)
	}
	resp, err := c.ExecuteTransactionBlock(
		ctx,
		ExecuteTransactionBlockRequest{
			TxDataBytes: req.TxDataBytes,
			Signatures:  []*iotasigner.Signature{signature},
			Options:     req.Options,
			RequestType: iotajsonrpc.TxnRequestTypeWaitForLocalExecution,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}

	if !isResponseComplete(resp, req.Options) {
		if c.WaitUntilEffectsVisible == nil {
			return resp, fmt.Errorf("failed to execute transaction: %s", resp.Digest)
		}

		resp, err = c.GetTransactionBlock(
			ctx, GetTransactionBlockRequest{
				Digest:  &resp.Digest,
				Options: req.Options,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("GetTransactionBlock failed: %w", err)
		}
	}

	return resp, err
}

func (c *Client) PublishContract(
	ctx context.Context,
	signer iotasigner.Signer,
	modules []*iotago.Base64Data,
	dependencies []*iotago.Address,
	gasBudget uint64,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, *iotago.PackageID, error) {
	txnBytes, err := c.Publish(
		context.Background(),
		PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: modules,
			Dependencies:    dependencies,
			GasBudget:       iotajsonrpc.NewBigInt(gasBudget),
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to publish move contract: %w", err)
	}
	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		&SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      signer,
			Options:     options,
		},
	)
	if err != nil || !txnResponse.Effects.Data.IsSuccess() {
		return nil, nil, fmt.Errorf("failed to sign move contract tx: %w", err)
	}

	packageID, err := txnResponse.GetPublishedPackageID()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get move contract package ID: %w", err)
	}
	return txnResponse, packageID, nil
}

func (c *Client) UpdateObjectRef(
	ctx context.Context,
	ref *iotago.ObjectRef,
) (*iotago.ObjectRef, error) {
	res, err := c.GetObject(
		context.Background(),
		GetObjectRequest{
			ObjectID: ref.ObjectID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get the object of ObjectRef: %w", err)
	}

	return &iotago.ObjectRef{
		ObjectID: res.Data.ObjectID,
		Version:  res.Data.Version.Uint64(),
		Digest:   res.Data.Digest,
	}, nil
}

func (c *Client) MintToken(
	ctx context.Context,
	signer iotasigner.Signer,
	packageID *iotago.PackageID,
	tokenName string,
	treasuryCap *iotago.ObjectRef,
	mintAmount uint64,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       packageID,
				Module:        tokenName,
				Function:      "mint",
				TypeArguments: []iotago.TypeTag{},
				Arguments: []iotago.Argument{
					ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: treasuryCap}),
					ptb.MustForceSeparatePure(mintAmount),
					ptb.MustForceSeparatePure(signer.Address()),
				},
			},
		},
	)
	pt := ptb.Finish()

	return c.SignAndExecuteTxWithRetry(ctx, signer, pt, nil, DefaultGasBudget, DefaultGasPrice, options)
}

// The assigned gasPayments, or the gasPayments got by FindCoinsForGasPayment may be outdated ObjectRef
// which would cause the execution of tx failed.
// This func can retry a few time
func (c *Client) SignAndExecuteTxWithRetry(
	ctx context.Context,
	signer iotasigner.Signer,
	pt iotago.ProgrammableTransaction,
	gasCoin *iotago.ObjectRef,
	gasBudget uint64,
	gasPrice uint64,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	var err error
	var txnBytes []byte
	var txnResponse *iotajsonrpc.IotaTransactionBlockResponse
	var gasPayments []*iotago.ObjectRef
	for i := 0; i < c.WaitUntilEffectsVisible.Attempts; i++ {
		if gasCoin == nil {
			gasPayments, err = c.FindCoinsForGasPayment(ctx, signer.Address(), pt, gasPrice, gasBudget)
			if err != nil {
				return nil, fmt.Errorf("failed to find gas payment: %w", err)
			}
		} else {
			gasCoin, err = c.UpdateObjectRef(ctx, gasCoin)
			if err != nil {
				return nil, fmt.Errorf("failed to update gas payment: %w", err)
			}
			gasPayments = []*iotago.ObjectRef{gasCoin}
		}

		tx := iotago.NewProgrammable(
			signer.Address(),
			pt,
			gasPayments,
			gasBudget,
			gasPrice,
		)
		txnBytes, err = bcs.Marshal(&tx)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tx: %w", err)
		}

		txnResponse, err = c.SignAndExecuteTransaction(
			ctx, &SignAndExecuteTransactionRequest{
				TxDataBytes: txnBytes,
				Signer:      signer,
				Options:     options,
			},
		)
		if err == nil {
			return txnResponse, nil
		}
		time.Sleep(c.WaitUntilEffectsVisible.DelayBetweenAttempts)
	}
	return nil, fmt.Errorf("can't execute the transaction in time: %w", err)
}

func (c *Client) FindCoinsForGasPayment(
	ctx context.Context,
	owner *iotago.Address,
	pt iotago.ProgrammableTransaction,
	gasPrice uint64,
	gasBudget uint64,
) ([]*iotago.ObjectRef, error) {
	coinType := iotajsonrpc.IotaCoinType.String()
	coinPage, err := c.GetCoins(
		ctx, GetCoinsRequest{
			CoinType: &coinType,
			Owner:    owner,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch coins for gas payment: %w", err)
	}
	gasPayments, err := iotajsonrpc.PickupCoinsWithFilter(
		coinPage.Data,
		gasBudget,
		func(c *iotajsonrpc.Coin) bool { return !pt.IsInInputObjects(c.CoinObjectID) },
	)
	return gasPayments.CoinRefs(), err
}

func (c *Client) MergeCoinsAndExecute(
	ctx context.Context,
	owner iotasigner.Signer,
	destinationCoin *iotago.ObjectRef,
	sourceCoins []*iotago.ObjectRef,
	gasBudget uint64,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	var argCoins []iotago.Argument
	for _, sourceCoin := range sourceCoins {
		argCoins = append(argCoins, ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: sourceCoin}))
	}
	ptb.Command(
		iotago.Command{
			MergeCoins: &iotago.ProgrammableMergeCoins{
				Destination: ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: destinationCoin}),
				Sources:     argCoins,
			},
		},
	)
	pt := ptb.Finish()
	gasPayments, err := c.FindCoinsForGasPayment(ctx, owner.Address(), pt, DefaultGasPrice, DefaultGasBudget)
	if err != nil {
		return nil, fmt.Errorf("failed to find gas payment: %w", err)
	}
	tx := iotago.NewProgrammable(
		owner.Address(),
		pt,
		gasPayments,
		DefaultGasBudget,
		DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}
	txnResponse, err := c.SignAndExecuteTransaction(
		ctx, &SignAndExecuteTransactionRequest{
			TxDataBytes: txBytes,
			Signer:      owner,
			Options:     &iotajsonrpc.IotaTransactionBlockResponseOptions{ShowEffects: true, ShowObjectChanges: true},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}

// NOTE: This a copy the query limit from our Rust JSON RPC backend, this needs to be kept in sync!
const QUERY_MAX_RESULT_LIMIT = 50

// GetIotaCoinsOwnedByAddress This function will retrieve a maximum of 200 coins.
func (c *Client) GetIotaCoinsOwnedByAddress(ctx context.Context, address *iotago.Address) (iotajsonrpc.Coins, error) {
	page, err := c.GetCoins(
		ctx, GetCoinsRequest{
			Owner: address,
			Limit: 200,
		},
	)
	if err != nil {
		return nil, err
	}
	return page.Data, nil
}

// BatchGetObjectsOwnedByAddress @param filterType You can specify filtering out the specified resources, this will fetch all resources if it is not empty ""
func (c *Client) BatchGetObjectsOwnedByAddress(
	ctx context.Context,
	address *iotago.Address,
	options *iotajsonrpc.IotaObjectDataOptions,
	filterType string,
) ([]iotajsonrpc.IotaObjectResponse, error) {
	filterType = strings.TrimSpace(filterType)
	return c.BatchGetFilteredObjectsOwnedByAddress(
		ctx, address, options, func(sod *iotajsonrpc.IotaObjectData) bool {
			return filterType == "" || filterType == *sod.Type
		},
	)
}

func (c *Client) BatchGetFilteredObjectsOwnedByAddress(
	ctx context.Context,
	address *iotago.Address,
	options *iotajsonrpc.IotaObjectDataOptions,
	filter func(*iotajsonrpc.IotaObjectData) bool,
) ([]iotajsonrpc.IotaObjectResponse, error) {
	filteringObjs, err := c.GetOwnedObjects(
		ctx, GetOwnedObjectsRequest{
			Address: address,
			Query: &iotajsonrpc.IotaObjectResponseQuery{
				Options: &iotajsonrpc.IotaObjectDataOptions{
					ShowType: true,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	objIds := make([]*iotago.ObjectID, 0)
	for _, obj := range filteringObjs.Data {
		if obj.Data == nil {
			continue // error obj
		}
		if filter != nil && !filter(obj.Data) {
			continue // ignore objects if non-specified type
		}
		objIds = append(objIds, obj.Data.ObjectID)
	}

	return c.MultiGetObjects(
		ctx, MultiGetObjectsRequest{
			ObjectIDs: objIds,
			Options:   options,
		},
	)
}

////// PTB impl

func BCS_RequestAddStake(
	signer *iotago.Address,
	coins []*iotago.ObjectRef,
	amount *iotajsonrpc.BigInt,
	validator *iotago.Address,
	gasBudget, gasPrice uint64,
) ([]byte, error) {
	// build with BCS
	ptb := iotago.NewProgrammableTransactionBuilder()
	amtArg, err := ptb.Pure(amount.Uint64())
	if err != nil {
		return nil, err
	}
	arg0, err := ptb.Obj(iotago.IotaSystemMutObj)
	if err != nil {
		return nil, err
	}
	arg1 := ptb.Command(
		iotago.Command{
			SplitCoins: &iotago.ProgrammableSplitCoins{
				Coin:    iotago.Argument{GasCoin: &serialization.EmptyEnum{}},
				Amounts: []iotago.Argument{amtArg},
			},
		},
	) // the coin is split result argument
	arg2, err := ptb.Pure(validator)
	if err != nil {
		return nil, err
	}

	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:  iotago.IotaPackageIDIotaSystem,
				Module:   iotago.IotaSystemModuleName,
				Function: iotago.AddStakeFunName,
				Arguments: []iotago.Argument{
					arg0, arg1, arg2,
				},
			},
		},
	)
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		signer, pt, coins, gasBudget, gasPrice,
	)
	return bcs.Marshal(&tx)
}

func BCS_RequestWithdrawStake(
	signer *iotago.Address,
	stakedIotaRef iotago.ObjectRef,
	gas []*iotago.ObjectRef,
	gasBudget, gasPrice uint64,
) ([]byte, error) {
	// build with BCS
	ptb := iotago.NewProgrammableTransactionBuilder()
	arg0, err := ptb.Obj(iotago.IotaSystemMutObj)
	if err != nil {
		return nil, err
	}
	arg1, err := ptb.Obj(
		iotago.ObjectArg{
			ImmOrOwnedObject: &stakedIotaRef,
		},
	)
	if err != nil {
		return nil, err
	}
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:  iotago.IotaPackageIDIotaSystem,
				Module:   iotago.IotaSystemModuleName,
				Function: iotago.WithdrawStakeFunName,
				Arguments: []iotago.Argument{
					arg0, arg1,
				},
			},
		},
	)
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		signer, pt, gas, gasBudget, gasPrice,
	)
	return bcs.Marshal(&tx)
}
