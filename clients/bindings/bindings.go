package bindings

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/bindings/iota_sdk_ffi"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

type BindingClient struct {
	RpcURL  string
	qclient *iota_sdk_ffi.GraphQlClient
}

func NewBindingClient(rpcUrl string) *BindingClient {
	var client BindingClient

	switch rpcUrl {
	case iotaconn.LocalnetEndpointURL:
		client.qclient = iota_sdk_ffi.GraphQlClientNewLocalhost()
	case iotaconn.TestnetEndpointURL:
		client.qclient = iota_sdk_ffi.GraphQlClientNewTestnet()
	case iotaconn.DevnetEndpointURL:
		client.qclient = iota_sdk_ffi.GraphQlClientNewDevnet()
	default:
		qclient, err := iota_sdk_ffi.NewGraphQlClient(rpcUrl)
		if err != nil {
			panic(err)
		}
		client.qclient = qclient
	}
	client.RpcURL = rpcUrl
	return &client
}

// L2ClientInterface defines the interface that matches clients.L2Client
// We use interface{} to avoid circular import issues
type L2ClientInterface any

// Helpers: convert between iotago and FFI types
func toFfiAddress(addr *iotago.Address) (*iota_sdk_ffi.Address, error) {
	if addr == nil {
		return nil, nil
	}
	return iota_sdk_ffi.AddressFromHex(addr.String())
}

func toFfiObjectID(id *iotago.ObjectID) (*iota_sdk_ffi.ObjectId, error) {
	if id == nil {
		return nil, nil
	}
	return iota_sdk_ffi.ObjectIdFromHex(id.String())
}

func fromFfiObjectID(id *iota_sdk_ffi.ObjectId) (*iotago.ObjectID, error) {
	if id == nil {
		return nil, nil
	}
	return iotago.ObjectIDFromHex(id.ToHex())
}

func fromFfiDigest(d *iota_sdk_ffi.Digest) (*iotago.Digest, error) {
	if d == nil {
		return nil, nil
	}
	return iotago.NewDigest(d.ToBase58())
}

func mapFfiObjectToIotaResponse(obj **iota_sdk_ffi.Object) (*iotajsonrpc.IotaObjectResponse, error) {
	if obj == nil || *obj == nil {
		return &iotajsonrpc.IotaObjectResponse{}, nil
	}
	o := *obj

	oid, err := fromFfiObjectID(o.ObjectId())
	if err != nil {
		return nil, err
	}
	ver := o.Version()
	dg, err := fromFfiDigest(o.Digest())
	if err != nil {
		return nil, err
	}
	var typeStr *string
	if ot := o.ObjectType(); ot != nil {
		s := ot.String()
		typeStr = &s
	}
	var prev *iotago.TransactionDigest
	if pd := o.PreviousTransaction(); pd != nil {
		prev, _ = fromFfiDigest(pd)
	}

	data := &iotajsonrpc.IotaObjectData{
		ObjectID:            oid,
		Version:             iotajsonrpc.NewBigInt(ver),
		Digest:              dg,
		Type:                typeStr,
		PreviousTransaction: prev,
		StorageRebate:       iotajsonrpc.NewBigInt(o.StorageRebate()),
	}
	return &iotajsonrpc.IotaObjectResponse{Data: data}, nil
}

func (c *BindingClient) GetDynamicFieldObject(ctx context.Context, req iotaclient.GetDynamicFieldObjectRequest) (*iotajsonrpc.IotaObjectResponse, error) {
	return nil, errors.New("GetDynamicFieldObject not supported by FFI bindings yet")
}

func (c *BindingClient) GetDynamicFields(ctx context.Context, req iotaclient.GetDynamicFieldsRequest) (*iotajsonrpc.DynamicFieldPage, error) {
	return nil, errors.New("GetDynamicFields not supported by FFI bindings yet")
}

func (c *BindingClient) GetOwnedObjects(ctx context.Context, req iotaclient.GetOwnedObjectsRequest) (*iotajsonrpc.ObjectsPage, error) {
	if req.Address == nil {
		return &iotajsonrpc.ObjectsPage{}, nil
	}
	owner, err := toFfiAddress(req.Address)
	if err != nil {
		return nil, err
	}
	// Best-effort: list coins for the owner and map to objects
	cp, err := c.qclient.Coins(owner, nil, nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("Coins failed: %w", err)
	}
	var page iotajsonrpc.ObjectsPage
	for _, coin := range cp.Data {
		obj, err := c.qclient.Object(coin.Id(), nil)
		if err.(*iota_sdk_ffi.SdkFfiError) != nil {
			return nil, fmt.Errorf("Object failed: %w", err)
		}
		resp, err := mapFfiObjectToIotaResponse(obj)
		if err != nil {
			return nil, err
		}
		page.Data = append(page.Data, *resp)
	}
	page.HasNextPage = cp.PageInfo.HasNextPage
	return &page, nil
}

func (c *BindingClient) QueryEvents(ctx context.Context, req iotaclient.QueryEventsRequest) (*iotajsonrpc.EventPage, error) {
	ep, err := c.qclient.Events(nil, nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL Events failed: %w", err)
	}
	// Not mapping individual events; return page info only.
	return &iotajsonrpc.EventPage{HasNextPage: ep.PageInfo.HasNextPage}, nil
}

func (c *BindingClient) QueryTransactionBlocks(ctx context.Context, req iotaclient.QueryTransactionBlocksRequest) (*iotajsonrpc.TransactionBlocksPage, error) {
	return nil, errors.New("QueryTransactionBlocks not supported by FFI bindings yet")
}

func (c *BindingClient) ResolveNameServiceAddress(ctx context.Context, iotaName string) (*iotago.Address, error) {
	return nil, errors.New("ResolveNameServiceAddress not implemented for BindingClient")
}

func (c *BindingClient) ResolveNameServiceNames(ctx context.Context, req iotaclient.ResolveNameServiceNamesRequest) (*iotajsonrpc.IotaNamePage, error) {
	return nil, errors.New("ResolveNameServiceNames not implemented for BindingClient")
}

func (c *BindingClient) DevInspectTransactionBlock(ctx context.Context, req iotaclient.DevInspectTransactionBlockRequest) (*iotajsonrpc.DevInspectResults, error) {
	return nil, errors.New("DevInspectTransactionBlock not supported by FFI bindings yet")
}

func (c *BindingClient) DryRunTransaction(ctx context.Context, txDataBytes iotago.Base64Data) (*iotajsonrpc.DryRunTransactionBlockResponse, error) {
	return nil, errors.New("DryRunTransaction not supported by FFI bindings yet")
}

func (c *BindingClient) ExecuteTransactionBlock(ctx context.Context, req iotaclient.ExecuteTransactionBlockRequest) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	return nil, errors.New("ExecuteTransactionBlock not supported by FFI bindings yet")
}

func (c *BindingClient) GetCommitteeInfo(ctx context.Context, epoch *iotajsonrpc.BigInt) (*iotajsonrpc.CommitteeInfo, error) {
	var epochPtr *uint64
	if epoch != nil {
		e := epoch.Uint64()
		epochPtr = &e
	}

	validators, err := c.qclient.ActiveValidators(epochPtr, nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("ActiveValidators failed: %w", err)
	}

	committeeInfo := &iotajsonrpc.CommitteeInfo{
		EpochId: epoch,
	}

	for _, validator := range validators.Data {
		var stake uint64
		if validator.StakingPoolIotaBalance != nil {
			stake = *validator.StakingPoolIotaBalance
		} else if validator.NextEpochStake != nil {
			stake = *validator.NextEpochStake
		}

		var publicKey *iotago.Base64Data
		if validator.Credentials != nil && validator.Credentials.ProtocolPubKey != nil {
			pkStr := string(*validator.Credentials.ProtocolPubKey)
			publicKey, _ = iotago.NewBase64Data(pkStr)
		}

		committeeInfo.Validators = append(committeeInfo.Validators, iotajsonrpc.Validator{
			PublicKey: publicKey,
			Stake:     iotajsonrpc.NewBigInt(stake),
		})
	}

	return committeeInfo, nil
}

func (c *BindingClient) GetLatestIotaSystemState(ctx context.Context) (*iotajsonrpc.IotaSystemStateSummary, error) {
	return nil, errors.New("GetLatestIotaSystemState not supported by FFI bindings yet")
}

func (c *BindingClient) GetReferenceGasPrice(ctx context.Context) (*iotajsonrpc.BigInt, error) {
	// Direct mapping to GraphQL client's ReferenceGasPrice method
	gasPrice, err := c.qclient.ReferenceGasPrice(nil) // Use current epoch
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL ReferenceGasPrice failed: %w", err)
	}
	if gasPrice == nil {
		return iotajsonrpc.NewBigInt(0), nil
	}
	return iotajsonrpc.NewBigInt(*gasPrice), nil
}

func (c *BindingClient) GetStakes(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.DelegatedStake, error) {
	return nil, errors.New("GetStakes not supported by FFI bindings yet")
}

func (c *BindingClient) GetStakesByIds(ctx context.Context, stakedIotaIds []iotago.ObjectID) ([]*iotajsonrpc.DelegatedStake, error) {
	return nil, errors.New("GetStakesByIds not supported by FFI bindings yet")
}

