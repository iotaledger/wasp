package l1

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"

	"fortio.org/safecast"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

const (
	defaultGasPrice = uint64(1000)
	faucetAmount    = uint64(2_000_000_000)
)

type FakeIotaClient struct {
	Store         *ObjectStore
	Executor      *Executor
	faucetMu      sync.Mutex
	faucetCounter uint64 // monotonic counter for unique faucet coin IDs
}

var _ iotagraphql.IotaClient = (*FakeIotaClient)(nil)

func NewFakeIotaClient(store *ObjectStore, executor *Executor) *FakeIotaClient {
	return &FakeIotaClient{Store: store, Executor: executor}
}

func (c *FakeIotaClient) ExecuteTransactionBlock(
	_ context.Context,
	txDataBytes iotago.Base64Data,
	signatures []*iotasigner.Signature,
) (*graphqltypes.ExecuteTransactionBlockResponse, error) {
	tx, err := bcs.Unmarshal[iotago.TransactionData](txDataBytes)
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: BCS unmarshal TransactionData: %w", err)
	}
	if tx.V1 == nil {
		return nil, fmt.Errorf("FakeIotaClient: TransactionData.V1 is nil")
	}

	validated, err := ValidateTransaction(c.Store, tx.V1)
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: validation failed: %w", err)
	}

	result, err := c.Executor.Execute(&tx, validated)
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: execution failed: %w", err)
	}

	var gasCoinID *iotago.ObjectID
	if len(validated.GasCoins) > 0 {
		id := validated.GasCoins[0].ID
		gasCoinID = &id
	}

	sigs := make([]iotago.Base64Data, len(signatures))
	for i, s := range signatures {
		if s != nil {
			sigs[i] = iotago.Base64Data(s.Bytes())
		}
	}

	resp := BuildExecuteResponse(result, txDataBytes, tx.V1.Sender, sigs, gasCoinID)

	c.Store.StoreTx(result.TxDigest, &StoredTx{
		TxData:     txDataBytes,
		Effects:    resp,
		Sender:     tx.V1.Sender,
		Digest:     result.TxDigest,
		Signatures: sigs,
	})

	return resp, nil
}

func (c *FakeIotaClient) SignAndExecuteTransaction(
	ctx context.Context,
	txnBytes []byte,
	signer iotasigner.Signer,
) (*graphqltypes.ExecuteTransactionBlockResponse, error) {
	signature, err := signer.SignTransactionBlock(txnBytes, iotasigner.DefaultIntent())
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: sign failed: %w", err)
	}
	return c.ExecuteTransactionBlock(ctx, txnBytes, []*iotasigner.Signature{signature})
}

func (c *FakeIotaClient) SignAndExecuteTxWithRetry(
	ctx context.Context,
	signer iotasigner.Signer,
	pt iotago.ProgrammableTransaction,
	gasCoin *iotago.ObjectRef,
	gasBudget uint64,
	gasPrice uint64,
) (*iotagraphql.ExecuteTransactionBlockResponse, error) {
	var gasPayments []*iotago.ObjectRef
	if gasCoin != nil {
		updated, err := c.UpdateObjectRef(ctx, gasCoin)
		if err != nil {
			return nil, fmt.Errorf("FakeIotaClient: update gas ref: %w", err)
		}
		gasPayments = []*iotago.ObjectRef{updated}
	} else {
		addr := signer.Address()
		if addr == nil {
			return nil, fmt.Errorf("FakeIotaClient: signer has no address")
		}
		coins := c.Store.GetCoinsByOwner(*addr, IotaCoinTypeStr)
		if len(coins) == 0 {
			return nil, fmt.Errorf("FakeIotaClient: no gas coins for %s", addr.String())
		}
		for _, coin := range coins {
			gasPayments = append(gasPayments, coin.Ref())
		}
	}

	addr := signer.Address()
	if addr == nil {
		return nil, fmt.Errorf("FakeIotaClient: signer has no address")
	}

	tx := iotago.NewProgrammable(addr, pt, gasPayments, gasBudget, gasPrice)
	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: BCS marshal tx: %w", err)
	}
	return c.SignAndExecuteTransaction(ctx, txBytes, signer)
}

