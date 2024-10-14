package iotaclient

import (
	"context"
	"fmt"
	"math/big"
	"strings"

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
	pickedCoins, err := iotajsonrpc.PickupCoins(coins, new(big.Int).SetUint64(targetAmount), 0, 0, 0)
	if err != nil {
		return nil, err
	}
	return pickedCoins.Coins, nil
}

func (c *Client) SignAndExecuteTransaction(
	ctx context.Context,
	signer iotasigner.Signer,
	txBytes iotago.Base64Data,
	options *iotajsonrpc.SuiTransactionBlockResponseOptions,
) (*iotajsonrpc.SuiTransactionBlockResponse, error) {
	// FIXME we need to support other intent
	signature, err := signer.SignTransactionBlock(txBytes, iotasigner.DefaultIntent())
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction block: %w", err)
	}
	resp, err := c.ExecuteTransactionBlock(
		ctx,
		ExecuteTransactionBlockRequest{
			TxDataBytes: txBytes,
			Signatures:  []*iotasigner.Signature{signature},
			Options:     options,
			RequestType: iotajsonrpc.TxnRequestTypeWaitForLocalExecution,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}
	if options.ShowEffects && !resp.Effects.Data.IsSuccess() {
		return resp, fmt.Errorf("failed to execute transaction: %v", resp.Effects.Data.V1.Status)
	}
	return resp, nil
}

func (c *Client) PublishContract(
	ctx context.Context,
	signer iotasigner.Signer,
	modules []*iotago.Base64Data,
	dependencies []*iotago.Address,
	gasBudget uint64,
	options *iotajsonrpc.SuiTransactionBlockResponseOptions,
) (*iotajsonrpc.SuiTransactionBlockResponse, *iotago.PackageID, error) {
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
	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes.TxBytes, options)
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
	options *iotajsonrpc.SuiTransactionBlockResponseOptions,
) (*iotajsonrpc.SuiTransactionBlockResponse, error) {
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

	txnResponse, err := c.SignAndExecuteTransaction(ctx, signer, txnBytes.TxBytes, options)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}

// NOTE: This a copy the query limit from our Rust JSON RPC backend, this needs to be kept in sync!
const QUERY_MAX_RESULT_LIMIT = 50

// GetSuiCoinsOwnedByAddress This function will retrieve a maximum of 200 coins.
func (c *Client) GetSuiCoinsOwnedByAddress(ctx context.Context, address *iotago.Address) (iotajsonrpc.Coins, error) {
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
	options *iotajsonrpc.SuiObjectDataOptions,
	filterType string,
) ([]iotajsonrpc.SuiObjectResponse, error) {
	filterType = strings.TrimSpace(filterType)
	return c.BatchGetFilteredObjectsOwnedByAddress(
		ctx, address, options, func(sod *iotajsonrpc.SuiObjectData) bool {
			return filterType == "" || filterType == *sod.Type
		},
	)
}

func (c *Client) BatchGetFilteredObjectsOwnedByAddress(
	ctx context.Context,
	address *iotago.Address,
	options *iotajsonrpc.SuiObjectDataOptions,
	filter func(*iotajsonrpc.SuiObjectData) bool,
) ([]iotajsonrpc.SuiObjectResponse, error) {
	filteringObjs, err := c.GetOwnedObjects(
		ctx, GetOwnedObjectsRequest{
			Address: address,
			Query: &iotajsonrpc.SuiObjectResponseQuery{
				Options: &iotajsonrpc.SuiObjectDataOptions{
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
	arg0, err := ptb.Obj(iotago.SuiSystemMutObj)
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
				Package:  iotago.IotaPackageIdIotaSystem,
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
	stakedSuiRef iotago.ObjectRef,
	gas []*iotago.ObjectRef,
	gasBudget, gasPrice uint64,
) ([]byte, error) {
	// build with BCS
	ptb := iotago.NewProgrammableTransactionBuilder()
	arg0, err := ptb.Obj(iotago.SuiSystemMutObj)
	if err != nil {
		return nil, err
	}
	arg1, err := ptb.Obj(
		iotago.ObjectArg{
			ImmOrOwnedObject: &stakedSuiRef,
		},
	)
	if err != nil {
		return nil, err
	}
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:  iotago.IotaPackageIdIotaSystem,
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
