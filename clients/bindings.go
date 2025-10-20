package clients

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/clients/bindings/iota_sdk_ffi"
	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

// BindingClient wraps the Rust FFI GraphQL client for IOTA blockchain interaction.
//
// IMPORTANT: This client uses the GraphQL API and requires endpoints that support GraphQL.
// It works with:
//   - Testnet: iotaconn.TestnetEndpointURL
//   - Devnet: iotaconn.DevnetEndpointURL
//   - Custom GraphQL endpoints
//
// It does NOT work with:
//   - Local test nodes (which typically only expose JSON-RPC)
//   - For local testing, use the regular iotaclient.Client instead
type BindingClient struct {
	RpcURL  string
	qclient *iota_sdk_ffi.GraphQlClient
}

// NewBindingClient creates a new BindingClient connected to the specified GraphQL endpoint.
// For standard networks, use iotaconn.TestnetEndpointURL or iotaconn.DevnetEndpointURL.
// For custom endpoints, the URL should point to a GraphQL-enabled IOTA node.
func NewBindingClient(rpcUrl string) *BindingClient {
	var client BindingClient

	switch rpcUrl {
	case iotaconn.LocalnetEndpointURL:
		client.qclient = iota_sdk_ffi.GraphQlClientNewLocalnet()
	case iotaconn.TestnetEndpointURL:
		client.qclient = iota_sdk_ffi.GraphQlClientNewTestnet()
	case iotaconn.DevnetEndpointURL:
		client.qclient = iota_sdk_ffi.GraphQlClientNewDevnet()
	default:
		// For custom URLs, append /graphql if not already present
		graphqlUrl := rpcUrl
		if !strings.HasSuffix(graphqlUrl, "/graphql") {
			graphqlUrl = rpcUrl + "/graphql"
		}
		qclient, err := iota_sdk_ffi.NewGraphQlClient(graphqlUrl)
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

func fromFfiOwner(owner *iota_sdk_ffi.Owner) (*iotago.Owner, error) {
	if owner == nil {
		return nil, nil
	}

	result := &iotago.Owner{}
	if owner.IsAddress() {
		addr := owner.AsAddress()
		iotaAddr, err := iotago.AddressFromHex(addr.ToHex())
		if err != nil {
			return nil, err
		}
		result.AddressOwner = iotaAddr
	} else if owner.IsObject() {
		objID := owner.AsObject()
		iotaAddr, err := iotago.AddressFromHex(objID.ToHex())
		if err != nil {
			return nil, err
		}
		result.ObjectOwner = iotaAddr
	} else if owner.IsShared() {
		version := owner.AsShared()
		result.Shared = &struct {
			InitialSharedVersion iotago.SequenceNumber `json:"initial_shared_version"`
		}{
			InitialSharedVersion: iotago.SequenceNumber(version),
		}
	} else if owner.IsImmutable() {
		result.Immutable = &serialization.EmptyEnum{}
	}

	return result, nil
}

func iotagoOwnerToObjectOwner(owner *iotago.Owner) iotajsonrpc.ObjectOwner {
	if owner == nil {
		return iotajsonrpc.ObjectOwner{}
	}

	ownerInternal := &iotajsonrpc.ObjectOwnerInternal{}
	if owner.AddressOwner != nil {
		ownerInternal.AddressOwner = owner.AddressOwner
	} else if owner.ObjectOwner != nil {
		ownerInternal.ObjectOwner = owner.ObjectOwner
	} else if owner.Shared != nil {
		seq := owner.Shared.InitialSharedVersion
		ownerInternal.Shared = &struct {
			InitialSharedVersion *iotago.SequenceNumber `json:"initial_shared_version"`
		}{
			InitialSharedVersion: &seq,
		}
	}

	return iotajsonrpc.ObjectOwner{ObjectOwnerInternal: ownerInternal}
}

func toFfiTypeTag(tag *iotago.TypeTag) (*iota_sdk_ffi.TypeTag, error) {
	if tag == nil {
		return nil, nil
	}

	if tag.Bool != nil {
		return iota_sdk_ffi.TypeTagNewBool(), nil
	} else if tag.U8 != nil {
		return iota_sdk_ffi.TypeTagNewU8(), nil
	} else if tag.U16 != nil {
		return iota_sdk_ffi.TypeTagNewU16(), nil
	} else if tag.U32 != nil {
		return iota_sdk_ffi.TypeTagNewU32(), nil
	} else if tag.U64 != nil {
		return iota_sdk_ffi.TypeTagNewU64(), nil
	} else if tag.U128 != nil {
		return iota_sdk_ffi.TypeTagNewU128(), nil
	} else if tag.U256 != nil {
		return iota_sdk_ffi.TypeTagNewU256(), nil
	} else if tag.Address != nil {
		return iota_sdk_ffi.TypeTagNewAddress(), nil
	} else if tag.Signer != nil {
		return iota_sdk_ffi.TypeTagNewSigner(), nil
	} else if tag.Vector != nil {
		innerTag, err := toFfiTypeTag(tag.Vector)
		if err != nil {
			return nil, fmt.Errorf("failed to convert vector inner type: %w", err)
		}
		return iota_sdk_ffi.TypeTagNewVector(innerTag), nil
	} else if tag.Struct != nil {
		// Convert struct tag
		ffiAddr, err := toFfiAddress(tag.Struct.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to convert struct address: %w", err)
		}

		moduleId, err := iota_sdk_ffi.NewIdentifier(string(tag.Struct.Module))
		if err != nil {
			return nil, fmt.Errorf("failed to create module identifier: %w", err)
		}

		nameId, err := iota_sdk_ffi.NewIdentifier(string(tag.Struct.Name))
		if err != nil {
			return nil, fmt.Errorf("failed to create name identifier: %w", err)
		}

		// Convert type parameters recursively
		var typeParams []*iota_sdk_ffi.TypeTag
		for _, param := range tag.Struct.TypeParams {
			ffiParam, err := toFfiTypeTag(&param)
			if err != nil {
				return nil, fmt.Errorf("failed to convert type parameter: %w", err)
			}
			typeParams = append(typeParams, ffiParam)
		}

		structTag := iota_sdk_ffi.NewStructTag(ffiAddr, moduleId, nameId, typeParams)
		return iota_sdk_ffi.TypeTagNewStruct(structTag), nil
	}

	return nil, fmt.Errorf("unknown TypeTag variant")
}

func mapFfiObjectToIotaResponse(obj **iota_sdk_ffi.Object, options *iotajsonrpc.IotaObjectDataOptions, bcsBytes *[]byte) (*iotajsonrpc.IotaObjectResponse, error) {
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
	ot := o.ObjectType()
	if ot != nil {
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

	// Add owner if requested
	if options != nil && options.ShowOwner {
		if owner := o.Owner(); owner != nil {
			iotaOwner, err := fromFfiOwner(owner)
			if err != nil {
				return nil, fmt.Errorf("failed to convert owner: %w", err)
			}
			if iotaOwner != nil {
				ownerInternal := &iotajsonrpc.ObjectOwnerInternal{
					AddressOwner: iotaOwner.AddressOwner,
					ObjectOwner:  iotaOwner.ObjectOwner,
				}
				// Convert Shared field if present
				if iotaOwner.Shared != nil {
					ownerInternal.Shared = &struct {
						InitialSharedVersion *iotago.SequenceNumber `json:"initial_shared_version"`
					}{
						InitialSharedVersion: &iotaOwner.Shared.InitialSharedVersion,
					}
				}
				data.Owner = &iotajsonrpc.ObjectOwner{
					ObjectOwnerInternal: ownerInternal,
				}
			}
		}
	}

	// Add BCS data if requested and available
	if options != nil && options.ShowBcs && bcsBytes != nil && *bcsBytes != nil {
		bcsData := iotago.Base64Data(*bcsBytes)

		// For move objects, we need to populate the IotaRawMoveObject structure
		if typeStr != nil {
			structTag, err := iotago.StructTagFromString(extractTypeTag(*typeStr))
			if err == nil {
				data.Bcs = &serialization.TagJson[iotajsonrpc.IotaRawData]{
					Data: iotajsonrpc.IotaRawData{
						MoveObject: &iotajsonrpc.IotaRawMoveObject{
							Type:              *structTag,
							HasPublicTransfer: false, // This would need to be determined from object data
							Version:           iotago.SequenceNumber(ver),
							BcsBytes:          bcsData,
						},
					},
				}
			}
		}
	}

	return &iotajsonrpc.IotaObjectResponse{Data: data}, nil
}

var moveTagRe = regexp.MustCompile(
	`^(?:Struct\()?(0x[0-9a-fA-F]+::[A-Za-z_][A-Za-z0-9_]*::[A-Za-z_][A-Za-z0-9_]*)(?:\))?$`,
)

func extractTypeTag(s string) string {
	s = strings.TrimSpace(s)
	if m := moveTagRe.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	// Fallback: if parentheses exist, peel the first (...) pair.
	if l := strings.IndexByte(s, '('); l >= 0 {
		if r := strings.IndexByte(s[l+1:], ')'); r >= 0 {
			return s[l+1 : l+1+r]
		}
	}
	// Otherwise just return as-is if non-empty.
	if s != "" {
		return s
	}
	return ""
}

// formatExecutionError formats an ExecutionError with type information
func formatExecutionError(err iota_sdk_ffi.ExecutionError) string {
	switch e := err.(type) {
	case iota_sdk_ffi.ExecutionErrorUnusedValueWithoutDrop:
		return fmt.Sprintf("UnusedValueWithoutDrop(result: %d, subresult: %d)", e.Result, e.Subresult)
	case iota_sdk_ffi.ExecutionErrorInsufficientGas:
		return "InsufficientGas"
	case iota_sdk_ffi.ExecutionErrorInvalidGasObject:
		return "InvalidGasObject"
	case iota_sdk_ffi.ExecutionErrorInvariantViolation:
		return "InvariantViolation"
	case iota_sdk_ffi.ExecutionErrorFeatureNotYetSupported:
		return "FeatureNotYetSupported"
	case iota_sdk_ffi.ExecutionErrorObjectTooBig:
		return fmt.Sprintf("ObjectTooBig(size: %d, max: %d)", e.ObjectSize, e.MaxObjectSize)
	case iota_sdk_ffi.ExecutionErrorPackageTooBig:
		return fmt.Sprintf("PackageTooBig(size: %d, max: %d)", e.ObjectSize, e.MaxObjectSize)
	case iota_sdk_ffi.ExecutionErrorMoveAbort:
		return fmt.Sprintf("MoveAbort(location: %+v, code: %d)", e.Location, e.Code)
	case iota_sdk_ffi.ExecutionErrorCommandArgument:
		return fmt.Sprintf("CommandArgument(argument: %d, kind: %+v)", e.Argument, e.Kind)
	case iota_sdk_ffi.ExecutionErrorTypeArgument:
		return fmt.Sprintf("TypeArgument(typeArgument: %d, kind: %+v)", e.TypeArgument, e.Kind)
	case iota_sdk_ffi.ExecutionErrorInvalidPublicFunctionReturnType:
		return fmt.Sprintf("InvalidPublicFunctionReturnType(index: %d)", e.Index)
	default:
		return fmt.Sprintf("%T: %+v", err, err)
	}
}

// convertTransactionEffects converts FFI TransactionEffects to IotaTransactionBlockEffects
func convertTransactionEffects(effects *iota_sdk_ffi.TransactionEffects) (*iotajsonrpc.IotaTransactionBlockEffects, error) {
	if effects == nil {
		return nil, nil
	}

	if !effects.IsV1() {
		return nil, fmt.Errorf("unsupported TransactionEffects version")
	}

	v1 := effects.AsV1()
	// Convert execution status
	status := iotajsonrpc.ExecutionStatus{}
	switch s := v1.Status.(type) {
	case iota_sdk_ffi.ExecutionStatusSuccess:
		status.Status = "success"
	case iota_sdk_ffi.ExecutionStatusFailure:
		status.Status = "failure"
		// Extract error details from failure with proper type information
		if s.Error != nil {
			status.Error = formatExecutionError(s.Error)
		}
	default:
		status.Status = "unknown"
	}

	// Convert gas cost summary
	gasUsed := iotajsonrpc.GasCostSummary{
		ComputationCost:         iotajsonrpc.NewBigInt(v1.GasUsed.ComputationCost),
		StorageCost:             iotajsonrpc.NewBigInt(v1.GasUsed.StorageCost),
		StorageRebate:           iotajsonrpc.NewBigInt(v1.GasUsed.StorageRebate),
		NonRefundableStorageFee: iotajsonrpc.NewBigInt(v1.GasUsed.NonRefundableStorageFee),
	}

	// Convert transaction digest
	var txDigest iotago.TransactionDigest
	if v1.TransactionDigest != nil {
		d, err := fromFfiDigest(v1.TransactionDigest)
		if err == nil && d != nil {
			txDigest = *d
		}
	}

	// Convert dependencies
	var dependencies []iotago.TransactionDigest
	for _, dep := range v1.Dependencies {
		if d, err := fromFfiDigest(dep); err == nil && d != nil {
			dependencies = append(dependencies, *d)
		}
	}

	// Convert events digest
	var eventsDigest *iotago.TransactionEventsDigest
	if v1.EventsDigest != nil && *v1.EventsDigest != nil {
		if d, err := fromFfiDigest(*v1.EventsDigest); err == nil && d != nil {
			eventsDigest = (*iotago.TransactionEventsDigest)(d)
		}
	}

	// Convert ChangedObjects to object change fields
	var created []iotajsonrpc.OwnedObjectRef
	var mutated []iotajsonrpc.OwnedObjectRef
	var unwrapped []iotajsonrpc.OwnedObjectRef
	var deleted []iotajsonrpc.IotaObjectRef
	var unwrappedThenDeleted []iotajsonrpc.IotaObjectRef
	var wrapped []iotajsonrpc.IotaObjectRef
	var gasObject iotajsonrpc.OwnedObjectRef

	for i, changedObj := range v1.ChangedObjects {
		objID, err := fromFfiObjectID(changedObj.ObjectId)
		if err != nil || objID == nil {
			continue
		}

		// Determine object state changes
		inputIsMissing := false
		outputIsMissing := false
		var outputDigest *iotago.Digest
		var outputOwner *iotago.Owner

		// Check input state
		switch changedObj.InputState.(type) {
		case iota_sdk_ffi.ObjectInMissing:
			inputIsMissing = true
		case iota_sdk_ffi.ObjectInData:
			// Input exists
		}

		// Check output state
		switch output := changedObj.OutputState.(type) {
		case iota_sdk_ffi.ObjectOutMissing:
			outputIsMissing = true
		case iota_sdk_ffi.ObjectOutObjectWrite:
			if d, err := fromFfiDigest(output.Digest); err == nil && d != nil {
				outputDigest = d
			}
			if o, err := fromFfiOwner(output.Owner); err == nil && o != nil {
				outputOwner = o
			}
		case iota_sdk_ffi.ObjectOutPackageWrite:
			// Package writes are treated similarly to object writes
			if d, err := fromFfiDigest(output.Digest); err == nil && d != nil {
				outputDigest = d
			}
		}

		// Build object reference
		objRef := iotajsonrpc.IotaObjectRef{
			ObjectID: objID,
			Version:  v1.LamportVersion,
		}
		if outputDigest != nil {
			objRef.Digest = *outputDigest
		}

		ownedObjRef := iotajsonrpc.OwnedObjectRef{
			Reference: objRef,
		}
		if outputOwner != nil {
			ownedObjRef.Owner = serialization.TagJson[iotago.Owner]{Data: *outputOwner}
		}

		// Categorize based on state transitions
		switch changedObj.IdOperation {
		case iota_sdk_ffi.IdOperationCreated:
			created = append(created, ownedObjRef)
		case iota_sdk_ffi.IdOperationDeleted:
			if inputIsMissing {
				unwrappedThenDeleted = append(unwrappedThenDeleted, objRef)
			} else {
				deleted = append(deleted, objRef)
			}
		case iota_sdk_ffi.IdOperationNone:
			if inputIsMissing && !outputIsMissing {
				unwrapped = append(unwrapped, ownedObjRef)
			} else if !inputIsMissing && outputIsMissing {
				wrapped = append(wrapped, objRef)
			} else if !inputIsMissing && !outputIsMissing {
				mutated = append(mutated, ownedObjRef)
			}
		}

		// Check if this is the gas object
		if v1.GasObjectIndex != nil && *v1.GasObjectIndex == uint32(i) {
			gasObject = ownedObjRef
		}
	}

	// Build the V1 effects
	effectsV1 := &iotajsonrpc.IotaTransactionBlockEffectsV1{
		Status:               status,
		ExecutedEpoch:        iotajsonrpc.NewBigInt(v1.Epoch),
		GasUsed:              gasUsed,
		TransactionDigest:    txDigest,
		Dependencies:         dependencies,
		EventsDigest:         eventsDigest,
		Created:              created,
		Mutated:              mutated,
		Unwrapped:            unwrapped,
		Deleted:              deleted,
		UnwrappedThenDeleted: unwrappedThenDeleted,
		Wrapped:              wrapped,
		GasObject:            gasObject,
	}

	return &iotajsonrpc.IotaTransactionBlockEffects{
		V1: effectsV1,
	}, nil
}

// PopulateObjectTypesInChanges enriches ObjectChanges with object types by querying the chain.
// This should be called after SignAndExecuteTransaction to populate ObjectType fields.
//
// Note: For newly created objects, there may be an indexing delay. Consider adding a sleep
// before calling this function, or use DryRunTransaction which populates types automatically.
//
// Example:
//
//	txnResponse, err := client.SignAndExecuteTransaction(ctx, &req)
//	if err != nil { return err }
//	time.Sleep(2 * time.Second) // Wait for indexing
//	err = client.PopulateObjectTypesInChanges(ctx, txnResponse.ObjectChanges)
func (c *BindingClient) PopulateObjectTypesInChanges(ctx context.Context, objectChanges []serialization.TagJson[iotajsonrpc.ObjectChange]) error {
	for i := range objectChanges {
		change := &objectChanges[i].Data

		var objectID *iotago.ObjectID
		var objectTypePtr *string

		// Determine which type of change and get the object ID
		if change.Created != nil {
			objectID = &change.Created.ObjectID
			objectTypePtr = &change.Created.ObjectType
		} else if change.Mutated != nil {
			objectID = &change.Mutated.ObjectID
			objectTypePtr = &change.Mutated.ObjectType
		} else if change.Deleted != nil {
			objectID = &change.Deleted.ObjectID
			objectTypePtr = &change.Deleted.ObjectType
		} else if change.Wrapped != nil {
			objectID = &change.Wrapped.ObjectID
			objectTypePtr = &change.Wrapped.ObjectType
		}

		if objectID == nil {
			continue
		}

		// Query the object to get its type
		objResp, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: objectID})
		if err != nil {
			// For debugging: print the error
			fmt.Printf("Failed to get object %s: %v\n", objectID.String(), err)
			continue
		}
		if objResp.Data == nil {
			fmt.Printf("Object %s: Data is nil\n", objectID.String())
			continue
		}
		if objResp.Data.Type == nil {
			fmt.Printf("Object %s: Type is nil\n", objectID.String())
			continue
		}

		// Populate the ObjectType field
		fmt.Printf("Setting ObjectType for %s: %s\n", objectID.String(), *objResp.Data.Type)
		*objectTypePtr = *objResp.Data.Type
	}
	return nil
}

// populateObjectTypesFromDryRun extracts object types from DryRunResult and populates them in ChangedObjects
func populateObjectTypesFromDryRun(effects *iota_sdk_ffi.TransactionEffects, dryRunResult *iota_sdk_ffi.DryRunResult) {
	if effects == nil || dryRunResult == nil || !effects.IsV1() {
		return
	}

	// Build a map of object ID (bytes) -> type string
	objectTypes := make(map[string]string)

	for _, result := range dryRunResult.Results {
		// Extract types from mutated references
		for _, mutation := range result.MutatedReferences {
			if mutation.TypeTag != nil {
				typeStr := mutation.TypeTag.String()
				// Extract object ID from BCS if it's a coin/object type
				// The BCS data starts with the object ID (32 bytes)
				if len(mutation.Bcs) >= 32 {
					objectID := string(mutation.Bcs[:32])
					objectTypes[objectID] = typeStr
				}
			}
		}

		// Extract types from return values
		for _, ret := range result.ReturnValues {
			if ret.TypeTag != nil {
				typeStr := ret.TypeTag.String()
				// Extract object ID from BCS if available
				if len(ret.Bcs) >= 32 {
					objectID := string(ret.Bcs[:32])
					objectTypes[objectID] = typeStr
				}
			}
		}
	}

	// Now populate the ChangedObjects with the types
	v1 := effects.AsV1()
	for i := range v1.ChangedObjects {
		objIDBytes := v1.ChangedObjects[i].ObjectId.ToBytes()
		if typeStr, ok := objectTypes[string(objIDBytes)]; ok {
			v1.ChangedObjects[i].ObjectType = &typeStr
		}
	}
}

// convertChangedObjectsToObjectChanges converts FFI TransactionEffects ChangedObjects to ObjectChanges
func (c *BindingClient) convertChangedObjectsToObjectChanges(effects *iota_sdk_ffi.TransactionEffects) ([]serialization.TagJson[iotajsonrpc.ObjectChange], error) {
	time.Sleep(400 * time.Millisecond)
	if effects == nil {
		return nil, nil
	}

	if !effects.IsV1() {
		return nil, fmt.Errorf("unsupported TransactionEffects version")
	}

	v1 := effects.AsV1()
	var objectChanges []serialization.TagJson[iotajsonrpc.ObjectChange]

	// Note: Sender address is not available in TransactionEffects.
	// We'll create a zero address as a placeholder.
	zeroAddr := iotago.Address{}
	for _, changedObj := range v1.ChangedObjects {
		fmt.Println("*******changedObj.IdOperation: ", changedObj.IdOperation)
		fmt.Println("*******changedObj.ObjectType: ", changedObj.ObjectType)
		fmt.Println("*******changedObj.ObjectId: ", changedObj.ObjectId.ToHex())
		if changedObj.ObjectType != nil {
			fmt.Println("*******!changedObj: ", *changedObj.ObjectType)
		}
		objID, err := fromFfiObjectID(changedObj.ObjectId)
		if err != nil || objID == nil {
			continue
		}

		// Determine object state changes
		inputIsMissing := false
		outputIsMissing := false
		isPackageWrite := false
		var outputDigest *iotago.Digest
		var outputOwner *iotago.Owner
		var outputVersion uint64

		// Check input state
		switch changedObj.InputState.(type) {
		case iota_sdk_ffi.ObjectInMissing:
			inputIsMissing = true
		case iota_sdk_ffi.ObjectInData:
			// Input exists
		}

		// Check output state
		switch output := changedObj.OutputState.(type) {
		case iota_sdk_ffi.ObjectOutMissing:
			outputIsMissing = true
		case iota_sdk_ffi.ObjectOutObjectWrite:
			if d, err := fromFfiDigest(output.Digest); err == nil && d != nil {
				outputDigest = d
			}
			if o, err := fromFfiOwner(output.Owner); err == nil && o != nil {
				outputOwner = o
			}
		case iota_sdk_ffi.ObjectOutPackageWrite:
			isPackageWrite = true
			outputVersion = output.Version
			if d, err := fromFfiDigest(output.Digest); err == nil && d != nil {
				outputDigest = d
			}
		}

		// Create ObjectChange based on state transitions
		var change iotajsonrpc.ObjectChange
		var changeObjectType string
		if changedObj.ObjectType != nil {
			changeObjectType = *changedObj.ObjectType
		}

		switch changedObj.IdOperation {
		case iota_sdk_ffi.IdOperationCreated:
			if isPackageWrite && !outputIsMissing && outputDigest != nil {
				// Package creation -> Published
				change.Published = &struct {
					PackageId iotago.ObjectID     `json:"packageId"`
					Version   *iotajsonrpc.BigInt `json:"version"`
					Digest    iotago.ObjectDigest `json:"digest"`
					Nodules   []string            `json:"nodules"`
				}{
					PackageId: *objID,
					Version:   iotajsonrpc.NewBigInt(outputVersion),
					Digest:    *outputDigest,
					Nodules:   []string{}, // TODO: Extract module names if available
				}
			} else if !outputIsMissing && outputDigest != nil && outputOwner != nil {
				resGetObject, err := c.GetObject(context.TODO(), iotaclient.GetObjectRequest{ObjectID: objID, Options: &iotajsonrpc.IotaObjectDataOptions{
					ShowType:                true,
					ShowContent:             true,
					ShowBcs:                 true,
					ShowOwner:               true,
					ShowPreviousTransaction: true,
					ShowStorageRebate:       true,
					ShowDisplay:             true,
				}})
				if err != nil {
					panic(err)
				}

				change.Created = &struct {
					Sender     iotago.Address          `json:"sender"`
					Owner      iotajsonrpc.ObjectOwner `json:"owner"`
					ObjectType string                  `json:"objectType"`
					ObjectID   iotago.ObjectID         `json:"objectId"`
					Version    *iotajsonrpc.BigInt     `json:"version"`
					Digest     iotago.ObjectDigest     `json:"digest"`
				}{
					Sender:     zeroAddr,
					Owner:      iotagoOwnerToObjectOwner(outputOwner),
					ObjectType: *resGetObject.Data.Type,
					ObjectID:   *objID,
					Version:    iotajsonrpc.NewBigInt(v1.LamportVersion),
					Digest:     *outputDigest,
				}
			}
		case iota_sdk_ffi.IdOperationDeleted:
			if inputIsMissing {
				// UnwrappedThenDeleted - not represented in ObjectChange
				continue
			} else {
				change.Deleted = &struct {
					Sender     iotago.Address      `json:"sender"`
					ObjectType string              `json:"objectType"`
					ObjectID   iotago.ObjectID     `json:"objectId"`
					Version    *iotajsonrpc.BigInt `json:"version"`
				}{
					Sender:     zeroAddr,
					ObjectType: changeObjectType,
					ObjectID:   *objID,
					Version:    iotajsonrpc.NewBigInt(v1.LamportVersion),
				}
			}
		case iota_sdk_ffi.IdOperationNone:
			if !inputIsMissing && outputIsMissing {
				change.Wrapped = &struct {
					Sender     iotago.Address      `json:"sender"`
					ObjectType string              `json:"objectType"`
					ObjectID   iotago.ObjectID     `json:"objectId"`
					Version    *iotajsonrpc.BigInt `json:"version"`
				}{
					Sender:     zeroAddr,
					ObjectType: changeObjectType,
					ObjectID:   *objID,
					Version:    iotajsonrpc.NewBigInt(v1.LamportVersion),
				}
			} else if !inputIsMissing && !outputIsMissing && outputDigest != nil && outputOwner != nil {
				change.Mutated = &struct {
					Sender          iotago.Address          `json:"sender"`
					Owner           iotajsonrpc.ObjectOwner `json:"owner"`
					ObjectType      string                  `json:"objectType"`
					ObjectID        iotago.ObjectID         `json:"objectId"`
					Version         *iotajsonrpc.BigInt     `json:"version"`
					PreviousVersion *iotajsonrpc.BigInt     `json:"previousVersion"`
					Digest          iotago.ObjectDigest     `json:"digest"`
				}{
					Sender:          zeroAddr,
					Owner:           iotagoOwnerToObjectOwner(outputOwner),
					ObjectType:      changeObjectType,
					ObjectID:        *objID,
					Version:         iotajsonrpc.NewBigInt(v1.LamportVersion),
					PreviousVersion: iotajsonrpc.NewBigInt(v1.LamportVersion - 1), // Approximation
					Digest:          *outputDigest,
				}
			}
		}
		// Only add non-empty changes
		if change.Created != nil || change.Deleted != nil || change.Mutated != nil || change.Wrapped != nil || change.Published != nil {
			objectChanges = append(objectChanges, serialization.TagJson[iotajsonrpc.ObjectChange]{Data: change})
		}
	}

	return objectChanges, nil
}

// convertTransactionEffectsToBalanceChanges converts FFI TransactionEffects to BalanceChanges
//
// CURRENT LIMITATION: This function requires FFI enhancement to work properly.
//
// The Rust SDK FFI needs to add a method like:
//   - `TransactionEffects.GetBalanceChanges() -> Vec<BalanceChange>`
//   - or `DryRunResult.GetBalanceChanges() -> Vec<BalanceChange>`
//
// Why the current approach doesn't work:
//   - For dry runs: Changed objects don't exist on-chain yet, so we can't query their balances
//   - DryRunResult.Results[] contains BCS-encoded coin data with balances
//   - BUT it's indexed by TransactionArgument (input/result indices), not ObjectId
//   - Mapping TransactionArgument -> ObjectId requires complex transaction structure analysis
//
// Recommended FFI enhancement:
//
//	Add to iota_sdk_ffi.udl:
//	  interface TransactionEffects {
//	    sequence<BalanceChange> balance_changes();
//	  };
//
//	This would calculate balance changes server-side where all data is available.
func (c *BindingClient) convertTransactionEffectsToBalanceChanges(ctx context.Context, effects *iota_sdk_ffi.TransactionEffects, dryRunResult *iota_sdk_ffi.DryRunResult) ([]iotajsonrpc.BalanceChange, error) {
	// FUNDAMENTAL LIMITATION: Balance changes cannot be calculated without FFI enhancement
	//
	// The problem:
	// 1. For dry runs: Objects don't exist on-chain yet, so we can't query their balances
	// 2. DryRunResult contains balance data in mutations[], BUT:
	//    - Mutations are indexed by TransactionArgument (input[0], result[1][0], etc.)
	//    - ChangedObjects are indexed by ObjectId
	//    - There's no mapping between TransactionArgument and ObjectId
	//
	// The ONLY solution is to add an FFI method that calculates balance changes server-side.
	// See FFI_BALANCE_CHANGES_ENHANCEMENT.md for the specification.
	//
	// Tests will fail until FFI enhancement is implemented in the upstream iota-rust-sdk.

	return []iotajsonrpc.BalanceChange{}, nil
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
		resp, err := mapFfiObjectToIotaResponse(obj, nil, nil)
		if err != nil {
			return nil, err
		}
		page.Data = append(page.Data, *resp)
	}
	page.HasNextPage = cp.PageInfo.HasNextPage
	return &page, nil
}

