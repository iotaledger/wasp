// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
)

func TestLogIndex(t *testing.T) {
	require.Equal(t, uint32(1), cmtLog.NilLogIndex().Next().AsUint32())
}
