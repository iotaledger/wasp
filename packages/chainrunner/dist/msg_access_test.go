// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dist

import (
	"math"
	"math/rand"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
)

func TestMsgAccessSerialization(t *testing.T) {
	msg := &msgAccess{
		gpa.BasicMessage{},
		rand.Intn(math.MaxUint32 + 1),
		rand.Intn(math.MaxUint32 + 1),
		[]isc.ChainID{isctest.RandomChainID(), isctest.RandomChainID()},
		[]isc.ChainID{isctest.RandomChainID(), isctest.RandomChainID()},
	}

	bcs.TestCodec(t, msg)

	msg = &msgAccess{
		gpa.BasicMessage{},
		math.MaxUint32,
		math.MaxUint32,
		[]isc.ChainID{isctest.RandomChainID(), isctest.RandomChainID()},
		[]isc.ChainID{isctest.RandomChainID(), isctest.RandomChainID()},
	}

	bcs.TestCodec(t, msg)
}
