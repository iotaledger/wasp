// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/chain/cmt_log"
)

func TestLogIndex(t *testing.T) {
	require.Equal(t, uint32(1), cmt_log.NilLogIndex().Next().AsUint32())
}
