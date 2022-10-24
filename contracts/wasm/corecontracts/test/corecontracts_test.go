// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/wasm/corecontracts/go/corecontracts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

func setup(t *testing.T) *wasmsolo.SoloContext {
	ctx := wasmsolo.NewSoloContext(t, corecontracts.ScName, corecontracts.OnDispatch)
	require.NoError(t, ctx.ContractExists(corecontracts.ScName))
	return ctx
}

func TestDeploy(t *testing.T) {
	setup(t)
}