func (c *FakeIotaClient) DryRunTransaction(
	_ context.Context,
	txDataBytes iotago.Base64Data,
) (*graphqltypes.DryRunTransactionBlockResponse, error) {
	tx, err := bcs.Unmarshal[iotago.TransactionData](txDataBytes)
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: BCS unmarshal: %w", err)
	}
	if tx.V1 == nil {
		return nil, fmt.Errorf("FakeIotaClient: TransactionData.V1 is nil")
	}

	validated, err := ValidateTransaction(c.Store, tx.V1)
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: validation: %w", err)
	}

	result, err := c.Executor.Execute(&tx, validated)
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: dry run execution: %w", err)
	}

	var gasCoinID *iotago.ObjectID
	if len(validated.GasCoins) > 0 {
		id := validated.GasCoins[0].ID
		gasCoinID = &id
	}

	execResp := BuildExecuteResponse(result, txDataBytes, tx.V1.Sender, nil, gasCoinID)
	return &graphqltypes.DryRunTransactionBlockResponse{
		DryRunTransactionBlock: graphqltypes.DryRunTransactionBlockDryRunTransactionBlockDryRunResult{
			Transaction: graphqltypes.TxBlockData{
				TX_CORE: execResp.ExecuteTransactionBlock.Effects.TransactionBlock.TX_CORE,
				Effects: execResp.ExecuteTransactionBlock.Effects,
			},
		},
	}, nil
}

func (c *FakeIotaClient) GetObject(_ context.Context, objectID iotago.ObjectID) (*graphqltypes.GetObjectResponse, error) {
	obj, ok := c.Store.Get(objectID)
	if !ok {
		return BuildNotFoundResponse(), nil
	}
	return BuildGetObjectResponse(obj), nil
}

func (c *FakeIotaClient) GetTransactionBlock(_ context.Context, digest iotago.TransactionDigest) (*graphqltypes.GetTransactionBlockResponse, error) {
	tx, ok := c.Store.GetTx(digest)
	if !ok {
		return nil, fmt.Errorf("FakeIotaClient: transaction %s not found", digest.String())
	}
	return BuildGetTransactionBlockResponse(tx), nil
}

func (c *FakeIotaClient) TryGetPastObject(
	_ context.Context,
	objectID iotago.ObjectID,
	version uint64,
) (*iotagraphql.TryGetPastObjectResponse, error) {
	obj, ok := c.Store.GetAtVersion(objectID, version)
	if !ok {
		current, curOk := c.Store.Get(objectID)
		if !curOk {
			return &graphqltypes.TryGetPastObjectResponse{
				Current: graphqltypes.TryGetPastObjectCurrentObject{
					Address: objectID,
				},
			}, nil
		}
		return &graphqltypes.TryGetPastObjectResponse{
			Current: graphqltypes.TryGetPastObjectCurrentObject{
				Address: current.ID,
				Version: current.Version,
			},
		}, nil
	}

	digestStr := obj.Digest.String()
	return &graphqltypes.TryGetPastObjectResponse{
		Current: graphqltypes.TryGetPastObjectCurrentObject{
			Address: obj.ID,
			Version: obj.Version,
		},
		Object: graphqltypes.TryGetPastObjectObject{
			RPC_OBJECT_FIELDS: graphqltypes.RPC_OBJECT_FIELDS{
				ObjectId: obj.ID,
				Version:  obj.Version,
				Status:   graphqltypes.ObjectKindIndexed,
				AsMoveObjectType: graphqltypes.RPC_OBJECT_FIELDSAsMoveObjectTypeMoveObject{
					Contents: graphqltypes.RPC_OBJECT_FIELDSAsMoveObjectTypeMoveObjectContentsMoveValue{
						Type: graphqltypes.RPC_OBJECT_FIELDSAsMoveObjectTypeMoveObjectContentsMoveValueTypeMoveType{
							Repr: obj.Type,
						},
					},
				},
				AsMoveObject: graphqltypes.RPC_OBJECT_FIELDSAsMoveObject{
					Contents: graphqltypes.RPC_OBJECT_FIELDSAsMoveObjectContentsMoveValue{
						Bcs: obj.Data,
						Type: graphqltypes.RPC_OBJECT_FIELDSAsMoveObjectContentsMoveValueTypeMoveType{
							Repr: obj.Type,
						},
					},
				},
				Owner:         buildRPCObjectOwner(obj.Owner),
				StorageRebate: *graphqltypes.NewBigInt(0),
				Digest:        digestStr,
				PreviousTransactionBlock: graphqltypes.RPC_OBJECT_FIELDSPreviousTransactionBlock{
					Digest: obj.PreviousTx.String(),
				},
			},
		},
	}, nil
}

