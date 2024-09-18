// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc_test

import (
	"crypto/rand"
	mathrand "math/rand"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestAliasOutputWithIDSerialization(t *testing.T) {
	output := iotago.AliasOutput{
		Amount:        mathrand.Uint64(),
		StateIndex:    mathrand.Uint32(),
		StateMetadata: []byte{1, 2, 3},
	}
	rand.Read(output.AliasID[:])
	outputID := iotago.OutputID{}
	rand.Read(outputID[:])
	aliasOutputWithID := isc.NewAliasOutputWithID(&output, outputID)
	bcs.TestCodec(t, output)
	rwutil.ReadWriteTest(t, aliasOutputWithID, new(isc.AliasOutputWithID))
	rwutil.BytesTest(t, aliasOutputWithID, isc.AliasOutputWithIDFromBytes)
}
