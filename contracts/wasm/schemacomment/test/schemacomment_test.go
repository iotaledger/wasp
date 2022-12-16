// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/schemacomment/go/schemacomment"
	"github.com/iotaledger/wasp/contracts/wasm/schemacomment/go/schemacommentimpl"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, schemacomment.ScName, schemacommentimpl.OnDispatch)
	require.NoError(t, ctx.ContractExists(schemacomment.ScName))
}
