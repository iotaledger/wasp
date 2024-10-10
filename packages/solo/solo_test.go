package solo

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestMain(m *testing.M) {
	flag.Parse()
	stv := l1starter.Start(context.Background(), l1starter.DefaultConfig)
	defer stv.Stop()
	m.Run()
}

func TestSoloBasic1(t *testing.T) {
	env := New(t, &InitOptions{Debug: true, PrintStackTrace: true})
	ch := env.NewChain()
	require.EqualValues(env.T, DefaultCommonAccountBaseTokens, ch.L2CommonAccountAssets().Coins.BaseTokens())
	require.EqualValues(env.T, DefaultChainOriginatorBaseTokens, ch.L2BaseTokens(ch.OriginatorAgentID))
}
