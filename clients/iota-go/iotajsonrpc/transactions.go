package iotajsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/serialization"
	"github.com/samber/lo"
)

type ExecuteTransactionRequestType string

const (
	TxnRequestTypeWaitForEffectsCert    ExecuteTransactionRequestType = "WaitForEffectsCert"
	TxnRequestTypeWaitForLocalExecution ExecuteTransactionRequestType = "WaitForLocalExecution"
)

type EpochId = uint64

type GasCostSummary struct {
	ComputationCost         *BigInt `json:"computationCost"`
	StorageCost             *BigInt `json:"storageCost"`
	StorageRebate           *BigInt `json:"storageRebate"`
	NonRefundableStorageFee *BigInt `json:"nonRefundableStorageFee"`
}

const (
	ExecutionStatusSuccess = "success"
	ExecutionStatusFailure = "failure"
)

type ExecutionStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type OwnedObjectRef struct {
	Owner     serialization.TagJson[iotago.Owner] `json:"owner"`
	Reference IotaObjectRef                       `json:"reference"`
}

type IotaTransactionBlockEffectsModifiedAtVersions struct {
	ObjectID       iotago.ObjectID `json:"objectId"`
	SequenceNumber *BigInt         `json:"sequenceNumber"`
}

type IotaTransactionBlockEffectsV1 struct {
	/** The status of the execution */
	Status ExecutionStatus `json:"status"`
	/** The epoch when this transaction was executed */
	ExecutedEpoch *BigInt        `json:"executedEpoch"`
	GasUsed       GasCostSummary `json:"gasUsed"`
	/** The version that every modified (mutated or deleted) object had before it was modified by this transaction. **/
	ModifiedAtVersions []IotaTransactionBlockEffectsModifiedAtVersions `json:"modifiedAtVersions,omitempty"`
	/** The object references of the shared objects used in this transaction. Empty if no shared objects were used. */
	SharedObjects []IotaObjectRef `json:"sharedObjects,omitempty"`
	/** The transaction digest */
	TransactionDigest iotago.TransactionDigest `json:"transactionDigest"`
	/** ObjectRef and owner of new objects created */
	Created []OwnedObjectRef `json:"created,omitempty"`
	/** ObjectRef and owner of mutated objects, including gas object */
	Mutated []OwnedObjectRef `json:"mutated,omitempty"`
	/**
	 * ObjectRef and owner of objects that are unwrapped in this transaction.
	 * Unwrapped objects are objects that were wrapped into other objects in the past,
	 * and just got extracted out.
	 */
	Unwrapped []OwnedObjectRef `json:"unwrapped,omitempty"`
	/** Object Refs of objects now deleted (the old refs) */
	Deleted []IotaObjectRef `json:"deleted,omitempty"`
	/** Object Refs of objects now deleted (the old refs) */
	UnwrappedThenDeleted []IotaObjectRef `json:"unwrapped_then_deleted,omitempty"`
	/** Object refs of objects now wrapped in other objects */
	Wrapped []IotaObjectRef `json:"wrapped,omitempty"`
	/**
	 * The updated gas object reference. Have a dedicated field for convenient access.
	 * It's also included in mutated.
	 */
	GasObject OwnedObjectRef `json:"gasObject"`
	/** The events emitted during execution. Note that only successful transactions emit events */
	EventsDigest *iotago.TransactionEventsDigest `json:"eventsDigest,omitempty"`
	/** The set of transaction digests this transaction depends on */
	Dependencies []iotago.TransactionDigest `json:"dependencies,omitempty"`
}

type IotaTransactionBlockEffects struct {
	V1 *IotaTransactionBlockEffectsV1 `json:"v1"`
}

func (t IotaTransactionBlockEffects) Tag() string {
	return "messageVersion"
}

func (t IotaTransactionBlockEffects) Content() string {
	return ""
}

func (t IotaTransactionBlockEffects) GasFee() int64 {
	return t.V1.GasUsed.StorageCost.Int64() -
		t.V1.GasUsed.StorageRebate.Int64() +
		t.V1.GasUsed.ComputationCost.Int64()
}

