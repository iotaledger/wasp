// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreroot"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func setupRoot(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreroot.ScName, coreroot.OnLoad)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestRootXxx(t *testing.T) {
	ctx := setupRoot(t)
	require.NoError(t, ctx.Err)
}
