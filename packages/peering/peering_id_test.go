// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering_test

import (
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/peering"
)

func TestPeeringIDSerialization(t *testing.T) {
	peeringID := peering.RandomPeeringID()

	bcs.TestCodec(t, &peeringID)
}
