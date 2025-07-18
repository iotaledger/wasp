package solo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestSoloBasic1(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	ch := env.NewChain(false)
	require.Zero(env.T, ch.L2CommonAccountAssets().Coins.BaseTokens())
	require.Zero(env.T, ch.L2BaseTokens(ch.AdminAgentID()))

	err := ch.DepositBaseTokensToL2(solo.DefaultChainAdminBaseTokens, nil)
	require.NoError(env.T, err)
	require.NotZero(env.T, ch.L2BaseTokens(ch.AdminAgentID()))
}
