package graphqltypes

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type TransactionResponse struct {
	// Core fields
	Digest     iotago.TransactionDigest
	Sender     iotago.Address
	Signatures [][]byte

	// Optional fields
	RawInput       []byte
	Effects        *TransactionEffects
	Events         []Event
	BalanceChanges []BalanceChange
	ObjectChanges  []ObjectChange
	TimestampMS    *uint64
	CheckpointSeq  *uint64
	Errors         []string
}

type TransactionEffects struct {
	// Execution status and metadata
	Status            ExecutionStatus // Type-safe execution status
	ExecutedEpoch     uint64          // Direct uint64, not BigInt
	GasUsed           GasCostSummary  // Flat structure with direct uint64 fields
	TransactionDigest iotago.TransactionDigest

	// Object references
	Created              []OwnedObjectRef
	Mutated              []OwnedObjectRef
	Unwrapped            []OwnedObjectRef
	Deleted              []iotago.ObjectRef
	UnwrappedThenDeleted []iotago.ObjectRef
	Wrapped              []iotago.ObjectRef
	GasObject            OwnedObjectRef
	Dependencies         []iotago.TransactionDigest
	EventsDigest         *iotago.TransactionEventsDigest
}

type ExecutionStatus struct {
	Success bool
	Error   string
}

const (
	ExecutionStatusSuccess = "Success"
	ExecutionStatusFailure = "Failure"
)

func (s ExecutionStatus) StatusString() string {
	if s.Success {
		return ExecutionStatusSuccess
	}
	return ExecutionStatusFailure
}

type GasCostSummary struct {
	ComputationCost         uint64
	StorageCost             uint64
	StorageRebate           uint64
	NonRefundableStorageFee uint64
}

type OwnedObjectRef struct {
	ObjectID iotago.ObjectID
	Version  uint64
	Digest   iotago.ObjectDigest
	Owner    Owner
}

func (t TransactionEffects) IsSuccess() bool {
	return t.Status.Success
}

func (t TransactionEffects) GasFee() int64 {
	return int64(t.GasUsed.StorageCost) - int64(t.GasUsed.StorageRebate) + int64(t.GasUsed.ComputationCost)
}

// GetCreatedObjectByName searches for a created object by its type name.
// This is a helper method for finding specific objects in the transaction effects.
func (t *TransactionResponse) GetCreatedObjectByName(moduleName, objectName string) (*iotago.ObjectRef, error) {
	if t.Effects == nil {
		return nil, ErrNoEffects
	}

	targetType := moduleName + "::" + objectName
	for _, created := range t.Effects.Created {
		// For now, we'll need to match by checking object changes
		// since we don't have the full type information in OwnedObjectRef
		for _, change := range t.ObjectChanges {
			if change.ObjectID == created.ObjectID &&
				change.Type == ObjectChangeCreated &&
				change.Created != nil {
				// Simple type matching - you may need to adjust this based on your type naming
				if containsTypeName(change.Created.ObjectType, targetType) {
					digest := created.Digest
					return &iotago.ObjectRef{
						ObjectID: &created.ObjectID,
						Version:  created.Version,
						Digest:   &digest,
					}, nil
				}
			}
		}
	}
	return nil, ErrObjectNotFound
}

// Helper function to check if a type string contains the target type name
func containsTypeName(fullType, targetType string) bool {
	// This is a simple substring match - adjust as needed for your type naming conventions
	return len(fullType) > 0 && (fullType == targetType ||
		len(fullType) > len(targetType) && fullType[len(fullType)-len(targetType):] == targetType)
}

func (t *TransactionResponse) GetPublishedPackageID() (*iotago.PackageID, error) {
	if t.Effects == nil {
		return nil, ErrNoEffects
	}

	for _, change := range t.ObjectChanges {
		if change.Type == ObjectChangePublished && change.Published != nil {
			return &change.Published.PackageID, nil
		}
	}

	return nil, ErrPackageIDNotFound
}

var (
	ErrNoEffects         = fmt.Errorf("no effects in transaction response")
	ErrObjectNotFound    = fmt.Errorf("object not found in created objects")
	ErrPackageIDNotFound = fmt.Errorf("package ID not found in transaction")
)

type BalanceChange struct {
	Owner    iotago.Address
	CoinType string
	Amount   int64
}

type Event struct {
	TxDigest iotago.TransactionDigest
	EventSeq uint64
	// Add more fields as needed
}
