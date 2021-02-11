package test

import (
	"github.com/iotaledger/wasp/contracts/rust/dividend"
	"github.com/iotaledger/wasp/contracts/testenv"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeployDividend(t *testing.T) {
	t.SkipNow()
	te := testenv.NewTestEnv(t, dividend.ScName)
	_, err := te.Chain.FindContract(dividend.ScName)
	require.NoError(t, err)
}

func TestAddMemberOk(t *testing.T) {
	t.SkipNow()
	te := testenv.NewTestEnv(t, dividend.ScName)
	user1 := te.Env.NewSignatureSchemeWithFunds()
	_ = te.NewCallParams(dividend.FuncMember,
		dividend.ParamAddress, user1.Address(),
		dividend.ParamFactor, 100,
	).Post(0)
}

func TestAddMemberFailMissingAddress(t *testing.T) {
	t.SkipNow()
	te := testenv.NewTestEnv(t, dividend.ScName)
	_ = te.NewCallParams(dividend.FuncMember,
		dividend.ParamFactor, 100,
	).PostFail(0)
}

func TestAddMemberFailMissingFactor(t *testing.T) {
	t.SkipNow()
	te := testenv.NewTestEnv(t, dividend.ScName)
	user1 := te.Env.NewSignatureSchemeWithFunds()
	_ = te.NewCallParams(dividend.FuncMember,
		dividend.ParamAddress, user1.Address(),
	).PostFail(0)
}
