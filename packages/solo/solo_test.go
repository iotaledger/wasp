package solo

import (
	"testing"

	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
)

func TestSoloBasic1(t *testing.T) {
	corecontracts.PrintWellKnownHnames()
	env := New(t)
	_ = env.NewChain()
}

func TestSoloBasic2(t *testing.T) {
	corecontracts.PrintWellKnownHnames()
	env := New(t, &InitOptions{
		Debug: true,
	})
	_ = env.NewChain()
}
