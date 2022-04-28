// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreblocklog"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func setupBlockLog(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreblocklog.ScName, coreblocklog.OnLoad)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestBlockLogXxx(t *testing.T) {
	ctx := setupBlockLog(t)
	require.NoError(t, ctx.Err)
}