func (c *BindingClient) GetValidatorsApy(ctx context.Context) (*iotajsonrpc.ValidatorsApy, error) {
	return nil, errors.New("GetValidatorsApy not implemented for BindingClient")
}

func (c *BindingClient) BatchTransaction(ctx context.Context, req iotaclient.BatchTransactionRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget
	builder.SetGasBudget(req.GasBudget)

	// Add gas coins
	if req.Gas != nil {
		// Get the gas coin object
		gasObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.Gas})
		if err != nil {
			return nil, fmt.Errorf("failed to get gas object: %w", err)
		}
		if gasObj.Data == nil {
			return nil, fmt.Errorf("gas object not found")
		}

		gasRef := &iotago.ObjectRef{
			ObjectID: gasObj.Data.ObjectID,
			Version:  gasObj.Data.Version.Uint64(),
			Digest:   gasObj.Data.Digest,
		}
		gasInput, err := convertObjectRefToUnresolvedInput(gasRef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder.AddGasObjects([]*iota_sdk_ffi.UnresolvedInput{gasInput})
	} else {
		// Find suitable gas coins
		err := c.addGasCoinsToBuilder(ctx, builder, req.Signer, req.GasBudget, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to add gas coins: %w", err)
		}
	}

	// Process each transaction parameter
	for i, txParam := range req.TxnParams {
		cmdType, ok := txParam["command"].(string)
		if !ok {
			return nil, fmt.Errorf("transaction parameter %d missing command type", i)
		}

		switch cmdType {
		case "MoveCall":
			// Extract MoveCall parameters
			pkg, ok := txParam["package"].(string)
			if !ok {
				return nil, fmt.Errorf("MoveCall missing package")
			}
			module, ok := txParam["module"].(string)
			if !ok {
				return nil, fmt.Errorf("MoveCall missing module")
			}
			function, ok := txParam["function"].(string)
			if !ok {
				return nil, fmt.Errorf("MoveCall missing function")
			}

			// Build Function
			packageAddr, err := iota_sdk_ffi.AddressFromHex(pkg)
			if err != nil {
				return nil, fmt.Errorf("failed to convert package address: %w", err)
			}

			moduleId, err := iota_sdk_ffi.NewIdentifier(module)
			if err != nil {
				return nil, fmt.Errorf("failed to create module identifier: %w", err)
			}

			functionId, err := iota_sdk_ffi.NewIdentifier(function)
			if err != nil {
				return nil, fmt.Errorf("failed to create function identifier: %w", err)
			}

			moveFunction := iota_sdk_ffi.Function{
				Package:  packageAddr,
				Module:   moduleId,
				Function: functionId,
				TypeArgs: []*iota_sdk_ffi.TypeTag{}, // TODO: Handle type args
			}

			// Convert arguments (simplified - real implementation would be more complex)
			var args []*iota_sdk_ffi.Argument
			if argsParam, ok := txParam["arguments"].([]interface{}); ok {
				for _, arg := range argsParam {
					switch v := arg.(type) {
					case string:
						pureInput := iota_sdk_ffi.UnresolvedInputNewPure([]byte(v))
						args = append(args, builder.Input(pureInput))
					default:
						// For object references, this would need more complex handling
						return nil, fmt.Errorf("unsupported argument type in batch transaction")
					}
				}
			}

			builder.MoveCall(moveFunction, args)

		case "TransferObjects":
			// Handle transfer objects command (simplified)
			return nil, fmt.Errorf("TransferObjects not implemented in BatchTransaction yet")

		case "SplitCoins":
			// Handle split coins command (simplified)
			return nil, fmt.Errorf("SplitCoins not implemented in BatchTransaction yet")

		case "MergeCoins":
			// Handle merge coins command (simplified)
			return nil, fmt.Errorf("MergeCoins not implemented in BatchTransaction yet")

		default:
			return nil, fmt.Errorf("unsupported command type: %s", cmdType)
		}
	}

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) MergeCoins(ctx context.Context, req iotaclient.MergeCoinsRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// Add gas coin if specified, otherwise find suitable gas coins
	if req.Gas != nil {
		// Get the gas coin object
		gasObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.Gas})
		if err != nil {
			return nil, fmt.Errorf("failed to get gas object: %w", err)
		}
		if gasObj.Data == nil {
			return nil, fmt.Errorf("gas object not found")
		}

		gasRef := &iotago.ObjectRef{
			ObjectID: gasObj.Data.ObjectID,
			Version:  gasObj.Data.Version.Uint64(),
			Digest:   gasObj.Data.Digest,
		}
		gasInput, err := convertObjectRefToUnresolvedInput(gasRef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder.AddGasObjects([]*iota_sdk_ffi.UnresolvedInput{gasInput})
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		err := c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to add gas coins: %w", err)
		}
	}

	// Get primary coin object
	primaryObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.PrimaryCoin})
	if err != nil {
		return nil, fmt.Errorf("failed to get primary coin object: %w", err)
	}
	if primaryObj.Data == nil {
		return nil, fmt.Errorf("primary coin object not found")
	}

	primaryRef := &iotago.ObjectRef{
		ObjectID: primaryObj.Data.ObjectID,
		Version:  primaryObj.Data.Version.Uint64(),
		Digest:   primaryObj.Data.Digest,
	}
	primaryInput, err := convertObjectRefToUnresolvedInput(primaryRef)
	if err != nil {
		return nil, fmt.Errorf("failed to convert primary coin: %w", err)
	}
	primaryArg := builder.Input(primaryInput)

	// Get coin to merge object
	coinToMergeObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.CoinToMerge})
	if err != nil {
		return nil, fmt.Errorf("failed to get coin to merge object: %w", err)
	}
	if coinToMergeObj.Data == nil {
		return nil, fmt.Errorf("coin to merge object not found")
	}

	coinToMergeRef := &iotago.ObjectRef{
		ObjectID: coinToMergeObj.Data.ObjectID,
		Version:  coinToMergeObj.Data.Version.Uint64(),
		Digest:   coinToMergeObj.Data.Digest,
	}
	coinToMergeInput, err := convertObjectRefToUnresolvedInput(coinToMergeRef)
	if err != nil {
		return nil, fmt.Errorf("failed to convert coin to merge: %w", err)
	}
	coinToMergeArg := builder.Input(coinToMergeInput)

	// Add merge coins command
	builder.MergeCoins(primaryArg, []*iota_sdk_ffi.Argument{coinToMergeArg})

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) MoveCall(ctx context.Context, req iotaclient.MoveCallRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if req.PackageID == nil {
		return nil, fmt.Errorf("package ID is required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins
	if req.Gas != nil {
		// Get the gas coin object
		gasObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.Gas})
		if err != nil {
			return nil, fmt.Errorf("failed to get gas object: %w", err)
		}
		if gasObj.Data == nil {
			return nil, fmt.Errorf("gas object not found")
		}

		gasRef := &iotago.ObjectRef{
			ObjectID: gasObj.Data.ObjectID,
			Version:  gasObj.Data.Version.Uint64(),
			Digest:   gasObj.Data.Digest,
		}
		gasInput, err := convertObjectRefToUnresolvedInput(gasRef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder.AddGasObjects([]*iota_sdk_ffi.UnresolvedInput{gasInput})
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		err := c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to add gas coins: %w", err)
		}
	}

	// Build the Function struct
	packageAddr, err := iota_sdk_ffi.AddressFromHex(req.PackageID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert package address: %w", err)
	}

	moduleId, err := iota_sdk_ffi.NewIdentifier(req.Module)
	if err != nil {
		return nil, fmt.Errorf("failed to create module identifier: %w", err)
	}

	functionId, err := iota_sdk_ffi.NewIdentifier(req.Function)
	if err != nil {
		return nil, fmt.Errorf("failed to create function identifier: %w", err)
	}

	// TODO: Handle TypeArgs - this would require parsing string types into TypeTag
	var typeArgs []*iota_sdk_ffi.TypeTag

	function := iota_sdk_ffi.Function{
		Package:  packageAddr,
		Module:   moduleId,
		Function: functionId,
		TypeArgs: typeArgs,
	}

	// Convert arguments
	var args []*iota_sdk_ffi.Argument
	for _, arg := range req.Arguments {
		switch v := arg.(type) {
		case string:
			// Try to parse as address
			if addr, err := iotago.AddressFromHex(v); err == nil {
				pureInput := iota_sdk_ffi.UnresolvedInputNewPure(addr.Bytes())
				args = append(args, builder.Input(pureInput))
			} else {
				// Treat as string literal - encode as BCS
				// For now, treat as raw bytes
				pureInput := iota_sdk_ffi.UnresolvedInputNewPure([]byte(v))
				args = append(args, builder.Input(pureInput))
			}
		case uint64:
			// Encode uint64 as BCS bytes
			// Simple big-endian encoding for now
			bytes := make([]byte, 8)
			for i := 7; i >= 0; i-- {
				bytes[i] = byte(v)
				v >>= 8
			}
			pureInput := iota_sdk_ffi.UnresolvedInputNewPure(bytes)
			args = append(args, builder.Input(pureInput))
		case *iotago.ObjectID:
			// Get object and convert to input
			obj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: v})
			if err != nil {
				return nil, fmt.Errorf("failed to get argument object: %w", err)
			}
			if obj.Data == nil {
				return nil, fmt.Errorf("argument object not found")
			}

			objRef := &iotago.ObjectRef{
				ObjectID: obj.Data.ObjectID,
				Version:  obj.Data.Version.Uint64(),
				Digest:   obj.Data.Digest,
			}
			objInput, err := convertObjectRefToUnresolvedInput(objRef)
			if err != nil {
				return nil, fmt.Errorf("failed to convert argument object: %w", err)
			}
			args = append(args, builder.Input(objInput))
		default:
			return nil, fmt.Errorf("unsupported argument type: %T", arg)
		}
	}

	// Add move call command
	builder.MoveCall(function, args)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) Pay(ctx context.Context, req iotaclient.PayRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if len(req.Recipients) != len(req.Amount) {
		return nil, fmt.Errorf("recipients and amounts must have same length")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins
	if req.Gas != nil {
		// Get the gas coin object
		gasObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.Gas})
		if err != nil {
			return nil, fmt.Errorf("failed to get gas object: %w", err)
		}
		if gasObj.Data == nil {
			return nil, fmt.Errorf("gas object not found")
		}

		gasRef := &iotago.ObjectRef{
			ObjectID: gasObj.Data.ObjectID,
			Version:  gasObj.Data.Version.Uint64(),
			Digest:   gasObj.Data.Digest,
		}
		gasInput, err := convertObjectRefToUnresolvedInput(gasRef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder.AddGasObjects([]*iota_sdk_ffi.UnresolvedInput{gasInput})
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		err := c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to add gas coins: %w", err)
		}
	}

	// Strategy: Use one of the input coins as source, split it for required amounts, then transfer
	if len(req.InputCoins) == 0 {
		return nil, fmt.Errorf("no input coins provided")
	}

	// Get the first input coin to use as primary
	primaryCoinObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.InputCoins[0]})
	if err != nil {
		return nil, fmt.Errorf("failed to get primary coin object: %w", err)
	}
	if primaryCoinObj.Data == nil {
		return nil, fmt.Errorf("primary coin object not found")
	}

	primaryCoinRef := &iotago.ObjectRef{
		ObjectID: primaryCoinObj.Data.ObjectID,
		Version:  primaryCoinObj.Data.Version.Uint64(),
		Digest:   primaryCoinObj.Data.Digest,
	}
	primaryCoinInput, err := convertObjectRefToUnresolvedInput(primaryCoinRef)
	if err != nil {
		return nil, fmt.Errorf("failed to convert primary coin: %w", err)
	}
	primaryCoinArg := builder.Input(primaryCoinInput)

	// Create amount arguments for splitting
	var amountArgs []*iota_sdk_ffi.Argument
	for _, amount := range req.Amount {
		// Encode amount as BCS bytes (little-endian u64)
		amountVal := amount.Uint64()
		bytes := make([]byte, 8)
		for i := 0; i < 8; i++ {
			bytes[i] = byte(amountVal)
			amountVal >>= 8
		}
		amountInput := iota_sdk_ffi.UnresolvedInputNewPure(bytes)
		amountArgs = append(amountArgs, builder.Input(amountInput))
	}

	// Split the primary coin
	splitResult := builder.SplitCoins(primaryCoinArg, amountArgs)

	// Transfer each split result to corresponding recipient
	for i, recipient := range req.Recipients {
		// Create recipient argument
		recipientInput := iota_sdk_ffi.UnresolvedInputNewPure(recipient.Bytes())
		recipientArg := builder.Input(recipientInput)

		// Get the i-th split coin using nested access
		coinToTransferPtr := splitResult.GetNestedResult(uint16(i))
		if coinToTransferPtr == nil || *coinToTransferPtr == nil {
			return nil, fmt.Errorf("failed to get split coin result %d", i)
		}
		coinToTransfer := *coinToTransferPtr
		builder.TransferObjects([]*iota_sdk_ffi.Argument{coinToTransfer}, recipientArg)
	}

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) PayAllIota(ctx context.Context, req iotaclient.PayAllIotaRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if req.Recipient == nil {
		return nil, fmt.Errorf("recipient is required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// For PayAllIota, we use the input coins directly and transfer them all
	if len(req.InputCoins) == 0 {
		return nil, fmt.Errorf("no input coins provided")
	}

	// Get all input coin objects and convert to arguments
	var coinArgs []*iota_sdk_ffi.Argument
	for _, coinID := range req.InputCoins {
		coinObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: coinID})
		if err != nil {
			return nil, fmt.Errorf("failed to get coin object: %w", err)
		}
		if coinObj.Data == nil {
			return nil, fmt.Errorf("coin object not found")
		}

		coinRef := &iotago.ObjectRef{
			ObjectID: coinObj.Data.ObjectID,
			Version:  coinObj.Data.Version.Uint64(),
			Digest:   coinObj.Data.Digest,
		}
		coinInput, err := convertObjectRefToUnresolvedInput(coinRef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert coin: %w", err)
		}
		coinArgs = append(coinArgs, builder.Input(coinInput))
	}

	// Create recipient argument
	recipientInput := iota_sdk_ffi.UnresolvedInputNewPure(req.Recipient.Bytes())
	recipientArg := builder.Input(recipientInput)

	// Transfer all coins to recipient
	builder.TransferObjects(coinArgs, recipientArg)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) PayIota(ctx context.Context, req iotaclient.PayIotaRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if len(req.Recipients) != len(req.Amount) {
		return nil, fmt.Errorf("recipients and amounts must have same length")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins (auto-selected since PayIota doesn't specify gas coins)
	gasBudget := uint64(1000000) // default gas budget
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}
	err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to add gas coins: %w", err)
	}

	// Use first input coin as primary for splitting
	if len(req.InputCoins) == 0 {
		return nil, fmt.Errorf("no input coins provided")
	}

	// Get the first input coin to use as primary
	primaryCoinObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.InputCoins[0]})
	if err != nil {
		return nil, fmt.Errorf("failed to get primary coin object: %w", err)
	}
	if primaryCoinObj.Data == nil {
		return nil, fmt.Errorf("primary coin object not found")
	}

	primaryCoinRef := &iotago.ObjectRef{
		ObjectID: primaryCoinObj.Data.ObjectID,
		Version:  primaryCoinObj.Data.Version.Uint64(),
		Digest:   primaryCoinObj.Data.Digest,
	}
	primaryCoinInput, err := convertObjectRefToUnresolvedInput(primaryCoinRef)
	if err != nil {
		return nil, fmt.Errorf("failed to convert primary coin: %w", err)
	}
	primaryCoinArg := builder.Input(primaryCoinInput)

	// Create amount arguments for splitting
	var amountArgs []*iota_sdk_ffi.Argument
	for _, amount := range req.Amount {
		// Encode amount as BCS bytes (little-endian u64)
		amountVal := amount.Uint64()
		bytes := make([]byte, 8)
		for i := 0; i < 8; i++ {
			bytes[i] = byte(amountVal)
			amountVal >>= 8
		}
		amountInput := iota_sdk_ffi.UnresolvedInputNewPure(bytes)
		amountArgs = append(amountArgs, builder.Input(amountInput))
	}

	// Split the primary coin
	splitResult := builder.SplitCoins(primaryCoinArg, amountArgs)

	// Transfer each split result to corresponding recipient
	for i, recipient := range req.Recipients {
		// Create recipient argument
		recipientInput := iota_sdk_ffi.UnresolvedInputNewPure(recipient.Bytes())
		recipientArg := builder.Input(recipientInput)

		// Get the i-th split coin using nested access
		coinToTransferPtr := splitResult.GetNestedResult(uint16(i))
		if coinToTransferPtr == nil || *coinToTransferPtr == nil {
			return nil, fmt.Errorf("failed to get split coin result %d", i)
		}
		coinToTransfer := *coinToTransferPtr
		builder.TransferObjects([]*iota_sdk_ffi.Argument{coinToTransfer}, recipientArg)
	}

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) Publish(ctx context.Context, req iotaclient.PublishRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Sender == nil {
		return nil, fmt.Errorf("sender is required")
	}
	if len(req.CompiledModules) == 0 {
		return nil, fmt.Errorf("compiled modules are required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins
	if req.Gas != nil {
		// Get the gas coin object
		gasObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.Gas})
		if err != nil {
			return nil, fmt.Errorf("failed to get gas object: %w", err)
		}
		if gasObj.Data == nil {
			return nil, fmt.Errorf("gas object not found")
		}

		gasRef := &iotago.ObjectRef{
			ObjectID: gasObj.Data.ObjectID,
			Version:  gasObj.Data.Version.Uint64(),
			Digest:   gasObj.Data.Digest,
		}
		gasInput, err := convertObjectRefToUnresolvedInput(gasRef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder.AddGasObjects([]*iota_sdk_ffi.UnresolvedInput{gasInput})
	} else {
		// Find suitable gas coins
		gasBudget := uint64(10000000) // higher default gas budget for publish
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		err := c.addGasCoinsToBuilder(ctx, builder, req.Sender, gasBudget, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to add gas coins: %w", err)
		}
	}

	// Convert compiled modules to byte slices
	var modules [][]byte
	for _, module := range req.CompiledModules {
		moduleBytes := module.Data()
		modules = append(modules, moduleBytes)
	}

	// Convert dependencies to FFI ObjectIds
	var dependencies []*iota_sdk_ffi.ObjectId
	for _, dep := range req.Dependencies {
		ffiDepId, err := toFfiObjectID(dep)
		if err != nil {
			return nil, fmt.Errorf("failed to convert dependency: %w", err)
		}
		dependencies = append(dependencies, ffiDepId)
	}

	// Add publish command
	builder.Publish(modules, dependencies)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) RequestAddStake(ctx context.Context, req iotaclient.RequestAddStakeRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if req.Validator == nil {
		return nil, fmt.Errorf("validator address is required")
	}
	if req.Amount == nil {
		return nil, fmt.Errorf("stake amount is required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins
	gasBudget := uint64(10000000) // higher gas budget for staking
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}
	err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to add gas coins: %w", err)
	}

	// Split gas coin to get stake amount
	gasArg := builder.Gas()
	stakeAmountVal := req.Amount.Uint64()
	stakeAmountBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		stakeAmountBytes[i] = byte(stakeAmountVal)
		stakeAmountVal >>= 8
	}
	stakeAmountInput := iota_sdk_ffi.UnresolvedInputNewPure(stakeAmountBytes)
	stakeAmountArg := builder.Input(stakeAmountInput)

	// Split the gas coin to get the stake amount
	splitResult := builder.SplitCoins(gasArg, []*iota_sdk_ffi.Argument{stakeAmountArg})
	stakeTokenPtr := splitResult.GetNestedResult(0)
	if stakeTokenPtr == nil || *stakeTokenPtr == nil {
		return nil, fmt.Errorf("failed to split coin for stake amount")
	}
	stakeTokenArg := *stakeTokenPtr

	// Build the add stake function call to 0x2::iota_system::request_add_stake
	systemPackageAddr, err := iota_sdk_ffi.AddressFromHex("0x0000000000000000000000000000000000000000000000000000000000000002")
	if err != nil {
		return nil, fmt.Errorf("failed to create system package address: %w", err)
	}

	moduleId, err := iota_sdk_ffi.NewIdentifier("iota_system")
	if err != nil {
		return nil, fmt.Errorf("failed to create module identifier: %w", err)
	}

	functionId, err := iota_sdk_ffi.NewIdentifier("request_add_stake")
	if err != nil {
		return nil, fmt.Errorf("failed to create function identifier: %w", err)
	}

	addStakeFunction := iota_sdk_ffi.Function{
		Package:  systemPackageAddr,
		Module:   moduleId,
		Function: functionId,
		TypeArgs: []*iota_sdk_ffi.TypeTag{},
	}

	// Add validator address as argument
	validatorInput := iota_sdk_ffi.UnresolvedInputNewPure(req.Validator.Bytes())
	validatorArg := builder.Input(validatorInput)

	// Call the add stake function
	args := []*iota_sdk_ffi.Argument{stakeTokenArg, validatorArg}
	builder.MoveCall(addStakeFunction, args)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) RequestWithdrawStake(ctx context.Context, req iotaclient.RequestWithdrawStakeRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if req.StakedIotaID == nil {
		return nil, fmt.Errorf("staked IOTA ID is required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins
	gasBudget := uint64(5000000) // gas budget for withdraw stake
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}
	err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to add gas coins: %w", err)
	}

	// Get the staked IOTA object
	stakedObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.StakedIotaID})
	if err != nil {
		return nil, fmt.Errorf("failed to get staked IOTA object: %w", err)
	}
	if stakedObj.Data == nil {
		return nil, fmt.Errorf("staked IOTA object not found")
	}

	stakedRef := &iotago.ObjectRef{
		ObjectID: stakedObj.Data.ObjectID,
		Version:  stakedObj.Data.Version.Uint64(),
		Digest:   stakedObj.Data.Digest,
	}
	stakedInput, err := convertObjectRefToUnresolvedInput(stakedRef)
	if err != nil {
		return nil, fmt.Errorf("failed to convert staked object: %w", err)
	}
	stakedArg := builder.Input(stakedInput)

	// Build the withdraw stake function call to 0x2::iota_system::request_withdraw_stake
	systemPackageAddr, err := iota_sdk_ffi.AddressFromHex("0x0000000000000000000000000000000000000000000000000000000000000002")
	if err != nil {
		return nil, fmt.Errorf("failed to create system package address: %w", err)
	}

	moduleId, err := iota_sdk_ffi.NewIdentifier("iota_system")
	if err != nil {
		return nil, fmt.Errorf("failed to create module identifier: %w", err)
	}

	functionId, err := iota_sdk_ffi.NewIdentifier("request_withdraw_stake")
	if err != nil {
		return nil, fmt.Errorf("failed to create function identifier: %w", err)
	}

	withdrawStakeFunction := iota_sdk_ffi.Function{
		Package:  systemPackageAddr,
		Module:   moduleId,
		Function: functionId,
		TypeArgs: []*iota_sdk_ffi.TypeTag{},
	}

	// Call the withdraw stake function with staked object
	args := []*iota_sdk_ffi.Argument{stakedArg}
	builder.MoveCall(withdrawStakeFunction, args)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) SplitCoin(ctx context.Context, req iotaclient.SplitCoinRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if req.Coin == nil {
		return nil, fmt.Errorf("coin is required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins
	if req.Gas != nil {
		// Get the gas coin object
		gasObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.Gas})
		if err != nil {
			return nil, fmt.Errorf("failed to get gas object: %w", err)
		}
		if gasObj.Data == nil {
			return nil, fmt.Errorf("gas object not found")
		}

		gasRef := &iotago.ObjectRef{
			ObjectID: gasObj.Data.ObjectID,
			Version:  gasObj.Data.Version.Uint64(),
			Digest:   gasObj.Data.Digest,
		}
		gasInput, err := convertObjectRefToUnresolvedInput(gasRef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder.AddGasObjects([]*iota_sdk_ffi.UnresolvedInput{gasInput})
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		err := c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to add gas coins: %w", err)
		}
	}

	// Get the coin object to split
	coinObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.Coin})
	if err != nil {
		return nil, fmt.Errorf("failed to get coin object: %w", err)
	}
	if coinObj.Data == nil {
		return nil, fmt.Errorf("coin object not found")
	}

	coinRef := &iotago.ObjectRef{
		ObjectID: coinObj.Data.ObjectID,
		Version:  coinObj.Data.Version.Uint64(),
		Digest:   coinObj.Data.Digest,
	}
	coinInput, err := convertObjectRefToUnresolvedInput(coinRef)
	if err != nil {
		return nil, fmt.Errorf("failed to convert coin: %w", err)
	}
	coinArg := builder.Input(coinInput)

	// Create amount arguments for splitting
	var amountArgs []*iota_sdk_ffi.Argument
	for _, amount := range req.SplitAmounts {
		// Encode amount as BCS bytes (little-endian u64)
		amountVal := amount.Uint64()
		bytes := make([]byte, 8)
		for i := 0; i < 8; i++ {
			bytes[i] = byte(amountVal)
			amountVal >>= 8
		}
		amountInput := iota_sdk_ffi.UnresolvedInputNewPure(bytes)
		amountArgs = append(amountArgs, builder.Input(amountInput))
	}

	// Split the coin
	builder.SplitCoins(coinArg, amountArgs)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) SplitCoinEqual(ctx context.Context, req iotaclient.SplitCoinEqualRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if req.Coin == nil {
		return nil, fmt.Errorf("coin is required")
	}
	if req.SplitCount == nil {
		return nil, fmt.Errorf("split count is required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins
	if req.Gas != nil {
		// Get the gas coin object
		gasObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.Gas})
		if err != nil {
			return nil, fmt.Errorf("failed to get gas object: %w", err)
		}
		if gasObj.Data == nil {
			return nil, fmt.Errorf("gas object not found")
		}

		gasRef := &iotago.ObjectRef{
			ObjectID: gasObj.Data.ObjectID,
			Version:  gasObj.Data.Version.Uint64(),
			Digest:   gasObj.Data.Digest,
		}
		gasInput, err := convertObjectRefToUnresolvedInput(gasRef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder.AddGasObjects([]*iota_sdk_ffi.UnresolvedInput{gasInput})
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		err := c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to add gas coins: %w", err)
		}
	}

	// Get the coin object to split
	coinObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.Coin})
	if err != nil {
		return nil, fmt.Errorf("failed to get coin object: %w", err)
	}
	if coinObj.Data == nil {
		return nil, fmt.Errorf("coin object not found")
	}

	coinRef := &iotago.ObjectRef{
		ObjectID: coinObj.Data.ObjectID,
		Version:  coinObj.Data.Version.Uint64(),
		Digest:   coinObj.Data.Digest,
	}
	coinInput, err := convertObjectRefToUnresolvedInput(coinRef)
	if err != nil {
		return nil, fmt.Errorf("failed to convert coin: %w", err)
	}
	coinArg := builder.Input(coinInput)

	// For SplitCoinEqual, we need to get the coin balance and divide by split count
	// This is a simplification - in a real implementation, you'd want to call a Move function
	// that handles equal splitting properly
	splitCount := req.SplitCount.Uint64()
	if splitCount == 0 {
		return nil, fmt.Errorf("split count must be greater than 0")
	}

	// Create count-1 amount arguments (the last piece stays with the original coin)
	var amountArgs []*iota_sdk_ffi.Argument
	for i := uint64(0); i < splitCount-1; i++ {
		// For now, use a default equal amount (this should be calculated from balance/count)
		amountVal := uint64(1000000) // 1 IOTA per split
		bytes := make([]byte, 8)
		for j := 0; j < 8; j++ {
			bytes[j] = byte(amountVal)
			amountVal >>= 8
		}
		amountInput := iota_sdk_ffi.UnresolvedInputNewPure(bytes)
		amountArgs = append(amountArgs, builder.Input(amountInput))
	}

	// Split the coin
	builder.SplitCoins(coinArg, amountArgs)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) TransferObject(ctx context.Context, req iotaclient.TransferObjectRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if req.ObjectID == nil {
		return nil, fmt.Errorf("object ID is required")
	}
	if req.Recipient == nil {
		return nil, fmt.Errorf("recipient is required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins
	if req.Gas != nil {
		// Get the gas coin object
		gasObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.Gas})
		if err != nil {
			return nil, fmt.Errorf("failed to get gas object: %w", err)
		}
		if gasObj.Data == nil {
			return nil, fmt.Errorf("gas object not found")
		}

		gasRef := &iotago.ObjectRef{
			ObjectID: gasObj.Data.ObjectID,
			Version:  gasObj.Data.Version.Uint64(),
			Digest:   gasObj.Data.Digest,
		}
		gasInput, err := convertObjectRefToUnresolvedInput(gasRef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder.AddGasObjects([]*iota_sdk_ffi.UnresolvedInput{gasInput})
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		err := c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to add gas coins: %w", err)
		}
	}

	// Get the object to transfer
	objToTransfer, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.ObjectID})
	if err != nil {
		return nil, fmt.Errorf("failed to get object to transfer: %w", err)
	}
	if objToTransfer.Data == nil {
		return nil, fmt.Errorf("object to transfer not found")
	}

	objRef := &iotago.ObjectRef{
		ObjectID: objToTransfer.Data.ObjectID,
		Version:  objToTransfer.Data.Version.Uint64(),
		Digest:   objToTransfer.Data.Digest,
	}
	objInput, err := convertObjectRefToUnresolvedInput(objRef)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object: %w", err)
	}
	objArg := builder.Input(objInput)

	// Create recipient argument
	recipientInput := iota_sdk_ffi.UnresolvedInputNewPure(req.Recipient.Bytes())
	recipientArg := builder.Input(recipientInput)

	// Transfer the object
	builder.TransferObjects([]*iota_sdk_ffi.Argument{objArg}, recipientArg)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) TransferIota(ctx context.Context, req iotaclient.TransferIotaRequest) (*iotajsonrpc.TransactionBytes, error) {
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if req.ObjectID == nil {
		return nil, fmt.Errorf("object ID is required")
	}
	if req.Recipient == nil {
		return nil, fmt.Errorf("recipient is required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder.SetGasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins (auto-selected since TransferIota doesn't specify gas coins)
	gasBudget := uint64(1000000) // default gas budget
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}
	err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to add gas coins: %w", err)
	}

	// Get the coin object to transfer
	coinObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.ObjectID})
	if err != nil {
		return nil, fmt.Errorf("failed to get coin object: %w", err)
	}
	if coinObj.Data == nil {
		return nil, fmt.Errorf("coin object not found")
	}

	coinRef := &iotago.ObjectRef{
		ObjectID: coinObj.Data.ObjectID,
		Version:  coinObj.Data.Version.Uint64(),
		Digest:   coinObj.Data.Digest,
	}
	coinInput, err := convertObjectRefToUnresolvedInput(coinRef)
	if err != nil {
		return nil, fmt.Errorf("failed to convert coin: %w", err)
	}
	coinArg := builder.Input(coinInput)

	if req.Amount != nil {
		// If amount is specified, split the coin first
		amountVal := req.Amount.Uint64()
		bytes := make([]byte, 8)
		for i := 0; i < 8; i++ {
			bytes[i] = byte(amountVal)
			amountVal >>= 8
		}
		amountInput := iota_sdk_ffi.UnresolvedInputNewPure(bytes)
		amountArg := builder.Input(amountInput)

		// Split the coin
		splitResult := builder.SplitCoins(coinArg, []*iota_sdk_ffi.Argument{amountArg})

		// Get the first split result
		coinToTransferPtr := splitResult.GetNestedResult(0)
		if coinToTransferPtr == nil || *coinToTransferPtr == nil {
			return nil, fmt.Errorf("failed to get split coin result")
		}
		coinArg = *coinToTransferPtr
	}

	// Create recipient argument
	recipientInput := iota_sdk_ffi.UnresolvedInputNewPure(req.Recipient.Bytes())
	recipientArg := builder.Input(recipientInput)

	// Transfer the coin
	builder.TransferObjects([]*iota_sdk_ffi.Argument{coinArg}, recipientArg)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          []iotago.ObjectRef{},
		InputObjects: []iotajsonrpc.InputObjectKind{},
		TxBytes:      iotago.Base64Data(txBytes),
	}, nil
}