func (t IotaTransactionBlockEffects) IsSuccess() bool {
	return t.V1.Status.Status == ExecutionStatusSuccess
}

func (t IotaTransactionBlockEffects) IsFailed() bool {
	return t.V1.Status.Status == ExecutionStatusFailure
}

const (
	IotaTransactionBlockKindIotaChangeEpoch             = "ChangeEpoch"
	IotaTransactionBlockKindIotaConsensusCommitPrologue = "ConsensusCommitPrologue"
	IotaTransactionBlockKindGenesis                     = "Genesis"
	IotaTransactionBlockKindProgrammableTransaction     = "ProgrammableTransaction"
)

type IotaTransactionBlockKind = serialization.TagJson[TransactionBlockKind]

type TransactionBlockKind struct {
	// A system transaction that will update epoch information on-chain.
	ChangeEpoch *IotaChangeEpoch `json:"ChangeEpoch,omitempty"`
	// A system transaction used for initializing the initial state of the chain.
	Genesis *IotaGenesisTransaction `json:"Genesis,omitempty"`
	// A system transaction marking the start of a series of transactions scheduled as part of a
	// checkpoint
	ConsensusCommitPrologue *IotaConsensusCommitPrologue `json:"ConsensusCommitPrologue,omitempty"`
	// A series of transactions where the results of one transaction can be used in future
	// transactions
	ProgrammableTransaction *IotaProgrammableTransactionBlock `json:"ProgrammableTransaction,omitempty"`
	// .. more transaction types go here
}

func (t TransactionBlockKind) Tag() string {
	return "kind"
}

func (t TransactionBlockKind) Content() string {
	return ""
}

type IotaChangeEpoch struct {
	Epoch                 *BigInt `json:"epoch"`
	StorageCharge         uint64  `json:"storage_charge"`
	ComputationCharge     uint64  `json:"computation_charge"`
	StorageRebate         uint64  `json:"storage_rebate"`
	EpochStartTimestampMs uint64  `json:"epoch_start_timestamp_ms"`
}

type IotaGenesisTransaction struct {
	Objects []iotago.ObjectID `json:"objects"`
}

type IotaConsensusCommitPrologue struct {
	Epoch             uint64 `json:"epoch"`
	Round             uint64 `json:"round"`
	CommitTimestampMs uint64 `json:"commit_timestamp_ms"`
}

type IotaProgrammableTransactionBlock struct {
	Inputs []interface{} `json:"inputs"`
	// The transactions to be executed sequentially. A failure in any transaction will
	// result in the failure of the entire programmable transaction block.
	Commands []interface{} `json:"transactions"`
}

type IotaTransactionBlockDataV1 struct {
	Transaction IotaTransactionBlockKind `json:"transaction"`
	Sender      iotago.Address           `json:"sender"`
	GasData     IotaGasData              `json:"gasData"`
}

type IotaTransactionBlockData struct {
	V1 *IotaTransactionBlockDataV1 `json:"v1,omitempty"`
}

func (t IotaTransactionBlockData) Tag() string {
	return "messageVersion"
}

func (t IotaTransactionBlockData) Content() string {
	return ""
}

type IotaTransactionBlock struct {
	Data         serialization.TagJson[IotaTransactionBlockData] `json:"data"`
	TxSignatures []string                                        `json:"txSignatures"`
}