func (c *FakeIotaClient) GetDynamicFieldObject(
	_ context.Context,
	req iotagraphql.GetDynamicFieldObjectRequest,
) (*iotagraphql.GetDynamicFieldObjectResponse, error) {
	return nil, fmt.Errorf("FakeIotaClient: GetDynamicFieldObject not implemented")
}

func (c *FakeIotaClient) GetDynamicFields(
	_ context.Context,
	req iotagraphql.GetDynamicFieldsRequest,
) (*graphqltypes.GetDynamicFieldsResponse, error) {
	if req.ParentObjectID == nil {
		return nil, fmt.Errorf("FakeIotaClient: GetDynamicFields: parentObjectID is nil")
	}

	fields := c.Store.GetDynamicFields(*req.ParentObjectID)
	nodes := make([]graphqltypes.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicField, 0, len(fields))

	for _, df := range fields {
		valObj, ok := c.Store.Get(df.ValueObjID)
		if !ok {
			continue
		}

		node := graphqltypes.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicField{
			Name: graphqltypes.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldNameMoveValue{
				Json: df.Name.JSON,
				Type: graphqltypes.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldNameMoveValueTypeMoveType{
					Repr: df.Name.TypeRepr,
				},
			},
			Value: &graphqltypes.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveObject{
				Typename: "MoveObject",
				Contents: graphqltypes.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveObjectContentsMoveValue{
					Type: graphqltypes.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnectionNodesDynamicFieldValueMoveObjectContentsMoveValueTypeMoveType{
						Repr: valObj.Type,
					},
					Json: buildBalanceJSON(valObj),
				},
				Address: valObj.ID,
				Digest:  valObj.Digest.String(),
				Version: valObj.Version,
			},
		}
		nodes = append(nodes, node)
	}

	return &graphqltypes.GetDynamicFieldsResponse{
		Owner: graphqltypes.GetDynamicFieldsOwner{
			DynamicFields: graphqltypes.GetDynamicFieldsOwnerDynamicFieldsDynamicFieldConnection{
				Nodes: nodes,
			},
		},
	}, nil
}

func (c *FakeIotaClient) GetOwnedObjects(
	_ context.Context,
	req iotagraphql.GetOwnedObjectsRequest,
) (*graphqltypes.GetOwnedObjectsResponse, error) {
	if req.Address == nil {
		return nil, fmt.Errorf("FakeIotaClient: GetOwnedObjects: address is nil")
	}

	var objs []*SimObject
	if req.Filter != nil && req.Filter.Type != nil {
		objs = c.Store.GetByOwnerAndType(*req.Address, *req.Filter.Type)
	} else {
		objs = c.Store.GetByOwner(*req.Address)
	}

	nodes := make([]graphqltypes.GetOwnedObjectsAddressObjectsMoveObjectConnectionNodesMoveObject, 0, len(objs))
	for _, obj := range objs {
		nodes = append(nodes, graphqltypes.GetOwnedObjectsAddressObjectsMoveObjectConnectionNodesMoveObject{
			RPC_MOVE_OBJECT_FIELDS: buildRPCMoveObjectFields(obj),
		})
	}

	return &graphqltypes.GetOwnedObjectsResponse{
		Address: graphqltypes.GetOwnedObjectsAddress{
			Objects: graphqltypes.GetOwnedObjectsAddressObjectsMoveObjectConnection{
				Nodes: nodes,
			},
		},
	}, nil
}

func (c *FakeIotaClient) GetCoins(_ context.Context, req iotagraphql.GetCoinsRequest) (*iotagraphql.GetCoinsResponse, error) {
	if req.Owner == nil {
		return nil, fmt.Errorf("FakeIotaClient: GetCoins: owner is nil")
	}

	coinType := IotaCoinTypeStr
	if req.CoinType != nil {
		coinType = string(*req.CoinType)
	}

	coins := c.Store.GetCoinsByOwner(*req.Owner, coinType)
	coinNodes := make([]graphqltypes.CoinData, 0, len(coins))
	for _, coin := range coins {
		coinNodes = append(coinNodes, buildCoinData(coin, coinType))
	}

	return &graphqltypes.GetCoinsResponse{
		Address: graphqltypes.GetCoinsAddress{
			Address: *req.Owner,
			Coins: graphqltypes.GetCoinsAddressCoinsCoinConnection{
				Nodes: coinNodes,
			},
		},
	}, nil
}

