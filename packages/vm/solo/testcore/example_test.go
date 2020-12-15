// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/solo"
	"github.com/stretchr/testify/require"
	"testing"
)

// a very simple test using 'alone' tool

func TestNew(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	chain.CheckBase()
	chain.Infof("\n%s\n", chain.String())

	req := solo.NewCall(blob.Interface.Name, "init")
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)
}
