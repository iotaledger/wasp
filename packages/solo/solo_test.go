package solo

import (
	"testing"

	"github.com/iotaledger/wasp/packages/vm/core"
)

func TestSoloBasic(t *testing.T) {
	core.PrintWellKnownHnames()
	env := New(t, true, false)
	_ = env.NewChain(nil, "ch1")
}
