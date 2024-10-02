// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package am_dist

import (
	"math"
	"math/rand"
	"testing"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestMsgAccessSerialization(t *testing.T) {
	msg := &msgAccess{
		gpa.BasicMessage{},
		rand.Intn(math.MaxUint32 + 1),
		rand.Intn(math.MaxUint32 + 1),
		[]isc.ChainID{isc.RandomChainID(), isc.RandomChainID()},
		[]isc.ChainID{isc.RandomChainID(), isc.RandomChainID()},
	}

	bcs.TestCodec(t, msg)

	msg = &msgAccess{
		gpa.BasicMessage{},
		math.MaxUint32,
		math.MaxUint32,
		[]isc.ChainID{isc.RandomChainID(), isc.RandomChainID()},
		[]isc.ChainID{isc.RandomChainID(), isc.RandomChainID()},
	}

	bcs.TestCodec(t, msg)
}
