// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"math/rand"
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testwasmlib/go/testwasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

const (
	ExhaustiveLimit  = 2048
	ExtremeLimit     = 66666
	LogOp            = false
	SkipWasm         = false
	TestExhaustive   = false
	TestExtreme      = false
	TestRandomValues = false
	UpperLimit       = 1_000_000_000
)

func setupBigIntTest(t *testing.T) *wasmsolo.SoloContext {
	//*wasmsolo.TsWasm = true
	if SkipWasm {
		return nil
	}
	return setupTest(t)
}

func testLoop(bigOp64 func(lhs, rhs uint64)) {
	if TestRandomValues {
		bigOp64(0xffffffffffffffff, 0xffff)
		for i := 0; i <= 100000; i++ {
			bigOp64(rand.Uint64(), (rand.Uint64()&0xffff)+1)
		}
	}
	if TestExhaustive {
		for lhs := 0; lhs <= ExhaustiveLimit; lhs++ {
			for rhs := 1; rhs <= lhs; rhs++ {
				bigOp64(uint64(lhs), uint64(rhs))
			}
		}
	}
	if TestExtreme {
		for lhs := 0; lhs <= ExtremeLimit; lhs += 17 {
			for rhs := 1; rhs <= lhs; rhs += 13 {
				bigOp64(uint64(lhs), uint64(rhs))
			}
		}
	}
	for lhs := 3; lhs < UpperLimit; lhs = lhs*2 + 1 {
		for rhs := 1; rhs < lhs; rhs = rhs*2 + 1 {
			bigOp64(uint64(lhs), uint64(rhs))
		}
	}
}

func TestSetupOverhead(t *testing.T) {
	_ = setupBigIntTest(t)
}

func TestBigAdd(t *testing.T) {
	ctx := setupBigIntTest(t)

	res := bigAdd(t, ctx, wasmtypes.NewScBigInt(), wasmtypes.NewScBigInt())
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigAdd(t, ctx, wasmtypes.NewScBigInt(1), wasmtypes.NewScBigInt(1))
	require.False(t, res.IsZero())
	require.EqualValues(t, "2", res.String())

	testLoop(func(lhs, rhs uint64) { bigAdd64(t, ctx, lhs, rhs) })
}

func TestBigSub(t *testing.T) {
	ctx := setupBigIntTest(t)

	res := bigSub(t, ctx, wasmtypes.NewScBigInt(), wasmtypes.NewScBigInt())
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigSub(t, ctx, wasmtypes.NewScBigInt(1), wasmtypes.NewScBigInt(1))
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())

	testLoop(func(lhs, rhs uint64) { bigSub64(t, ctx, lhs, rhs) })
}

func TestBigMul(t *testing.T) {
	ctx := setupBigIntTest(t)

	res := bigMul(t, ctx, wasmtypes.NewScBigInt(), wasmtypes.NewScBigInt())
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigMul(t, ctx, wasmtypes.NewScBigInt(1), wasmtypes.NewScBigInt(1))
	require.False(t, res.IsZero())
	require.EqualValues(t, "1", res.String())

	testLoop(func(lhs, rhs uint64) { bigMul64(t, ctx, lhs, rhs) })
}

func TestBigDiv(t *testing.T) {
	ctx := setupBigIntTest(t)

	bigDiv64(t, ctx, 536870911, 511)

	res := bigDiv(t, ctx, wasmtypes.NewScBigInt(), wasmtypes.NewScBigInt(1))
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigDiv(t, ctx, wasmtypes.NewScBigInt(1), wasmtypes.NewScBigInt(1))
	require.False(t, res.IsZero())
	require.EqualValues(t, "1", res.String())

	testLoop(func(lhs, rhs uint64) { bigDiv64(t, ctx, lhs, rhs) })
}

