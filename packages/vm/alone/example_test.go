// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package alone

import (
	"testing"

	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/stretchr/testify/require"
)

// a very simple test using 'alone' tool

func TestNew(t *testing.T) {
	glb := New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	chain.CheckBase()
	chain.Infof("\n%s\n", chain.String())

	req := NewCall(blob.Interface.Name, "init")
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)
}