func (c *BindingClient) GetCoinObjsForTargetAmount(ctx context.Context, address *iotago.Address, targetAmount uint64, gasAmount uint64) (iotajsonrpc.Coins, error) {
	if address == nil {
		return nil, nil
	}

	// Get all coins for the address
	coinPage, err := c.GetCoins(ctx, iotaclient.GetCoinsRequest{
		Owner: address,
		Limit: 200,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get coins: %w", err)
	}

	// Use iotajsonrpc.PickupCoins to select minimal set
	pickedCoins, err := iotajsonrpc.PickupCoins(coinPage, new(big.Int).SetUint64(targetAmount), gasAmount, 0, 25)
	if err != nil {
		return nil, fmt.Errorf("failed to pickup coins: %w", err)
	}

	return pickedCoins.Coins, nil
}

func (c *BindingClient) SignAndExecuteTransaction(ctx context.Context, req *iotaclient.SignAndExecuteTransactionRequest) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if req.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if len(req.TxDataBytes) == 0 {
		return nil, fmt.Errorf("transaction data bytes are required")
	}

	// The challenge here is that FFI doesn't have a "TransactionFromBytes" method
	// We need to reconstruct the transaction from bytes, but this is complex without
	// a deserialization method in the FFI. For now, return an error indicating
	// this limitation.

	// In a real implementation, you would need either:
	// 1. An FFI method to deserialize transaction bytes back to Transaction
	// 2. Or use the buildTransactionAndExecute helper with a signer directly

	return nil, fmt.Errorf("SignAndExecuteTransaction not supported by FFI bindings - no transaction deserialization method available. Use individual transaction builder methods instead")
}

