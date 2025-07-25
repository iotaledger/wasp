package iotago

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
)

// https://sdk.mystenlabs.com/typescript/transaction-building/basics#object-references
// https://docs.sui.io/concepts/transactions/prog-txn-blocks
type Command struct {
	MoveCall        *ProgrammableMoveCall
	TransferObjects *ProgrammableTransferObjects
	SplitCoins      *ProgrammableSplitCoins
	MergeCoins      *ProgrammableMergeCoins
	// `Publish` publishes a Move package. Returns the upgrade capability object.
	Publish *ProgrammablePublish
	// `MakeMoveVec` constructs a vector of objects that can be passed into a moveCall.
	// This is required as there’s no way to define a vector as an input.
	MakeMoveVec *ProgrammableMakeMoveVec
	// upgrades a Move package
	Upgrade *ProgrammableUpgrade
}

func (c Command) IsBcsEnum() {}

func (c Command) String() string {
	if c.MoveCall != nil {
		return c.MoveCall.String()
	}
	if c.TransferObjects != nil {
		return c.TransferObjects.String()
	}
	if c.SplitCoins != nil {
		return c.SplitCoins.String()
	}
	if c.MergeCoins != nil {
		return c.MergeCoins.String()
	}
	if c.Publish != nil {
		return c.Publish.String()
	}
	if c.MakeMoveVec != nil {
		return c.MakeMoveVec.String()
	}
	if c.Upgrade != nil {
		return c.Upgrade.String()
	}
	panic("invalid command")
}

type Argument struct {
	/// The gas coin. The gas coin can only be used by-ref, except for with
	/// `TransferObjects`, which can use it by-value.
	GasCoin *serialization.EmptyEnum
	// One of the input objects or primitive values (from `ProgrammableTransaction` inputs)
	Input *uint16
	// The result of another transaction (from `ProgrammableTransaction` transactions)
	Result *uint16
	// Like a `Result` but it accesses a nested result. Currently, the only usage of this is to access a
	// value from a Move call with multiple return values.
	NestedResult *NestedResult
}
type NestedResult struct {
	Cmd    uint16 // command index
	Result uint16 // result index
}

func (nr *NestedResult) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as an array first
	var arr []uint16
	if err := json.Unmarshal(data, &arr); err == nil {
		if len(arr) == 2 {
			nr.Cmd = arr[0]
			nr.Result = arr[1]
			return nil
		}
		return fmt.Errorf("NestedResult array must have exactly 2 elements, got %d", len(arr))
	}

	// If not an array, try as a regular struct
	type NestedResultAlias NestedResult
	return json.Unmarshal(data, (*NestedResultAlias)(nr))
}

func (a Argument) IsBcsEnum() {}

func (a Argument) String() string {
	if a.GasCoin != nil {
		return "GasCoin"
	}
	if a.Input != nil {
		return fmt.Sprintf("Input(%d)", *a.Input)
	}
	if a.Result != nil {
		return fmt.Sprintf("Result(%d)", *a.Result)
	}
	if a.NestedResult != nil {
		return fmt.Sprintf("NestedResult(%d, %d)", a.NestedResult.Cmd, a.NestedResult.Result)
	}
	panic("invalid argument")
}

func GetArgumentGasCoin() Argument {
	return Argument{GasCoin: &serialization.EmptyEnum{}}
}

type ProgrammableMoveCall struct {
	Package       *PackageID
	Module        Identifier
	Function      Identifier
	TypeArguments []TypeTag `json:"type_arguments"`
	Arguments     []Argument
}

func (p *ProgrammableMoveCall) String() string {
	return fmt.Sprintf(
		"MoveCall: %s::%s<%s>(%s)",
		p.Module,
		p.Function,
		strings.Join(lo.Map(p.TypeArguments, func(t TypeTag, _ int) string { return t.String() }), ", "),
		strings.Join(lo.Map(p.Arguments, func(a Argument, _ int) string { return a.String() }), ", "),
	)
}

type ProgrammableTransferObjects struct {
	Objects []Argument
	Address Argument
}

func (p *ProgrammableTransferObjects) String() string {
	return "TransferObjects"
}

type ProgrammableSplitCoins struct {
	Coin    Argument
	Amounts []Argument
}

func (p *ProgrammableSplitCoins) String() string {
	return "SplitCoins"
}

type ProgrammableMergeCoins struct {
	Destination Argument
	Sources     []Argument
}

func (p *ProgrammableMergeCoins) String() string {
	return "MergeCoins"
}

type ProgrammablePublish struct {
	Modules      [][]byte
	Dependencies []*ObjectID
}

func (p *ProgrammablePublish) String() string {
	return "Publish"
}

type ProgrammableMakeMoveVec struct {
	Type    *TypeTag `bcs:"optional"`
	Objects []Argument
}

func (p *ProgrammableMakeMoveVec) String() string {
	return "MakeMoveVec"
}

type ProgrammableUpgrade struct {
	Modules      [][]byte
	Dependencies []*ObjectID
	PackageID    *PackageID
	Ticket       Argument
}

func (p *ProgrammableUpgrade) String() string {
	return "Upgrade"
}
