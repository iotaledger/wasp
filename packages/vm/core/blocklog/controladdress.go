package blocklog

import (
	iotago "github.com/iotaledger/iota.go/v3"
)

type ControlAddresses struct {
	StateAddress     iotago.Address
	GoverningAddress iotago.Address
	SinceBlockIndex  uint32
}
