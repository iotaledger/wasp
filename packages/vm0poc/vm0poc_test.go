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
			VMRunner:                  NewVMRunner(),
			SkipStardustVMInitRequest: true,
		})
	})
	t.Run("send 1 request", func(t *testing.T) {
		env := solo.New(t)
		ch := env.NewChain(nil, "ch1", solo.InitChainOptions{
			VMRunner:                  NewVMRunner(),
			SkipStardustVMInitRequest: true,
		})
		req := solo.NewCallParams("dummy", "dummy", ParamDeltaInt64, int64(10))
		_, err := ch.PostRequestOffLedger(req, nil)
		require.NoError(t, err)
	})
}