func TestBigDivMod(t *testing.T) {
	ctx := setupBigIntTest(t)

	bigDivMod64(t, ctx, 536870911, 511)

	quo, rem := bigDivMod(t, ctx, wasmtypes.NewScBigInt(), wasmtypes.NewScBigInt(1))
	require.True(t, quo.IsZero())
	require.EqualValues(t, "0", quo.String())
	require.True(t, rem.IsZero())
	require.EqualValues(t, "0", rem.String())

	quo, rem = bigDivMod(t, ctx, wasmtypes.NewScBigInt(1), wasmtypes.NewScBigInt(1))
	require.False(t, quo.IsZero())
	require.EqualValues(t, "1", quo.String())
	require.True(t, rem.IsZero())
	require.EqualValues(t, "0", rem.String())

	bigDivModString(t, ctx, "0", "1", "0", "0")
	bigDivModString(t, ctx, "1", "1", "1", "0")
	bigDivModString(t, ctx, "123456789012345678901234567", "63531", "1943252727209483227105", "26812")

	testLoop(func(lhs, rhs uint64) { bigDivMod64(t, ctx, lhs, rhs) })
}

func TestBigMod(t *testing.T) {
	ctx := setupBigIntTest(t)

	res := bigMod(t, ctx, wasmtypes.NewScBigInt(), wasmtypes.NewScBigInt(1))
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigMod(t, ctx, wasmtypes.NewScBigInt(1), wasmtypes.NewScBigInt(1))
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())

	testLoop(func(lhs, rhs uint64) { bigMod64(t, ctx, lhs, rhs) })
}

func TestBigShl(t *testing.T) {
	ctx := setupBigIntTest(t)

	res := bigShl(t, ctx, wasmtypes.NewScBigInt(), 0)
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigShl(t, ctx, wasmtypes.NewScBigInt(1), 1)
	require.False(t, res.IsZero())
	require.EqualValues(t, "2", res.String())

	testLoop(func(lhs, rhs uint64) { bigShl64(t, ctx, lhs, uint32(rhs)) })
}

func TestBigShr(t *testing.T) {
	ctx := setupBigIntTest(t)

	res := bigShr(t, ctx, wasmtypes.NewScBigInt(), 0)
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())
	res = bigShr(t, ctx, wasmtypes.NewScBigInt(1), 1)
	require.True(t, res.IsZero())
	require.EqualValues(t, "0", res.String())

	testLoop(func(lhs, rhs uint64) { bigShr64(t, ctx, lhs, uint32(rhs)) })
}

