// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package alone

import (
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/stretchr/testify/require"
	"testing"
)

// a very simple test using 'alone' tool

func TestNew(t *testing.T) {
	e := New(t, false, false)
	e.CheckBase()
	e.Infof("\n%s\n", e.String())

	req := NewCall(blob.Interface.Name, "init")
	_, err := e.PostRequest(req, nil)
	require.Error(t, err)
}
