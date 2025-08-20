// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package peering_test

import (
	"crypto/ed25519"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/testutil/testval"
)

func TestPeeringIDSerialization(t *testing.T) {
	peeringID := peering.RandomPeeringID()
	bcs.TestCodec(t, &peeringID)
	bcs.TestCodecAndHash(t, peering.PeeringID(testval.TestBytes(ed25519.PublicKeySize)), "b4ff315a20ce")
}