func bigAdd(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs wasmtypes.ScBigInt) wasmtypes.ScBigInt {
	if SkipWasm {
		return lhs.Add(rhs)
	}
	f := testwasmlib.ScFuncs.BigIntAdd(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Rhs().SetValue(rhs)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
}

func bigSub(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs wasmtypes.ScBigInt) wasmtypes.ScBigInt {
	if SkipWasm {
		return lhs.Sub(rhs)
	}
	f := testwasmlib.ScFuncs.BigIntSub(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Rhs().SetValue(rhs)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
}

func bigMul(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs wasmtypes.ScBigInt) wasmtypes.ScBigInt {
	if SkipWasm {
		return lhs.Mul(rhs)
	}
	f := testwasmlib.ScFuncs.BigIntMul(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Rhs().SetValue(rhs)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
}

func bigDiv(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs wasmtypes.ScBigInt) wasmtypes.ScBigInt {
	if SkipWasm {
		return lhs.Div(rhs)
	}
	f := testwasmlib.ScFuncs.BigIntDiv(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Rhs().SetValue(rhs)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
}

func bigDivMod(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs wasmtypes.ScBigInt) (wasmtypes.ScBigInt, wasmtypes.ScBigInt) {
	if SkipWasm {
		return lhs.DivMod(rhs)
	}
	f := testwasmlib.ScFuncs.BigIntDivMod(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Rhs().SetValue(rhs)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Quo().Value(), f.Results.Remainder().Value()
}

func bigMod(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs wasmtypes.ScBigInt) wasmtypes.ScBigInt {
	if SkipWasm {
		return lhs.Modulo(rhs)
	}
	f := testwasmlib.ScFuncs.BigIntMod(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Rhs().SetValue(rhs)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
}

func bigShl(t *testing.T, ctx *wasmsolo.SoloContext, lhs wasmtypes.ScBigInt, shift uint32) wasmtypes.ScBigInt {
	if SkipWasm {
		return lhs.Shl(shift)
	}
	f := testwasmlib.ScFuncs.BigIntShl(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Shift().SetValue(shift)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
}

func bigShr(t *testing.T, ctx *wasmsolo.SoloContext, lhs wasmtypes.ScBigInt, shift uint32) wasmtypes.ScBigInt {
	if SkipWasm {
		return lhs.Shr(shift)
	}
	f := testwasmlib.ScFuncs.BigIntShr(ctx)
	f.Params.Lhs().SetValue(lhs)
	f.Params.Shift().SetValue(shift)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	return f.Results.Res().Value()
}

func bigAdd64(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs uint64) {
	expect := lhs + rhs
	if LogOp {
		t.Logf("%d + %d = %d\n", lhs, rhs, expect)
	}
	res := bigAdd(t, ctx, wasmtypes.NewScBigInt(lhs), wasmtypes.NewScBigInt(rhs))
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}

func bigSub64(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs uint64) {
	expect := lhs - rhs
	if LogOp {
		t.Logf("%d - %d = %d\n", lhs, rhs, expect)
	}
	res := bigSub(t, ctx, wasmtypes.NewScBigInt(lhs), wasmtypes.NewScBigInt(rhs))
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}

func bigMul64(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs uint64) {
	expect := lhs * rhs
	if LogOp {
		t.Logf("%d * %d = %d\n", lhs, rhs, expect)
	}
	res := bigMul(t, ctx, wasmtypes.NewScBigInt(lhs), wasmtypes.NewScBigInt(rhs))
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}

func bigDiv64(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs uint64) {
	expect := lhs / rhs
	if LogOp {
		t.Logf("%d / %d = %d\n", lhs, rhs, expect)
	}
	res := bigDiv(t, ctx, wasmtypes.NewScBigInt(lhs), wasmtypes.NewScBigInt(rhs))
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}

func bigDivMod64(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs uint64) {
	expectQuo := lhs / rhs
	expectRemainder := lhs - rhs*expectQuo
	if LogOp {
		t.Logf("%d / %d = %d...%d\n", lhs, rhs, expectQuo, expectRemainder)
	}
	quo, remainder := bigDivMod(t, ctx, wasmtypes.NewScBigInt(lhs), wasmtypes.NewScBigInt(rhs))
	require.EqualValues(t, expectQuo, quo.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expectQuo), quo.String())
	require.EqualValues(t, 0, quo.Cmp(wasmtypes.NewScBigInt(expectQuo)))
	require.EqualValues(t, expectRemainder, remainder.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expectRemainder), remainder.String())
	require.EqualValues(t, 0, remainder.Cmp(wasmtypes.NewScBigInt(expectRemainder)))
}

func bigDivModString(t *testing.T, ctx *wasmsolo.SoloContext, num, denom, quo, rem string) {
	if LogOp {
		t.Logf("%s / %s = %s...%s\n", num, denom, quo, rem)
	}
	quotient, remainder := bigDivMod(t, ctx, wasmtypes.BigIntFromString(num), wasmtypes.BigIntFromString(denom))
	require.EqualValues(t, quo, quotient.String())
	require.EqualValues(t, rem, remainder.String())
}

func bigMod64(t *testing.T, ctx *wasmsolo.SoloContext, lhs, rhs uint64) {
	expect := lhs % rhs
	if LogOp {
		t.Logf("%d %% %d = %d\n", lhs, rhs, expect)
	}
	res := bigMod(t, ctx, wasmtypes.NewScBigInt(lhs), wasmtypes.NewScBigInt(rhs))
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}

func bigShl64(t *testing.T, ctx *wasmsolo.SoloContext, lhs uint64, shift uint32) {
	if shift > 1_000 {
		// 2^shift is getting silly
		return
	}
	expect := lhs << shift
	if LogOp {
		t.Logf("%d << %d = %d\n", lhs, shift, expect)
	}
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
	if shift > 1_000 {
		// 2^shift is getting silly
		return
	}
	expect := lhs >> shift
	if LogOp {
		t.Logf("%d >> %d = %d\n", lhs, shift, expect)
	}
	res := bigShr(t, ctx, wasmtypes.NewScBigInt(lhs), shift)
	require.EqualValues(t, expect, res.Uint64())
	require.EqualValues(t, wasmtypes.Uint64ToString(expect), res.String())
	require.EqualValues(t, 0, res.Cmp(wasmtypes.NewScBigInt(expect)))
}