type ObjectChange struct {
	Published *struct {
		PackageId iotago.ObjectID     `json:"packageId"`
		Version   *BigInt             `json:"version"`
		Digest    iotago.ObjectDigest `json:"digest"`
		Nodules   []string            `json:"nodules"`
	} `json:"published,omitempty"`
	// Transfer objects to new address / wrap in another object
	Transferred *struct {
		Sender     iotago.Address      `json:"sender"`
		Recipient  ObjectOwner         `json:"recipient"`
		ObjectType string              `json:"objectType"`
		ObjectID   iotago.ObjectID     `json:"objectId"`
		Version    *BigInt             `json:"version"`
		Digest     iotago.ObjectDigest `json:"digest"`
	} `json:"transferred,omitempty"`
	// Object mutated.
	Mutated *struct {
		Sender          iotago.Address      `json:"sender"`
		Owner           ObjectOwner         `json:"owner"`
		ObjectType      string              `json:"objectType"`
		ObjectID        iotago.ObjectID     `json:"objectId"`
		Version         *BigInt             `json:"version"`
		PreviousVersion *BigInt             `json:"previousVersion"`
		Digest          iotago.ObjectDigest `json:"digest"`
	} `json:"mutated,omitempty"`
	// Delete object j
	Deleted *struct {
		Sender     iotago.Address  `json:"sender"`
		ObjectType string          `json:"objectType"`
		ObjectID   iotago.ObjectID `json:"objectId"`
		Version    *BigInt         `json:"version"`
	} `json:"deleted,omitempty"`
	// Wrapped object
	Wrapped *struct {
		Sender     iotago.Address  `json:"sender"`
		ObjectType string          `json:"objectType"`
		ObjectID   iotago.ObjectID `json:"objectId"`
		Version    *BigInt         `json:"version"`
	} `json:"wrapped,omitempty"`
	// New object creation
	Created *struct {
		Sender     iotago.Address      `json:"sender"`
		Owner      ObjectOwner         `json:"owner"`
		ObjectType string              `json:"objectType"`
		ObjectID   iotago.ObjectID     `json:"objectId"`
		Version    *BigInt             `json:"version"`
		Digest     iotago.ObjectDigest `json:"digest"`
	} `json:"created,omitempty"`
}

func (o ObjectChange) Tag() string {
	return "type"
}

func (o ObjectChange) Content() string {
	return ""
}

func (o ObjectChange) String() string {
	s := ""
	if o.Published != nil {
		s = fmt.Sprintf("Published: %v", o.Published)
	}
	if o.Transferred != nil {
		s = fmt.Sprintf("Transferred: %v", o.Transferred)
	}
	if o.Mutated != nil {
		s = fmt.Sprintf("Mutated: %v", o.Mutated)
	}
	if o.Deleted != nil {
		s = fmt.Sprintf("Deleted: %v", o.Deleted)
	}
	if o.Wrapped != nil {
		s = fmt.Sprintf("Wrapped: %v", o.Wrapped)
	}
	if o.Created != nil {
		s = fmt.Sprintf("Created: %v", o.Created)
	}
	return s
}

type BalanceChange struct {
	Owner    ObjectOwner `json:"owner"`
	CoinType string      `json:"coinType"`
	/* Coin balance change(positive means receive, negative means send) */
	Amount string `json:"amount"`
}

type IotaTransactionBlockResponse struct {
	Digest                  iotago.TransactionDigest                            `json:"digest"`
	Transaction             *IotaTransactionBlock                               `json:"transaction,omitempty"`
	RawTransaction          iotago.Base64Data                                   `json:"rawTransaction,omitempty"` // enable by show_raw_input
	Effects                 *serialization.TagJson[IotaTransactionBlockEffects] `json:"effects,omitempty"`
	Events                  []*IotaEvent                                        `json:"events,omitempty"`
	TimestampMs             *BigInt                                             `json:"timestampMs,omitempty"`
	Checkpoint              *BigInt                                             `json:"checkpoint,omitempty"`
	ConfirmedLocalExecution *bool                                               `json:"confirmedLocalExecution,omitempty"`
	ObjectChanges           []serialization.TagJson[ObjectChange]               `json:"objectChanges,omitempty"`
	BalanceChanges          []BalanceChange                                     `json:"balanceChanges,omitempty"`
	Errors                  []string                                            `json:"errors,omitempty"` // Errors that occurred in fetching/serializing the transaction.
	// FIXME datatype may be wrong
	RawEffects []int `json:"rawEffects,omitempty"` // enable by show_raw_effects
}

