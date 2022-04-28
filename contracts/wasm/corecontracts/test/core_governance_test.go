// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coregovernance"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func setupGovernance(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coregovernance.ScName, coregovernance.OnLoad)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestGovernanceXxx(t *testing.T) {
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)
}
