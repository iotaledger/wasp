package l1

import (
	"time"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/graphqltypes"
)

// BuildExecuteResponse constructs a graphqltypes.ExecuteTransactionBlockResponse from execution results.
func BuildExecuteResponse(
	result *ExecutionResult,
	txData []byte,
	sender iotago.Address,
	signatures []iotago.Base64Data,
	gasCoinID *iotago.ObjectID,
) *graphqltypes.ExecuteTransactionBlockResponse {
	objectChanges := buildObjectChanges(result)
	gasEffects := buildGasEffects(result, gasCoinID)

	txDigestStr := result.TxDigest.String()

	effects := graphqltypes.TxEffects{
		TX_EFFECTS: graphqltypes.TX_EFFECTS{
			Status:        graphqltypes.ExecutionStatusSuccess,
			Timestamp:     time.Now(),
			GasEffects:    gasEffects,
			ObjectChanges: objectChanges,
			TransactionBlock: graphqltypes.TxBlockCore{
				TX_CORE: graphqltypes.TX_CORE{
					Digest:     txDigestStr,
					Bcs:        txData,
					Sender:     graphqltypes.TX_CORESenderAddress{Address: sender},
					Signatures: signatures,
				},
			},
		},
	}

	return &graphqltypes.ExecuteTransactionBlockResponse{
		ExecuteTransactionBlock: graphqltypes.ExecuteTransactionBlockExecuteTransactionBlockExecutionResult{
			Effects: effects,
		},
	}
}

