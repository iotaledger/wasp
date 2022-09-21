// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/corecontracts/go/corecontracts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) *wasmsolo.SoloContext {
	ctx := wasmsolo.NewSoloContext(t, corecontracts.ScName, corecontracts.OnLoad)
	require.NoError(t, ctx.ContractExists(corecontracts.ScName))
	return ctx
}

func TestDeploy(t *testing.T) {
	setup(t)
}
