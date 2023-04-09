// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreerrors"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

func setupErrors(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreerrors.ScName, coreerrors.OnDispatch)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestRegisterErrorAndGetErrorMessageFormat(t *testing.T) {
	ctx := setupErrors(t)
	require.NoError(t, ctx.Err)

	errorFmt := "this \"%v\" is error format"
	fRegister := coreerrors.ScFuncs.RegisterError(ctx)
	fRegister.Params.Template().SetValue(errorFmt)
	fRegister.Func.Post()
	require.NoError(t, ctx.Err)
	errCodeBytes := fRegister.Results.ErrorCode().Value()

	fGet := coreerrors.ScFuncs.GetErrorMessageFormat(ctx)
	fGet.Params.ErrorCode().SetValue(errCodeBytes)
	fGet.Func.Call()
	require.NoError(t, ctx.Err)
	resErrorFmt := fGet.Results.Template().String()
	require.Equal(t, errorFmt, resErrorFmt)
}
