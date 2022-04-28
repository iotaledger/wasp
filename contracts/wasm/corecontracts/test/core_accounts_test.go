// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func setupAccounts(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreaccounts.ScName, coreaccounts.OnLoad)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestAccountsXxx(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent()
	f := coreaccounts.ScFuncs.Withdraw(ctx.Sign(user))
	f.Func.TransferIotas(10_000).Post()
	require.NoError(t, ctx.Err)
}