func (c *BindingClient) QueryEvents(ctx context.Context, req iotaclient.QueryEventsRequest) (*iotajsonrpc.EventPage, error) {
	// Convert the request filter to FFI EventFilter
	var ffiFilter *iota_sdk_ffi.EventFilter
	if req.Query != nil {
		ffiFilter = &iota_sdk_ffi.EventFilter{}

		// Map Transaction digest
		if req.Query.Transaction != nil {
			txDigest := req.Query.Transaction.String()
			ffiFilter.TransactionDigest = &txDigest
		}

		// Map Sender address
		if req.Query.Sender != nil {
			ffiAddr, err := toFfiAddress(req.Query.Sender)
			if err != nil {
				return nil, fmt.Errorf("failed to convert sender address: %w", err)
			}
			ffiFilter.Sender = &ffiAddr
		}

		// Map MoveEventType to EventType
		if req.Query.MoveEventType != nil {
			eventType := req.Query.MoveEventType.String()
			ffiFilter.EventType = &eventType
		}

		// Map MoveModule
		if req.Query.MoveModule != nil {
			module := fmt.Sprintf("%s::%s", req.Query.MoveModule.Package.String(), req.Query.MoveModule.Module)
			ffiFilter.EmittingModule = &module
		}
	}

	ep, err := c.qclient.Events(ffiFilter, nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL Events failed: %w", err)
	}

	// Map the events to the expected format
	events := make([]iotajsonrpc.IotaEvent, 0, len(ep.Data))
	for _, event := range ep.Data {
		iotaEvent := iotajsonrpc.IotaEvent{
			ParsedJson: []byte(event.Json),
		}

		// Map package ID
		if event.PackageId != nil {
			pkgID, err := fromFfiObjectID(event.PackageId)
			if err != nil {
				return nil, fmt.Errorf("failed to convert package ID: %w", err)
			}
			iotaEvent.PackageId = pkgID
		}

		// Map transaction module
		iotaEvent.TransactionModule = iotago.Identifier(event.Module)

		// Map sender
		if event.Sender != nil {
			sender, err := iotago.AddressFromHex(event.Sender.ToHex())
			if err != nil {
				return nil, fmt.Errorf("failed to convert sender address: %w", err)
			}
			iotaEvent.Sender = sender
		}

		// Map type
		if event.Type != "" {
			structTag, err := iotago.StructTagFromString(event.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to parse struct tag: %w", err)
			}
			iotaEvent.Type = structTag
		}

		events = append(events, iotaEvent)
	}

	return &iotajsonrpc.EventPage{
		Data:        events,
		HasNextPage: ep.PageInfo.HasNextPage,
	}, nil
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
	if len(txDataBytes) == 0 {
		return nil, fmt.Errorf("transaction data bytes are required")
	}

	// Unmarshal transaction data
	txData, err := bcs.Unmarshal[iotago.TransactionData](txDataBytes.Data())
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal transaction data: %w", err)
	}

	// Convert to FFI Transaction
	tx, err := convertTransactionDataToTransaction(&txData)
	if err != nil {
		return nil, fmt.Errorf("can't convert to Transaction: %w", err)
	}

	// Call DryRunTx
	skipChecks := false
	dryRunResult, err := c.qclient.DryRunTx(tx, &skipChecks)
	if err != nil {
		if sdkErr, ok := err.(*iota_sdk_ffi.SdkFfiError); ok && sdkErr != nil {
			return nil, fmt.Errorf("failed to dry run tx: %w", err)
		}
	}

	// Build response
	response := &iotajsonrpc.DryRunTransactionBlockResponse{}

	// Convert effects
	if dryRunResult.Effects != nil && *dryRunResult.Effects != nil {
		// Populate object types from dry run results
		populateObjectTypesFromDryRun(*dryRunResult.Effects, &dryRunResult)

		convertedEffects, err := convertTransactionEffects(*dryRunResult.Effects)
		if err != nil {
			return nil, fmt.Errorf("failed to convert transaction effects: %w", err)
		}
		response.Effects = serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{Data: *convertedEffects}

		// Convert object changes
		objectChanges, err := c.convertChangedObjectsToObjectChanges(*dryRunResult.Effects)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object changes: %w", err)
		}
		response.ObjectChanges = objectChanges

		// Convert balance changes
		balanceChanges, err := c.convertTransactionEffectsToBalanceChanges(ctx, *dryRunResult.Effects, &dryRunResult)
		if err != nil {
			return nil, fmt.Errorf("failed to convert balance changes: %w", err)
		}
		response.BalanceChanges = balanceChanges
	}

	// TODO: Convert events if needed
	// response.Events = ...
	// response.Input = ... (would need conversion from iotago types to iotajsonrpc types)

	return response, nil
}