func (c *FakeIotaClient) GetAllCoins(_ context.Context, req iotagraphql.GetAllCoinsRequest) (*iotagraphql.GetAllCoinsResponse, error) {
	if req.Owner == nil {
		return nil, fmt.Errorf("FakeIotaClient: GetAllCoins: owner is nil")
	}

	allObjs := c.Store.GetByOwner(*req.Owner)
	coinNodes := make([]graphqltypes.CoinData, 0)
	for _, obj := range allObjs {
		if ct, ok := extractCoinType(obj.Type); ok {
			coinNodes = append(coinNodes, buildCoinData(obj, ct))
		}
	}

	return &graphqltypes.GetAllCoinsResponse{
		Address: graphqltypes.GetAllCoinsAddress{
			Address: *req.Owner,
			Coins: graphqltypes.GetAllCoinsAddressCoinsCoinConnection{
				Nodes: coinNodes,
			},
		},
	}, nil
}

func (c *FakeIotaClient) GetBalance(_ context.Context, req iotagraphql.GetBalanceRequest) (*iotagraphql.Balance, error) {
	if req.Owner == nil {
		return nil, fmt.Errorf("FakeIotaClient: GetBalance: owner is nil")
	}

	coinType := IotaCoinTypeStr
	if req.CoinType != "" {
		coinType = string(req.CoinType)
	}

	coins := c.Store.GetCoinsByOwner(*req.Owner, coinType)
	var total uint64
	for _, coin := range coins {
		total += DecodeCoinObjectBalance(coin.Data)
	}

	return &graphqltypes.Balance{
		CoinType:        graphqltypes.CoinType(coinType),
		CoinObjectCount: graphqltypes.NewBigInt(uint64(len(coins))),
		TotalBalance:    graphqltypes.NewBigInt(total),
	}, nil
}

func (c *FakeIotaClient) GetAllBalances(_ context.Context, owner iotago.Address) ([]*iotagraphql.Balance, error) {
	balances := c.Store.GetAllCoinBalances(owner)
	result := make([]*graphqltypes.Balance, 0, len(balances))

	coinCounts := make(map[string]int)
	objs := c.Store.GetByOwner(owner)
	for _, obj := range objs {
		if ct, ok := extractCoinType(obj.Type); ok {
			coinCounts[ct]++
		}
	}

	for ct, total := range balances {
		result = append(result, &graphqltypes.Balance{
			CoinType:        graphqltypes.CoinType(ct),
			CoinObjectCount: graphqltypes.NewBigInt(safecast.MustConvert[uint64](coinCounts[ct])),
			TotalBalance:    graphqltypes.NewBigInt(total),
		})
	}
	return result, nil
}

func (c *FakeIotaClient) GetCoinMetadata(_ context.Context, coinType iotagraphql.CoinType) (*iotagraphql.IotaCoinMetadata, error) {
	return &iotagraphql.IotaCoinMetadata{
		Name:     "IOTA",
		Symbol:   "IOTA",
		Decimals: 9,
	}, nil
}

func (c *FakeIotaClient) GetTotalSupply(_ context.Context, coinType iotagraphql.CoinType) (*iotagraphql.Supply, error) {
	var total uint64
	c.Store.mu.RLock()
	for _, obj := range c.Store.objects {
		if ct, ok := extractCoinType(obj.Type); ok && ct == string(coinType) {
			total += DecodeCoinObjectBalance(obj.Data)
		}
	}
	c.Store.mu.RUnlock()

	return &iotagraphql.Supply{
		Value: graphqltypes.NewBigInt(total),
	}, nil
}

