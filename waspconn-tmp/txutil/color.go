package txutil

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"golang.org/x/crypto/blake2b"
)

func MintedColorFromOutput(out ledgerstate.Output) ledgerstate.Color {
	return blake2b.Sum256(out.ID().Bytes())
}
