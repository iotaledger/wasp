package iotaclient

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/util/bcs"
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

func (c *Client) retryGetTransactionBlock(
	ctx context.Context,
	digest *iotago.TransactionDigest,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if c.WaitUntilEffectsVisible == nil {
		return nil, fmt.Errorf("waitUntilEffectsVisible is nil, retry is disabled")
	}

	for attempt := 0; attempt <= c.WaitUntilEffectsVisible.Attempts; attempt++ {
		res, err := c.GetTransactionBlock(
			ctx, GetTransactionBlockRequest{
				Digest:  digest,
				Options: options,
			},
		)

		// Return immediately on context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// On last attempt, return whatever error we got
		if attempt == c.WaitUntilEffectsVisible.Attempts {
			if err != nil {
				return nil, fmt.Errorf(
					"retryGetTransactionBlock failed after %d attempts: %v",
					c.WaitUntilEffectsVisible.Attempts,
					err,
				)
			}
			// If it's the last attempt and we have a response but it's incomplete,
			// return it anyway
			return res, nil
		}

		// If we got an error or incomplete response, wait and retry
		if err != nil || !isResponseComplete(res, options) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.WaitUntilEffectsVisible.DelayBetweenAttempts):
				continue
			}
		}

		// We have a complete response, return it
		return res, nil
	}

	// This should never be reached due to the returns in the loop
	return nil, fmt.Errorf("unexpected error in retry logic")
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
		
		resp, err = c.retryGetTransactionBlock(ctx, &resp.Digest, req.Options)
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
	treasuryCap *iotago.ObjectID,
	mintAmount uint64,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	txnBytes, err := c.MoveCall(
		ctx,
		MoveCallRequest{
			Signer:    signer.Address(),
			PackageID: packageID,
			Module:    tokenName,
			Function:  "mint",
			TypeArgs:  []string{},
			Arguments: []any{treasuryCap.String(), fmt.Sprintf("%d", mintAmount), signer.Address().String()},
			GasBudget: iotajsonrpc.NewBigInt(DefaultGasBudget),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call mint() move call: %w", err)
	}

	txnResponse, err := c.SignAndExecuteTransaction(
		ctx, &SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      signer,
			Options:     options,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
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
			Options:     &iotajsonrpc.IotaTransactionBlockResponseOptions{ShowEffects: true},
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
