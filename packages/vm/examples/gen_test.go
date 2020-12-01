package examples

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"testing"
)

// it is needed only to generate dummy hash code
func TestGenData(t *testing.T) {
	h := hashing.HashStrings("dummy builtin program")
	t.Logf("dummy builtin program hash = %s", h.String())
}