// DryRunTransactionRaw performs a dry run and returns the raw FFI types
//
// This method exposes the underlying iota_sdk_ffi.DryRunResult directly,
// allowing you to access all the raw data including:
// - dryRunResult.Results[] - contains BCS-encoded mutation data with coin balances
// - dryRunResult.Effects - the transaction effects
// - dryRunResult.Transaction - the signed transaction
//
// This is useful for:
// 1. Debugging and exploring the raw FFI data structure
// 2. Implementing custom balance change extraction logic
// 3. Accessing data not exposed through the standard DryRunTransaction method
//
// Example usage:
//
//	rawResult, err := client.DryRunTransactionRaw(ctx, txDataBytes)
//	if err != nil { ... }
//
//	// Access raw mutation data
//	for _, result := range rawResult.Results {
//	    for _, mutRef := range result.MutatedReferences {
//	        // mutRef.Bcs contains the BCS-encoded coin data
//	        // mutRef.Input is the TransactionArgument (e.g., Input{Ix: 0})
//	    }
//	}
func (c *BindingClient) DryRunTransactionRaw(ctx context.Context, txDataBytes iotago.Base64Data) (*iota_sdk_ffi.DryRunResult, error) {
	if len(txDataBytes) == 0 {
		return nil, fmt.Errorf("transaction data bytes are required")
	}

	// Unmarshal transaction data
	txData, err := bcs.Unmarshal[iotago.TransactionData](txDataBytes.Data())
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal transaction data: %w", err)
	}

	// Convert to FFI Transaction
	tx, err := convertTransactionDataToTransaction(&txData)
	if err != nil {
		return nil, fmt.Errorf("can't convert to Transaction: %w", err)
	}

	// Call DryRunTx and return the raw result
	skipChecks := false
	dryRunResult, err := c.qclient.DryRunTx(tx, &skipChecks)
	if err != nil {
		if sdkErr, ok := err.(*iota_sdk_ffi.SdkFfiError); ok && sdkErr != nil {
			return nil, fmt.Errorf("failed to dry run tx: %w", err)
		}
	}

	return &dryRunResult, nil
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
	epoch, err := c.qclient.Epoch(nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL GetLatestIotaSystemState failed: %w", err)
	}

	// Helper function to convert string pointer to BigInt
	strToBigInt := func(s *string) *iotajsonrpc.BigInt {
		if s == nil {
			return nil
		}
		bi := &iotajsonrpc.BigInt{Int: new(big.Int)}
		bi.SetString(*s, 10)
		return bi
	}

	// Convert ObjectId pointer to iotago.ObjectID
	objIDConv := func(id **iota_sdk_ffi.ObjectId) iotago.ObjectID {
		if id == nil || *id == nil {
			return iotago.ObjectID{}
		}
		bytes := (*id).ToBytes()
		if len(bytes) != 32 {
			return iotago.ObjectID{}
		}
		var arr [32]byte
		copy(arr[:], bytes)
		return iotago.ObjectID(arr)
	}

	// Convert int32 pointer to BigInt
	int32ToBigInt := func(i *int32) *iotajsonrpc.BigInt {
		if i == nil {
			return nil
		}
		return iotajsonrpc.NewBigIntInt64(int64(*i))
	}

	// Convert []int32 pointer to []*BigInt
	int32SliceToBigIntSlice := func(s *[]int32) []*iotajsonrpc.BigInt {
		if s == nil {
			return nil
		}
		result := make([]*iotajsonrpc.BigInt, len(*s))
		for i, v := range *s {
			result[i] = iotajsonrpc.NewBigIntInt64(int64(v))
		}
		return result
	}

	summary := &iotajsonrpc.IotaSystemStateSummary{
		Epoch:                 iotajsonrpc.NewBigInt(epoch.EpochId),
		ReferenceGasPrice:     strToBigInt(epoch.ReferenceGasPrice),
		EpochStartTimestampMs: iotajsonrpc.NewBigInt(epoch.StartTimestamp),
	}

	// Map SystemStateVersion if available
	if epoch.SystemStateVersion != nil {
		summary.SystemStateVersion = iotajsonrpc.NewBigInt(*epoch.SystemStateVersion)
	}

	// Map ProtocolVersion from ProtocolConfigs if available
	if epoch.ProtocolConfigs != nil {
		summary.ProtocolVersion = iotajsonrpc.NewBigInt(epoch.ProtocolConfigs.ProtocolVersion)

		// Extract config attributes from ProtocolConfigs.Configs
		for _, attr := range epoch.ProtocolConfigs.Configs {
			if attr.Value == nil {
				continue
			}

			switch attr.Key {
			case "epoch_duration_ms", "epochDurationMs", "epoch-duration-ms":
				if val, err := strconv.ParseUint(*attr.Value, 10, 64); err == nil {
					summary.EpochDurationMs = iotajsonrpc.NewBigInt(val)
				}
			case "min_validator_count", "minValidatorCount":
				if val, err := strconv.ParseUint(*attr.Value, 10, 64); err == nil {
					summary.MinValidatorCount = iotajsonrpc.NewBigInt(val)
				}
			case "max_validator_count", "maxValidatorCount":
				if val, err := strconv.ParseUint(*attr.Value, 10, 64); err == nil {
					summary.MaxValidatorCount = iotajsonrpc.NewBigInt(val)
				}
			}
		}
	}

	// FIXME: EpochDurationMs is not available in the current GraphQL API response.
	// The epoch_duration_ms key does not exist in ProtocolConfigs.Configs attributes.
	// Using a hardcoded default value of 24 hours (86400000 ms) as a workaround.
	// This should be replaced with the actual value from the protocol config once available.
	if summary.EpochDurationMs == nil {
		summary.EpochDurationMs = iotajsonrpc.NewBigInt(86400000) // 24 hours in milliseconds
	}

	// Map ValidatorSet fields if available
	if epoch.ValidatorSet != nil {
		vs := epoch.ValidatorSet
		summary.TotalStake = strToBigInt(vs.TotalStake)
		summary.PendingActiveValidatorsId = objIDConv(vs.PendingActiveValidatorsId)
		summary.PendingActiveValidatorsSize = int32ToBigInt(vs.PendingActiveValidatorsSize)
		summary.PendingRemovals = int32SliceToBigIntSlice(vs.PendingRemovals)
		summary.StakingPoolMappingsId = objIDConv(vs.StakingPoolMappingsId)
		summary.StakingPoolMappingsSize = int32ToBigInt(vs.StakingPoolMappingsSize)
		summary.InactivePoolsId = objIDConv(vs.InactivePoolsId)
		summary.InactivePoolsSize = int32ToBigInt(vs.InactivePoolsSize)
		summary.ValidatorCandidatesId = objIDConv(vs.ValidatorCandidatesId)
		summary.ValidatorCandidatesSize = int32ToBigInt(vs.ValidatorCandidatesSize)
	}

	metadata, err := c.qclient.CoinMetadata("0x2::iota::IOTA")
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL CoinMetadata failed: %w", err)
	}
	summary.IotaTotalSupply = iotajsonrpc.NewBigIntFromString(*metadata.Supply)
	// Note: Many fields in IotaSystemStateSummary don't have corresponding fields in Epoch
	// and are left as nil or use defaults:
	// - StorageFundTotalObjectStorageRebates, StorageFundNonRefundableBalance
	// - SafeMode, SafeModeStorageCharges, SafeModeStorageRewards, SafeModeComputationRewards
	// - SafeModeStorageRebates, SafeModeNonRefundableStorageFee
	// - EpochDurationMs (using hardcoded 24h default - see FIXME above)
	// - MinValidatorCount, MaxValidatorCount (may not be available in ProtocolConfigs)
	// - StakeSubsidy* fields, Validator* threshold fields
	// - ActiveValidators, AtRiskValidators, ValidatorReportRecords

	return summary, nil
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

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget
	builder = builder.GasBudget(req.GasBudget)

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

		ffiObjID, err := toFfiObjectID(gasObj.Data.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder = builder.Gas(ffiObjID)
	} else {
		// Find suitable gas coins
		builder, err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, req.GasBudget, 0)
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

			// Build Function parameters
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

			// Parse type arguments if provided
			var typeArgs []*iota_sdk_ffi.TypeTag
			if typeArgsParam, ok := txParam["typeArguments"].([]interface{}); ok {
				for _, typeArgInterface := range typeArgsParam {
					typeArgStr, ok := typeArgInterface.(string)
					if !ok {
						return nil, fmt.Errorf("type argument must be a string")
					}
					iotagoTypeTag, err := iotago.TypeTagFromString(typeArgStr)
					if err != nil {
						return nil, fmt.Errorf("failed to parse type argument %q: %w", typeArgStr, err)
					}
					ffiTypeTag, err := toFfiTypeTag(iotagoTypeTag)
					if err != nil {
						return nil, fmt.Errorf("failed to convert type argument %q: %w", typeArgStr, err)
					}
					typeArgs = append(typeArgs, ffiTypeTag)
				}
			}

			// Convert arguments (simplified - real implementation would be more complex)
			var args []*iota_sdk_ffi.PtbArgument
			if argsParam, ok := txParam["arguments"].([]interface{}); ok {
				for _, arg := range argsParam {
					switch v := arg.(type) {
					case string:
						args = append(args, iota_sdk_ffi.PtbArgumentString(v))
					default:
						// For object references, this would need more complex handling
						return nil, fmt.Errorf("unsupported argument type in batch transaction")
					}
				}
			}

			builder = builder.MoveCall(packageAddr, moduleId, functionId, args, typeArgs, nil)

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
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
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

		ffiObjID, err := toFfiObjectID(gasObj.Data.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder = builder.Gas(ffiObjID)
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		builder, err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
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

	primaryObjID, err := toFfiObjectID(primaryObj.Data.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert primary coin: %w", err)
	}

	// Get coin to merge object
	coinToMergeObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: req.CoinToMerge})
	if err != nil {
		return nil, fmt.Errorf("failed to get coin to merge object: %w", err)
	}
	if coinToMergeObj.Data == nil {
		return nil, fmt.Errorf("coin to merge object not found")
	}

	coinToMergeObjID, err := toFfiObjectID(coinToMergeObj.Data.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert coin to merge: %w", err)
	}

	// Add merge coins command
	builder = builder.MergeCoins(primaryObjID, []*iota_sdk_ffi.ObjectId{coinToMergeObjID})

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
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

		ffiObjID, err := toFfiObjectID(gasObj.Data.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder = builder.Gas(ffiObjID)
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		builder, err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to add gas coins: %w", err)
		}
	}

	gasPrice, err := c.qclient.ReferenceGasPrice(nil) // Use current epoch
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL ReferenceGasPrice failed: %w", err)
	}
	if gasPrice == nil {
		tmp := uint64(1000)
		gasPrice = &tmp
	}
	builder = builder.GasPrice(*gasPrice)

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

	// Parse TypeArgs from strings to TypeTag
	var typeArgs []*iota_sdk_ffi.TypeTag
	for _, typeArgStr := range req.TypeArgs {
		iotagoTypeTag, err := iotago.TypeTagFromString(typeArgStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse type argument %q: %w", typeArgStr, err)
		}
		ffiTypeTag, err := toFfiTypeTag(iotagoTypeTag)
		if err != nil {
			return nil, fmt.Errorf("failed to convert type argument %q: %w", typeArgStr, err)
		}
		typeArgs = append(typeArgs, ffiTypeTag)
	}

	// Convert arguments
	var args []*iota_sdk_ffi.PtbArgument
	for _, arg := range req.Arguments {
		// Check if arg is a slice/array using reflection
		argValue := reflect.ValueOf(arg)
		if argValue.Kind() == reflect.Slice || argValue.Kind() == reflect.Array {
			// Handle slices and arrays by BCS encoding them
			// We need to handle the specific types we support
			switch v := arg.(type) {
			case []string:
				// BCS encode each string separately
				var encodedStrings [][]byte
				for _, s := range v {
					bcsBytes, err := bcs.Marshal(&s)
					if err != nil {
						return nil, fmt.Errorf("failed to BCS encode string argument: %w", err)
					}
					encodedStrings = append(encodedStrings, bcsBytes)
				}
				args = append(args, iota_sdk_ffi.PtbArgumentVector(encodedStrings))
			case []uint64:
				// BCS encode each uint64 separately
				var encodedUints [][]byte
				for _, u := range v {
					bcsBytes, err := bcs.Marshal(&u)
					if err != nil {
						return nil, fmt.Errorf("failed to BCS encode uint64 argument: %w", err)
					}
					encodedUints = append(encodedUints, bcsBytes)
				}
				args = append(args, iota_sdk_ffi.PtbArgumentVector(encodedUints))
			case [][]byte:
				args = append(args, iota_sdk_ffi.PtbArgumentVector(v))
			default:
				return nil, fmt.Errorf("unsupported slice/array argument type: %T", arg)
			}
			continue
		}

		switch v := arg.(type) {
		case string:
			// Try to parse as address
			if addr, err := iotago.AddressFromHex(v); err == nil {
				ffiAddr, err := toFfiAddress(addr)
				if err != nil {
					return nil, fmt.Errorf("failed to convert address: %w", err)
				}
				args = append(args, iota_sdk_ffi.PtbArgumentAddress(ffiAddr))
			} else {
				// Treat as string literal
				args = append(args, iota_sdk_ffi.PtbArgumentString(v))
			}
		case uint64:
			args = append(args, iota_sdk_ffi.PtbArgumentU64(v))
		case *iotago.ObjectID:
			// Get object and convert to input
			obj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: v})
			if err != nil {
				return nil, fmt.Errorf("failed to get argument object: %w", err)
			}
			if obj.Data == nil {
				return nil, fmt.Errorf("argument object not found")
			}

			ffiObjID, err := toFfiObjectID(obj.Data.ObjectID)
			if err != nil {
				return nil, fmt.Errorf("failed to convert argument object: %w", err)
			}
			args = append(args, iota_sdk_ffi.PtbArgumentObjectId(ffiObjID))
		default:
			return nil, fmt.Errorf("unsupported argument type: %T", arg)
		}
	}

	// Add move call command
	builder = builder.MoveCall(packageAddr, moduleId, functionId, args, typeArgs, nil)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
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

		ffiObjID, err := toFfiObjectID(gasObj.Data.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder = builder.Gas(ffiObjID)
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		builder, err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
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

	primaryCoinObjID, err := toFfiObjectID(primaryCoinObj.Data.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert primary coin: %w", err)
	}

	// Create amounts array for SplitCoins
	var amounts []uint64
	splitNames := make([]string, len(req.Amount))
	for i, amount := range req.Amount {
		amounts = append(amounts, amount.Uint64())
		splitNames[i] = fmt.Sprintf("split_%d", i)
	}

	// Split the primary coin
	builder = builder.SplitCoins(primaryCoinObjID, amounts, splitNames)

	// Transfer each split result to corresponding recipient
	for i, recipient := range req.Recipients {
		// Create recipient argument
		ffiRecipient, err := toFfiAddress(recipient)
		if err != nil {
			return nil, fmt.Errorf("failed to convert recipient address: %w", err)
		}

		// Get the i-th split coin using result reference
		coinToTransfer := iota_sdk_ffi.PtbArgumentRes(splitNames[i])
		builder = builder.TransferObjects(ffiRecipient, []*iota_sdk_ffi.PtbArgument{coinToTransfer})
	}

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
	}

	// For PayAllIota, we use the input coins directly and transfer them all
	if len(req.InputCoins) == 0 {
		return nil, fmt.Errorf("no input coins provided")
	}

	// Get all input coin objects and convert to arguments
	var coinArgs []*iota_sdk_ffi.PtbArgument
	for _, coinID := range req.InputCoins {
		coinObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: coinID})
		if err != nil {
			return nil, fmt.Errorf("failed to get coin object: %w", err)
		}
		if coinObj.Data == nil {
			return nil, fmt.Errorf("coin object not found")
		}

		ffiObjID, err := toFfiObjectID(coinObj.Data.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert coin: %w", err)
		}
		coinArgs = append(coinArgs, iota_sdk_ffi.PtbArgumentObjectId(ffiObjID))
	}

	// Create recipient argument
	ffiRecipient, err := toFfiAddress(req.Recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to convert recipient address: %w", err)
	}

	// Transfer all coins to recipient
	builder = builder.TransferObjects(ffiRecipient, coinArgs)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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

	// Set sender
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins (auto-selected since PayIota doesn't specify gas coins)
	gasBudget := uint64(1000000) // default gas budget
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}
	builder, err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
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

	primaryCoinObjID, err := toFfiObjectID(primaryCoinObj.Data.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert primary coin: %w", err)
	}

	// Create amounts array for SplitCoins
	var amounts []uint64
	splitNames := make([]string, len(req.Amount))
	for i, amount := range req.Amount {
		amounts = append(amounts, amount.Uint64())
		splitNames[i] = fmt.Sprintf("split_%d", i)
	}

	// Split the primary coin
	builder = builder.SplitCoins(primaryCoinObjID, amounts, splitNames)

	// Transfer each split result to corresponding recipient
	for i, recipient := range req.Recipients {
		// Create recipient argument
		ffiRecipient, err := toFfiAddress(recipient)
		if err != nil {
			return nil, fmt.Errorf("failed to convert recipient address: %w", err)
		}

		// Get the i-th split coin using result reference
		coinToTransfer := iota_sdk_ffi.PtbArgumentRes(splitNames[i])
		builder = builder.TransferObjects(ffiRecipient, []*iota_sdk_ffi.PtbArgument{coinToTransfer})
	}

	var gasRefs []iotago.ObjectRef
	for i, gasCoin := range req.InputCoins {
		if i == 0 {
			continue
		}
		gasCoinObj, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: gasCoin})
		if err != nil {
			return nil, fmt.Errorf("failed to get primary coin object: %w", err)
		}
		gasRefs = append(gasRefs, gasCoinObj.Data.Ref())
		ffiObjectID, err := toFfiObjectID(gasCoin)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gasCoin to FFI type: %w", err)
		}
		builder.Gas(ffiObjectID)
		break
	}

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("failed to finish transaction: %w", err)
	}

	txBytes, err := tx.BcsSerialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Build TransactionBytes response
	return &iotajsonrpc.TransactionBytes{
		Gas:          gasRefs,
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
	senderAddr, err := toFfiAddress(req.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
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
		ffiObjID, err := toFfiObjectID(gasObj.Data.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder = builder.Gas(ffiObjID)
	} else {
		// Find suitable gas coins
		gasBudget := uint64(10000000) // higher default gas budget for publish
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		builder, err = c.addGasCoinsToBuilder(ctx, builder, req.Sender, gasBudget, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to add gas coins: %w", err)
		}
	}

	gasPrice, err := c.qclient.ReferenceGasPrice(nil) // Use current epoch
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL ReferenceGasPrice failed: %w", err)
	}
	if gasPrice == nil {
		tmp := uint64(1000)
		gasPrice = &tmp
	}
	builder = builder.GasPrice(*gasPrice)

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

	// Add publish command with upgrade capability name
	upgradeCapName := "upgrade_cap"
	builder = builder.Publish(modules, dependencies, upgradeCapName)

	// Transfer upgrade capability to sender
	ffiSender, err := toFfiAddress(req.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender: %w", err)
	}
	// Get the upgrade capability result by name
	upgradeCapArg := iota_sdk_ffi.PtbArgumentRes(upgradeCapName)
	builder = builder.TransferObjects(ffiSender, []*iota_sdk_ffi.PtbArgument{upgradeCapArg})

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins
	gasBudget := uint64(10000000) // higher gas budget for staking
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}

	// Find gas coins to use for both gas and splitting
	gasCoins, err := c.FindCoinsForGasPayment(ctx, req.Signer, iotago.ProgrammableTransaction{}, 0, gasBudget)
	if err != nil {
		return nil, fmt.Errorf("failed to find gas coins: %w", err)
	}
	if len(gasCoins) == 0 {
		return nil, fmt.Errorf("no gas coins available")
	}

	// Add gas coins to builder
	for _, coin := range gasCoins {
		ffiObjID, err := toFfiObjectID(coin.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder = builder.Gas(ffiObjID)
	}

	// Split the first gas coin to get stake amount
	ffiGasCoinID, err := toFfiObjectID(gasCoins[0].ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert gas coin for splitting: %w", err)
	}
	stakeAmountVal := req.Amount.Uint64()
	splitCoinName := "stake_coin"
	builder = builder.SplitCoins(ffiGasCoinID, []uint64{stakeAmountVal}, []string{splitCoinName})

	// Reference the split result by name
	stakeTokenArg := iota_sdk_ffi.PtbArgumentRes(splitCoinName)

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

	// Add validator address as argument
	ffiValidator, err := toFfiAddress(req.Validator)
	if err != nil {
		return nil, fmt.Errorf("failed to convert validator: %w", err)
	}
	validatorArg := iota_sdk_ffi.PtbArgumentAddress(ffiValidator)

	// Call the add stake function
	args := []*iota_sdk_ffi.PtbArgument{stakeTokenArg, validatorArg}
	builder = builder.MoveCall(systemPackageAddr, moduleId, functionId, args, []*iota_sdk_ffi.TypeTag{}, []string{})

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins
	gasBudget := uint64(5000000) // gas budget for withdraw stake
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}
	builder, err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
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
	ffiStakedID, err := toFfiObjectID(stakedRef.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert staked object: %w", err)
	}
	stakedArg := iota_sdk_ffi.PtbArgumentObjectId(ffiStakedID)

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

	// Call the withdraw stake function with staked object
	args := []*iota_sdk_ffi.PtbArgument{stakedArg}
	builder = builder.MoveCall(systemPackageAddr, moduleId, functionId, args, []*iota_sdk_ffi.TypeTag{}, []string{})

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
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
		ffiObjID, err := toFfiObjectID(gasObj.Data.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder = builder.Gas(ffiObjID)
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		builder, err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
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
	ffiCoinID, err := toFfiObjectID(coinRef.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert coin: %w", err)
	}

	// Create amount arguments for splitting
	var amounts []uint64
	var names []string
	for i, amount := range req.SplitAmounts {
		amounts = append(amounts, amount.Uint64())
		names = append(names, fmt.Sprintf("split_%d", i))
	}

	// Split the coin
	builder = builder.SplitCoins(ffiCoinID, amounts, names)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
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
		ffiObjID, err := toFfiObjectID(gasObj.Data.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder = builder.Gas(ffiObjID)
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		builder, err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
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
	ffiCoinID, err := toFfiObjectID(coinRef.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert coin: %w", err)
	}

	// For SplitCoinEqual, we need to get the coin balance and divide by split count
	// This is a simplification - in a real implementation, you'd want to call a Move function
	// that handles equal splitting properly
	splitCount := req.SplitCount.Uint64()
	if splitCount == 0 {
		return nil, fmt.Errorf("split count must be greater than 0")
	}

	// Create count-1 amount arguments (the last piece stays with the original coin)
	var amounts []uint64
	var names []string
	for i := uint64(0); i < splitCount-1; i++ {
		// For now, use a default equal amount (this should be calculated from balance/count)
		amounts = append(amounts, 1000000) // 1 IOTA per split
		names = append(names, fmt.Sprintf("equal_split_%d", i))
	}

	// Split the coin
	builder = builder.SplitCoins(ffiCoinID, amounts, names)

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
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
		ffiObjID, err := toFfiObjectID(gasObj.Data.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder = builder.Gas(ffiObjID)
	} else {
		// Find suitable gas coins
		gasBudget := uint64(1000000) // default gas budget
		if req.GasBudget != nil {
			gasBudget = req.GasBudget.Uint64()
		}
		builder, err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
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
	ffiObjID, err := toFfiObjectID(objRef.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object: %w", err)
	}
	objArg := iota_sdk_ffi.PtbArgumentObjectId(ffiObjID)

	// Create recipient argument
	ffiRecipient, err := toFfiAddress(req.Recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to convert recipient: %w", err)
	}

	// Transfer the object
	builder = builder.TransferObjects(ffiRecipient, []*iota_sdk_ffi.PtbArgument{objArg})

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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
	senderAddr, err := toFfiAddress(req.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Set gas budget if provided
	if req.GasBudget != nil {
		builder = builder.GasBudget(req.GasBudget.Uint64())
	}

	// Add gas coins (auto-selected since TransferIota doesn't specify gas coins)
	gasBudget := uint64(1000000) // default gas budget
	if req.GasBudget != nil {
		gasBudget = req.GasBudget.Uint64()
	}
	builder, err = c.addGasCoinsToBuilder(ctx, builder, req.Signer, gasBudget, 0)
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
	ffiCoinID, err := toFfiObjectID(coinRef.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert coin: %w", err)
	}

	var coinArg *iota_sdk_ffi.PtbArgument
	if req.Amount != nil {
		// If amount is specified, split the coin first
		amountVal := req.Amount.Uint64()
		splitCoinName := "transfer_coin"
		builder = builder.SplitCoins(ffiCoinID, []uint64{amountVal}, []string{splitCoinName})

		// Reference the split result by name
		coinArg = iota_sdk_ffi.PtbArgumentRes(splitCoinName)
	} else {
		// Transfer the whole coin
		coinArg = iota_sdk_ffi.PtbArgumentObjectId(ffiCoinID)
	}

	// Create recipient argument
	ffiRecipient, err := toFfiAddress(req.Recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to convert recipient: %w", err)
	}

	// Transfer the coin
	builder = builder.TransferObjects(ffiRecipient, []*iota_sdk_ffi.PtbArgument{coinArg})

	// Finish the transaction to get bytes
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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

	txData, err := bcs.Unmarshal[iotago.TransactionData](req.TxDataBytes.Data())
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal")
	}
	tx, err := convertTransactionDataToTransaction(&txData)
	if err != nil {
		return nil, fmt.Errorf("can't convert to Transaction: %w", err)
	}
	signedDigest, err := req.Signer.Sign(tx.SigningDigest())
	if err != nil {
		return nil, fmt.Errorf("can't sign digest: %w", err)
	}

	// Convert to FFI signature
	ffiSig, err := iota_sdk_ffi.UserSignatureFromBytes(signedDigest.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to create FFI signature: %w", err)
	}

	txEffects, err := c.qclient.ExecuteTx([]*iota_sdk_ffi.UserSignature{ffiSig}, tx)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("failed to execute tx: %w", err)
	}

	// Build response
	digest, _ := fromFfiDigest(tx.Digest())
	response := &iotajsonrpc.IotaTransactionBlockResponse{
		Digest: *digest,
	}

	// If options request effects or object changes, use effects from execution
	if req.Options != nil && (req.Options.ShowEffects || req.Options.ShowObjectChanges || req.Options.ShowBalanceChanges) {
		if txEffects == nil || *txEffects == nil {
			return nil, fmt.Errorf("transaction effects are nil after successful execution")
		}

		if req.Options.ShowEffects {
			convertedEffects, err := convertTransactionEffects(*txEffects)
			if err != nil {
				return nil, fmt.Errorf("failed to convert transaction effects: %w", err)
			}
			response.Effects = &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{Data: *convertedEffects}
		}
		if req.Options.ShowObjectChanges {
			objectChanges, err := c.convertChangedObjectsToObjectChanges(*txEffects)
			if err != nil {
				return nil, fmt.Errorf("failed to convert object changes: %w", err)
			}
			response.ObjectChanges = objectChanges
		}
		if req.Options.ShowBalanceChanges {
			balanceChanges, err := c.convertTransactionEffectsToBalanceChanges(ctx, *txEffects, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to convert balance changes: %w", err)
			}
			response.BalanceChanges = balanceChanges
		}
	}

	return response, nil
}

// SignAndExecuteTransactionRaw signs and executes a transaction, returning raw FFI types
//
// This method exposes the underlying iota_sdk_ffi.TransactionEffects directly,
// allowing you to access all the raw data from the executed transaction including:
// - All changed objects with their input/output states
// - Raw object data for extracting coin balances
// - Complete transaction effects without conversion overhead
//
// Returns:
// - digest: The transaction digest
// - effects: The raw transaction effects from FFI
// - error: Any error that occurred
//
// This is useful for:
// 1. Accessing raw transaction effects data
// 2. Implementing custom balance change extraction from executed transactions
// 3. Debugging transaction execution results
//
// Example usage:
//
//	digest, rawEffects, err := client.SignAndExecuteTransactionRaw(ctx, req)
//	if err != nil { ... }
//
//	// Access changed objects
//	v1 := rawEffects.AsV1()
//	for _, changedObj := range v1.ChangedObjects {
//	    // changedObj contains input/output state and operation type
//	}
func (c *BindingClient) SignAndExecuteTransactionRaw(ctx context.Context, req *iotaclient.SignAndExecuteTransactionRequest) (*iotago.TransactionDigest, *iota_sdk_ffi.TransactionEffects, error) {
	if req == nil {
		return nil, nil, fmt.Errorf("request is required")
	}
	if req.Signer == nil {
		return nil, nil, fmt.Errorf("signer is required")
	}
	if len(req.TxDataBytes) == 0 {
		return nil, nil, fmt.Errorf("transaction data bytes are required")
	}

	txData, err := bcs.Unmarshal[iotago.TransactionData](req.TxDataBytes.Data())
	if err != nil {
		return nil, nil, fmt.Errorf("can't unmarshal: %w", err)
	}

	tx, err := convertTransactionDataToTransaction(&txData)
	if err != nil {
		return nil, nil, fmt.Errorf("can't convert to Transaction: %w", err)
	}

	signedDigest, err := req.Signer.Sign(tx.SigningDigest())
	if err != nil {
		return nil, nil, fmt.Errorf("can't sign digest: %w", err)
	}

	// Convert to FFI signature
	ffiSig, err := iota_sdk_ffi.UserSignatureFromBytes(signedDigest.Bytes())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create FFI signature: %w", err)
	}

	// Execute transaction
	txEffects, err := c.qclient.ExecuteTx([]*iota_sdk_ffi.UserSignature{ffiSig}, tx)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, nil, fmt.Errorf("failed to execute tx: %w", err)
	}

	if txEffects == nil || *txEffects == nil {
		return nil, nil, fmt.Errorf("transaction effects are nil after execution")
	}

	digest, _ := fromFfiDigest(tx.Digest())
	return digest, *txEffects, nil
}