func (c *FakeIotaClient) GetCoinObjsForTargetAmount(
	ctx context.Context,
	address iotago.Address,
	targetAmount uint64,
	gasAmount uint64,
) (iotagraphql.Coins, error) {
	coins, err := c.GetCoins(ctx, iotagraphql.GetCoinsRequest{
		Owner: &address,
		Limit: 50,
	})
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: GetCoinObjsForTargetAmount: %w", err)
	}
	pickedCoins, err := graphqltypes.PickupCoins(
		graphqltypes.Coins(coins.Address.Coins.Nodes),
		new(big.Int).SetUint64(targetAmount),
		gasAmount, 0, 25,
	)
	if err != nil {
		return nil, err
	}
	return pickedCoins.Coins, nil
}

func (c *FakeIotaClient) PayIota(_ context.Context, req iotagraphql.PayIotaRequest) (*iotagraphql.TransactionBytes, error) {
	if req.Signer == nil || len(req.InputCoins) == 0 || len(req.Recipients) == 0 || len(req.Amount) == 0 {
		return nil, fmt.Errorf("FakeIotaClient: PayIota: missing required fields")
	}

	ptb := iotago.NewProgrammableTransactionBuilder()

	amounts := make([]iotago.Argument, len(req.Amount))
	for i, amt := range req.Amount {
		amounts[i] = ptb.MustPure(amt.Uint64())
	}
	splitResults := ptb.Command(iotago.Command{
		SplitCoins: &iotago.ProgrammableSplitCoins{
			Coin:    iotago.GetArgumentGasCoin(),
			Amounts: amounts,
		},
	})

	for i, recipient := range req.Recipients {
		if recipient == nil {
			continue
		}
		idx := safecast.MustConvert[uint16](i)
		ptb.Command(iotago.Command{
			TransferObjects: &iotago.ProgrammableTransferObjects{
				Objects: []iotago.Argument{{NestedResult: &iotago.NestedResult{Cmd: *splitResults.Result, Result: idx}}},
				Address: ptb.MustPure(*recipient),
			},
		})
	}

	pt := ptb.Finish()

	gasPayments := make([]*iotago.ObjectRef, 0, len(req.InputCoins))
	for _, coinID := range req.InputCoins {
		obj, ok := c.Store.Get(coinID)
		if !ok {
			return nil, fmt.Errorf("FakeIotaClient: PayIota: coin %s not found", coinID.String())
		}
		gasPayments = append(gasPayments, obj.Ref())
	}

	gasBudget := uint64(50_000_000)
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}

	tx := iotago.NewProgrammable(req.Signer, pt, gasPayments, gasBudget, defaultGasPrice)
	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: PayIota: BCS marshal: %w", err)
	}

	return &graphqltypes.TransactionBytes{TxBytes: txBytes}, nil
}

func (c *FakeIotaClient) MergeCoins(_ context.Context, req iotagraphql.MergeCoinsRequest) (*iotagraphql.TransactionBytes, error) {
	return nil, fmt.Errorf("FakeIotaClient: MergeCoins not implemented")
}

func (c *FakeIotaClient) PayAllIota(_ context.Context, req iotagraphql.PayAllIotaRequest) (*iotagraphql.TransactionBytes, error) {
	return nil, fmt.Errorf("FakeIotaClient: PayAllIota not implemented")
}

func (c *FakeIotaClient) Publish(_ context.Context, req iotagraphql.PublishRequest) (*iotagraphql.TransactionBytes, error) {
	if req.Sender == nil {
		return nil, fmt.Errorf("FakeIotaClient: Publish: sender is nil")
	}

	modules := make([][]byte, len(req.CompiledModules))
	for i, module := range req.CompiledModules {
		if module == nil {
			continue
		}
		modules[i] = module.Data()
	}

	ptb := iotago.NewProgrammableTransactionBuilder()
	capArg := ptb.PublishUpgradeable(modules, req.Dependencies)
	ptb.TransferArgs(req.Sender, []iotago.Argument{capArg})
	pt := ptb.Finish()

	gasBudget := uint64(50_000_000)
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}

	coins := c.Store.GetCoinsByOwner(*req.Sender, IotaCoinTypeStr)
	if len(coins) == 0 {
		return nil, fmt.Errorf("FakeIotaClient: Publish: no gas coins for %s", req.Sender.String())
	}

	gasPayments := []*iotago.ObjectRef{coins[0].Ref()}

	tx := iotago.NewProgrammable(req.Sender, pt, gasPayments, gasBudget, defaultGasPrice)
	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: Publish: BCS marshal: %w", err)
	}

	return &graphqltypes.TransactionBytes{TxBytes: txBytes}, nil
}

