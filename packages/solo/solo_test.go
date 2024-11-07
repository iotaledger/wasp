package solo

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestSoloBasic1(t *testing.T) {
	env := New(t, &InitOptions{Debug: true, PrintStackTrace: true})
	ch := env.NewChain(false)
	require.EqualValues(env.T, DefaultCommonAccountBaseTokens, ch.L2CommonAccountAssets().Coins.BaseTokens())

	err := ch.DepositBaseTokensToL2(DefaultChainOriginatorBaseTokens, nil)
	require.NoError(env.T, err)
	require.EqualValues(env.T, DefaultChainOriginatorBaseTokens, ch.L2BaseTokens(ch.OriginatorAgentID))
}
