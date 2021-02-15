// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/iotaledger/wasp/contracts/testenv"
	"github.com/iotaledger/wasp/packages/vm/wasmlib"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeployHelloWorld(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_, err := te.Chain.FindContract(ScName)
	require.NoError(t, err)
}

func TestHelloWorld(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncHelloWorld).Post(0)
}

func TestGetHelloWorld(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	res := te.CallView(ViewGetHelloWorld)
	result := te.Results(res)
	hw := result.GetString(wasmlib.Key(VarHelloWorld))
	require.EqualValues(t, "Hello, world!", hw.Value())
}
