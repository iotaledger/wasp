package iotago

import (
	"fmt"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/packages/hashing"
)

type TransactionData struct {
	V1 *TransactionDataV1
}

func (t TransactionData) IsBcsEnum() {}

// Generates the transaction digest of the current TransactionData
func (t TransactionData) Digest() (*Digest, error) {
	buf := []byte("TransactionData::")
	txnBytes, err := bcs.Marshal(&t)
	if err != nil {
		return nil, fmt.Errorf("failed to encode into BCS: %w", err)
	}
	buf = append(buf, txnBytes...)
	h := hashing.HashDataBlake2b(buf)
	return DigestFromBytes(h.Bytes()), nil
}

type TransactionDataV1 struct {
	Kind       TransactionKind
	Sender     Address
	GasData    GasData
	Expiration TransactionExpiration
}

type TransactionExpiration struct {
	None  *serialization.EmptyEnum
	Epoch *EpochId
}

func (t TransactionExpiration) IsBcsEnum() {}

type GasData struct {
	Payment []*ObjectRef
	Owner   *Address
	Price   uint64
	Budget  uint64
}

type TransactionKind struct {
	ProgrammableTransaction *ProgrammableTransaction
	ChangeEpoch             *ChangeEpoch
	Genesis                 *GenesisTransaction
	ConsensusCommitPrologue *ConsensusCommitPrologue
}

func (t TransactionKind) IsBcsEnum() {}

type ConsensusCommitPrologue struct {
	Epoch             uint64
	Round             uint64
	CommitTimestampMs CheckpointTimestamp
}

type ProgrammableTransaction struct {
	Inputs   []CallArg
	Commands []Command
}

func (p ProgrammableTransaction) Print(prefix string) {
	fmt.Print(prefix + "ProgrammableTransaction:\n")
	for i, input := range p.Inputs {
		fmt.Printf("%s  input %2d: %s\n", prefix, i, input.String())
	}
	for i, cmd := range p.Commands {
		fmt.Printf("%s  command %2d: %s\n", prefix, i, cmd.String())
	}
}

func (p ProgrammableTransaction) IsInInputObjects(targetID *ObjectID) bool {
	for _, input := range p.Inputs {
		if input.Object != nil && input.Object.ID().Equals(*targetID) {
			return true
		}
	}
	return false
}

func (p ProgrammableTransaction) EstimateGasBudget(gasPrice uint64) uint64 {
	// TODO: Implement this function!
	return 10_000_000 // iotaclient.DefaultGasBudget
}

type TransferObject struct {
	Recipient Address
	ObjectRef ObjectRef
}

type MoveModulePublish struct {
	Modules [][]byte
}

type MoveCall struct {
	Package       ObjectID
	Module        string
	Function      string
	TypeArguments []*TypeTag
	Arguments     []*CallArg
}

type TransferIota struct {
	Recipient Address
	Amount    *uint64 `bcs:"optional"`
}

type Pay struct {
	Coins      []*ObjectRef
	Recipients []*Address
	Amounts    []*uint64
}

type PayIota = Pay

type PayAllIota struct {
	Coins     []*ObjectRef
	Recipient Address
}

type ChangeEpoch struct {
	Epoch                   EpochId
	ProtocolVersion         ProtocolVersion
	StorageCharge           uint64
	ComputationCharge       uint64
	StorageRebate           uint64
	NonRefundableStorageFee uint64
	EpochStartTimestampMs   uint64
	SystemPackages          []*struct {
		SequenceNumber SequenceNumber
		Bytes          [][]uint8
		Objects        []*ObjectID
	}
}

type GenesisTransaction struct {
	Objects []GenesisObject
}

type GenesisObject struct {
	RawObject *struct {
		Data  Data
		Owner Owner
	}
}

type CallArg struct {
	Pure   *[]byte
	Object *ObjectArg
}

func (c CallArg) IsBcsEnum() {}

func (c CallArg) String() string {
	if c.Pure != nil {
		return fmt.Sprintf("Pure(0x%x)", *c.Pure)
	}
	if c.Object != nil {
		return fmt.Sprintf("Object(%v)", *c.Object)
	}
	panic("invalid argument")
}

type ObjectArg struct {
	ImmOrOwnedObject *ObjectRef
	SharedObject     *SharedObjectArg
	Receiving        *ObjectRef
}

type SharedObjectArg struct {
	Id                   *ObjectID
	InitialSharedVersion SequenceNumber
	Mutable              bool
}

func (o ObjectArg) IsBcsEnum() {}

func (o ObjectArg) ID() *ObjectID {
	switch {
	case o.ImmOrOwnedObject != nil:
		return o.ImmOrOwnedObject.ObjectID
	case o.SharedObject != nil:
		return o.SharedObject.Id
	case o.Receiving != nil:
		return o.Receiving.ObjectID
	default:
		return &ObjectID{}
	}
}
