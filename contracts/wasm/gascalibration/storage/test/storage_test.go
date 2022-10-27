// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/wasm/gascalibration"
	"github.com/iotaledger/wasp/contracts/wasm/gascalibration/storage/go/storage"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

var force = flag.Bool("force", false, "")

func TestCallF(t *testing.T) {
	if !*force {
		t.SkipNow()
	}

	wasmlib.ConnectHost(nil)
	ctx := wasmsolo.NewSoloContext(t, storage.ScName, storage.OnDispatch)
	require.NoError(t, ctx.Err)

	results := make(map[uint32]uint64)
	f := storage.ScFuncs.F(ctx)
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

	filePath := "../pkg/storage_" + contractVersion + ".json"
	gascalibration.SaveTestResultAsJSON(filePath, results)
}
