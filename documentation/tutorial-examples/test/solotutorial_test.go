// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/documentation/tutorial-examples/go/solotutorial"
	"github.com/iotaledger/wasp/documentation/tutorial-examples/go/solotutorialimpl"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

func TestDeploy(t *testing.T) {
	ctx := wasmsolo.NewSoloContext(t, solotutorial.ScName, solotutorialimpl.OnDispatch)
	require.NoError(t, ctx.ContractExists(solotutorial.ScName))
}