func (c *FakeIotaClient) TransferIota(_ context.Context, req iotagraphql.TransferIotaRequest) (*iotagraphql.TransactionBytes, error) {
	return nil, fmt.Errorf("FakeIotaClient: TransferIota not implemented")
}

func (c *FakeIotaClient) TransferObject(_ context.Context, req iotagraphql.TransferObjectRequest) (*iotagraphql.TransactionBytes, error) {
	return nil, fmt.Errorf("FakeIotaClient: TransferObject not implemented")
}

func (c *FakeIotaClient) GetReferenceGasPrice(_ context.Context) (*iotagraphql.BigInt, error) {
	return graphqltypes.NewBigInt(defaultGasPrice), nil
}

func (c *FakeIotaClient) GetLatestIotaSystemState(_ context.Context) (*iotagraphql.GetLatestIotaSystemStateResponse, error) {
	now := time.Now()
	epochStart := now.Add(-1 * time.Hour)
	epochEnd := now.Add(23 * time.Hour)

	return &graphqltypes.GetLatestIotaSystemStateResponse{
		Epoch: graphqltypes.GetLatestIotaSystemStateEpoch{
			EpochId:            0,
			StartTimestamp:     epochStart,
			EndTimestamp:       epochEnd,
			ReferenceGasPrice:  *graphqltypes.NewBigInt(defaultGasPrice),
			IotaTotalSupply:    *graphqltypes.NewBigInt(10_000_000_000_000_000_000), // 10B IOTA
			SystemStateVersion: 1,
			ProtocolConfigs: graphqltypes.GetLatestIotaSystemStateEpochProtocolConfigs{
				ProtocolVersion: 1,
			},
			SafeMode: graphqltypes.GetLatestIotaSystemStateEpochSafeMode{
				GasSummary: graphqltypes.GetLatestIotaSystemStateEpochSafeModeGasSummaryGasCostSummary{
					ComputationCost:         *graphqltypes.NewBigInt(0),
					NonRefundableStorageFee: *graphqltypes.NewBigInt(0),
					StorageCost:             *graphqltypes.NewBigInt(0),
					StorageRebate:           *graphqltypes.NewBigInt(0),
				},
			},
			StorageFund: graphqltypes.GetLatestIotaSystemStateEpochStorageFund{
				NonRefundableBalance:      *graphqltypes.NewBigInt(0),
				TotalObjectStorageRebates: *graphqltypes.NewBigInt(0),
			},
			SystemParameters: graphqltypes.GetLatestIotaSystemStateEpochSystemParameters{
				MinValidatorCount:              4,
				MaxValidatorCount:              150,
				MinValidatorJoiningStake:       *graphqltypes.NewBigInt(30_000_000_000_000),
				DurationMs:                     *graphqltypes.NewBigInt(86_400_000),
				ValidatorLowStakeThreshold:     *graphqltypes.NewBigInt(20_000_000_000_000),
				ValidatorLowStakeGracePeriod:   *graphqltypes.NewBigInt(7),
				ValidatorVeryLowStakeThreshold: *graphqltypes.NewBigInt(15_000_000_000_000),
			},
			ValidatorSet: graphqltypes.GetLatestIotaSystemStateEpochValidatorSet{
				TotalStake: *graphqltypes.NewBigInt(1_000_000_000_000_000),
			},
		},
	}, nil
}

func (c *FakeIotaClient) UpdateObjectRef(_ context.Context, ref *iotago.ObjectRef) (*iotago.ObjectRef, error) {
	if ref == nil || ref.ObjectID == nil {
		return ref, nil
	}
	obj, ok := c.Store.Get(*ref.ObjectID)
	if !ok {
		return ref, nil
	}
	return obj.Ref(), nil
}

