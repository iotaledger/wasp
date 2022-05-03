// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreblob"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func setupBlob(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreblob.ScName, coreblob.OnLoad)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestBlobXxx(t *testing.T) {
	ctx := setupBlob(t)
	require.NoError(t, ctx.Err)
}
