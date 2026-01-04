package iotagraphql

import (
	"encoding/base64"
	"fmt"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

const (
	ExecutionStatusSuccess = "success"
	ExecutionStatusFailure = "failure"
)

func convertExecuteTransactionBlockToGraphQL(
	resp *ExecuteTransactionBlockResponse,
	options *iotajsonrpc.IotaTransactionBlockResponseOptions,
) (*graphqltypes.TransactionResponse, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}

	if len(resp.ExecuteTransactionBlock.Errors) > 0 {
		return &graphqltypes.TransactionResponse{
			Errors: resp.ExecuteTransactionBlock.Errors,
		}, nil
	}

	txBlock := &resp.ExecuteTransactionBlock.Effects.TransactionBlock.RPC_TRANSACTION_FIELDS

	digest, err := iotago.NewDigest(txBlock.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction digest: %w", err)
	}

	// Convert signatures to [][]byte
	sigs := make([][]byte, len(txBlock.Signatures))
	for i, sig := range txBlock.Signatures {
		sigs[i] = sig.Data()
	}

	txResp := &graphqltypes.TransactionResponse{
		Digest:     *digest,
		Sender:     txBlock.Sender.Address,
		Signatures: sigs,
		Errors:     resp.ExecuteTransactionBlock.Errors,
	}

	// Decode effects from BCS
	if len(txBlock.Effects.Bcs) > 0 {
		effects, err := decodeEffectsFromBCS(txBlock.Effects.Bcs.Data())
		if err == nil {
			txResp.Effects = effects
		}
	}

	// Create minimal Effects if not decoded
	if txResp.Effects == nil {
		txResp.Effects = &graphqltypes.TransactionEffects{
			Status: graphqltypes.ExecutionStatus{
				Success: true,
			},
		}
	}

	// Convert object changes from GraphQL response
	if options != nil && options.ShowObjectChanges {
		txResp.ObjectChanges = convertObjectChanges(resp.ExecuteTransactionBlock.Effects.ObjectChanges.Nodes)
	}

	// Convert balance changes from GraphQL response
	if options != nil && options.ShowBalanceChanges {
		txResp.BalanceChanges = convertBalanceChanges(resp.ExecuteTransactionBlock.Effects.BalanceChanges.Nodes)
	}

	return txResp, nil
}

// decodeEffectsFromBCS decodes BCS effects data to graphqltypes.TransactionEffects
func decodeEffectsFromBCS(bcsData []byte) (*graphqltypes.TransactionEffects, error) {
	if len(bcsData) == 0 {
		return nil, nil
	}

	// Try to decode base64 if needed
	data := bcsData
	if decoded, err := base64.StdEncoding.DecodeString(string(bcsData)); err == nil {
		data = decoded
	}

	// Decode BCS to iotajsonrpc type first
	rawEffects, err := bcs.Unmarshal[iotajsonrpc.IotaTransactionBlockEffects](data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal BCS: %w", err)
	}

	if rawEffects.V1 == nil {
		return nil, fmt.Errorf("effects.V1 is nil")
	}
	v1 := rawEffects.V1

	effects := &graphqltypes.TransactionEffects{
		Status: graphqltypes.ExecutionStatus{
			Success: v1.Status.Status == ExecutionStatusSuccess,
			Error:   v1.Status.Error,
		},
		ExecutedEpoch:     v1.ExecutedEpoch.Uint64(),
		TransactionDigest: v1.TransactionDigest,
		GasUsed: graphqltypes.GasCostSummary{
			ComputationCost:         v1.GasUsed.ComputationCost.Uint64(),
			StorageCost:             v1.GasUsed.StorageCost.Uint64(),
			StorageRebate:           v1.GasUsed.StorageRebate.Uint64(),
			NonRefundableStorageFee: v1.GasUsed.NonRefundableStorageFee.Uint64(),
		},
	}

	// Convert owned refs
	if len(v1.Created) > 0 {
		effects.Created = make([]graphqltypes.OwnedObjectRef, len(v1.Created))
		for i, ref := range v1.Created {
			owner, err := convertOwnerFromBCS(ref.Owner)
			if err != nil {
				return nil, err
			}
			effects.Created[i] = graphqltypes.OwnedObjectRef{
				ObjectID: *ref.Reference.ObjectID,
				Version:  ref.Reference.Version,
				Digest:   ref.Reference.Digest,
				Owner:    *owner,
			}
		}
	}

	if len(v1.Mutated) > 0 {
		effects.Mutated = make([]graphqltypes.OwnedObjectRef, len(v1.Mutated))
		for i, ref := range v1.Mutated {
			owner, err := convertOwnerFromBCS(ref.Owner)
			if err != nil {
				return nil, err
			}
			effects.Mutated[i] = graphqltypes.OwnedObjectRef{
				ObjectID: *ref.Reference.ObjectID,
				Version:  ref.Reference.Version,
				Digest:   ref.Reference.Digest,
				Owner:    *owner,
			}
		}
	}

	// Convert gas object
	gasOwner, err := convertOwnerFromBCS(v1.GasObject.Owner)
	if err != nil {
		return nil, err
	}
	effects.GasObject = graphqltypes.OwnedObjectRef{
		ObjectID: *v1.GasObject.Reference.ObjectID,
		Version:  v1.GasObject.Reference.Version,
		Digest:   v1.GasObject.Reference.Digest,
		Owner:    *gasOwner,
	}

	// Convert simple refs
	if len(v1.Deleted) > 0 {
		effects.Deleted = make([]iotago.ObjectRef, len(v1.Deleted))
		for i, ref := range v1.Deleted {
			objDigest := iotago.ObjectDigest(ref.Digest)
			effects.Deleted[i] = iotago.ObjectRef{
				ObjectID: ref.ObjectID,
				Version:  ref.Version,
				Digest:   &objDigest,
			}
		}
	}

	effects.Dependencies = v1.Dependencies
	if v1.EventsDigest != nil {
		effects.EventsDigest = v1.EventsDigest
	}

	return effects, nil
}

