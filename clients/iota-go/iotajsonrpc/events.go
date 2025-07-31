package iotajsonrpc

import (
	"encoding/json"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type EventId struct {
	TxDigest iotago.TransactionDigest `json:"txDigest"`
	EventSeq *BigInt                  `json:"eventSeq"`
}

type IotaEvent struct {
	Id EventId `json:"id"`
	// Move package where this event was emitted.
	PackageId *iotago.ObjectID `json:"packageId"`
	// Move module where this event was emitted.
	TransactionModule iotago.Identifier `json:"transactionModule"`
	// Sender's Iota iotago.address.
	Sender *iotago.Address `json:"sender"`
	// Move event type.
	Type *iotago.StructTag `json:"type"`
	// Parsed json value of the event
	ParsedJson interface{} `json:"parsedJson,omitempty"`
	// Base 64 encoded bcs bytes of the move event
	Bcs         iotago.Base64Data `json:"bcs"`
	TimestampMs *BigInt           `json:"timestampMs,omitempty"`
}

type EventPage = Page[IotaEvent, EventId]

type EventFilter struct {
	/// Query by sender address
	Sender *iotago.Address `json:"Sender,omitempty"`
	/// Return events emitted by the given transaction
	///digest of the transaction, as base-64 encoded string
	Transaction *iotago.TransactionDigest `json:"Transaction,omitempty"`
	/// Return events emitted in a specified Package.
	Package *iotago.ObjectID `json:"Package,omitempty"`
	/// Return events emitted in a specified Move module.
	/// If the event is defined in Module A but emitted in a tx with Module B,
	/// query `MoveModule` by module B returns the event.
	/// Query `MoveEventModule` by module A returns the event too.
	MoveModule *EventFilterMoveModule `json:"MoveModule,omitempty"`
	/// Return events with the given Move event struct name (struct tag).
	/// For example, if the event is defined in `0xabcd::MyModule`, and named
	/// `Foo`, then the struct tag is `0xabcd::MyModule::Foo`.
	MoveEventType  *iotago.StructTag          `json:"MoveEventType,omitempty"`
	MoveEventField *EventFilterMoveEventField `json:"MoveEventField,omitempty"`
	// Return events emitted in [start_time, end_time] interval
	TimeRange *EventFilterTimeRange `json:"TimeRange,omitempty"`

	All *[]EventFilter    `json:"All,omitempty"`
	Any *[]EventFilter    `json:"Any,omitempty"`
	And *AndOrEventFilter `json:"And,omitempty"`
	Or  *AndOrEventFilter `json:"Or,omitempty"`
}

type EventFilterMoveModule struct {
	// the Move package ID
	Package *iotago.ObjectID `json:"package"`
	// the module name
	Module iotago.Identifier `json:"module"`
}

type EventFilterMoveEventField struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type EventFilterTimeRange struct {
	// left endpoint of time interval, milliseconds since epoch, inclusive
	StartTime *BigInt `json:"startTime"`
	// right endpoint of time interval, milliseconds since epoch, exclusive
	EndTime *BigInt `json:"endTime"`
}

type AndOrEventFilter struct {
	Filter1 *EventFilter
	Filter2 *EventFilter
}

func (f AndOrEventFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]interface{}{f.Filter1, f.Filter2})
}