func (c *FakeIotaClient) MintToken(
	ctx context.Context,
	signer iotasigner.Signer,
	packageID iotago.PackageID,
	tokenName string,
	treasuryCap *iotago.ObjectRef,
	mintAmount uint64,
	_ int,
) (*graphqltypes.ExecuteTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb.Command(iotago.Command{
		MoveCall: &iotago.ProgrammableMoveCall{
			Package:       &packageID,
			Module:        tokenName,
			Function:      "mint",
			TypeArguments: []iotago.TypeTag{},
			Arguments: []iotago.Argument{
				ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: treasuryCap}),
				ptb.MustForceSeparatePure(mintAmount),
				ptb.MustForceSeparatePure(signer.Address()),
			},
		},
	})
	pt := ptb.Finish()

	coins := c.Store.GetCoinsByOwner(*signer.Address(), IotaCoinTypeStr)
	if len(coins) == 0 {
		return nil, fmt.Errorf("FakeIotaClient: MintToken: no gas coins")
	}
	gasPayments := []*iotago.ObjectRef{coins[0].Ref()}

	tx := iotago.NewProgrammable(signer.Address(), pt, gasPayments, iotagraphql.DefaultGasBudget, defaultGasPrice)
	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("FakeIotaClient: MintToken: %w", err)
	}

	return c.SignAndExecuteTransaction(ctx, txBytes, signer)
}

func (c *FakeIotaClient) RequestFundsFromFaucet(_ context.Context, receiverAddress iotago.Address) error {
	c.faucetMu.Lock()
	defer c.faucetMu.Unlock()

	const FaucetCoinsAmount = 5

	for i := 0; i < FaucetCoinsAmount; i++ {
		buf := make([]byte, 32+8)
		copy(buf, receiverAddress[:])
		binary.LittleEndian.PutUint64(buf[32:], c.faucetCounter)
		c.faucetCounter++
		txDigest := ComputeDigest(buf)
		coinID := FreshID(txDigest, &c.faucetCounter)
		c.Store.PresetCoinObject(coinID, receiverAddress, IotaCoinTypeStr, faucetAmount, txDigest)
	}

	return nil
}

func buildCoinData(obj *SimObject, coinType string) graphqltypes.CoinData {
	balance := DecodeCoinObjectBalance(obj.Data)
	return graphqltypes.CoinData{
		COIN_DATA: graphqltypes.COIN_DATA{
			Address:     obj.ID,
			Version:     obj.Version,
			Digest:      obj.Digest.String(),
			CoinBalance: *graphqltypes.NewBigInt(balance),
			Contents: graphqltypes.COIN_DATAContentsMoveValue{
				Type: graphqltypes.COIN_DATAContentsMoveValueTypeMoveType{
					Repr: CoinTypeString(coinType),
				},
			},
		},
	}
}

func buildRPCMoveObjectFields(obj *SimObject) graphqltypes.RPC_MOVE_OBJECT_FIELDS {
	return graphqltypes.RPC_MOVE_OBJECT_FIELDS{
		ObjectId: obj.ID,
		Bcs:      obj.Data,
		Status:   graphqltypes.ObjectKindIndexed,
		Contents_type: graphqltypes.RPC_MOVE_OBJECT_FIELDSContents_typeMoveValue{
			Type: graphqltypes.RPC_MOVE_OBJECT_FIELDSContents_typeMoveValueTypeMoveType{
				Repr: obj.Type,
			},
		},
		Contents_content: graphqltypes.RPC_MOVE_OBJECT_FIELDSContents_contentMoveValue{
			Type: graphqltypes.RPC_MOVE_OBJECT_FIELDSContents_contentMoveValueTypeMoveType{
				Repr: obj.Type,
			},
		},
		Contents: graphqltypes.RPC_MOVE_OBJECT_FIELDSContentsMoveValue{
			Bcs: obj.Data,
			Type: graphqltypes.RPC_MOVE_OBJECT_FIELDSContentsMoveValueTypeMoveType{
				Repr: obj.Type,
			},
		},
		Owner:         buildRPCMoveObjectOwner(obj.Owner),
		StorageRebate: *graphqltypes.NewBigInt(0),
		Digest:        obj.Digest.String(),
		PreviousTransactionBlock: graphqltypes.RPC_MOVE_OBJECT_FIELDSPreviousTransactionBlock{
			Digest: obj.PreviousTx.String(),
		},
	}
}

func buildBalanceJSON(obj *SimObject) json.RawMessage {
	balance := DecodeBalanceValue(obj.Data)
	return json.RawMessage(fmt.Sprintf(`{"value":"%d"}`, balance))
}
