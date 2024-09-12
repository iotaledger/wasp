package cryptolib_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestKeypairCodec(t *testing.T) {
	bcs.TestCodec(t, cryptolib.NewKeyPair())
}
