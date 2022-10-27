package util

import (
	"os"

	"github.com/iotaledger/wasp/packages/log"
)

func ReadFile(fname string) []byte {
	b, err := os.ReadFile(fname)
	log.Check(err)
	return b
}
