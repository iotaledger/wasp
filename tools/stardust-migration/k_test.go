package main

import (
	"testing"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
)

func TestA(t *testing.T) {
	lo.Must(iotago.NewDigest("MIGRATED"))
}
