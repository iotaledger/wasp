package solo

import (
	"testing"

	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
)

func TestSoloBasic1(t *testing.T) {
	corecontracts.PrintWellKnownHnames()
	env := New(t, &InitOptions{Debug: true, PrintStackTrace: true})
	_ = env.NewChain()
}
