// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"flag"
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/gascalibration"
	"github.com/iotaledger/wasp/contracts/wasm/gascalibration/memory/go/memory"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

var force = flag.Bool("force", false, "")

func deployContract(t *testing.T) *wasmsolo.SoloContext {
	ctx := wasmsolo.NewSoloContext(t, memory.ScName, memory.OnLoad)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestCallF(t *testing.T) {
	// running these tests should be intentional
	if !*force {
		t.SkipNow()
	}
	wasmlib.ConnectHost(nil)
	ctx := deployContract(t)
	f := memory.ScFuncs.F(ctx)

	results := make(map[uint32]uint64)
	for i := uint32(1); i <= 100; i++ {
		n := i * 10
		f.Params.N().SetValue(n)
		f.Func.Post()
		require.NoError(t, ctx.Err)
		t.Logf("n = %d, gas = %d\n", n, ctx.Gas)
		results[n] = ctx.Gas
	}

	contractVersion := ""
	if *wasmsolo.GoWasm {
		contractVersion = "go"
	} else if *wasmsolo.TsWasm {
		contractVersion = "ts"
	} else if *wasmsolo.RsWasm {
		contractVersion = "rs"
	}
	t.Logf("Running %s version of contract", contractVersion)

	filePath := "../pkg/memory_" + contractVersion + ".json"
	gascalibration.SaveTestResultAsJSON(filePath, results)
}
