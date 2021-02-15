// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/iotaledger/wasp/contracts/testenv"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeployDividend(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_, err := te.Chain.FindContract(ScName)
	require.NoError(t, err)
}

func TestAddMemberOk(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	user1 := te.Env.NewSignatureSchemeWithFunds()
	_ = te.NewCallParams(FuncMember,
		ParamAddress, user1.Address(),
		ParamFactor, 100,
	).Post(0)
}

func TestAddMemberFailMissingAddress(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	_ = te.NewCallParams(FuncMember,
		ParamFactor, 100,
	).PostFail(0)
}

func TestAddMemberFailMissingFactor(t *testing.T) {
	te := testenv.NewTestEnv(t, ScName)
	user1 := te.Env.NewSignatureSchemeWithFunds()
	_ = te.NewCallParams(FuncMember,
		ParamAddress, user1.Address(),
	).PostFail(0)
}