func (c *BindingClient) PublishContract(ctx context.Context, signer iotasigner.Signer, modules []*iotago.Base64Data, dependencies []*iotago.Address, gasBudget uint64, options *iotajsonrpc.IotaTransactionBlockResponseOptions) (*iotajsonrpc.IotaTransactionBlockResponse, *iotago.PackageID, error) {
	if signer == nil {
		return nil, nil, fmt.Errorf("signer is required")
	}
	if len(modules) == 0 {
		return nil, nil, fmt.Errorf("modules are required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Add gas coins
	err := c.addGasCoinsToBuilder(ctx, builder, signer.Address(), gasBudget, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add gas coins: %w", err)
	}

	// Convert modules to byte slices
	var moduleBytes [][]byte
	for _, module := range modules {
		moduleBytes = append(moduleBytes, module.Data())
	}

	// Convert dependencies to FFI ObjectIds
	var ffiDeps []*iota_sdk_ffi.ObjectId
	for _, dep := range dependencies {
		depPackageID := &iotago.PackageID{}
		copy(depPackageID[:], dep.Bytes())
		ffiDepId, err := toFfiObjectID((*iotago.ObjectID)(depPackageID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert dependency: %w", err)
		}
		ffiDeps = append(ffiDeps, ffiDepId)
	}

	// Add publish command
	publishResult := builder.Publish(moduleBytes, ffiDeps)

	// Transfer upgrade capability to sender (as per requirements)
	senderInput := iota_sdk_ffi.UnresolvedInputNewPure(signer.Address().Bytes())
	senderArg := builder.Input(senderInput)
	builder.TransferObjects([]*iota_sdk_ffi.Argument{publishResult}, senderArg)

	// Build, sign, and execute
	response, err := c.buildTransactionAndExecute(ctx, signer, builder, gasBudget, 0, options)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Extract package ID from transaction effects
	// This would require parsing the transaction effects to find the published package
	packageID := &iotago.PackageID{}

	return response, packageID, nil
}

func (c *BindingClient) UpdateObjectRef(ctx context.Context, ref *iotago.ObjectRef) (*iotago.ObjectRef, error) {
	if ref == nil || ref.ObjectID == nil {
		return nil, nil
	}

	ffiObjectID, err := toFfiObjectID(ref.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ObjectID: %w", err)
	}

	obj, err := c.qclient.Object(ffiObjectID, nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("Object query failed: %w", err)
	}

	if obj == nil || *obj == nil {
		return nil, fmt.Errorf("object not found")
	}

	o := *obj
	digest, err := fromFfiDigest(o.Digest())
	if err != nil {
		return nil, fmt.Errorf("failed to convert digest: %w", err)
	}

	return &iotago.ObjectRef{
		ObjectID: ref.ObjectID,
		Version:  o.Version(),
		Digest:   digest,
	}, nil
}

func (c *BindingClient) MintToken(ctx context.Context, signer iotasigner.Signer, packageID *iotago.PackageID, tokenName string, treasuryCap *iotago.ObjectRef, mintAmount uint64, options *iotajsonrpc.IotaTransactionBlockResponseOptions) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if packageID == nil {
		return nil, fmt.Errorf("package ID is required")
	}
	if treasuryCap == nil {
		return nil, fmt.Errorf("treasury cap is required")
	}

	builder := iota_sdk_ffi.NewTransactionBuilder()

	// Add gas coins
	gasBudget := uint64(5000000) // higher gas budget for mint
	err := c.addGasCoinsToBuilder(ctx, builder, signer.Address(), gasBudget, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to add gas coins: %w", err)
	}

	// Build the mint function call
	packageAddr, err := iota_sdk_ffi.AddressFromHex((*iotago.Address)(packageID).String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert package address: %w", err)
	}

	moduleId, err := iota_sdk_ffi.NewIdentifier(tokenName)
	if err != nil {
		return nil, fmt.Errorf("failed to create module identifier: %w", err)
	}

	mintFuncId, err := iota_sdk_ffi.NewIdentifier("mint")
	if err != nil {
		return nil, fmt.Errorf("failed to create mint function identifier: %w", err)
	}

	mintFunction := iota_sdk_ffi.Function{
		Package:  packageAddr,
		Module:   moduleId,
		Function: mintFuncId,
		TypeArgs: []*iota_sdk_ffi.TypeTag{}, // TODO: Add proper type args if needed
	}

	// Add treasury cap as argument
	treasuryCapInput, err := convertObjectRefToUnresolvedInput(treasuryCap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert treasury cap: %w", err)
	}
	treasuryCapArg := builder.Input(treasuryCapInput)

	// Add mint amount as argument
	amountBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		amountBytes[i] = byte(mintAmount)
		mintAmount >>= 8
	}
	amountInput := iota_sdk_ffi.UnresolvedInputNewPure(amountBytes)
	amountArg := builder.Input(amountInput)

	// Add recipient (sender) as argument
	recipientInput := iota_sdk_ffi.UnresolvedInputNewPure(signer.Address().Bytes())
	recipientArg := builder.Input(recipientInput)

	// Call the mint function
	args := []*iota_sdk_ffi.Argument{treasuryCapArg, amountArg, recipientArg}
	builder.MoveCall(mintFunction, args)

	// Build, sign, and execute
	response, err := c.buildTransactionAndExecute(ctx, signer, builder, gasBudget, 0, options)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *BindingClient) GetIotaCoinsOwnedByAddress(ctx context.Context, address *iotago.Address) (iotajsonrpc.Coins, error) {
	if address == nil {
		return nil, nil
	}

	coinType := iotajsonrpc.IotaCoinType.String()
	coinPage, err := c.GetCoins(ctx, iotaclient.GetCoinsRequest{
		Owner:    address,
		CoinType: &coinType,
	})
	if err != nil {
		return nil, err
	}

	return coinPage.Data, nil
}

func (c *BindingClient) BatchGetObjectsOwnedByAddress(ctx context.Context, address *iotago.Address, options *iotajsonrpc.IotaObjectDataOptions, filterType string) ([]iotajsonrpc.IotaObjectResponse, error) {
	if address == nil {
		return nil, nil
	}

	ffiAddr, err := toFfiAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to convert address: %w", err)
	}

	// Build ObjectFilter with Owner set
	filter := &iota_sdk_ffi.ObjectFilter{
		Owner: &ffiAddr,
	}

	objectPage, err := c.qclient.Objects(filter, nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("Objects query failed: %w", err)
	}

	var results []iotajsonrpc.IotaObjectResponse
	for _, obj := range objectPage.Data {
		resp, err := mapFfiObjectToIotaResponse(&obj)
		if err != nil {
			return nil, fmt.Errorf("failed to map object: %w", err)
		}
		results = append(results, *resp)
	}

	return results, nil
}

