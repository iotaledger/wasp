package vm0poc

import (
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	t.Run("create chain", func(t *testing.T) {
		env := solo.New(t)
		_ = env.NewChain(nil, "ch1", solo.InitChainOptions{
			VMRunner:         NewVMRunner(),
			BypassStardustVM: true,
		})
	})
	t.Run("send 1 request", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "ch1", solo.InitChainOptions{
			VMRunner:         NewVMRunner(),
			BypassStardustVM: true,
		})
		req := solo.NewCallParams("dummy", "dummy", ParamDeltaInt64, int64(10))
		_, err := ch.PostRequestOffLedger(req, nil)
		require.NoError(t, err)
	})
	t.Run("send 5 requests", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "ch1", solo.InitChainOptions{
			VMRunner:         NewVMRunner(),
			BypassStardustVM: true,
		})
		req := solo.NewCallParams("dummy", "dummy", ParamDeltaInt64, int64(1))
		for i := 0; i < 5; i++ {
			_, err := ch.PostRequestOffLedger(req, nil)
			require.NoError(t, err)
		}
	})
}