func (c *BindingClient) PublishContract(ctx context.Context, signer iotasigner.Signer, modules []*iotago.Base64Data, dependencies []*iotago.Address, gasBudget uint64, options *iotajsonrpc.IotaTransactionBlockResponseOptions) (*iotajsonrpc.IotaTransactionBlockResponse, *iotago.PackageID, error) {
	if signer == nil {
		return nil, nil, fmt.Errorf("signer is required")
	}
	if len(modules) == 0 {
		return nil, nil, fmt.Errorf("modules are required")
	}

	senderAddr, err := toFfiAddress(signer.Address())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Add gas coins
	builder, err = c.addGasCoinsToBuilder(ctx, builder, signer.Address(), gasBudget, 0)
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

	// Add publish command with upgrade capability name
	upgradeCapName := "upgrade_cap"
	builder = builder.Publish(moduleBytes, ffiDeps, upgradeCapName)

	// Transfer upgrade capability to sender (as per requirements)
	ffiSender, err := toFfiAddress(signer.Address())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert sender: %w", err)
	}
	// Get the upgrade capability result by name
	upgradeCapArg := iota_sdk_ffi.PtbArgumentRes(upgradeCapName)
	builder = builder.TransferObjects(ffiSender, []*iota_sdk_ffi.PtbArgument{upgradeCapArg})

	// Build, sign, and execute (request ObjectChanges to extract package ID)
	if options == nil {
		options = &iotajsonrpc.IotaTransactionBlockResponseOptions{}
	}
	options.ShowObjectChanges = true

	response, err := c.buildTransactionAndExecute(ctx, signer, builder, gasBudget, 0, options)
	if err != nil {
		return nil, nil, err
	}

	// Extract package ID from transaction response
	packageID, err := response.GetPublishedPackageID()
	if err != nil {
		return response, nil, fmt.Errorf("failed to extract published package ID: %w", err)
	}

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

	senderAddr, err := toFfiAddress(signer.Address())
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	builder := iota_sdk_ffi.TransactionBuilderInit(senderAddr, c.qclient)

	// Add gas coins
	gasBudget := uint64(5000000) // higher gas budget for mint
	builder, err = c.addGasCoinsToBuilder(ctx, builder, signer.Address(), gasBudget, 0)
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

	// Add treasury cap as argument
	ffiTreasuryCapID, err := toFfiObjectID(treasuryCap.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert treasury cap: %w", err)
	}
	treasuryCapArg := iota_sdk_ffi.PtbArgumentObjectId(ffiTreasuryCapID)

	// Add mint amount as argument
	amountArg := iota_sdk_ffi.PtbArgumentU64(mintAmount)

	// Add recipient (sender) as argument
	ffiRecipient, err := toFfiAddress(signer.Address())
	if err != nil {
		return nil, fmt.Errorf("failed to convert recipient: %w", err)
	}
	recipientArg := iota_sdk_ffi.PtbArgumentAddress(ffiRecipient)

	// Call the mint function
	args := []*iota_sdk_ffi.PtbArgument{treasuryCapArg, amountArg, recipientArg}
	builder = builder.MoveCall(packageAddr, moduleId, mintFuncId, args, []*iota_sdk_ffi.TypeTag{}, []string{})

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
		resp, err := mapFfiObjectToIotaResponse(&obj, options, nil)
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

	// Get coin count
	coins, err := c.qclient.Coins(addr, nil, nil)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("Failed to get coins: %v", err)
	}

	coinCount := 0
	if coins.Data != nil {
		coinCount = len(coins.Data)
	}

	// Convert to response format
	var result []*iotajsonrpc.Balance
	if balance != nil {
		balanceResp := &iotajsonrpc.Balance{
			CoinType:        iotajsonrpc.CoinType(coinType),
			TotalBalance:    iotajsonrpc.NewBigInt(*balance),
			CoinObjectCount: iotajsonrpc.NewBigInt(uint64(coinCount)),
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
	var coinType *string
	if req.CoinType != nil {
		tmp := fmt.Sprintf("0x2::coin::Coin<%s>", *req.CoinType)
		coinType = &tmp
	}

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

	// Fetch BCS data if requested
	var bcsBytes *[]byte
	if req.Options != nil && req.Options.ShowBcs {
		// Use MoveObjectContentsBcs to get just the Move object contents (not the full object wrapper)
		// This matches what the JSON-RPC API returns for showBcs and allows direct deserialization
		bcs, err := c.qclient.MoveObjectContentsBcs(oid, nil)
		if err.(*iota_sdk_ffi.SdkFfiError) == nil && bcs != nil {
			bcsBytes = bcs
		}
		// Silently ignore BCS fetch errors - BCS data might not be available for all objects
	}

	return mapFfiObjectToIotaResponse(obj, req.Options, bcsBytes)
}

func (c *BindingClient) GetProtocolConfig(ctx context.Context, version *iotajsonrpc.BigInt) (*iotajsonrpc.ProtocolConfig, error) {
	// Convert version parameter
	var versionUint64 *uint64
	if version != nil {
		v := version.Uint64()
		versionUint64 = &v
	}

	// Call GraphQL client's ProtocolConfig method
	protocolConfigs, err := c.qclient.ProtocolConfig(versionUint64)
	if err != nil {
		return nil, fmt.Errorf("GraphQL ProtocolConfig failed: %w", err)
	}

	if protocolConfigs == nil {
		return nil, fmt.Errorf("no protocol config found")
	}

	// Convert back to iota-go response format
	response := &iotajsonrpc.ProtocolConfig{
		ProtocolVersion: iotajsonrpc.NewBigInt(protocolConfigs.ProtocolVersion),
	}

	// Convert feature flags from slice to map
	if len(protocolConfigs.FeatureFlags) > 0 {
		response.FeatureFlags = make(map[string]bool)
		for _, flag := range protocolConfigs.FeatureFlags {
			response.FeatureFlags[flag.Key] = flag.Value
		}
	}

	// Convert config attributes from slice to map
	if len(protocolConfigs.Configs) > 0 {
		response.Attributes = make(map[string]iotajsonrpc.ProtocolConfigValue)
		for _, attr := range protocolConfigs.Configs {
			if attr.Value != nil {
				if configValue := parseProtocolConfigValue(*attr.Value); configValue != nil {
					response.Attributes[attr.Key] = *configValue
				}
			}
		}
	}

	return response, nil
}

// parseProtocolConfigValue attempts to parse a string value into the appropriate numeric type
func parseProtocolConfigValue(value string) *iotajsonrpc.ProtocolConfigValue {
	// Try parsing as uint64
	if u64, err := strconv.ParseUint(value, 10, 64); err == nil {
		// Check if it fits in smaller types
		if u64 <= math.MaxUint16 {
			u16 := uint16(u64)
			return &iotajsonrpc.ProtocolConfigValue{U16: &u16}
		}
		if u64 <= math.MaxUint32 {
			u32 := uint32(u64)
			return &iotajsonrpc.ProtocolConfigValue{U32: &u32}
		}
		return &iotajsonrpc.ProtocolConfigValue{U64: &u64}
	}

	// Try parsing as float64
	if f64, err := strconv.ParseFloat(value, 64); err == nil {
		return &iotajsonrpc.ProtocolConfigValue{F64: &f64}
	}

	// If parsing fails, return nil
	return nil
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
	if req.Digest == nil {
		return nil, fmt.Errorf("digest is required")
	}

	digest, err := iota_sdk_ffi.DigestFromBase58(req.Digest.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert digest: %w", err)
	}

	// Call GraphQL client's Transaction method
	_, err = c.qclient.Transaction(digest)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("GraphQL Transaction failed: %w", err)
	}

	// Convert back to iota-go response format
	// Note: Mapping SignedTransaction to IotaTransactionBlockResponse requires detailed
	// field-by-field conversion. The transaction is available but comprehensive mapping
	// should be implemented based on which fields are actually needed by the application.
	// For now, return minimal response with digest populated.
	response := &iotajsonrpc.IotaTransactionBlockResponse{
		Digest: *req.Digest,
	}

	return response, nil
}