func (c *BindingClient) BatchGetFilteredObjectsOwnedByAddress(ctx context.Context, address *iotago.Address, options *iotajsonrpc.IotaObjectDataOptions, filter func(*iotajsonrpc.IotaObjectData) bool) ([]iotajsonrpc.IotaObjectResponse, error) {
	// First get all objects owned by the address
	allObjects, err := c.BatchGetObjectsOwnedByAddress(ctx, address, options, "")
	if err != nil {
		return nil, err
	}

	// Filter in-memory using the provided predicate
	var filtered []iotajsonrpc.IotaObjectResponse
	for _, obj := range allObjects {
		if obj.Data != nil && filter(obj.Data) {
			filtered = append(filtered, obj)
		}
	}

	return filtered, nil
}

func (c *BindingClient) GetAllBalances(ctx context.Context, owner *iotago.Address) ([]*iotajsonrpc.Balance, error) {
	if owner == nil {
		return nil, nil
	}
	addr, err := toFfiAddress(owner)
	if err != nil {
		return nil, err
	}
	// Currently support IOTA coin only
	coinType := "0x2::iota::IOTA"
	balance, err := c.qclient.Balance(addr, nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("Failed to get balance: %v", err)
	}

	// Convert to response format
	var result []*iotajsonrpc.Balance
	if balance != nil {
		balanceResp := &iotajsonrpc.Balance{
			CoinType:        iotajsonrpc.CoinType(coinType),
			TotalBalance:    iotajsonrpc.NewBigInt(*balance),
			CoinObjectCount: iotajsonrpc.NewBigInt(1), // TODO: get actual count
		}
		result = append(result, balanceResp)
	}

	return result, nil
}