func convertOwnerFromBCS(ownerTag serialization.TagJson[iotago.Owner]) (*graphqltypes.Owner, error) {
	owner := &graphqltypes.Owner{}

	if ownerTag.Data.AddressOwner != nil {
		owner.Type = graphqltypes.OwnerTypeAddress
		owner.Address = ownerTag.Data.AddressOwner
	} else if ownerTag.Data.ObjectOwner != nil {
		owner.Type = graphqltypes.OwnerTypeObject
		owner.Address = ownerTag.Data.ObjectOwner
	} else if ownerTag.Data.Shared != nil {
		owner.Type = graphqltypes.OwnerTypeShared
		version := ownerTag.Data.Shared.InitialSharedVersion
		owner.InitialSharedVersion = &version
	} else {
		owner.Type = graphqltypes.OwnerTypeImmutable
	}

	return owner, nil
}

func convertObjectChanges(nodes []ExecuteTransactionBlockExecuteTransactionBlockExecutionResultEffectsTransactionBlockEffectsObjectChangesObjectChangeConnectionNodesObjectChange) []graphqltypes.ObjectChange {
	if len(nodes) == 0 {
		return nil
	}

	changes := make([]graphqltypes.ObjectChange, 0, len(nodes))
	for _, node := range nodes {
		objID := iotago.ObjectID(node.Address)
		change := graphqltypes.ObjectChange{
			ObjectID: objID,
		}

		// Determine change type and populate type-specific details
		if node.IdCreated {
			change.Type = graphqltypes.ObjectChangeCreated
			if node.OutputState.Version > 0 {
				objectType := ""
				if node.OutputState.AsMoveObject.Contents.Type.Repr != "" {
					objectType = node.OutputState.AsMoveObject.Contents.Type.Repr
				}

				owner := extractOwnerFromGraphQL(node.OutputState.Owner)

				digest, _ := iotago.NewDigest(node.OutputState.Digest)
				change.Created = &graphqltypes.CreatedChange{
					ObjectType: objectType,
					Version:    node.OutputState.Version,
					Digest:     *digest,
					Owner:      owner,
				}
			}
		} else if node.IdDeleted {
			change.Type = graphqltypes.ObjectChangeDeleted
			if node.InputState.Version > 0 {
				digest, _ := iotago.NewDigest(node.InputState.Digest)
				change.Deleted = &graphqltypes.DeletedChange{
					Version: node.InputState.Version,
					Digest:  *digest,
				}
			}
		} else if node.OutputState.Version > 0 {
			change.Type = graphqltypes.ObjectChangeModified
			objectType := ""
			if node.OutputState.AsMoveObject.Contents.Type.Repr != "" {
				objectType = node.OutputState.AsMoveObject.Contents.Type.Repr
			}

			owner := extractOwnerFromGraphQL(node.OutputState.Owner)

			digest, _ := iotago.NewDigest(node.OutputState.Digest)
			change.Modified = &graphqltypes.ModifiedChange{
				ObjectType:      objectType,
				PreviousVersion: node.InputState.Version,
				Version:         node.OutputState.Version,
				Digest:          *digest,
				Owner:           owner,
			}
		}

		changes = append(changes, change)
	}

	return changes
}

func convertBalanceChanges(nodes []ExecuteTransactionBlockExecuteTransactionBlockExecutionResultEffectsTransactionBlockEffectsBalanceChangesBalanceChangeConnectionNodesBalanceChange) []graphqltypes.BalanceChange {
	if len(nodes) == 0 {
		return nil
	}

	changes := make([]graphqltypes.BalanceChange, 0, len(nodes))
	for _, node := range nodes {
		// Extract owner address
		var ownerAddr *iotago.Address
		if node.Owner.AsAddress.Address != (iotago.Address{}) {
			addr := node.Owner.AsAddress.Address
			ownerAddr = &addr
		}

		if ownerAddr != nil {
			changes = append(changes, graphqltypes.BalanceChange{
				Owner:    *ownerAddr,
				CoinType: node.CoinType.Repr,
				Amount:   node.Amount.Int.Int64(),
			})
		}
	}

	return changes
}

// extractOwnerFromGraphQL extracts owner information from the GraphQL ObjectOwner type
// Returns nil if owner cannot be determined (e.g., Immutable or Shared objects)
func extractOwnerFromGraphQL(owner ExecuteTransactionBlockExecuteTransactionBlockExecutionResultEffectsTransactionBlockEffectsObjectChangesObjectChangeConnectionNodesObjectChangeOutputStateObjectOwner) *graphqltypes.Owner {
	if owner == nil {
		return nil
	}

	// For now, we'll return nil since the owner extraction from GraphQL union types
	// is complex and requires more investigation of the generated types
	// The BCS-decoded effects already contain accurate owner information
	// TODO: Implement proper owner extraction from GraphQL types
	return nil
}
