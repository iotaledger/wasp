package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"testing"
)

// it is needed only to generate dummy hash code
func TestGenData(t *testing.T) {
	textToHash := "Wasm VM PoC program"
	h := hashing.HashStrings(textToHash)
	t.Logf("Hash of '%s' = %s", textToHash, h.String())
}
