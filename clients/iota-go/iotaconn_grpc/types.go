package iotaconn_grpc

import "github.com/iotaledger/wasp/clients/iota-go/iotago"

type StructTag struct {
	Address [32]byte
	Module  string
	Name    string
	Params  []iotago.TypeTag
}

type RpcEventID struct {
	Digest        iotago.Base58
	EventSequence uint64
}

type IotaRpcEvent struct {
	EventID RpcEventID
	// Move module where this event was emitted.
	PackageID iotago.PackageID
	// Sender's Iota iotago.address.
	TransactionModule string
	// Move event type.
	Sender iotago.Address
	// Parseof the event
	Type       StructTag
	ParsedJson []byte
	Bcs        []byte
	Timestamp  *uint64 `bcs:"optional"`
}