func (c *BindingClient) Transaction(digest *iota_sdk_ffi.Digest) (*iota_sdk_ffi.SignedTransaction, error) {
	signedTx, err := c.qclient.Transaction(digest)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return signedTx, nil
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
		resp, err := mapFfiObjectToIotaResponse(&obj, req.Options, nil)
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
			// Note: Mapping FFI TransactionEffects to IotaTransactionBlockEffects
			// requires detailed field conversion. Can be implemented based on
			// application needs. Effects are available but not mapped.
			_ = effects
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
	faucetURL := iotaconn.FaucetURL(c.RpcURL)

	return iotaclient.RequestFundsFromFaucet(ctx, address.AsIotaAddress(), faucetURL)
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

func (c *BindingClient) L2() L2Client {
	return NewBindingClientL2(c.RpcURL, c)
}

func (c *BindingClient) IotaClient() L1Client {
	return c
}

func (c *BindingClient) DeployISCContracts(ctx context.Context, signer iotasigner.Signer) (iotago.PackageID, error) {
	iscBytecode := contracts.ISC()
	txnBytes, err := c.Publish(ctx, iotaclient.PublishRequest{
		Sender:          signer.Address(),
		CompiledModules: iscBytecode.Modules,
		Dependencies:    iscBytecode.Dependencies,
		GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 10),
	})
	if err != nil {
		return iotago.PackageID{}, err
	}

	txnResponse, err := c.SignAndExecuteTransaction(
		ctx,
		&iotaclient.SignAndExecuteTransactionRequest{
			TxDataBytes: txnBytes.TxBytes,
			Signer:      signer,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	if err != nil {
		return iotago.PackageID{}, err
	}

	if !txnResponse.Effects.Data.IsSuccess() {
		return iotago.PackageID{}, errors.New("publish ISC contracts failed")
	}
	packageID, err := txnResponse.GetPublishedPackageID()
	if err != nil {
		return iotago.PackageID{}, err
	}
	return *packageID, nil
}

func (c *BindingClient) FindCoinsForGasPayment(ctx context.Context, owner *iotago.Address, pt iotago.ProgrammableTransaction, gasPrice uint64, gasBudget uint64) ([]*iotago.ObjectRef, error) {
	if owner == nil {
		return nil, nil
	}

	// Get IOTA coins for the owner with retry logic
	// Coins may not be immediately available after faucet request or previous transactions
	// due to indexing delays in the GraphQL endpoint
	coinType := iotajsonrpc.IotaCoinType.String()
	var coinPage *iotajsonrpc.CoinPage
	var err error

	maxRetries := 30
	var selectedCoins []*iotajsonrpc.Coin
	for i := 0; i < maxRetries; i++ {
		coinPage, err = c.GetCoins(ctx, iotaclient.GetCoinsRequest{
			Owner:    owner,
			CoinType: &coinType,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get coins: %w", err)
		}

		// Filter out coins that are already used as inputs in the transaction
		// and try to select usable gas coins
		if len(coinPage.Data) > 0 {
			selectedCoins, err = iotajsonrpc.PickupCoinsWithFilter(coinPage.Data, gasBudget, func(c *iotajsonrpc.Coin) bool {
				return !pt.IsInInputObjects(c.CoinObjectID)
			})
			// If we successfully found usable coins, break out of retry loop
			if err == nil && len(selectedCoins) > 0 {
				break
			}
		}

		// Wait before retrying (exponential backoff up to 1 second)
		if i < maxRetries-1 {
			waitTime := time.Duration(100*(i+1)) * time.Millisecond
			if waitTime > time.Second {
				waitTime = time.Second
			}
			time.Sleep(waitTime)
		}
	}

	// Final check: if still no usable coins after retries
	if len(selectedCoins) == 0 {
		return nil, fmt.Errorf("no coins found for address %s after %d retries", owner.String(), maxRetries)
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
	var err error
	var txnBytes []byte
	var txnResponse *iotajsonrpc.IotaTransactionBlockResponse
	var gasPayments []*iotago.ObjectRef
	for i := 0; i < 5; i++ {
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
			ctx, &iotaclient.SignAndExecuteTransactionRequest{
				TxDataBytes: txnBytes,
				Signer:      signer,
				Options:     options,
			},
		)
		if err == nil {
			return txnResponse, nil
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("can't execute the transaction in time: %w", err)
}

func (c *BindingClient) WaitForNextVersionForTesting(ctx context.Context, timeout time.Duration, logger log.Logger, currentRef *iotago.ObjectRef, cb func()) (*iotago.ObjectRef, error) {
	// Some 'sugar' to make dynamic refs handling easier (where refs can be nil or set depending on state)
	if currentRef == nil {
		cb()
		return currentRef, nil
	}

	cb()

	// Create a ticker for polling
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	// Add timeout to context if not already set
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("WaitForNextVersionForTesting: context deadline exceeded while waiting for object version change: %v", currentRef)
		case <-ticker.C:
			// Poll for object update
			newRef, err := c.GetObject(ctx, iotaclient.GetObjectRequest{ObjectID: currentRef.ObjectID})
			if err != nil {
				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: error getting object: %v, retrying...", err)
				}
				continue
			}

			if newRef.Error != nil {
				// The provided object got consumed and is gone. We can return.
				if newRef.Error.Data.Deleted != nil || newRef.Error.Data.NotExists != nil {
					return currentRef, nil
				}

				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: object error: %v, retrying...", newRef.Error)
				}
				continue
			}

			if newRef.Data.Ref().Version > currentRef.Version {
				if logger != nil {
					logger.LogInfof("WaitForNextVersionForTesting: Found the updated version of %v, which is: %v", currentRef, newRef.Data.Ref())
				}

				ref := newRef.Data.Ref()
				return &ref, nil
			}

			if logger != nil {
				logger.LogInfof("WaitForNextVersionForTesting: Getting the same version ref as before. Retrying. %v", currentRef)
			}
		}
	}
}

// Transaction builder helper functions

// convertObjectRefToUnresolvedInput converts an iotago.ObjectRef to FFI UnresolvedInput
// NOTE: Commented out because UnresolvedInput is no longer available in the new FFI API
// func convertObjectRefToUnresolvedInput(ref *iotago.ObjectRef) (*iota_sdk_ffi.UnresolvedInput, error) {
// 	if ref == nil {
// 		return nil, fmt.Errorf("ObjectRef is nil")
// 	}
//
// 	ffiObjID, err := toFfiObjectID(ref.ObjectID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to convert ObjectID: %w", err)
// 	}
//
// 	ffiDigest, err := iota_sdk_ffi.DigestFromBase58(ref.Digest.String())
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to convert digest: %w", err)
// 	}
//
// 	return iota_sdk_ffi.UnresolvedInputNewOwned(ffiObjID, ref.Version, ffiDigest), nil
// }

// buildTransactionAndExecute builds, signs, and executes a transaction
func (c *BindingClient) buildTransactionAndExecute(_ context.Context, signer iotasigner.Signer, builder *iota_sdk_ffi.TransactionBuilder, gasBudget uint64, gasPrice uint64, options *iotajsonrpc.IotaTransactionBlockResponseOptions) (*iotajsonrpc.IotaTransactionBlockResponse, error) {
	// Set gas budget and price
	builder = builder.GasBudget(gasBudget)
	if gasPrice > 0 {
		builder = builder.GasPrice(gasPrice)
	}

	// Note: Sender is already set during TransactionBuilderInit

	// Finish the transaction
	tx, err := builder.Finish()
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
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
	txEffects, err := c.qclient.ExecuteTx([]*iota_sdk_ffi.UserSignature{ffiSig}, tx)
	if err.(*iota_sdk_ffi.SdkFfiError) != nil {
		return nil, fmt.Errorf("ExecuteTx failed: %w", err)
	}

	// Build response
	digest, _ := fromFfiDigest(tx.Digest())
	response := &iotajsonrpc.IotaTransactionBlockResponse{
		Digest: *digest,
	}

	// If options request effects or object changes, use effects from execution
	if options != nil && (options.ShowEffects || options.ShowObjectChanges) {
		if txEffects == nil || *txEffects == nil {
			return nil, fmt.Errorf("transaction effects are nil after successful execution")
		}

		if options.ShowEffects {
			convertedEffects, err := convertTransactionEffects(*txEffects)
			if err != nil {
				return nil, fmt.Errorf("failed to convert transaction effects: %w", err)
			}
			response.Effects = &serialization.TagJson[iotajsonrpc.IotaTransactionBlockEffects]{Data: *convertedEffects}
		}
		if options.ShowObjectChanges {
			objectChanges, err := c.convertChangedObjectsToObjectChanges(*txEffects)
			if err != nil {
				return nil, fmt.Errorf("failed to convert object changes: %w", err)
			}
			response.ObjectChanges = objectChanges
		}
	}

	return response, nil
}

// addGasCoinsToBuilder adds gas coins to the transaction builder
func (c *BindingClient) addGasCoinsToBuilder(ctx context.Context, builder *iota_sdk_ffi.TransactionBuilder, owner *iotago.Address, gasBudget, gasPrice uint64) (*iota_sdk_ffi.TransactionBuilder, error) {
	// Find gas coins
	gasCoins, err := c.FindCoinsForGasPayment(ctx, owner, iotago.ProgrammableTransaction{}, gasPrice, gasBudget)
	if err != nil {
		return nil, fmt.Errorf("failed to find gas coins: %w", err)
	}

	// Add gas objects to builder using the new fluent API
	for _, coin := range gasCoins {
		ffiObjID, err := toFfiObjectID(coin.ObjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gas coin: %w", err)
		}
		builder = builder.Gas(ffiObjID)
	}
	return builder, nil
}

// convertTransactionDataToTransaction converts iotago.TransactionData to iota_sdk_ffi.Transaction
// This is useful when you have a TransactionData structure and need to create an FFI Transaction
func convertTransactionDataToTransaction(td *iotago.TransactionData) (*iota_sdk_ffi.Transaction, error) {
	if td == nil {
		return nil, fmt.Errorf("TransactionData is nil")
	}
	if td.V1 == nil {
		return nil, fmt.Errorf("TransactionData.V1 is nil")
	}

	v1 := td.V1

	// Convert TransactionKind
	kind, err := convertTransactionKind(&v1.Kind)
	if err != nil {
		return nil, fmt.Errorf("failed to convert TransactionKind: %w", err)
	}

	// Convert Sender
	sender, err := toFfiAddress(&v1.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sender address: %w", err)
	}

	// Convert GasData to GasPayment
	gasPayment, err := convertGasDataToGasPayment(&v1.GasData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert GasData: %w", err)
	}

	// Convert Expiration
	expiration, err := convertTransactionExpiration(&v1.Expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to convert expiration: %w", err)
	}

	return iota_sdk_ffi.NewTransaction(kind, sender, gasPayment, expiration), nil
}

// convertGasDataToGasPayment converts iotago.GasData to iota_sdk_ffi.GasPayment
func convertGasDataToGasPayment(gasData *iotago.GasData) (iota_sdk_ffi.GasPayment, error) {
	if gasData == nil {
		return iota_sdk_ffi.GasPayment{}, fmt.Errorf("GasData is nil")
	}

	// Convert payment ObjectRefs to ObjectReferences
	objects := make([]iota_sdk_ffi.ObjectReference, len(gasData.Payment))
	for i, ref := range gasData.Payment {
		objRef, err := convertObjectRefToObjectReference(ref)
		if err != nil {
			return iota_sdk_ffi.GasPayment{}, fmt.Errorf("failed to convert payment object %d: %w", i, err)
		}
		objects[i] = objRef
	}

	// Convert owner address
	owner, err := toFfiAddress(gasData.Owner)
	if err != nil {
		return iota_sdk_ffi.GasPayment{}, fmt.Errorf("failed to convert owner address: %w", err)
	}

	return iota_sdk_ffi.GasPayment{
		Objects: objects,
		Owner:   owner,
		Price:   gasData.Price,
		Budget:  gasData.Budget,
	}, nil
}

// convertObjectRefToObjectReference converts iotago.ObjectRef to iota_sdk_ffi.ObjectReference
func convertObjectRefToObjectReference(ref *iotago.ObjectRef) (iota_sdk_ffi.ObjectReference, error) {
	if ref == nil {
		return iota_sdk_ffi.ObjectReference{}, fmt.Errorf("ObjectRef is nil")
	}

	ffiObjID, err := toFfiObjectID(ref.ObjectID)
	if err != nil {
		return iota_sdk_ffi.ObjectReference{}, fmt.Errorf("failed to convert ObjectID: %w", err)
	}

	ffiDigest, err := iota_sdk_ffi.DigestFromBase58(ref.Digest.String())
	if err != nil {
		return iota_sdk_ffi.ObjectReference{}, fmt.Errorf("failed to convert digest: %w", err)
	}

	return iota_sdk_ffi.ObjectReference{
		ObjectId: ffiObjID,
		Version:  uint64(ref.Version),
		Digest:   ffiDigest,
	}, nil
}

// convertTransactionExpiration converts iotago.TransactionExpiration to iota_sdk_ffi.TransactionExpiration
func convertTransactionExpiration(exp *iotago.TransactionExpiration) (iota_sdk_ffi.TransactionExpiration, error) {
	if exp == nil {
		return iota_sdk_ffi.TransactionExpirationNone{}, nil
	}

	if exp.None != nil {
		return iota_sdk_ffi.TransactionExpirationNone{}, nil
	}

	if exp.Epoch != nil {
		return iota_sdk_ffi.TransactionExpirationEpoch{
			Field0: *exp.Epoch,
		}, nil
	}

	return iota_sdk_ffi.TransactionExpirationNone{}, nil
}

// convertTransactionKind converts iotago.TransactionKind to iota_sdk_ffi.TransactionKind
func convertTransactionKind(kind *iotago.TransactionKind) (*iota_sdk_ffi.TransactionKind, error) {
	if kind == nil {
		return nil, fmt.Errorf("TransactionKind is nil")
	}

	if kind.ProgrammableTransaction != nil {
		pt, err := convertProgrammableTransaction(kind.ProgrammableTransaction)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ProgrammableTransaction: %w", err)
		}
		return iota_sdk_ffi.TransactionKindNewProgrammableTransaction(pt), nil
	}

	if kind.ChangeEpoch != nil {
		// Note: ChangeEpoch conversion would require additional implementation
		return nil, fmt.Errorf("ChangeEpoch conversion not implemented")
	}

	if kind.Genesis != nil {
		// Note: Genesis conversion would require additional implementation
		return nil, fmt.Errorf("Genesis conversion not implemented")
	}

	if kind.ConsensusCommitPrologue != nil {
		// Note: ConsensusCommitPrologue conversion would require additional implementation
		return nil, fmt.Errorf("ConsensusCommitPrologue conversion not implemented")
	}

	return nil, fmt.Errorf("unknown TransactionKind variant")
}

