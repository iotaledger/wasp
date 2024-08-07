package main

import iotago "github.com/iotaledger/iota.go/v3"

type foundryOutputRec struct {
	OutputID    iotago.OutputID
	Amount      uint64 // always storage deposit
	TokenScheme iotago.TokenScheme
	Metadata    []byte
}
