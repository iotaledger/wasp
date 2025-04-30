// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtlog_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/cmtlog"
)

func TestLogIndex(t *testing.T) {
	require.Equal(t, uint32(1), cmtlog.NilLogIndex().Next().AsUint32())
}
