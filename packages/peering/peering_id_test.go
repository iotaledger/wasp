// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestPeeringIDSerialization(t *testing.T) {
	peeringID := peering.RandomPeeringID()

	bcs.TestCodec(t, &peeringID)
}