func (c *BindingClient) GetAllCoins(ctx context.Context, req iotaclient.GetAllCoinsRequest) (*iotajsonrpc.CoinPage, error) {
	if req.Owner == nil {
		return &iotajsonrpc.CoinPage{}, nil
	}
	owner, err := toFfiAddress(req.Owner)
	if err != nil {
		return nil, err
	}
	var paginationFilter *iota_sdk_ffi.PaginationFilter
	var coinType *string
	coins, err := c.qclient.Coins(owner, paginationFilter, coinType)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL Coins failed: %w", err)
	}

	// Convert back to iota-go response format
	response := &iotajsonrpc.CoinPage{}
	for _, ccoin := range coins.Data {
		oid, err := fromFfiObjectID(ccoin.Id())
		if err != nil {
			return nil, err
		}
		// fetch object details for version/digest
		obj, err := c.qclient.Object(ccoin.Id(), nil)
		if err.(*iota_sdk_ffi.SdkFfiError) != nil {
			return nil, fmt.Errorf("Object failed: %w", err)
		}
		var version uint64
		var digest *iotago.ObjectDigest
		if obj != nil && *obj != nil {
			o := *obj
			version = o.Version()
			if dg, err := fromFfiDigest(o.Digest()); err == nil {
				digest = dg
			}
		}
		coinTypeStr := ccoin.CoinType().String()
		response.Data = append(response.Data, &iotajsonrpc.Coin{
			CoinType:            iotajsonrpc.CoinType(coinTypeStr),
			CoinObjectID:        oid,
			Version:             iotajsonrpc.NewBigInt(version),
			Digest:              digest,
			Balance:             iotajsonrpc.NewBigInt(ccoin.Balance()),
			PreviousTransaction: iotago.TransactionDigest{},
		})
	}
	response.HasNextPage = coins.PageInfo.HasNextPage

	return response, nil
}

func (c *BindingClient) GetBalance(ctx context.Context, req iotaclient.GetBalanceRequest) (*iotajsonrpc.Balance, error) {
	if req.Owner == nil {
		return &iotajsonrpc.Balance{CoinType: iotajsonrpc.CoinType("")}, nil
	}
	addr, err := toFfiAddress(req.Owner)
	if err != nil {
		return nil, err
	}
	var ct *string
	if req.CoinType != "" {
		ct = &req.CoinType
	}
	bal, err := c.qclient.Balance(addr, ct)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL Balance failed: %w", err)
	}
	b := uint64(0)
	if bal != nil {
		b = *bal
	}
	typ := req.CoinType
	if typ == "" {
		typ = "0x2::iota::IOTA"
	}
	return &iotajsonrpc.Balance{
		CoinType:        iotajsonrpc.CoinType(typ),
		CoinObjectCount: iotajsonrpc.NewBigInt(0),
		TotalBalance:    iotajsonrpc.NewBigInt(b),
	}, nil
}

