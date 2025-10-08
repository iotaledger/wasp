// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
)

func TestLogIndex(t *testing.T) {
	require.Equal(t, uint32(1), committeelog.NilLogIndex().Next().AsUint32())
}
