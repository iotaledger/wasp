// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func TestBigAdd(t *testing.T) {
	ctx := setupTest(t)

	res := bigAdd(t, ctx, wasmtypes.NewScBigInt(), wasmtypes.NewScBigInt())
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigAdd(t, ctx, wasmtypes.NewScBigInt(1), wasmtypes.NewScBigInt(1))
	require.False(t, res.IsZero())
	require.EqualValues(t, "2", res.String())

	//for lhs := 0; lhs <= 2048; lhs++ {
	//	for rhs := 1; rhs <= lhs; rhs++ {
	//		bigAdd64(t, ctx, uint64(lhs), uint64(rhs))
	//	}
	//}
	for lhs := 3; lhs < 1_000_000_000; lhs = lhs*2 + 1 {
		for rhs := 1; rhs < lhs; rhs = rhs*2 + 1 {
			bigAdd64(t, ctx, uint64(lhs), uint64(rhs))
		}
	}
}

func TestBigSub(t *testing.T) {
	ctx := setupTest(t)

	res := bigSub(t, ctx, wasmtypes.NewScBigInt(), wasmtypes.NewScBigInt())
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigSub(t, ctx, wasmtypes.NewScBigInt(1), wasmtypes.NewScBigInt(1))
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())

	//for lhs := 0; lhs <= 2048; lhs++ {
	//	for rhs := 1; rhs <= lhs; rhs++ {
	//		bigSub64(t, ctx, uint64(lhs), uint64(rhs))
	//	}
	//}
	for lhs := 3; lhs < 1_000_000_000; lhs = lhs*2 + 1 {
		for rhs := 1; rhs < lhs; rhs = rhs*2 + 1 {
			bigSub64(t, ctx, uint64(lhs), uint64(rhs))
		}
	}
}

func TestBigMul(t *testing.T) {
	ctx := setupTest(t)

	res := bigMul(t, ctx, wasmtypes.NewScBigInt(), wasmtypes.NewScBigInt())
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigMul(t, ctx, wasmtypes.NewScBigInt(1), wasmtypes.NewScBigInt(1))
	require.False(t, res.IsZero())
	require.EqualValues(t, "1", res.String())

	//for lhs := 0; lhs <= 2048; lhs++ {
	//	for rhs := 1; rhs <= lhs; rhs++ {
	//		bigMul64(t, ctx, uint64(lhs), uint64(rhs))
	//	}
	//}
	for lhs := 3; lhs < 1_000_000_000; lhs = lhs*2 + 1 {
		for rhs := 1; rhs < lhs; rhs = rhs*2 + 1 {
			bigMul64(t, ctx, uint64(lhs), uint64(rhs))
		}
	}
}

func TestBigDiv(t *testing.T) {
	ctx := setupTest(t)

	bigDiv64(t, ctx, 262143, 511)

	res := bigDiv(t, ctx, wasmtypes.NewScBigInt(), wasmtypes.NewScBigInt(1))
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigDiv(t, ctx, wasmtypes.NewScBigInt(1), wasmtypes.NewScBigInt(1))
	require.False(t, res.IsZero())
	require.EqualValues(t, "1", res.String())

	//for lhs := 0; lhs <= 2048; lhs++ {
	//	for rhs := 1; rhs <= lhs; rhs++ {
	//		bigDiv64(t, ctx, uint64(lhs), uint64(rhs))
	//	}
	//}
	//for lhs := 0; lhs <= 66666; lhs+= 17 {
	//	for rhs := 1; rhs <= lhs; rhs += 13 {
	//		bigDiv64(t, ctx, uint64(lhs), uint64(rhs))
	//	}
	//}
	for lhs := 3; lhs < 1_000_000_000; lhs = lhs*2 + 1 {
		for rhs := 1; rhs < lhs; rhs = rhs*2 + 1 {
			bigDiv64(t, ctx, uint64(lhs), uint64(rhs))
		}
	}
}

func TestBigMod(t *testing.T) {
	ctx := setupTest(t)

	res := bigMod(t, ctx, wasmtypes.NewScBigInt(), wasmtypes.NewScBigInt(1))
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigMod(t, ctx, wasmtypes.NewScBigInt(1), wasmtypes.NewScBigInt(1))
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())

	//for lhs := 0; lhs <= 2048; lhs++ {
	//	for rhs := 1; rhs <= lhs; rhs++ {
	//		bigMod64(t, ctx, uint64(lhs), uint64(rhs))
	//	}
	//}
	//for lhs := 0; lhs <= 66666; lhs+= 17 {
	//	for rhs := 1; rhs <= lhs; rhs += 13 {
	//		bigMod64(t, ctx, uint64(lhs), uint64(rhs))
	//	}
	//}
	for lhs := 3; lhs < 1_000_000_000; lhs = lhs*2 + 1 {
		for rhs := 1; rhs < lhs; rhs = rhs*2 + 1 {
			bigMod64(t, ctx, uint64(lhs), uint64(rhs))
		}
	}
}

