package isc

import (
	iotago "github.com/iotaledger/iota.go/v3"
)

const Million = uint64(1_000_000)

var emptyOutputID = iotago.OutputID{}

func IsEmptyOutputID(outputID iotago.OutputID) bool {
	return outputID == emptyOutputID
}
