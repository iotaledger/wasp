package l1

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type ValidatedInputs struct {
	Objects  map[iotago.ObjectID]*SimObject
	GasCoins []*SimObject
	Versions []uint64 // all input versions for Lamport computation
}

func ValidateTransaction(store *ObjectStore, tx *iotago.TransactionDataV1) (*ValidatedInputs, error) {
	pt := tx.Kind.ProgrammableTransaction
	if pt == nil {
		return nil, fmt.Errorf("transaction must be ProgrammableTransaction")
	}
	if tx.GasData.Budget == 0 {
		return nil, fmt.Errorf("gas budget must be > 0")
	}

	result := &ValidatedInputs{
		Objects: make(map[iotago.ObjectID]*SimObject),
	}

	if err := validateGasCoins(store, tx, result); err != nil {
		return nil, err
	}

	for _, input := range pt.Inputs {
		if input.Object == nil {
			continue
		}
		if err := validateObjectInput(store, tx.Sender, input.Object, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func validateGasCoins(store *ObjectStore, tx *iotago.TransactionDataV1, result *ValidatedInputs) error {
	for _, gasRef := range tx.GasData.Payment {
		if gasRef == nil || gasRef.ObjectID == nil {
			continue
		}
		obj, ok := store.Get(*gasRef.ObjectID)
		if !ok {
			if store.IsDeleted(*gasRef.ObjectID) {
				return fmt.Errorf("gas coin %s has been deleted", gasRef.ObjectID.String())
			}
			return fmt.Errorf("gas coin %s not found", gasRef.ObjectID.String())
		}
		if obj.Owner.AddressOwner == nil || *obj.Owner.AddressOwner != *tx.GasData.Owner {
			return fmt.Errorf("gas coin %s not owned by gas owner %s", gasRef.ObjectID.String(), tx.GasData.Owner.String())
		}
		result.GasCoins = append(result.GasCoins, obj)
		result.Objects[*gasRef.ObjectID] = obj
		result.Versions = append(result.Versions, obj.Version)
	}
	return nil
}

func validateObjectInput(store *ObjectStore, sender iotago.Address, objArg *iotago.ObjectArg, result *ValidatedInputs) error {
	switch {
	case objArg.ImmOrOwnedObject != nil:
		return validateImmOrOwned(store, sender, objArg.ImmOrOwnedObject, result)
	case objArg.SharedObject != nil:
		return validateShared(store, objArg.SharedObject, result)
	case objArg.Receiving != nil:
		return validateReceiving(store, objArg.Receiving, result)
	}
	return nil
}

func validateImmOrOwned(store *ObjectStore, sender iotago.Address, ref *iotago.ObjectRef, result *ValidatedInputs) error {
	if ref.ObjectID == nil {
		return nil
	}
	obj, ok := store.Get(*ref.ObjectID)
	if !ok {
		if store.IsDeleted(*ref.ObjectID) {
			return fmt.Errorf("object %s has been deleted", ref.ObjectID.String())
		}
		return fmt.Errorf("object %s not found", ref.ObjectID.String())
	}
	// Relaxed version check: only verify if a specific version was requested
	if ref.Version != 0 && obj.Version != ref.Version {
		return fmt.Errorf("object %s version mismatch: expected %d, got %d",
			ref.ObjectID.String(), ref.Version, obj.Version)
	}
	if obj.Owner.AddressOwner != nil && *obj.Owner.AddressOwner != sender {
		return fmt.Errorf("object %s owned by %s, not sender %s",
			ref.ObjectID.String(), obj.Owner.AddressOwner.String(), sender.String())
	}
	result.Objects[*ref.ObjectID] = obj
	result.Versions = append(result.Versions, obj.Version)
	return nil
}

func validateShared(store *ObjectStore, shared *iotago.SharedObjectArg, result *ValidatedInputs) error {
	if shared.Id == nil {
		return nil
	}
	obj, ok := store.Get(*shared.Id)
	if !ok {
		return fmt.Errorf("shared object %s not found", shared.Id.String())
	}
	if obj.Owner.Shared == nil {
		return fmt.Errorf("object %s is not shared", shared.Id.String())
	}
	result.Objects[*shared.Id] = obj
	result.Versions = append(result.Versions, obj.Version)
	return nil
}

func validateReceiving(store *ObjectStore, ref *iotago.ObjectRef, result *ValidatedInputs) error {
	if ref.ObjectID == nil {
		return nil
	}
	obj, ok := store.Get(*ref.ObjectID)
	if !ok {
		if store.IsDeleted(*ref.ObjectID) {
			return fmt.Errorf("receiving object %s has been deleted", ref.ObjectID.String())
		}
		return fmt.Errorf("receiving object %s not found", ref.ObjectID.String())
	}
	result.Objects[*ref.ObjectID] = obj
	result.Versions = append(result.Versions, obj.Version)
	return nil
}