func TestBigShl(t *testing.T) {
	ctx := setupTest(t)

	res := bigShl(t, ctx, wasmtypes.NewScBigInt(), 0)
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigShl(t, ctx, wasmtypes.NewScBigInt(1), 1)
	require.False(t, res.IsZero())
	require.EqualValues(t, "2", res.String())

	for lhs := 3; lhs < 1_000_000_000; lhs = lhs*2 + 1 {
		for rhs := 1; rhs < lhs && rhs < 256; rhs = rhs*2 + 1 {
			bigShl64(t, ctx, uint64(lhs), uint32(rhs))
		}
	}
}

func TestBigShr(t *testing.T) {
	ctx := setupTest(t)

	res := bigShr(t, ctx, wasmtypes.NewScBigInt(), 0)
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigShr(t, ctx, wasmtypes.NewScBigInt(1), 1)
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())

	for lhs := 3; lhs < 1_000_000_000; lhs = lhs*2 + 1 {
		for rhs := 1; rhs < lhs && rhs < 256; rhs = rhs*2 + 1 {
			bigShr64(t, ctx, uint64(lhs), uint32(rhs))
		}
	}
}

func bigAdd(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs wasmtypes.ScBigInt) wasmtypes.ScBigInt {
	f := testwasmlib.ScFuncs.BigIntAdd(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Rhs().SetValue(rhs)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
	// return lhs.Add(rhs)
}

func bigSub(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs wasmtypes.ScBigInt) wasmtypes.ScBigInt {
	f := testwasmlib.ScFuncs.BigIntSub(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Rhs().SetValue(rhs)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
	// return lhs.Sub(rhs)
}

func bigMul(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs wasmtypes.ScBigInt) wasmtypes.ScBigInt {
	f := testwasmlib.ScFuncs.BigIntMul(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Rhs().SetValue(rhs)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
	// return lhs.Mul(rhs)
}

func bigDiv(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs wasmtypes.ScBigInt) wasmtypes.ScBigInt {
	f := testwasmlib.ScFuncs.BigIntDiv(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Rhs().SetValue(rhs)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
	// return lhs.Div(rhs)
}

func bigMod(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs wasmtypes.ScBigInt) wasmtypes.ScBigInt {
	f := testwasmlib.ScFuncs.BigIntMod(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Rhs().SetValue(rhs)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
	// return lhs.Modulo(rhs)
}

func bigShl(t *testing.T, ctx *wasmsolo.SoloContext, lhs wasmtypes.ScBigInt, shift uint32) wasmtypes.ScBigInt {
	f := testwasmlib.ScFuncs.BigIntShl(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Shift().SetValue(shift)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
}

func bigShr(t *testing.T, ctx *wasmsolo.SoloContext, lhs wasmtypes.ScBigInt, shift uint32) wasmtypes.ScBigInt {
	f := testwasmlib.ScFuncs.BigIntShr(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Shift().SetValue(shift)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
}

func bigAdd64(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs uint64) {
	expect := lhs + rhs
	t.Logf("%d + %d = %d\n", lhs, rhs, expect)
	res := bigAdd(t, ctx, wasmtypes.NewScBigInt(lhs), wasmtypes.NewScBigInt(rhs))
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}

func bigSub64(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs uint64) {
	expect := lhs - rhs
	t.Logf("%d - %d = %d\n", lhs, rhs, expect)
	res := bigSub(t, ctx, wasmtypes.NewScBigInt(lhs), wasmtypes.NewScBigInt(rhs))
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}

func bigMul64(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs uint64) {
	expect := lhs * rhs
	t.Logf("%d * %d = %d\n", lhs, rhs, expect)
	res := bigMul(t, ctx, wasmtypes.NewScBigInt(lhs), wasmtypes.NewScBigInt(rhs))
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}

func bigDiv64(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs uint64) {
	expect := lhs / rhs
	t.Logf("%d / %d = %d\n", lhs, rhs, expect)
	res := bigDiv(t, ctx, wasmtypes.NewScBigInt(lhs), wasmtypes.NewScBigInt(rhs))
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}

func bigMod64(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs uint64) {
	expect := lhs % rhs
	t.Logf("%d %% %d = %d\n", lhs, rhs, expect)
	res := bigMod(t, ctx, wasmtypes.NewScBigInt(lhs), wasmtypes.NewScBigInt(rhs))
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}

func bigShl64(t *testing.T, ctx *wasmsolo.SoloContext, lhs uint64, shift uint32) {
	expect := lhs << shift
	t.Logf("%d << %d = %d\n", lhs, shift, expect)
	res := bigShl(t, ctx, wasmtypes.NewScBigInt(lhs), shift)
	if res.IsUint64() {
		require.EqualValues(t, expect, res.Uint64())
		require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
		require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
		return
	}
	// t.Logf("%d << %d = %s\n", lhs, shift, res.String())
}

func bigShr64(t *testing.T, ctx *wasmsolo.SoloContext, lhs uint64, shift uint32) {
	expect := lhs >> shift
	t.Logf("%d >> %d = %d\n", lhs, shift, expect)
	res := bigShr(t, ctx, wasmtypes.NewScBigInt(lhs), shift)
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}
