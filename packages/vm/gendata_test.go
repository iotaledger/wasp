package vm

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"testing"
)

const emptyProgram = "empty program"

func TestGendata(t *testing.T) {
	t.Logf("Program hash of '%s' = %s", emptyProgram, hashing.HashStrings(emptyProgram).String())
}
