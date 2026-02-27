package graphqltypes

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

func (v *OBJECT_REF) ObjectRef() (*iotago.ObjectRef, error) {
	digest, err := iotago.NewDigest(v.Digest)
	if err != nil {
		return nil, fmt.Errorf("invalid object digest %q: %w", v.Digest, err)
	}
	objectID := v.Address
	return &iotago.ObjectRef{
		ObjectID: &objectID,
		Version:  v.Version,
		Digest:   digest,
	}, nil
}

func (v *TxEffects) IsSuccess() bool {
	return v.GetStatus() == ExecutionStatusSuccess
}

func (v *TxEffects) IsFailed() bool {
	return !v.IsSuccess()
}

func (v *TxEffects) GasFee() int64 {
	s := v.GetGasEffects().GasSummary
	return s.ComputationCost.Int64() + s.StorageCost.Int64() - s.StorageRebate.Int64()
}

func (v *TxEffects) GetPublishedPackageID() (*iotago.PackageID, error) {
	for _, change := range v.GetObjectChanges().Nodes {
		if !change.GetIdCreated() {
			continue
		}
		// Check if this is a package (has modules)
		modules := change.GetOutputState().AsMovePackage.Modules.Nodes
		if len(modules) > 0 {
			packageID := change.GetAddress()
			return &packageID, nil
		}
	}
	return nil, fmt.Errorf("no published package found in transaction")
}

func (v *TxEffects) GetCreatedObjectByName(module string, objectName string) (*iotago.ObjectRef, error) {
	nodes := v.GetObjectChanges().Nodes
	for i := range nodes {
		if !nodes[i].IdCreated {
			continue
		}
		typeRepr := nodes[i].OutputState.AsMoveObject.Contents.Type.Repr
		if typeRepr == "" {
			continue
		}
		resource, err := iotago.NewResourceType(typeRepr)
		if err != nil {
			return nil, fmt.Errorf("invalid resource string %q: %w", typeRepr, err)
		}
		if resource.Contains(nil, module, objectName) {
			return nodes[i].OutputState.ObjectRef()
		}
	}
	return nil, fmt.Errorf("created object %s::%s not found", module, objectName)
}

func (v *TxEffects) GetCreatedCoinByType(module string, coinType string) (*iotago.ObjectRef, error) {
	nodes := v.GetObjectChanges().Nodes
	for i := range nodes {
		if !nodes[i].IdCreated {
			continue
		}
		typeRepr := nodes[i].OutputState.AsMoveObject.Contents.Type.Repr
		if typeRepr == "" {
			continue
		}
		resource, err := iotago.NewResourceType(typeRepr)
		if err != nil {
			return nil, fmt.Errorf("invalid resource string %q: %w", typeRepr, err)
		}
		if resource.Module == "coin" && resource.SubType1 != nil &&
			resource.SubType1.Module == module && resource.SubType1.ObjectName == coinType {
			return nodes[i].OutputState.ObjectRef()
		}
	}
	return nil, fmt.Errorf("created coin %s::%s not found", module, coinType)
}

func (v *TxEffects) GetMutatedObjectByID(objectID iotago.ObjectID) (*iotago.ObjectRef, error) {
	nodes := v.GetObjectChanges().Nodes
	for i := range nodes {
		if nodes[i].IdCreated || nodes[i].IdDeleted {
			continue
		}
		if nodes[i].Address == objectID {
			return nodes[i].OutputState.ObjectRef()
		}
	}
	return nil, fmt.Errorf("mutated object %s not found", objectID)
}

func (v *TxEffects) GetMutatedCoinByType(module string, coinType string) (*iotago.ObjectRef, error) {
	nodes := v.GetObjectChanges().Nodes
	for i := range nodes {
		if nodes[i].IdCreated || nodes[i].IdDeleted {
			continue
		}
		typeRepr := nodes[i].OutputState.AsMoveObject.Contents.Type.Repr
		if typeRepr == "" {
			continue
		}
		resource, err := iotago.NewResourceType(typeRepr)
		if err != nil {
			return nil, fmt.Errorf("invalid resource string %q: %w", typeRepr, err)
		}
		if resource.Module == "coin" && resource.SubType1 != nil &&
			resource.SubType1.Module == module && resource.SubType1.ObjectName == coinType {
			return nodes[i].OutputState.ObjectRef()
		}
	}
	return nil, fmt.Errorf("mutated coin %s::%s not found", module, coinType)
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

func (v *ExecuteTransactionBlockResponse) GetMutatedObjectByID(objectID iotago.ObjectID) (*iotago.ObjectRef, error) {
	return v.ExecuteTransactionBlock.Effects.GetMutatedObjectByID(objectID)
}

func (v *ExecuteTransactionBlockResponse) GetMutatedCoinByType(module string, coinType string) (*iotago.ObjectRef, error) {
	return v.ExecuteTransactionBlock.Effects.GetMutatedCoinByType(module, coinType)
}