// BuildGetObjectResponse constructs a GetObjectResponse for a SimObject.
func BuildGetObjectResponse(obj *SimObject) *graphqltypes.GetObjectResponse {
	digestStr := obj.Digest.String()
	return &graphqltypes.GetObjectResponse{
		Object: graphqltypes.GetObjectObject{
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
	}
}

// BuildNotFoundResponse returns a GetObjectResponse for a not-found object.
func BuildNotFoundResponse() *graphqltypes.GetObjectResponse {
	return &graphqltypes.GetObjectResponse{
		Object: graphqltypes.GetObjectObject{
			RPC_OBJECT_FIELDS: graphqltypes.RPC_OBJECT_FIELDS{
				ObjectId: iotago.Address{}, // zero address = not found
				Status:   graphqltypes.ObjectKindWrappedOrDeleted,
			},
		},
	}
}

// BuildGetTransactionBlockResponse constructs a response for GetTransactionBlock.
func BuildGetTransactionBlockResponse(tx *StoredTx) *graphqltypes.GetTransactionBlockResponse {
	if tx.Effects == nil {
		return nil
	}
	effects := tx.Effects.ExecuteTransactionBlock.Effects
	return &graphqltypes.GetTransactionBlockResponse{
		TransactionBlock: graphqltypes.TxBlockData{
			TX_CORE: graphqltypes.TX_CORE{
				Digest:     tx.Digest.String(),
				Bcs:        tx.TxData,
				Sender:     graphqltypes.TX_CORESenderAddress{Address: tx.Sender},
				Signatures: tx.Signatures,
			},
			Effects: effects,
		},
	}
}

func buildObjectChanges(result *ExecutionResult) graphqltypes.TX_EFFECTSObjectChangesObjectChangeConnection {
	var nodes []graphqltypes.ObjectChangeData

	for _, obj := range result.Created {
		nodes = append(nodes, buildObjectChangeNode(obj, true, false))
	}
	for _, obj := range result.Mutated {
		nodes = append(nodes, buildObjectChangeNode(obj, false, false))
	}
	for _, id := range result.Deleted {
		nodes = append(nodes, graphqltypes.ObjectChangeData{
			OBJECT_CHANGE: graphqltypes.OBJECT_CHANGE{
				Address:   id,
				IdCreated: false,
				IdDeleted: true,
			},
		})
	}

	return graphqltypes.TX_EFFECTSObjectChangesObjectChangeConnection{
		Nodes: nodes,
	}
}

func buildObjectChangeNode(obj *SimObject, created, deleted bool) graphqltypes.ObjectChangeData {
	digestStr := obj.Digest.String()

	outputState := graphqltypes.OBJECT_CHANGEOutputStateObject{
		OBJECT_REF: graphqltypes.OBJECT_REF{
			Address: obj.ID,
			Version: obj.Version,
			Digest:  digestStr,
		},
	}

	if obj.Type == "package" {
		// Packages are NOT move objects — AsMoveObject must remain empty.
		// Only AsMovePackage is populated (used by GetPublishedPackageID).
		outputState.AsMovePackage = graphqltypes.OBJECT_CHANGEOutputStateObjectAsMovePackage{
			Modules: graphqltypes.OBJECT_CHANGEOutputStateObjectAsMovePackageModulesMoveModuleConnection{
				Nodes: []graphqltypes.OBJECT_CHANGEOutputStateObjectAsMovePackageModulesMoveModuleConnectionNodesMoveModule{
					{Name: "anchor"},
					{Name: "request"},
					{Name: "assets_bag"},
				},
			},
		}
	} else {
		outputState.AsMoveObject = graphqltypes.OBJECT_CHANGEOutputStateObjectAsMoveObject{
			Contents: graphqltypes.OBJECT_CHANGEOutputStateObjectAsMoveObjectContentsMoveValue{
				Type: graphqltypes.OBJECT_CHANGEOutputStateObjectAsMoveObjectContentsMoveValueTypeMoveType{
					Repr: obj.Type,
				},
			},
		}
	}

	return graphqltypes.ObjectChangeData{
		OBJECT_CHANGE: graphqltypes.OBJECT_CHANGE{
			Address:     obj.ID,
			IdCreated:   created,
			IdDeleted:   deleted,
			OutputState: outputState,
		},
	}
}

func buildGasEffects(result *ExecutionResult, gasCoinID *iotago.ObjectID) graphqltypes.TX_EFFECTSGasEffects {
	gasObj := graphqltypes.TX_EFFECTSGasEffectsGasObject{}
	if gasCoinID != nil {
		gasObj.OBJECT_REF = graphqltypes.OBJECT_REF{
			Address: *gasCoinID,
		}
	}

	return graphqltypes.TX_EFFECTSGasEffects{
		GasObject: gasObj,
		GasSummary: graphqltypes.TX_EFFECTSGasEffectsGasSummaryGasCostSummary{
			ComputationCost:         *graphqltypes.NewBigInt(result.GasCost),
			ComputationCostBurned:   *graphqltypes.NewBigInt(0),
			StorageCost:             *graphqltypes.NewBigInt(0),
			StorageRebate:           *graphqltypes.NewBigInt(0),
			NonRefundableStorageFee: *graphqltypes.NewBigInt(0),
		},
	}
}

// Owner inner type constructors — shared by both RPC_OBJECT_FIELDS and RPC_MOVE_OBJECT_FIELDS wrappers,
// which are structurally identical but satisfy different genqlient interfaces.

func makeAddressOwnerInner(addr iotago.Address) graphqltypes.RPC_OBJECT_OWNER_FIELDSAddressOwner {
	return graphqltypes.RPC_OBJECT_OWNER_FIELDSAddressOwner{
		Typename: "AddressOwner",
		Owner: graphqltypes.RPC_OBJECT_OWNER_FIELDSOwner{
			AsAddress: graphqltypes.RPC_OBJECT_OWNER_FIELDSOwnerAsAddress{Address: addr},
			AsObject:  graphqltypes.RPC_OBJECT_OWNER_FIELDSOwnerAsObject{Address: addr},
		},
	}
}

func makeParentOwnerInner(addr iotago.Address) graphqltypes.RPC_OBJECT_OWNER_FIELDSParent {
	return graphqltypes.RPC_OBJECT_OWNER_FIELDSParent{
		Typename: "Parent",
		Parent:   graphqltypes.RPC_OBJECT_OWNER_FIELDSParentObject{Address: addr},
	}
}

func makeSharedOwnerInner(version uint64) graphqltypes.RPC_OBJECT_OWNER_FIELDSShared {
	return graphqltypes.RPC_OBJECT_OWNER_FIELDSShared{
		Typename:             "Shared",
		InitialSharedVersion: version,
	}
}

func makeImmutableOwnerInner() graphqltypes.RPC_OBJECT_OWNER_FIELDSImmutable {
	return graphqltypes.RPC_OBJECT_OWNER_FIELDSImmutable{
		Typename: "Immutable",
	}
}

func buildRPCObjectOwner(owner SimOwner) graphqltypes.RPC_OBJECT_FIELDSOwnerObjectOwner {
	switch {
	case owner.AddressOwner != nil:
		return &graphqltypes.RPC_OBJECT_FIELDSOwnerAddressOwner{
			Typename: "AddressOwner", RPC_OBJECT_OWNER_FIELDSAddressOwner: makeAddressOwnerInner(*owner.AddressOwner),
		}
	case owner.ObjectOwner != nil:
		return &graphqltypes.RPC_OBJECT_FIELDSOwnerParent{
			Typename: "Parent", RPC_OBJECT_OWNER_FIELDSParent: makeParentOwnerInner(*owner.ObjectOwner),
		}
	case owner.Shared != nil:
		return &graphqltypes.RPC_OBJECT_FIELDSOwnerShared{
			Typename: "Shared", RPC_OBJECT_OWNER_FIELDSShared: makeSharedOwnerInner(owner.Shared.InitialSharedVersion),
		}
	case owner.Immutable:
		return &graphqltypes.RPC_OBJECT_FIELDSOwnerImmutable{
			Typename: "Immutable", RPC_OBJECT_OWNER_FIELDSImmutable: makeImmutableOwnerInner(),
		}
	default:
		return nil
	}
}

func buildRPCMoveObjectOwner(owner SimOwner) graphqltypes.RPC_MOVE_OBJECT_FIELDSOwnerObjectOwner {
	switch {
	case owner.AddressOwner != nil:
		return &graphqltypes.RPC_MOVE_OBJECT_FIELDSOwnerAddressOwner{
			Typename: "AddressOwner", RPC_OBJECT_OWNER_FIELDSAddressOwner: makeAddressOwnerInner(*owner.AddressOwner),
		}
	case owner.ObjectOwner != nil:
		return &graphqltypes.RPC_MOVE_OBJECT_FIELDSOwnerParent{
			Typename: "Parent", RPC_OBJECT_OWNER_FIELDSParent: makeParentOwnerInner(*owner.ObjectOwner),
		}
	case owner.Shared != nil:
		return &graphqltypes.RPC_MOVE_OBJECT_FIELDSOwnerShared{
			Typename: "Shared", RPC_OBJECT_OWNER_FIELDSShared: makeSharedOwnerInner(owner.Shared.InitialSharedVersion),
		}
	case owner.Immutable:
		return &graphqltypes.RPC_MOVE_OBJECT_FIELDSOwnerImmutable{
			Typename: "Immutable", RPC_OBJECT_OWNER_FIELDSImmutable: makeImmutableOwnerInner(),
		}
	default:
		return nil
	}
}