// convertProgrammableTransaction converts iotago.ProgrammableTransaction to iota_sdk_ffi.ProgrammableTransaction
func convertProgrammableTransaction(pt *iotago.ProgrammableTransaction) (*iota_sdk_ffi.ProgrammableTransaction, error) {
	if pt == nil {
		return nil, fmt.Errorf("ProgrammableTransaction is nil")
	}

	// Convert Inputs
	inputs := make([]*iota_sdk_ffi.Input, len(pt.Inputs))
	for i, input := range pt.Inputs {
		ffiInput, err := convertCallArg(&input)
		if err != nil {
			return nil, fmt.Errorf("failed to convert input %d: %w", i, err)
		}
		inputs[i] = ffiInput
	}

	// Convert Commands
	commands := make([]*iota_sdk_ffi.Command, len(pt.Commands))
	for i, cmd := range pt.Commands {
		ffiCmd, err := convertCommand(&cmd)
		if err != nil {
			return nil, fmt.Errorf("failed to convert command %d: %w", i, err)
		}
		commands[i] = ffiCmd
	}

	return iota_sdk_ffi.NewProgrammableTransaction(inputs, commands), nil
}

// convertCallArg converts iotago.CallArg to iota_sdk_ffi.Input
func convertCallArg(arg *iotago.CallArg) (*iota_sdk_ffi.Input, error) {
	if arg == nil {
		return nil, fmt.Errorf("CallArg is nil")
	}

	// Handle Pure variant
	if arg.Pure != nil {
		return iota_sdk_ffi.InputNewPure(*arg.Pure), nil
	}

	// Handle Object variant
	if arg.Object != nil {
		return convertObjectArgToInput(arg.Object)
	}

	return nil, fmt.Errorf("unknown CallArg variant")
}