func (c *BindingClient) GetCoinMetadata(ctx context.Context, coinType string) (*iotajsonrpc.IotaCoinMetadata, error) {
	// Direct mapping to GraphQL client's CoinMetadata method
	metadata, err := c.qclient.CoinMetadata(coinType)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL CoinMetadata failed: %w", err)
	}

	// Convert back to iota-go response format
	response := &iotajsonrpc.IotaCoinMetadata{}
	if metadata != nil {
		// Map the fields from iota_sdk_ffi.CoinMetadata to iotajsonrpc.IotaCoinMetadata
		if metadata.Decimals != nil {
			response.Decimals = uint8(*metadata.Decimals)
		}
		if metadata.Name != nil {
			response.Name = *metadata.Name
		}
		if metadata.Symbol != nil {
			response.Symbol = *metadata.Symbol
		}
		if metadata.Description != nil {
			response.Description = *metadata.Description
		}
		if metadata.IconUrl != nil {
			response.IconUrl = *metadata.IconUrl
		}
	}

	return response, nil
}

func (c *BindingClient) GetCoins(ctx context.Context, req iotaclient.GetCoinsRequest) (*iotajsonrpc.CoinPage, error) {
	if req.Owner == nil {
		return &iotajsonrpc.CoinPage{}, nil
	}
	owner, err := toFfiAddress(req.Owner)
	if err != nil {
		return nil, err
	}
	var coinType *string = req.CoinType
	cps, err := c.qclient.Coins(owner, nil, coinType)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL Coins failed: %w", err)
	}
	out := &iotajsonrpc.CoinPage{}
	for _, ccoin := range cps.Data {
		oid, err := fromFfiObjectID(ccoin.Id())
		if err != nil {
			return nil, err
		}
		obj, err := c.qclient.Object(ccoin.Id(), nil)
		if err.(*iota_sdk_ffi.SdkFfiError) != nil {
			return nil, fmt.Errorf("Object failed: %w", err)
		}
		var version uint64
		var digest *iotago.ObjectDigest
		if obj != nil && *obj != nil {
			o := *obj
			version = o.Version()
			if dg, err := fromFfiDigest(o.Digest()); err == nil {
				digest = dg
			}
		}
		coinTypeStr := ccoin.CoinType().String()
		out.Data = append(out.Data, &iotajsonrpc.Coin{
			CoinType:     iotajsonrpc.CoinType(coinTypeStr),
			CoinObjectID: oid,
			Version:      iotajsonrpc.NewBigInt(version),
			Digest:       digest,
			Balance:      iotajsonrpc.NewBigInt(ccoin.Balance()),
		})
	}
	out.HasNextPage = cps.PageInfo.HasNextPage
	return out, nil
}

func (c *BindingClient) GetTotalSupply(ctx context.Context, coinType string) (*iotajsonrpc.Supply, error) {
	// Direct mapping to GraphQL client's TotalSupply method
	supply, err := c.qclient.TotalSupply(coinType)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL TotalSupply failed: %w", err)
	}

	// Convert back to iota-go response format
	response := &iotajsonrpc.Supply{}
	if supply != nil {
		response.Value = iotajsonrpc.NewBigInt(*supply)
	} else {
		response.Value = iotajsonrpc.NewBigInt(0)
	}

	return response, nil
}

func (c *BindingClient) GetChainIdentifier(ctx context.Context) (string, error) {
	// Direct mapping to GraphQL client's ChainId method
	chainId, err := c.qclient.ChainId()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return "", fmt.Errorf("GraphQL ChainId failed: %w", err)
	}
	return chainId, nil
}

func (c *BindingClient) GetCheckpoint(ctx context.Context, checkpointID *iotajsonrpc.BigInt) (*iotajsonrpc.Checkpoint, error) {
	// Convert checkpointID parameter
	var seqNum *uint64
	if checkpointID != nil {
		v := checkpointID.Uint64()
		seqNum = &v
	}

	// Call GraphQL client's Checkpoint method
	checkpoint, err := c.qclient.Checkpoint(nil, seqNum) // nil digest, use seqNum
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL Checkpoint failed: %w", err)
	}

	response := &iotajsonrpc.Checkpoint{}
	if checkpoint != nil && *checkpoint != nil {
		cs := *checkpoint
		response.Epoch = iotajsonrpc.NewBigInt(cs.Epoch())
		response.SequenceNumber = iotajsonrpc.NewBigInt(cs.SequenceNumber())
	}
	return response, nil
}

func (c *BindingClient) GetCheckpoints(ctx context.Context, req iotaclient.GetCheckpointsRequest) (*iotajsonrpc.CheckpointPage, error) {
	var paginationFilter *iota_sdk_ffi.PaginationFilter
	cps, err := c.qclient.Checkpoints(paginationFilter)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL Checkpoints failed: %w", err)
	}
	out := &iotajsonrpc.CheckpointPage{}
	for _, cs := range cps.Data {
		cp := &iotajsonrpc.Checkpoint{
			Epoch:          iotajsonrpc.NewBigInt(cs.Epoch()),
			SequenceNumber: iotajsonrpc.NewBigInt(cs.SequenceNumber()),
		}
		out.Data = append(out.Data, cp)
	}
	out.HasNextPage = cps.PageInfo.HasNextPage
	return out, nil
}

func (c *BindingClient) GetEvents(ctx context.Context, digest *iotago.TransactionDigest) ([]*iotajsonrpc.IotaEvent, error) {
	// Minimal placeholder: FFI mapping of events structure is non-trivial.
	_, err := c.qclient.Events(nil, nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL Events failed: %w", err)
	}
	return []*iotajsonrpc.IotaEvent{}, nil
}

func (c *BindingClient) GetLatestCheckpointSequenceNumber(ctx context.Context) (string, error) {
	// Direct mapping to GraphQL client's LatestCheckpointSequenceNumber method
	seqNum, err := c.qclient.LatestCheckpointSequenceNumber()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return "", fmt.Errorf("GraphQL LatestCheckpointSequenceNumber failed: %w", err)
	}
	if seqNum == nil {
		return "0", nil
	}
	return fmt.Sprintf("%d", *seqNum), nil
}

func (c *BindingClient) GetObject(ctx context.Context, req iotaclient.GetObjectRequest) (*iotajsonrpc.IotaObjectResponse, error) {
	if req.ObjectID == nil {
		return &iotajsonrpc.IotaObjectResponse{}, nil
	}
	oid, err := toFfiObjectID(req.ObjectID)
	if err != nil {
		return nil, err
	}
	obj, err := c.qclient.Object(oid, nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL Object failed: %w", err)
	}

	return mapFfiObjectToIotaResponse(obj)
}

func (c *BindingClient) GetProtocolConfig(ctx context.Context, version *iotajsonrpc.BigInt) (*iotajsonrpc.ProtocolConfig, error) {
	// Convert version parameter
	var versionUint64 *uint64
	if version != nil {
		v := version.Uint64()
		versionUint64 = &v
	}

	// Call GraphQL client's ProtocolConfig method
	config, err := c.qclient.ProtocolConfig(versionUint64)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL ProtocolConfig failed: %w", err)
	}

	// Convert back to iota-go response format
	response := &iotajsonrpc.ProtocolConfig{}
	if config != nil {
		// TODO: Map config fields properly
		// This would require detailed field mapping between ProtocolConfigs types
	}

	return response, nil
}

func (c *BindingClient) GetTotalTransactionBlocks(ctx context.Context) (string, error) {
	// Direct mapping to GraphQL client's TotalTransactionBlocks method
	total, err := c.qclient.TotalTransactionBlocks()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return "", fmt.Errorf("GraphQL TotalTransactionBlocks failed: %w", err)
	}
	if total == nil {
		return "0", nil
	}
	return fmt.Sprintf("%d", *total), nil
}

func (c *BindingClient) GetTransactionBlock(ctx context.Context, req iotaclient.GetTransactionBlockRequest) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	// Convert digest parameter
	digest := &iota_sdk_ffi.Digest{}
	// TODO: Convert req.Digest to iota_sdk_ffi.Digest

	// Call GraphQL client's Transaction method
	tx, err := c.qclient.Transaction(digest)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL Transaction failed: %w", err)
	}

	// Convert back to iota-go response format
	response := &iotajsonrpc.IotaTransactionBlockResponse{}
	if tx != nil {
		// TODO: Map transaction fields from iota_sdk_ffi.SignedTransaction to iotajsonrpc.IotaTransactionBlockResponse
	}

	return response, nil
}