// requires to set 'IotaTransactionBlockResponseOptions.ShowObjectChanges' to true
func (r *IotaTransactionBlockResponse) GetPublishedPackageID() (*iotago.PackageID, error) {
	if r.ObjectChanges == nil {
		return nil, errors.New("no 'IotaTransactionBlockResponse.ObjectChanges' object")
	}
	var packageID iotago.PackageID
	for _, change := range r.ObjectChanges {
		if change.Data.Published != nil {
			packageID = change.Data.Published.PackageId
			return &packageID, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

// requires `ShowObjectChanges: true`
func (r *IotaTransactionBlockResponse) GetCreatedObjectByName(module string, objectName string) (
	*iotago.ObjectRef,
	error,
) {
	if r.ObjectChanges == nil {
		return nil, errors.New("expected ObjectChanges != nil")
	}

	var ref *iotago.ObjectRef
	var prevCreatedObj any

	for _, change := range r.ObjectChanges {
		if change.Data.Created != nil {
			// some possible examples
			// * 0x2::coin::TreasuryCap<0x14c12b454ac6996024342312769e00bb98c70ad2f3546a40f62516c83aa0f0d4::testcoin::TESTCOIN>
			// * 0x14c12b454ac6996024342312769e00bb98c70ad2f3546a40f62516c83aa0f0d4::anchor::Anchor
			resource, err := iotago.NewResourceType(change.Data.Created.ObjectType)
			if err != nil {
				return nil, fmt.Errorf("invalid resource string: %w", err)
			}
			if resource.Contains(nil, module, objectName) {
				if ref != nil {
					return nil, fmt.Errorf("multiple created objects found for %s::%s: first = %v, second = %v",
						module, objectName,
						string(lo.Must(json.Marshal(prevCreatedObj))),
						string(lo.Must(json.Marshal(change.Data.Created))),
					)
				}

				ref = &iotago.ObjectRef{
					ObjectID: &change.Data.Created.ObjectID,
					Version:  change.Data.Created.Version.Uint64(),
					Digest:   &change.Data.Created.Digest,
				}
				prevCreatedObj = change.Data.Created
			}
		}
	}

	if ref == nil {
		return nil, fmt.Errorf("not found")
	}
	return ref, nil
}

func (r *IotaTransactionBlockResponse) GetMutatedObjectByName(module string, objectName string) (
	*iotago.ObjectRef,
	error,
) {
	if r.ObjectChanges == nil {
		return nil, errors.New("expected ObjectChanges != nil")
	}

	var ref *iotago.ObjectRef
	var prevMutatedObj any

	for _, change := range r.ObjectChanges {
		if change.Data.Mutated != nil {
			// some possible examples
			// * 0x2::coin::TreasuryCap<0x14c12b454ac6996024342312769e00bb98c70ad2f3546a40f62516c83aa0f0d4::testcoin::TESTCOIN>
			// * 0x14c12b454ac6996024342312769e00bb98c70ad2f3546a40f62516c83aa0f0d4::anchor::Anchor
			resource, err := iotago.NewResourceType(change.Data.Mutated.ObjectType)
			if err != nil {
				return nil, fmt.Errorf("invalid resource string: %w", err)
			}
			if resource.Contains(nil, module, objectName) {
				if ref != nil {
					return nil, fmt.Errorf("multiple mutated objects found for %s::%s: first = %v, second = %v",
						module, objectName,
						string(lo.Must(json.Marshal(prevMutatedObj))),
						string(lo.Must(json.Marshal(change.Data.Mutated))),
					)
				}

				ref = &iotago.ObjectRef{
					ObjectID: &change.Data.Mutated.ObjectID,
					Version:  change.Data.Mutated.Version.Uint64(),
					Digest:   &change.Data.Mutated.Digest,
				}
				prevMutatedObj = change.Data.Mutated
			}
		}
	}

	if ref == nil {
		return nil, fmt.Errorf("not found")
	}
	return ref, nil
}

func (r *IotaTransactionBlockResponse) GetMutatedObjectByID(objectID iotago.ObjectID) (
	*iotago.ObjectRef,
	error,
) {
	if r.ObjectChanges == nil {
		return nil, errors.New("expected ObjectChanges != nil")
	}

	var ref *iotago.ObjectRef
	var prevMutatedObj any

	for _, change := range r.ObjectChanges {
		if change.Data.Mutated != nil {
			if change.Data.Mutated.ObjectID == objectID {
				if ref != nil {
					return nil, fmt.Errorf("multiple mutated objects found for %v: first = %v, second = %v",
						objectID.String(),
						string(lo.Must(json.Marshal(prevMutatedObj))),
						string(lo.Must(json.Marshal(change.Data.Mutated))),
					)
				}

				ref = &iotago.ObjectRef{
					ObjectID: &change.Data.Mutated.ObjectID,
					Version:  change.Data.Mutated.Version.Uint64(),
					Digest:   &change.Data.Mutated.Digest,
				}
				prevMutatedObj = change.Data.Mutated
			}
		}
	}

	if ref == nil {
		return nil, fmt.Errorf("not found")
	}
	return ref, nil
}

// requires `ShowObjectChanges: true`
func (r *IotaTransactionBlockResponse) GetCreatedCoinByType(module string, coinType string) (
	*iotago.ObjectRef,
	error,
) {
	if r.ObjectChanges == nil {
		return nil, errors.New("expected ObjectChanges != nil")
	}

	var ref *iotago.ObjectRef
	var prevCreatedObj any

	for _, change := range r.ObjectChanges {
		if change.Data.Created != nil {
			resource, err := iotago.NewResourceType(change.Data.Created.ObjectType)
			if err != nil {
				return nil, fmt.Errorf("invalid resource string: %w", err)
			}
			if resource.Module == "coin" && resource.SubType1 != nil {
				if resource.SubType1.Module == module && resource.SubType1.ObjectName == coinType {
					if ref != nil {
						return nil, fmt.Errorf("multiple created coins found for %s::%s: first = %v, second = %v",
							module, coinType,
							string(lo.Must(json.Marshal(prevCreatedObj))),
							string(lo.Must(json.Marshal(change.Data.Created))),
						)
					}

					ref = &iotago.ObjectRef{
						ObjectID: &change.Data.Created.ObjectID,
						Version:  change.Data.Created.Version.Uint64(),
						Digest:   &change.Data.Created.Digest,
					}
					prevCreatedObj = change.Data.Created
				}
			}
		}
	}

	if ref == nil {
		return nil, fmt.Errorf("not found")
	}
	return ref, nil
}

// requires `ShowObjectChanges: true`
func (r *IotaTransactionBlockResponse) GetMutatedCoinByType(module string, coinType string) (
	*iotago.ObjectRef,
	error,
) {
	if r.ObjectChanges == nil {
		return nil, errors.New("expected ObjectChanges != nil")
	}

	var ref *iotago.ObjectRef
	var prevMutatedObj any

	for _, change := range r.ObjectChanges {
		if change.Data.Mutated != nil {
			resource, err := iotago.NewResourceType(change.Data.Mutated.ObjectType)
			if err != nil {
				return nil, fmt.Errorf("invalid resource string: %w", err)
			}
			if resource.Module == "coin" && resource.SubType1 != nil {
				if resource.SubType1.Module == module && resource.SubType1.ObjectName == coinType {
					if ref != nil {
						return nil, fmt.Errorf("multiple mutated coins found for %s::%s: first = %v, second = %v",
							module, coinType,
							string(lo.Must(json.Marshal(prevMutatedObj))),
							string(lo.Must(json.Marshal(change.Data.Mutated))),
						)
					}

					ref = &iotago.ObjectRef{
						ObjectID: &change.Data.Mutated.ObjectID,
						Version:  change.Data.Mutated.Version.Uint64(),
						Digest:   &change.Data.Mutated.Digest,
					}
					prevMutatedObj = change.Data.Mutated
				}
			}
		}
	}

	if ref == nil {
		return nil, errors.New("not found")
	}
	return ref, nil
}

type (
	ReturnValueType            interface{}
	MutableReferenceOutputType interface{}
	ExecutionResultType        struct {
		MutableReferenceOutputs []MutableReferenceOutputType `json:"mutableReferenceOutputs,omitempty"`
		ReturnValues            []ReturnValueType            `json:"returnValues,omitempty"`
	}
)

type DevInspectResults struct {
	Effects serialization.TagJson[IotaTransactionBlockEffects] `json:"effects"`
	Events  []IotaEvent                                        `json:"events"`
	Results []ExecutionResultType                              `json:"results,omitempty"`
	Error   string                                             `json:"error,omitempty"`
}

type TransactionFilter struct {
	/// Query by checkpoint.
	Checkpoint *BigInt `json:"Checkpoint,omitempty"`
	/// Query by move function.
	MoveFunction *TransactionFilterMoveFunction `json:"MoveFunction,omitempty"`
	/// Query by input object.
	InputObject *iotago.ObjectID `json:"InputObject,omitempty"`
	/// Query by changed object, including created, mutated and unwrapped objects.
	ChangedObject *iotago.ObjectID `json:"ChangedObject,omitempty"`
	/// Query by sender address.
	FromAddress *iotago.Address `json:"FromAddress,omitempty"`
	/// Query by recipient address.
	ToAddress *iotago.Address `json:"ToAddress,omitempty"`
	/// Query by sender and recipient address.
	FromAndToAddress *TransactionFilterFromAndToAddress `json:"FromAndToAddress,omitempty"`
	/// Query txs that have a given address as sender or recipient.
	FromOrToAddress *iotago.Address `json:"FromOrToAddress,omitempty"`
	/// Query by transaction kind
	TransactionKind *string `json:"TransactionKind,omitempty"`
	/// Query transactions of any given kind in the input.
	TransactionKindIn []string `json:"TransactionKindIn,omitempty"`
}

type TransactionFilterMoveFunction struct {
	Package  iotago.ObjectID   `json:"package"`
	Module   iotago.Identifier `json:"module,omitempty"`
	Function iotago.Identifier `json:"function,omitempty"`
}

type TransactionFilterFromAndToAddress struct {
	From *iotago.Address `json:"from"`
	To   *iotago.Address `json:"to"`
}

type IotaTransactionBlockResponseOptions struct {
	// Whether to show transaction input data. Default to be False
	ShowInput bool `json:"showInput,omitempty"`
	// Whether to show bcs-encoded transaction input data
	ShowRawInput bool `json:"showRawInput,omitempty"`
	// Whether to show transaction effects. Default to be False
	ShowEffects bool `json:"showEffects,omitempty"`
	// Whether to show transaction events. Default to be False
	ShowEvents bool `json:"showEvents,omitempty"`
	// Whether to show object_changes. Default to be False
	ShowObjectChanges bool `json:"showObjectChanges,omitempty"`
	// Whether to show balance_changes. Default to be False
	ShowBalanceChanges bool `json:"showBalanceChanges,omitempty"`
	// Whether to show raw transaction effects. Default to be False
	ShowRawEffects bool `json:"showRawEffects,omitempty"`
}

type IotaTransactionBlockResponseQuery struct {
	Filter  *TransactionFilter                   `json:"filter,omitempty"`
	Options *IotaTransactionBlockResponseOptions `json:"options,omitempty"`
}

type TransactionBlocksPage = Page[IotaTransactionBlockResponse, iotago.TransactionDigest]

type DryRunTransactionBlockResponse struct {
	Effects        serialization.TagJson[IotaTransactionBlockEffects] `json:"effects"`
	Events         []IotaEvent                                        `json:"events"`
	ObjectChanges  []serialization.TagJson[ObjectChange]              `json:"objectChanges"`
	BalanceChanges []BalanceChange                                    `json:"balanceChanges"`
	Input          serialization.TagJson[IotaTransactionBlockData]    `json:"input"`
}
