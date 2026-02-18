package l1

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

// ValidatedInputs holds the resolved objects for a transaction.
type ValidatedInputs struct {
	Objects  map[iotago.ObjectID]*SimObject
	GasCoins []*SimObject
	Versions []uint64 // all input versions for Lamport computation
}

// ValidateTransaction validates transaction inputs against the object store.
func ValidateTransaction(store *ObjectStore, tx *iotago.TransactionDataV1) (*ValidatedInputs, error) {
	result := &ValidatedInputs{
		Objects: make(map[iotago.ObjectID]*SimObject),
	}

	pt := tx.Kind.ProgrammableTransaction
	if pt == nil {
		return nil, fmt.Errorf("transaction must be ProgrammableTransaction")
	}

	if tx.GasData.Budget == 0 {
		return nil, fmt.Errorf("gas budget must be > 0")
	}

	for _, gasRef := range tx.GasData.Payment {
		if gasRef == nil || gasRef.ObjectID == nil {
			continue
		}
		obj, ok := store.Get(*gasRef.ObjectID)
		if !ok {
			if store.IsDeleted(*gasRef.ObjectID) {
				return nil, fmt.Errorf("gas coin %s has been deleted", gasRef.ObjectID.String())
			}
			return nil, fmt.Errorf("gas coin %s not found", gasRef.ObjectID.String())
		}
		if obj.Owner.AddressOwner == nil || *obj.Owner.AddressOwner != *tx.GasData.Owner {
			return nil, fmt.Errorf("gas coin %s not owned by gas owner %s", gasRef.ObjectID.String(), tx.GasData.Owner.String())
		}
		result.GasCoins = append(result.GasCoins, obj)
		result.Objects[*gasRef.ObjectID] = obj
		result.Versions = append(result.Versions, obj.Version)
	}

	for _, input := range pt.Inputs {
		if input.Object == nil {
			continue
		}
		objArg := input.Object

		switch {
		case objArg.ImmOrOwnedObject != nil:
			ref := objArg.ImmOrOwnedObject
			if ref.ObjectID == nil {
				continue
			}
			obj, ok := store.Get(*ref.ObjectID)
			if !ok {
				if store.IsDeleted(*ref.ObjectID) {
					return nil, fmt.Errorf("object %s has been deleted", ref.ObjectID.String())
				}
				return nil, fmt.Errorf("object %s not found", ref.ObjectID.String())
			}
			// Relaxed version check: only verify if a specific version was requested
			if ref.Version != 0 && obj.Version != ref.Version {
				return nil, fmt.Errorf("object %s version mismatch: expected %d, got %d",
					ref.ObjectID.String(), ref.Version, obj.Version)
			}
			if obj.Owner.AddressOwner != nil && *obj.Owner.AddressOwner != tx.Sender {
				return nil, fmt.Errorf("object %s owned by %s, not sender %s",
					ref.ObjectID.String(), obj.Owner.AddressOwner.String(), tx.Sender.String())
			}
			result.Objects[*ref.ObjectID] = obj
			result.Versions = append(result.Versions, obj.Version)

		case objArg.SharedObject != nil:
			shared := objArg.SharedObject
			if shared.Id == nil {
				continue
			}
			obj, ok := store.Get(*shared.Id)
			if !ok {
				return nil, fmt.Errorf("shared object %s not found", shared.Id.String())
			}
			if obj.Owner.Shared == nil {
				return nil, fmt.Errorf("object %s is not shared", shared.Id.String())
			}
			result.Objects[*shared.Id] = obj
			result.Versions = append(result.Versions, obj.Version)

		case objArg.Receiving != nil:
			ref := objArg.Receiving
			if ref.ObjectID == nil {
				continue
			}
			obj, ok := store.Get(*ref.ObjectID)
			if !ok {
				if store.IsDeleted(*ref.ObjectID) {
					return nil, fmt.Errorf("receiving object %s has been deleted", ref.ObjectID.String())
				}
				return nil, fmt.Errorf("receiving object %s not found", ref.ObjectID.String())
			}
			result.Objects[*ref.ObjectID] = obj
			result.Versions = append(result.Versions, obj.Version)
		}
	}

	return result, nil
}
