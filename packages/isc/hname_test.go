// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isc_test

import (
	"math/rand"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

func TestHnameSerialize(t *testing.T) {
	hname := isc.Hname(rand.Uint32())
	bcs.TestCodec(t, hname)
	rwutil.StringTest(t, hname, isc.HnameFromString)

	bcs.TestCodecAndHash(t, isc.Hname(12345678), "c24fc2b805ef")
}