func (c *BindingClient) MultiGetObjects(ctx context.Context, req iotaclient.MultiGetObjectsRequest) ([]iotajsonrpc.IotaObjectResponse, error) {
	if len(req.ObjectIDs) == 0 {
		return nil, nil
	}

	// Convert object IDs to FFI format
	var ffiObjectIds []*iota_sdk_ffi.ObjectId
	for _, objID := range req.ObjectIDs {
		ffiID, err := toFfiObjectID(objID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ObjectID: %w", err)
		}
		ffiObjectIds = append(ffiObjectIds, ffiID)
	}

	// Build ObjectFilter with ObjectIds set
	filter := &iota_sdk_ffi.ObjectFilter{
		ObjectIds: &ffiObjectIds,
	}

	objectPage, err := c.qclient.Objects(filter, nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("Objects query failed: %w", err)
	}

	var results []iotajsonrpc.IotaObjectResponse
	for _, obj := range objectPage.Data {
		resp, err := mapFfiObjectToIotaResponse(&obj)
		if err != nil {
			return nil, fmt.Errorf("failed to map object: %w", err)
		}
		results = append(results, *resp)
	}

	return results, nil
}

func (c *BindingClient) MultiGetTransactionBlocks(ctx context.Context, req iotaclient.MultiGetTransactionBlocksRequest) ([]*iotajsonrpc.IotaTransactionBlockResponse, error) {
	if len(req.Digests) == 0 {
		return nil, nil
	}

	var results []*iotajsonrpc.IotaTransactionBlockResponse
	for _, digest := range req.Digests {
		// Convert digest to FFI format
		ffiDigest, err := iota_sdk_ffi.DigestFromBase58(digest.String())
		if err != nil {
			return nil, fmt.Errorf("failed to convert digest: %w", err)
		}

		// Get transaction effects or transaction data effects based on options
		if req.Options != nil && req.Options.ShowEffects {
			effects, err := c.qclient.TransactionEffects(ffiDigest)
			if err.(*iota_sdk_ffi.SdkFfiError) != nil {
				return nil, fmt.Errorf("TransactionEffects failed: %w", err)
			}

			resp := &iotajsonrpc.IotaTransactionBlockResponse{
				Digest: *digest,
			}
			if effects != nil {
				// TODO: Map effects to IotaTransactionBlockEffects
			}
			results = append(results, resp)
		} else {
			// Minimal response with just digest
			results = append(results, &iotajsonrpc.IotaTransactionBlockResponse{
				Digest: *digest,
			})
		}
	}

	return results, nil
}

func (c *BindingClient) TryGetPastObject(ctx context.Context, req iotaclient.TryGetPastObjectRequest) (*iotajsonrpc.IotaPastObjectResponse, error) {
	return nil, errors.New("TryGetPastObject not implemented for BindingClient")
}

func (c *BindingClient) TryMultiGetPastObjects(ctx context.Context, req iotaclient.TryMultiGetPastObjectsRequest) ([]*iotajsonrpc.IotaPastObjectResponse, error) {
	return nil, errors.New("TryMultiGetPastObjects not implemented for BindingClient")
}

func (c *BindingClient) RequestFunds(ctx context.Context, address cryptolib.Address) error {
	return errors.New("RequestFunds not implemented for BindingClient")
}

func (c *BindingClient) Health(ctx context.Context) error {
	// Perform lightweight checks using ChainId and LatestCheckpointSequenceNumber
	_, err := c.qclient.ChainId()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return fmt.Errorf("health check failed on ChainId: %w", err)
	}

	_, err = c.qclient.LatestCheckpointSequenceNumber()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return fmt.Errorf("health check failed on LatestCheckpointSequenceNumber: %w", err)
	}

	return nil
}

func (c *BindingClient) L2() clients.L2Client {
	return nil
}

func (c *BindingClient) IotaClient() *iotaclient.Client {
	return nil
}

func (c *BindingClient) DeployISCContracts(ctx context.Context, signer iotasigner.Signer) (iotago.PackageID, error) {
	return iotago.PackageID{}, errors.New("DeployISCContracts not implemented for BindingClient")
}

func (c *BindingClient) FindCoinsForGasPayment(ctx context.Context, owner *iotago.Address, pt iotago.ProgrammableTransaction, gasPrice uint64, gasBudget uint64) ([]*iotago.ObjectRef, error) {
	if owner == nil {
		return nil, nil
	}

	// Get IOTA coins for the owner
	coinType := iotajsonrpc.IotaCoinType.String()
	coinPage, err := c.GetCoins(ctx, iotaclient.GetCoinsRequest{
		Owner:    owner,
		CoinType: &coinType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get coins: %w", err)
	}

	// Filter out coins that are already used as inputs in the transaction
	// TODO: implement IsInInputObjects logic similar to iota-go
	availableCoins := coinPage.Data

	// Use PickupCoinsWithFilter to select gas coins
	totalGasCost := gasPrice * gasBudget
	selectedCoins, err := iotajsonrpc.PickupCoinsWithFilter(availableCoins, totalGasCost, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to select gas coins: %w", err)
	}

	// Convert to ObjectRef slice
	var refs []*iotago.ObjectRef
	for _, coin := range selectedCoins {
		refs = append(refs, coin.Ref())
	}

	return refs, nil
}

func (c *BindingClient) MergeCoinsAndExecute(ctx context.Context, owner iotasigner.Signer, destinationCoin *iotago.ObjectRef, sourceCoins []*iotago.ObjectRef, gasBudget uint64) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	return nil, errors.New("MergeCoinsAndExecute not implemented for BindingClient")
}

func (c *BindingClient) SignAndExecuteTxWithRetry(ctx context.Context, signer iotasigner.Signer, pt iotago.ProgrammableTransaction, gasCoin *iotago.ObjectRef, gasBudget uint64, gasPrice uint64, options *iotajsonrpc.IotaTransactionBlockResponseOptions) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	return nil, errors.New("SignAndExecuteTxWithRetry not implemented for BindingClient")
}

func (c *BindingClient) WaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error) {
	return nil, errors.New("WaitForNextVersionForTesting not implemented for BindingClient")
}

// Transaction builder helper functions

// // convertObjectRefToUnresolvedInput converts an iotago.ObjectRef to FFI UnresolvedInput
func convertObjectRefToUnresolvedInput(ref *iotago.ObjectRef) (*iota_sdk_ffi.UnresolvedInput, error) {
	if ref == nil {
		return nil, fmt.Errorf("ObjectRef is nil")
	}

	ffiObjID, err := toFfiObjectID(ref.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ObjectID: %w", err)
	}

	ffiDigest, err := iota_sdk_ffi.DigestFromBase58(ref.Digest.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert digest: %w", err)
	}

	return iota_sdk_ffi.UnresolvedInputNewOwned(ffiObjID, ref.Version, ffiDigest), nil
}

// buildTransactionAndExecute builds, signs, and executes a transaction
func (c *BindingClient) buildTransactionAndExecute(ctx context.Context, signer iotasigner.Signer, builder *iota_sdk_ffi.TransactionBuilder, gasBudget uint64, gasPrice uint64, options *iotajsonrpc.IotaTransactionBlockResponseOptions) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	// Set gas budget and price
	builder.SetGasBudget(gasBudget)
	if gasPrice > 0 {
		builder.SetGasPrice(gasPrice)
	}

	// Set sender
	senderAddr, err := toFfiAddress(signer.Address())
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}
	builder.SetSender(senderAddr)

	// Finish the transaction
	tx, err := builder.Finish()
	if err != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	// Serialize and sign
	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}
	signature, err := signer.SignTransactionBlock(txBytes, iotasigner.DefaultIntent())
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Convert to FFI signature
	ffiSig, err := iota_sdk_ffi.UserSignatureFromBytes(signature.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to create FFI signature: %w", err)
	}

	// Execute transaction
	effects, err := c.qclient.ExecuteTx([]*iota_sdk_ffi.UserSignature{ffiSig}, tx)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("ExecuteTx failed: %w", err)
	}

	// Build response
	digest, _ := fromFfiDigest(tx.Digest())
	response := &iotajsonrpc.IotaTransactionBlockResponse{
		Digest: *digest,
	}

	// If options request effects, add them
	if options != nil && options.ShowEffects && effects != nil {
		// Fetch full effects
		txEffects, err := c.qclient.TransactionEffects(tx.Digest())
		if err.(*iota_sdk_ffi.SdkFfiError) != nil && txEffects != nil {
			// TODO: Map effects to IotaTransactionBlockEffects
			// This would require detailed field mapping
		}
	}

	return response, nil
}

// addGasCoinsToBuilder adds gas coins to the transaction builder
func (c *BindingClient) addGasCoinsToBuilder(ctx context.Context, builder *iota_sdk_ffi.TransactionBuilder, owner *iotago.Address, gasBudget, gasPrice uint64) error {
	// Find gas coins
	gasCoins, err := c.FindCoinsForGasPayment(ctx, owner, iotago.ProgrammableTransaction{}, gasPrice, gasBudget)
	if err != nil {
		return fmt.Errorf("failed to find gas coins: %w", err)
	}

	// Convert to UnresolvedInput
	var gasInputs []*iota_sdk_ffi.UnresolvedInput
	for _, coin := range gasCoins {
		unresolvedInput, err := convertObjectRefToUnresolvedInput(coin)
		if err != nil {
			return fmt.Errorf("failed to convert gas coin: %w", err)
		}
		gasInputs = append(gasInputs, unresolvedInput)
	}

	// Add gas objects to builder
	builder.AddGasObjects(gasInputs)
	return nil
}