// convertObjectArgToInput converts iotago.ObjectArg to iota_sdk_ffi.Input
func convertObjectArgToInput(obj *iotago.ObjectArg) (*iota_sdk_ffi.Input, error) {
	if obj == nil {
		return nil, fmt.Errorf("ObjectArg is nil")
	}

	if obj.ImmOrOwnedObject != nil {
		objRef, err := convertObjectRefToObjectReference(obj.ImmOrOwnedObject)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ImmOrOwnedObject: %w", err)
		}
		return iota_sdk_ffi.InputNewImmutableOrOwned(objRef), nil
	}

	if obj.SharedObject != nil {
		objID, err := toFfiObjectID(obj.SharedObject.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to convert SharedObject Id: %w", err)
		}
		return iota_sdk_ffi.InputNewShared(objID, uint64(obj.SharedObject.InitialSharedVersion), obj.SharedObject.Mutable), nil
	}

	if obj.Receiving != nil {
		objRef, err := convertObjectRefToObjectReference(obj.Receiving)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Receiving: %w", err)
		}
		return iota_sdk_ffi.InputNewReceiving(objRef), nil
	}

	return nil, fmt.Errorf("unknown ObjectArg variant")
}

// convertCommand converts iotago.Command to iota_sdk_ffi.Command
func convertCommand(cmd *iotago.Command) (*iota_sdk_ffi.Command, error) {
	if cmd == nil {
		return nil, fmt.Errorf("Command is nil")
	}

	// Note: Full command conversion would require implementing all command variants.
	// This is a placeholder that shows the pattern. Add cases as needed:

	if cmd.Publish != nil {
		// Convert Publish
		return convertPublishCommand(cmd.Publish)
	}

	if cmd.MoveCall != nil {
		// Convert MoveCall
		return convertMoveCallCommand(cmd.MoveCall)
	}

	if cmd.TransferObjects != nil {
		// Convert TransferObjects
		return convertTransferObjectsCommand(cmd.TransferObjects)
	}

	if cmd.SplitCoins != nil {
		// Convert SplitCoins
		return convertSplitCoinsCommand(cmd.SplitCoins)
	}

	if cmd.MergeCoins != nil {
		// Convert MergeCoins
		return convertMergeCoinsCommand(cmd.MergeCoins)
	}

	// Add other command types as needed (MakeMoveVec, Publish, Upgrade, etc.)
	return nil, fmt.Errorf("command conversion not fully implemented for this command type")
}

