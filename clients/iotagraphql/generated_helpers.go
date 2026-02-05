package iotagraphql

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

// Extension methods for generated GraphQL types.
// These are kept separate from generated.go so they survive code regeneration.

// IsSuccess returns true if the transaction effects indicate successful execution.
func (v *TxEffects) IsSuccess() bool {
	return v.GetStatus() == ExecutionStatusSuccess
}

// GasFee returns the total gas fee for the execute transaction result.
func (v *TxEffects) GasFee() int64 {
	s := v.GetGasEffects().GasSummary
	return s.ComputationCost.Int64() + s.StorageCost.Int64() - s.StorageRebate.Int64()
}

// GetPublishedPackageID finds the published package ID from the transaction's object changes.
func (v *TxEffects) GetPublishedPackageID() (*iotago.PackageID, error) {
	for _, change := range v.GetObjectChanges().Nodes {
		if !change.GetIdCreated() {
			continue
		}
		// Check if this is a package (has modules)
		modules := change.GetOutputState().AsMovePackage.Modules.Nodes
		if len(modules) > 0 {
			packageID := iotago.PackageID(change.GetAddress())
			return &packageID, nil
		}
	}
	return nil, fmt.Errorf("no published package found in transaction")
}

// GetCreatedObjectByName finds a created object by module and type name.
// TODO: implement properly
func (v *TxEffects) GetCreatedObjectByName(module string, objectName string) (*iotago.ObjectRef, error) {
	return nil, fmt.Errorf("GetCreatedObjectByName not yet implemented")
}

// GetCreatedCoinByType finds a created coin by module and type name.
// TODO: implement properly
func (v *TxEffects) GetCreatedCoinByType(module string, coinType string) (*iotago.ObjectRef, error) {
	return nil, fmt.Errorf("GetCreatedCoinByType not yet implemented")
}

// GetMutatedObjectByID finds a mutated object by its ID.
// TODO: implement properly
func (v *TxEffects) GetMutatedObjectByID(objectID iotago.ObjectID) (*iotago.ObjectRef, error) {
	return nil, fmt.Errorf("GetMutatedObjectByID not yet implemented")
}

func (v *ExecuteTransactionBlockResponse) IsSuccess() bool {
	return v.ExecuteTransactionBlock.Effects.IsSuccess()
}

func (v *ExecuteTransactionBlockResponse) GetPublishedPackageID() (*iotago.PackageID, error) {
	return v.ExecuteTransactionBlock.Effects.GetPublishedPackageID()
}

func (v *ExecuteTransactionBlockResponse) GetCreatedObjectByName(module string, objectName string) (*iotago.ObjectRef, error) {
	return v.ExecuteTransactionBlock.Effects.GetCreatedObjectByName(module, objectName)
}

func (v *ExecuteTransactionBlockResponse) GetCreatedCoinByType(module string, coinType string) (*iotago.ObjectRef, error) {
	return v.ExecuteTransactionBlock.Effects.GetCreatedCoinByType(module, coinType)
}

// Convenience methods for GetTransactionBlockResponse that delegate to Effects.

func (v *GetTransactionBlockResponse) GetMutatedObjectByID(objectID iotago.ObjectID) (*iotago.ObjectRef, error) {
	return v.TransactionBlock.Effects.GetMutatedObjectByID(objectID)
}

func (v *GetTransactionBlockResponse) GetCreatedObjectByName(module string, objectName string) (*iotago.ObjectRef, error) {
	return v.TransactionBlock.Effects.GetCreatedObjectByName(module, objectName)
}