// convertPublishCommand converts iotago.ProgrammablePublish to iota_sdk_ffi.Command
func convertPublishCommand(pub *iotago.ProgrammablePublish) (*iota_sdk_ffi.Command, error) {
	if pub == nil {
		return nil, fmt.Errorf("ProgrammablePublish is nil")
	}

	// Convert dependencies
	dependencies := make([]*iota_sdk_ffi.ObjectId, len(pub.Dependencies))
	for i, dep := range pub.Dependencies {
		ffiDep, err := toFfiObjectID(dep)
		if err != nil {
			return nil, fmt.Errorf("failed to convert dependency %d: %w", i, err)
		}
		dependencies[i] = ffiDep
	}

	publish := iota_sdk_ffi.NewPublish(pub.Modules, dependencies)
	return iota_sdk_ffi.CommandNewPublish(publish), nil
}

// convertMoveCallCommand converts iotago.ProgrammableMoveCall to iota_sdk_ffi.Command
func convertMoveCallCommand(mc *iotago.ProgrammableMoveCall) (*iota_sdk_ffi.Command, error) {
	if mc == nil {
		return nil, fmt.Errorf("ProgrammableMoveCall is nil")
	}

	packageID, err := toFfiObjectID(mc.Package)
	if err != nil {
		return nil, fmt.Errorf("failed to convert package ID: %w", err)
	}

	module, err := iota_sdk_ffi.NewIdentifier(mc.Module)
	if err != nil {
		return nil, fmt.Errorf("failed to convert module identifier: %w", err)
	}

	function, err := iota_sdk_ffi.NewIdentifier(mc.Function)
	if err != nil {
		return nil, fmt.Errorf("failed to convert function identifier: %w", err)
	}

	typeArgs := make([]*iota_sdk_ffi.TypeTag, len(mc.TypeArguments))
	for i, typeArg := range mc.TypeArguments {
		ffiTypeTag, err := toFfiTypeTag(&typeArg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert type argument %d: %w", i, err)
		}
		typeArgs[i] = ffiTypeTag
	}

	arguments := make([]*iota_sdk_ffi.Argument, len(mc.Arguments))
	for i, arg := range mc.Arguments {
		ffiArg, err := convertArgument(&arg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert argument %d: %w", i, err)
		}
		arguments[i] = ffiArg
	}

	moveCall := iota_sdk_ffi.NewMoveCall(packageID, module, function, typeArgs, arguments)
	return iota_sdk_ffi.CommandNewMoveCall(moveCall), nil
}

// convertArgument converts iotago.Argument to iota_sdk_ffi.Argument
func convertArgument(arg *iotago.Argument) (*iota_sdk_ffi.Argument, error) {
	if arg == nil {
		return nil, fmt.Errorf("Argument is nil")
	}

	if arg.GasCoin != nil {
		return iota_sdk_ffi.ArgumentNewGas(), nil
	}

	if arg.Input != nil {
		return iota_sdk_ffi.ArgumentNewInput(uint16(*arg.Input)), nil
	}

	if arg.Result != nil {
		return iota_sdk_ffi.ArgumentNewResult(uint16(*arg.Result)), nil
	}

	if arg.NestedResult != nil {
		return iota_sdk_ffi.ArgumentNewNestedResult(arg.NestedResult.Cmd, arg.NestedResult.Result), nil
	}

	return nil, fmt.Errorf("unknown Argument variant")
}

// convertTransferObjectsCommand converts iotago.ProgrammableTransferObjects to iota_sdk_ffi.Command
func convertTransferObjectsCommand(to *iotago.ProgrammableTransferObjects) (*iota_sdk_ffi.Command, error) {
	if to == nil {
		return nil, fmt.Errorf("ProgrammableTransferObjects is nil")
	}

	objects := make([]*iota_sdk_ffi.Argument, len(to.Objects))
	for i, obj := range to.Objects {
		ffiArg, err := convertArgument(&obj)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object %d: %w", i, err)
		}
		objects[i] = ffiArg
	}

	address, err := convertArgument(&to.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to convert address: %w", err)
	}

	transferObjects := iota_sdk_ffi.NewTransferObjects(objects, address)
	return iota_sdk_ffi.CommandNewTransferObjects(transferObjects), nil
}

// convertSplitCoinsCommand converts iotago.ProgrammableSplitCoins to iota_sdk_ffi.Command
func convertSplitCoinsCommand(sc *iotago.ProgrammableSplitCoins) (*iota_sdk_ffi.Command, error) {
	if sc == nil {
		return nil, fmt.Errorf("ProgrammableSplitCoins is nil")
	}

	coin, err := convertArgument(&sc.Coin)
	if err != nil {
		return nil, fmt.Errorf("failed to convert coin: %w", err)
	}

	amounts := make([]*iota_sdk_ffi.Argument, len(sc.Amounts))
	for i, amount := range sc.Amounts {
		ffiArg, err := convertArgument(&amount)
		if err != nil {
			return nil, fmt.Errorf("failed to convert amount %d: %w", i, err)
		}
		amounts[i] = ffiArg
	}

	splitCoins := iota_sdk_ffi.NewSplitCoins(coin, amounts)
	return iota_sdk_ffi.CommandNewSplitCoins(splitCoins), nil
}

// convertMergeCoinsCommand converts iotago.ProgrammableMergeCoins to iota_sdk_ffi.Command
func convertMergeCoinsCommand(mc *iotago.ProgrammableMergeCoins) (*iota_sdk_ffi.Command, error) {
	if mc == nil {
		return nil, fmt.Errorf("ProgrammableMergeCoins is nil")
	}

	destination, err := convertArgument(&mc.Destination)
	if err != nil {
		return nil, fmt.Errorf("failed to convert destination: %w", err)
	}

	sources := make([]*iota_sdk_ffi.Argument, len(mc.Sources))
	for i, source := range mc.Sources {
		ffiArg, err := convertArgument(&source)
		if err != nil {
			return nil, fmt.Errorf("failed to convert source %d: %w", i, err)
		}
		sources[i] = ffiArg
	}

	mergeCoins := iota_sdk_ffi.NewMergeCoins(destination, sources)
	return iota_sdk_ffi.CommandNewMergeCoins(mergeCoins), nil
}
